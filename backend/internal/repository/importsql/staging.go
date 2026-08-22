package importsql

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	fielddeviceimport "github.com/besart951/go_infra_link/backend/internal/application/fielddeviceimport"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	aggregatePageSize = 100
	cleanupBatchSize  = 100
	stagingRetention  = 90 * 24 * time.Hour
)

type sessionRecord struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	OwnerID   uuid.UUID `gorm:"type:uuid;not null;index"`
	Status    string    `gorm:"size:24;not null"`
	Manifest  *string   `gorm:"type:jsonb"`
	CreatedAt time.Time `gorm:"not null"`
	UpdatedAt time.Time `gorm:"not null;index"`
}

func (sessionRecord) TableName() string { return "facility_import_sessions" }

type rowRecord struct {
	ID       uuid.UUID `gorm:"type:uuid;primaryKey"`
	ImportID uuid.UUID `gorm:"type:uuid;not null;index:idx_facility_import_rows_owner,priority:1;index:idx_facility_import_rows_source,priority:1"`
	Kind     string    `gorm:"size:32;not null;index:idx_facility_import_rows_source,priority:2"`
	SourceID uuid.UUID `gorm:"type:uuid;not null;index:idx_facility_import_rows_source,priority:3"`
	OwnerID  uuid.UUID `gorm:"type:uuid;index:idx_facility_import_rows_owner,priority:2"`
	Payload  string    `gorm:"type:jsonb;not null"`
}

func (rowRecord) TableName() string { return "facility_import_rows" }

type Store struct{ db *gorm.DB }

func NewStore(db *gorm.DB) *Store { return &Store{db: db} }

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(&sessionRecord{}, &rowRecord{})
}

func (s *Store) Start(ctx context.Context, ownerID uuid.UUID) (uuid.UUID, fielddeviceimport.Session, error) {
	now := time.Now().UTC()
	if err := s.Cleanup(ctx, now.Add(-stagingRetention)); err != nil {
		return uuid.Nil, nil, err
	}
	record := sessionRecord{ID: uuid.New(), OwnerID: ownerID, Status: "staging", CreatedAt: now, UpdatedAt: now}
	if err := s.db.WithContext(ctx).Create(&record).Error; err != nil {
		return uuid.Nil, nil, err
	}
	return record.ID, &session{db: s.db, id: record.ID}, nil
}

func (s *Store) Cleanup(ctx context.Context, cutoff time.Time) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var ids []uuid.UUID
		if err := tx.Model(&sessionRecord{}).Where("updated_at < ?", cutoff).Order("updated_at ASC").Limit(cleanupBatchSize).Pluck("id", &ids).Error; err != nil || len(ids) == 0 {
			return err
		}
		if err := tx.Where("import_id IN ?", ids).Delete(&rowRecord{}).Error; err != nil {
			return err
		}
		return tx.Where("id IN ?", ids).Delete(&sessionRecord{}).Error
	})
}

type session struct {
	db *gorm.DB
	id uuid.UUID
}

func (s *session) Seal(ctx context.Context, manifest fielddeviceimport.Manifest) error {
	payload, err := json.Marshal(manifest)
	if err != nil {
		return err
	}
	return s.db.WithContext(ctx).Model(&sessionRecord{}).Where("id = ?", s.id).Updates(map[string]any{
		"manifest": string(payload), "status": "validating", "updated_at": time.Now().UTC(),
	}).Error
}

func (s *session) FieldDevices(ctx context.Context, values []domainFacility.FieldDevice) error {
	return appendRows(ctx, appendRequest[domainFacility.FieldDevice]{
		db: s.db, importID: s.id, kind: "field_device", values: values,
		ids: func(value domainFacility.FieldDevice) (uuid.UUID, uuid.UUID) { return value.ID, uuid.Nil },
	})
}

func (s *session) Specifications(ctx context.Context, values []domainFacility.Specification) error {
	return appendRows(ctx, appendRequest[domainFacility.Specification]{
		db: s.db, importID: s.id, kind: "specification", values: values,
		ids: func(value domainFacility.Specification) (uuid.UUID, uuid.UUID) {
			return value.ID, pointerUUID(value.FieldDeviceID)
		},
	})
}

func (s *session) BacnetObjects(ctx context.Context, values []domainFacility.BacnetObject) error {
	return appendRows(ctx, appendRequest[domainFacility.BacnetObject]{
		db: s.db, importID: s.id, kind: "bacnet_object", values: values,
		ids: func(value domainFacility.BacnetObject) (uuid.UUID, uuid.UUID) {
			return value.ID, pointerUUID(value.FieldDeviceID)
		},
	})
}

func (s *session) SoftwareReferences(ctx context.Context, values []fielddeviceimport.SoftwareReference) error {
	return appendRows(ctx, appendRequest[fielddeviceimport.SoftwareReference]{
		db: s.db, importID: s.id, kind: "software_reference", values: values,
		ids: func(value fielddeviceimport.SoftwareReference) (uuid.UUID, uuid.UUID) {
			return value.SourceObjectID, value.FieldDeviceID
		},
	})
}

func (s *session) AlarmValues(ctx context.Context, values []domainFacility.BacnetObjectAlarmValue) error {
	return appendRows(ctx, appendRequest[domainFacility.BacnetObjectAlarmValue]{
		db: s.db, importID: s.id, kind: "alarm_value", values: values,
		ids: func(value domainFacility.BacnetObjectAlarmValue) (uuid.UUID, uuid.UUID) {
			return value.ID, value.BacnetObjectID
		},
	})
}

type appendRequest[T any] struct {
	db       *gorm.DB
	importID uuid.UUID
	kind     string
	values   []T
	ids      func(T) (uuid.UUID, uuid.UUID)
}

func appendRows[T any](ctx context.Context, request appendRequest[T]) error {
	rows := make([]rowRecord, len(request.values))
	for index, value := range request.values {
		payload, err := json.Marshal(value)
		if err != nil {
			return err
		}
		sourceID, ownerID := request.ids(value)
		rows[index] = rowRecord{ID: uuid.New(), ImportID: request.importID, Kind: request.kind, SourceID: sourceID, OwnerID: ownerID, Payload: string(payload)}
	}
	if len(rows) == 0 {
		return nil
	}
	return request.db.WithContext(ctx).CreateInBatches(rows, 500).Error
}

func pointerUUID(value *uuid.UUID) uuid.UUID {
	if value == nil {
		return uuid.Nil
	}
	return *value
}

func (s *session) Validate(ctx context.Context) ([]fielddeviceimport.Issue, error) {
	checks := []func(context.Context) ([]fielddeviceimport.Issue, error){
		s.manifestIssues,
		s.duplicateIssues,
		s.ownerIssues,
		s.ownerCardinalityIssues,
		s.alarmOwnerIssues,
		s.softwareReferenceIssues,
		s.existingIDIssues,
		s.referenceDataIssues,
	}
	issues := make([]fielddeviceimport.Issue, 0)
	for _, check := range checks {
		found, err := check(ctx)
		if err != nil {
			return nil, err
		}
		issues = append(issues, found...)
	}
	return issues, nil
}

func (s *session) manifestIssues(ctx context.Context) ([]fielddeviceimport.Issue, error) {
	var record sessionRecord
	if err := s.db.WithContext(ctx).Where("id = ?", s.id).Take(&record).Error; err != nil {
		return nil, err
	}
	if record.Manifest == nil {
		return nil, fmt.Errorf("import manifest is not sealed")
	}
	var manifest fielddeviceimport.Manifest
	if err := json.Unmarshal([]byte(*record.Manifest), &manifest); err != nil {
		return nil, err
	}
	actual, err := s.rowCounts(ctx)
	if err != nil {
		return nil, err
	}
	expected := map[string]int64{
		"field_device": manifest.DeviceCount, "specification": manifest.Counts.Specifications,
		"bacnet_object": manifest.Counts.BacnetObjects, "software_reference": manifest.Counts.SoftwareReferences,
		"alarm_value": manifest.Counts.AlarmValues,
	}
	issues := make([]fielddeviceimport.Issue, 0)
	for kind, count := range expected {
		if actual[kind] != count {
			issues = append(issues, fielddeviceimport.Issue{Code: "manifest_count_mismatch", Entity: "manifest", Field: kind, Message: fmt.Sprintf("manifest declares %d rows, data contains %d", count, actual[kind])})
		}
	}
	return issues, nil
}

type kindCount struct {
	Kind  string
	Count int64
}

func (s *session) rowCounts(ctx context.Context) (map[string]int64, error) {
	var rows []kindCount
	err := s.db.WithContext(ctx).Model(&rowRecord{}).Select("kind, COUNT(*) AS count").Where("import_id = ?", s.id).Group("kind").Scan(&rows).Error
	counts := make(map[string]int64, len(rows))
	for _, row := range rows {
		counts[row.Kind] = row.Count
	}
	return counts, err
}

type issueRow struct {
	Kind     string
	SourceID uuid.UUID
	Field    string
}

func (s *session) duplicateIssues(ctx context.Context) ([]fielddeviceimport.Issue, error) {
	var rows []issueRow
	err := s.db.WithContext(ctx).Raw(`
		SELECT kind, source_id, 'source_id' AS field
		FROM facility_import_rows
		WHERE import_id = ?
		GROUP BY kind, source_id HAVING COUNT(*) > 1`, s.id).Scan(&rows).Error
	return mapIssues(rows, "duplicate_source_id", "source ID occurs more than once"), err
}

func (s *session) ownerIssues(ctx context.Context) ([]fielddeviceimport.Issue, error) {
	var rows []issueRow
	err := s.db.WithContext(ctx).Raw(`
		SELECT child.kind, child.source_id, 'owner_id' AS field
		FROM facility_import_rows child
		LEFT JOIN facility_import_rows owner ON owner.import_id = child.import_id
		 AND owner.kind = 'field_device' AND owner.source_id = child.owner_id
		WHERE child.import_id = ? AND child.kind IN ('specification', 'bacnet_object')
		 AND owner.id IS NULL`, s.id).Scan(&rows).Error
	return mapIssues(rows, "missing_owner", "field device owner is missing"), err
}

func (s *session) ownerCardinalityIssues(ctx context.Context) ([]fielddeviceimport.Issue, error) {
	var rows []issueRow
	err := s.db.WithContext(ctx).Raw(`
		SELECT kind, source_id, 'field_device_id' AS field FROM (
			SELECT kind, source_id, COUNT(*) OVER (PARTITION BY owner_id) AS owner_count
			FROM facility_import_rows WHERE import_id = ? AND kind = 'specification'
		) specifications WHERE owner_count > 1`, s.id).Scan(&rows).Error
	return mapIssues(rows, "multiple_specifications", "field device has more than one specification"), err
}

func (s *session) alarmOwnerIssues(ctx context.Context) ([]fielddeviceimport.Issue, error) {
	var rows []issueRow
	err := s.db.WithContext(ctx).Raw(`
		SELECT child.kind, child.source_id, 'bacnet_object_id' AS field
		FROM facility_import_rows child
		LEFT JOIN facility_import_rows owner ON owner.import_id = child.import_id
		 AND owner.kind = 'bacnet_object' AND owner.source_id = child.owner_id
		WHERE child.import_id = ? AND child.kind = 'alarm_value' AND owner.id IS NULL`, s.id).Scan(&rows).Error
	return mapIssues(rows, "missing_owner", "BACnet object owner is missing"), err
}

func (s *session) softwareReferenceIssues(ctx context.Context) ([]fielddeviceimport.Issue, error) {
	var rows []issueRow
	targetExpression := s.jsonText("ref.payload", "target_object_id")
	objectReference := s.jsonText("source.payload", "SoftwareReferenceID")
	query := fmt.Sprintf(`
		SELECT ref.kind, ref.source_id, 'target_object_id' AS field
		FROM facility_import_rows ref
		LEFT JOIN facility_import_rows source ON source.import_id = ref.import_id
		 AND source.kind = 'bacnet_object' AND source.source_id = ref.source_id
		LEFT JOIN facility_import_rows target ON target.import_id = ref.import_id
		 AND target.kind = 'bacnet_object' AND %s = %s
		WHERE ref.import_id = ? AND ref.kind = 'software_reference'
		 AND (source.id IS NULL OR target.id IS NULL OR source.owner_id <> target.owner_id
		 OR ref.owner_id <> source.owner_id OR COALESCE(%s, '') <> COALESCE(%s, ''))`,
		s.idText("target.source_id"), targetExpression, objectReference, targetExpression)
	if err := s.db.WithContext(ctx).Raw(query, s.id).Scan(&rows).Error; err != nil {
		return nil, err
	}
	issues := mapIssues(rows, "invalid_software_reference", "software reference must stay inside one field device")
	missing, err := s.missingSoftwareReferenceRows(ctx)
	return append(issues, missing...), err
}

func (s *session) missingSoftwareReferenceRows(ctx context.Context) ([]fielddeviceimport.Issue, error) {
	var rows []issueRow
	referenceExpression := s.jsonText("object.payload", "SoftwareReferenceID")
	query := fmt.Sprintf(`
		SELECT object.kind, object.source_id, 'software_reference_id' AS field
		FROM facility_import_rows object
		LEFT JOIN facility_import_rows ref ON ref.import_id = object.import_id
		 AND ref.kind = 'software_reference' AND ref.source_id = object.source_id
		WHERE object.import_id = ? AND object.kind = 'bacnet_object'
		 AND COALESCE(%s, '') <> '' AND ref.id IS NULL`, referenceExpression)
	err := s.db.WithContext(ctx).Raw(query, s.id).Scan(&rows).Error
	return mapIssues(rows, "missing_software_reference_row", "BACnet reference is missing from Data-SoftwareReferences"), err
}

func (s *session) existingIDIssues(ctx context.Context) ([]fielddeviceimport.Issue, error) {
	tables := map[string]string{
		"field_device": "field_devices", "specification": "specifications",
		"bacnet_object": "bacnet_objects", "alarm_value": "bacnet_object_alarm_values",
	}
	issues := make([]fielddeviceimport.Issue, 0)
	for kind, table := range tables {
		var rows []issueRow
		query := fmt.Sprintf(`SELECT r.kind, r.source_id, 'source_id' AS field FROM facility_import_rows r JOIN %s target ON target.id = r.source_id WHERE r.import_id = ? AND r.kind = ?`, table)
		if err := s.db.WithContext(ctx).Raw(query, s.id, kind).Scan(&rows).Error; err != nil {
			return nil, err
		}
		issues = append(issues, mapIssues(rows, "source_id_conflict", "source ID already exists")...)
	}
	return issues, nil
}

func (s *session) referenceDataIssues(ctx context.Context) ([]fielddeviceimport.Issue, error) {
	checks := []referenceCheck{
		{kind: "field_device", table: "sps_controller_system_types", field: "SPSControllerSystemTypeID", required: true},
		{kind: "field_device", table: "system_parts", field: "SystemPartID", required: true},
		{kind: "field_device", table: "apparats", field: "ApparatID", required: true},
		{kind: "bacnet_object", table: "state_texts", field: "StateTextID"},
		{kind: "bacnet_object", table: "notification_classes", field: "NotificationClassID"},
		{kind: "bacnet_object", table: "alarm_types", field: "AlarmTypeID"},
		{kind: "alarm_value", table: "alarm_type_fields", field: "AlarmTypeFieldID", required: true},
		{kind: "alarm_value", table: "units", field: "UnitID"},
	}
	issues := make([]fielddeviceimport.Issue, 0)
	for _, check := range checks {
		found, err := s.missingReferenceIssues(ctx, check)
		if err != nil {
			return nil, err
		}
		issues = append(issues, found...)
	}
	return issues, nil
}

type referenceCheck struct {
	kind, table, field string
	required           bool
}

func (s *session) missingReferenceIssues(ctx context.Context, check referenceCheck) ([]fielddeviceimport.Issue, error) {
	var rows []issueRow
	valueExpression := s.jsonText("r.payload", check.field)
	presentCondition := fmt.Sprintf("COALESCE(%s, '') <> '' AND ", valueExpression)
	if check.required {
		presentCondition = ""
	}
	query := fmt.Sprintf(`
		SELECT r.kind, r.source_id, '%s' AS field FROM facility_import_rows r
		LEFT JOIN %s target ON %s = %s
		WHERE r.import_id = ? AND r.kind = ? AND %s target.id IS NULL`, check.field, check.table, s.idText("target.id"), valueExpression, presentCondition)
	err := s.db.WithContext(ctx).Raw(query, s.id, check.kind).Scan(&rows).Error
	return mapIssues(rows, "invalid_reference", "referenced entity does not exist"), err
}

func (s *session) jsonText(column, field string) string {
	if s.db.Dialector.Name() == "sqlite" {
		return fmt.Sprintf("json_extract(%s, '$.%s')", column, field)
	}
	return fmt.Sprintf("CAST(%s AS jsonb)->>'%s'", column, field)
}

func (s *session) idText(column string) string {
	if s.db.Dialector.Name() == "sqlite" {
		return column
	}
	return column + "::text"
}

func mapIssues(rows []issueRow, code, message string) []fielddeviceimport.Issue {
	issues := make([]fielddeviceimport.Issue, len(rows))
	for index, row := range rows {
		issues[index] = fielddeviceimport.Issue{Code: code, Entity: row.Kind, SourceID: row.SourceID, Field: row.Field, Message: message}
	}
	return issues
}

func (s *session) Aggregates(ctx context.Context, cursor string) (fielddeviceimport.AggregatePage, error) {
	query := s.db.WithContext(ctx).Where("import_id = ? AND kind = ?", s.id, "field_device").Order("source_id ASC").Limit(aggregatePageSize + 1)
	if id, err := uuid.Parse(cursor); err == nil {
		query = query.Where("source_id > ?", id)
	}
	var devices []rowRecord
	if err := query.Find(&devices).Error; err != nil {
		return fielddeviceimport.AggregatePage{}, err
	}
	hasMore := len(devices) > aggregatePageSize
	if hasMore {
		devices = devices[:aggregatePageSize]
	}
	items, err := s.loadAggregates(ctx, devices)
	page := fielddeviceimport.AggregatePage{Items: items}
	if err == nil && hasMore {
		page.NextCursor = devices[len(devices)-1].SourceID.String()
	}
	return page, err
}

func (s *session) loadAggregates(ctx context.Context, devices []rowRecord) ([]fielddeviceimport.Aggregate, error) {
	items := make([]fielddeviceimport.Aggregate, len(devices))
	for index, record := range devices {
		if err := json.Unmarshal([]byte(record.Payload), &items[index].FieldDevice); err != nil {
			return nil, err
		}
		if err := s.attachOwnedRows(ctx, &items[index]); err != nil {
			return nil, err
		}
	}
	return items, nil
}

func (s *session) attachOwnedRows(ctx context.Context, aggregate *fielddeviceimport.Aggregate) error {
	deviceID := aggregate.FieldDevice.ID
	if err := s.loadOne(ctx, rowLoad{kind: "specification", ownerID: deviceID, target: &aggregate.Specification}); err != nil {
		return err
	}
	if err := s.loadMany(ctx, rowLoad{kind: "bacnet_object", ownerID: deviceID, target: &aggregate.BacnetObjects}); err != nil {
		return err
	}
	if err := s.loadMany(ctx, rowLoad{kind: "software_reference", ownerID: deviceID, target: &aggregate.SoftwareReferences}); err != nil {
		return err
	}
	return s.attachAlarmValues(ctx, aggregate.BacnetObjects)
}

type rowLoad struct {
	kind    string
	ownerID uuid.UUID
	target  any
}

func (s *session) loadOne(ctx context.Context, request rowLoad) error {
	var row rowRecord
	result := s.db.WithContext(ctx).Where("import_id = ? AND kind = ? AND owner_id = ?", s.id, request.kind, request.ownerID).Limit(1).Find(&row)
	if result.Error != nil || result.RowsAffected == 0 {
		return result.Error
	}
	return json.Unmarshal([]byte(row.Payload), request.target)
}

func (s *session) loadMany(ctx context.Context, request rowLoad) error {
	var rows []rowRecord
	if err := s.db.WithContext(ctx).Where("import_id = ? AND kind = ? AND owner_id = ?", s.id, request.kind, request.ownerID).Order("source_id ASC").Find(&rows).Error; err != nil {
		return err
	}
	payload, err := payloadArray(rows)
	if err != nil {
		return err
	}
	return json.Unmarshal(payload, request.target)
}

func payloadArray(rows []rowRecord) ([]byte, error) {
	values := make([]json.RawMessage, len(rows))
	for index := range rows {
		values[index] = json.RawMessage(rows[index].Payload)
	}
	return json.Marshal(values)
}

func (s *session) attachAlarmValues(ctx context.Context, objects []domainFacility.BacnetObject) error {
	for index := range objects {
		if err := s.loadMany(ctx, rowLoad{kind: "alarm_value", ownerID: objects[index].ID, target: &objects[index].AlarmValues}); err != nil {
			return err
		}
	}
	return nil
}

func (s *session) Complete(ctx context.Context) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("import_id = ?", s.id).Delete(&rowRecord{}).Error; err != nil {
			return err
		}
		return updateSessionStatus(ctx, tx, sessionStatusUpdate{id: s.id, status: "completed"})
	})
}

func (s *session) Discard(ctx context.Context) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("import_id = ?", s.id).Delete(&rowRecord{}).Error; err != nil {
			return err
		}
		return tx.Where("id = ?", s.id).Delete(&sessionRecord{}).Error
	})
}

func (s *session) setStatus(ctx context.Context, status string) error {
	return updateSessionStatus(ctx, s.db, sessionStatusUpdate{id: s.id, status: status})
}

type sessionStatusUpdate struct {
	id     uuid.UUID
	status string
}

func updateSessionStatus(ctx context.Context, db *gorm.DB, update sessionStatusUpdate) error {
	return db.WithContext(ctx).Model(&sessionRecord{}).Where("id = ?", update.id).Updates(map[string]any{
		"status": update.status, "updated_at": time.Now().UTC(),
	}).Error
}
