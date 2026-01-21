package facilitysql

import (
	"database/sql"
	"strings"
	"time"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/besart951/go_infra_link/backend/internal/repository/sqlutil"
	"github.com/google/uuid"
)

type bacnetObjectRepo struct {
	db      *sql.DB
	dialect sqlutil.Dialect
}

func NewBacnetObjectRepository(db *sql.DB, driver string) domainFacility.BacnetObjectStore {
	return &bacnetObjectRepo{db: db, dialect: sqlutil.DialectFromDriver(driver)}
}

func (r *bacnetObjectRepo) GetByIds(ids []uuid.UUID) ([]*domainFacility.BacnetObject, error) {
	if len(ids) == 0 {
		return []*domainFacility.BacnetObject{}, nil
	}

	q := "SELECT id, created_at, updated_at, deleted_at, text_fix, description, gms_visible, optional, text_individual, software_type, software_number, hardware_type, hardware_quantity, field_device_id, software_reference_id, state_text_id, notification_class_id, alarm_definition_id " +
		"FROM bacnet_objects WHERE deleted_at IS NULL AND id IN (" + sqlutil.Placeholders(len(ids)) + ")"
	q = sqlutil.Rebind(r.dialect, q)

	args := make([]any, 0, len(ids))
	for _, id := range ids {
		args = append(args, id)
	}

	rows, err := r.db.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	out := make([]*domainFacility.BacnetObject, 0, len(ids))
	for rows.Next() {
		var b domainFacility.BacnetObject
		var deletedAt sql.NullTime
		var description sql.NullString
		var textIndividual sql.NullString
		var fieldDeviceID sql.NullString
		var softwareRefID sql.NullString
		var stateTextID sql.NullString
		var notificationClassID sql.NullString
		var alarmDefinitionID sql.NullString
		if err := rows.Scan(
			&b.ID,
			&b.CreatedAt,
			&b.UpdatedAt,
			&deletedAt,
			&b.TextFix,
			&description,
			&b.GMSVisible,
			&b.Optional,
			&textIndividual,
			&b.SoftwareType,
			&b.SoftwareNumber,
			&b.HardwareType,
			&b.HardwareQuantity,
			&fieldDeviceID,
			&softwareRefID,
			&stateTextID,
			&notificationClassID,
			&alarmDefinitionID,
		); err != nil {
			return nil, err
		}
		if deletedAt.Valid {
			t := deletedAt.Time
			b.DeletedAt = &t
		}
		if description.Valid {
			v := description.String
			b.Description = &v
		}
		if textIndividual.Valid {
			v := textIndividual.String
			b.TextIndividual = &v
		}
		if fieldDeviceID.Valid {
			id, err := uuid.Parse(fieldDeviceID.String)
			if err != nil {
				return nil, err
			}
			b.FieldDeviceID = &id
		}
		if softwareRefID.Valid {
			id, err := uuid.Parse(softwareRefID.String)
			if err != nil {
				return nil, err
			}
			b.SoftwareReferenceID = &id
		}
		if stateTextID.Valid {
			id, err := uuid.Parse(stateTextID.String)
			if err != nil {
				return nil, err
			}
			b.StateTextID = &id
		}
		if notificationClassID.Valid {
			id, err := uuid.Parse(notificationClassID.String)
			if err != nil {
				return nil, err
			}
			b.NotificationClassID = &id
		}
		if alarmDefinitionID.Valid {
			id, err := uuid.Parse(alarmDefinitionID.String)
			if err != nil {
				return nil, err
			}
			b.AlarmDefinitionID = &id
		}

		out = append(out, &b)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *bacnetObjectRepo) Create(entity *domainFacility.BacnetObject) error {
	now := time.Now().UTC()
	if err := entity.Base.InitForCreate(now); err != nil {
		return err
	}

	q := "INSERT INTO bacnet_objects (id, created_at, updated_at, deleted_at, text_fix, description, gms_visible, optional, text_individual, software_type, software_number, hardware_type, hardware_quantity, field_device_id, software_reference_id, state_text_id, notification_class_id, alarm_definition_id) " +
		"VALUES (?, ?, ?, NULL, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)"
	q = sqlutil.Rebind(r.dialect, q)

	_, err := r.db.Exec(
		q,
		entity.ID,
		entity.CreatedAt,
		entity.UpdatedAt,
		entity.TextFix,
		argStringPtr(entity.Description),
		entity.GMSVisible,
		entity.Optional,
		argStringPtr(entity.TextIndividual),
		string(entity.SoftwareType),
		entity.SoftwareNumber,
		string(entity.HardwareType),
		entity.HardwareQuantity,
		argUUIDPtr(entity.FieldDeviceID),
		argUUIDPtr(entity.SoftwareReferenceID),
		argUUIDPtr(entity.StateTextID),
		argUUIDPtr(entity.NotificationClassID),
		argUUIDPtr(entity.AlarmDefinitionID),
	)
	return err
}

func (r *bacnetObjectRepo) Update(entity *domainFacility.BacnetObject) error {
	now := time.Now().UTC()
	entity.Base.TouchForUpdate(now)

	q := "UPDATE bacnet_objects SET updated_at = ?, text_fix = ?, description = ?, gms_visible = ?, optional = ?, text_individual = ?, software_type = ?, software_number = ?, hardware_type = ?, hardware_quantity = ?, field_device_id = ?, software_reference_id = ?, state_text_id = ?, notification_class_id = ?, alarm_definition_id = ? WHERE deleted_at IS NULL AND id = ?"
	q = sqlutil.Rebind(r.dialect, q)

	_, err := r.db.Exec(
		q,
		entity.UpdatedAt,
		entity.TextFix,
		argStringPtr(entity.Description),
		entity.GMSVisible,
		entity.Optional,
		argStringPtr(entity.TextIndividual),
		string(entity.SoftwareType),
		entity.SoftwareNumber,
		string(entity.HardwareType),
		entity.HardwareQuantity,
		argUUIDPtr(entity.FieldDeviceID),
		argUUIDPtr(entity.SoftwareReferenceID),
		argUUIDPtr(entity.StateTextID),
		argUUIDPtr(entity.NotificationClassID),
		argUUIDPtr(entity.AlarmDefinitionID),
		entity.ID,
	)
	return err
}

func (r *bacnetObjectRepo) DeleteByIds(ids []uuid.UUID) error {
	if len(ids) == 0 {
		return nil
	}

	now := time.Now().UTC()
	q := "UPDATE bacnet_objects SET deleted_at = ?, updated_at = ? WHERE deleted_at IS NULL AND id IN (" + sqlutil.Placeholders(len(ids)) + ")"
	q = sqlutil.Rebind(r.dialect, q)

	args := make([]any, 0, 2+len(ids))
	args = append(args, now, now)
	for _, id := range ids {
		args = append(args, id)
	}

	_, err := r.db.Exec(q, args...)
	return err
}

func (r *bacnetObjectRepo) GetPaginatedList(params domain.PaginationParams) (*domain.PaginatedList[domainFacility.BacnetObject], error) {
	page := params.Page
	limit := params.Limit
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 10
	}
	offset := (page - 1) * limit

	where := "deleted_at IS NULL"
	args := make([]any, 0, 6)
	if strings.TrimSpace(params.Search) != "" {
		pattern := "%" + params.Search + "%"
		where += " AND (text_fix " + sqlutil.LikeOperator(r.dialect) + " ?)"
		args = append(args, pattern)
	}

	countQ := "SELECT COUNT(*) FROM bacnet_objects WHERE " + where
	countQ = sqlutil.Rebind(r.dialect, countQ)
	var total int64
	if err := r.db.QueryRow(countQ, args...).Scan(&total); err != nil {
		return nil, err
	}

	dataQ := "SELECT id, created_at, updated_at, deleted_at, text_fix, description, gms_visible, optional, text_individual, software_type, software_number, hardware_type, hardware_quantity, field_device_id, software_reference_id, state_text_id, notification_class_id, alarm_definition_id " +
		"FROM bacnet_objects WHERE " + where + " ORDER BY created_at DESC LIMIT ? OFFSET ?"
	dataQ = sqlutil.Rebind(r.dialect, dataQ)
	dataArgs := append(append([]any{}, args...), limit, offset)

	rows, err := r.db.Query(dataQ, dataArgs...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	items := make([]domainFacility.BacnetObject, 0, limit)
	for rows.Next() {
		var b domainFacility.BacnetObject
		var deletedAt sql.NullTime
		var description sql.NullString
		var textIndividual sql.NullString
		var fieldDeviceID sql.NullString
		var softwareRefID sql.NullString
		var stateTextID sql.NullString
		var notificationClassID sql.NullString
		var alarmDefinitionID sql.NullString
		if err := rows.Scan(
			&b.ID,
			&b.CreatedAt,
			&b.UpdatedAt,
			&deletedAt,
			&b.TextFix,
			&description,
			&b.GMSVisible,
			&b.Optional,
			&textIndividual,
			&b.SoftwareType,
			&b.SoftwareNumber,
			&b.HardwareType,
			&b.HardwareQuantity,
			&fieldDeviceID,
			&softwareRefID,
			&stateTextID,
			&notificationClassID,
			&alarmDefinitionID,
		); err != nil {
			return nil, err
		}
		if deletedAt.Valid {
			t := deletedAt.Time
			b.DeletedAt = &t
		}
		if description.Valid {
			v := description.String
			b.Description = &v
		}
		if textIndividual.Valid {
			v := textIndividual.String
			b.TextIndividual = &v
		}
		if fieldDeviceID.Valid {
			id, err := uuid.Parse(fieldDeviceID.String)
			if err != nil {
				return nil, err
			}
			b.FieldDeviceID = &id
		}
		if softwareRefID.Valid {
			id, err := uuid.Parse(softwareRefID.String)
			if err != nil {
				return nil, err
			}
			b.SoftwareReferenceID = &id
		}
		if stateTextID.Valid {
			id, err := uuid.Parse(stateTextID.String)
			if err != nil {
				return nil, err
			}
			b.StateTextID = &id
		}
		if notificationClassID.Valid {
			id, err := uuid.Parse(notificationClassID.String)
			if err != nil {
				return nil, err
			}
			b.NotificationClassID = &id
		}
		if alarmDefinitionID.Valid {
			id, err := uuid.Parse(alarmDefinitionID.String)
			if err != nil {
				return nil, err
			}
			b.AlarmDefinitionID = &id
		}

		items = append(items, b)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &domain.PaginatedList[domainFacility.BacnetObject]{
		Items:      items,
		Total:      total,
		Page:       page,
		TotalPages: domain.CalculateTotalPages(total, limit),
	}, nil
}

func (r *bacnetObjectRepo) GetByFieldDeviceIDs(ids []uuid.UUID) ([]*domainFacility.BacnetObject, error) {
	if len(ids) == 0 {
		return []*domainFacility.BacnetObject{}, nil
	}

	q := "SELECT id, created_at, updated_at, deleted_at, text_fix, description, gms_visible, optional, text_individual, software_type, software_number, hardware_type, hardware_quantity, field_device_id, software_reference_id, state_text_id, notification_class_id, alarm_definition_id " +
		"FROM bacnet_objects WHERE deleted_at IS NULL AND field_device_id IN (" + sqlutil.Placeholders(len(ids)) + ")"
	q = sqlutil.Rebind(r.dialect, q)

	args := make([]any, 0, len(ids))
	for _, id := range ids {
		args = append(args, id)
	}

	rows, err := r.db.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	out := make([]*domainFacility.BacnetObject, 0, 16)
	for rows.Next() {
		var b domainFacility.BacnetObject
		var deletedAt sql.NullTime
		var description sql.NullString
		var textIndividual sql.NullString
		var fieldDeviceID sql.NullString
		var softwareRefID sql.NullString
		var stateTextID sql.NullString
		var notificationClassID sql.NullString
		var alarmDefinitionID sql.NullString
		if err := rows.Scan(
			&b.ID,
			&b.CreatedAt,
			&b.UpdatedAt,
			&deletedAt,
			&b.TextFix,
			&description,
			&b.GMSVisible,
			&b.Optional,
			&textIndividual,
			&b.SoftwareType,
			&b.SoftwareNumber,
			&b.HardwareType,
			&b.HardwareQuantity,
			&fieldDeviceID,
			&softwareRefID,
			&stateTextID,
			&notificationClassID,
			&alarmDefinitionID,
		); err != nil {
			return nil, err
		}
		if deletedAt.Valid {
			t := deletedAt.Time
			b.DeletedAt = &t
		}
		if description.Valid {
			v := description.String
			b.Description = &v
		}
		if textIndividual.Valid {
			v := textIndividual.String
			b.TextIndividual = &v
		}
		if fieldDeviceID.Valid {
			id, err := uuid.Parse(fieldDeviceID.String)
			if err != nil {
				return nil, err
			}
			b.FieldDeviceID = &id
		}
		if softwareRefID.Valid {
			id, err := uuid.Parse(softwareRefID.String)
			if err != nil {
				return nil, err
			}
			b.SoftwareReferenceID = &id
		}
		if stateTextID.Valid {
			id, err := uuid.Parse(stateTextID.String)
			if err != nil {
				return nil, err
			}
			b.StateTextID = &id
		}
		if notificationClassID.Valid {
			id, err := uuid.Parse(notificationClassID.String)
			if err != nil {
				return nil, err
			}
			b.NotificationClassID = &id
		}
		if alarmDefinitionID.Valid {
			id, err := uuid.Parse(alarmDefinitionID.String)
			if err != nil {
				return nil, err
			}
			b.AlarmDefinitionID = &id
		}

		out = append(out, &b)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return out, nil
}

func (r *bacnetObjectRepo) SoftDeleteByFieldDeviceIDs(ids []uuid.UUID) error {
	if len(ids) == 0 {
		return nil
	}

	now := time.Now().UTC()
	q := "UPDATE bacnet_objects SET deleted_at = ?, updated_at = ? WHERE deleted_at IS NULL AND field_device_id IN (" + sqlutil.Placeholders(len(ids)) + ")"
	q = sqlutil.Rebind(r.dialect, q)

	args := make([]any, 0, 2+len(ids))
	args = append(args, now, now)
	for _, id := range ids {
		args = append(args, id)
	}

	_, err := r.db.Exec(q, args...)
	return err
}
