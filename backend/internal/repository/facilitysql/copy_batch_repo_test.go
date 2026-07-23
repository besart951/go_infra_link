package facilitysql

import (
	"context"
	"testing"

	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func TestFieldDeviceSpecificationAssignmentUsesOneSetBasedUpdate(t *testing.T) {
	ctx := context.Background()
	db := newFieldDeviceRepoTestDB(t)
	repository := NewFieldDeviceRepository(db)

	first := createCopyTestFieldDevice(t, ctx, repository)
	second := createCopyTestFieldDevice(t, ctx, repository)
	firstSpecification := seedFacilityRecord(t, db, &domainFacility.Specification{FieldDeviceID: &first.ID})
	secondSpecification := seedFacilityRecord(t, db, &domainFacility.Specification{FieldDeviceID: &second.ID})

	updateStatements := registerUpdateCounter(t, db)
	if err := repository.AssignSpecificationIDs(ctx, map[uuid.UUID]uuid.UUID{
		first.ID:  firstSpecification.ID,
		second.ID: secondSpecification.ID,
	}); err != nil {
		t.Fatalf("assign specification IDs: %v", err)
	}
	if got := *updateStatements; got != 1 {
		t.Fatalf("update statements: got %d, want 1", got)
	}

	items, err := repository.GetByIds(ctx, []uuid.UUID{first.ID, second.ID})
	if err != nil {
		t.Fatalf("load assigned field devices: %v", err)
	}
	assigned := make(map[uuid.UUID]uuid.UUID, len(items))
	for _, item := range items {
		if item.SpecificationID == nil {
			t.Fatalf("field device %s has no specification ID", item.ID)
		}
		assigned[item.ID] = *item.SpecificationID
	}
	if assigned[first.ID] != firstSpecification.ID || assigned[second.ID] != secondSpecification.ID {
		t.Fatalf("assigned specification IDs: got %v", assigned)
	}
}

func TestFieldDeviceIDKeysetPagesAreOrderedBoundedAndSetScoped(t *testing.T) {
	const (
		fieldDeviceCount = 451
		pageSize         = 200
	)
	ctx := context.Background()
	db := newFieldDeviceRepoTestDB(t)
	repository := NewFieldDeviceRepository(db)
	parentID := uuid.New()
	excludedParentID := uuid.New()
	items := make([]*domainFacility.FieldDevice, 0, fieldDeviceCount+1)
	for i := 0; i < fieldDeviceCount; i++ {
		items = append(items, &domainFacility.FieldDevice{
			ApparatNr:                 i + 1,
			SPSControllerSystemTypeID: parentID,
			SystemPartID:              uuid.New(),
			ApparatID:                 uuid.New(),
		})
	}
	excluded := &domainFacility.FieldDevice{
		ApparatNr:                 1,
		SPSControllerSystemTypeID: excludedParentID,
		SystemPartID:              uuid.New(),
		ApparatID:                 uuid.New(),
	}
	items = append(items, excluded)
	if err := repository.BulkCreate(ctx, items, 100); err != nil {
		t.Fatalf("create field-device page fixtures: %v", err)
	}

	queryStatements := registerQueryCounter(t, db)
	var afterID *uuid.UUID
	got := make([]uuid.UUID, 0, fieldDeviceCount)
	for {
		page, err := repository.ListIDsBySPSControllerSystemTypeIDsAfter(
			ctx,
			[]uuid.UUID{parentID},
			afterID,
			pageSize,
		)
		if err != nil {
			t.Fatalf("load field-device ID page: %v", err)
		}
		if len(page) == 0 {
			break
		}
		if len(page) > pageSize {
			t.Fatalf("page length: got %d, max %d", len(page), pageSize)
		}
		got = append(got, page...)
		lastID := page[len(page)-1]
		afterID = &lastID
		if len(page) < pageSize {
			break
		}
	}

	if gotQueries := *queryStatements; gotQueries != 3 {
		t.Fatalf("page queries: got %d, want 3", gotQueries)
	}
	if len(got) != fieldDeviceCount {
		t.Fatalf("field-device IDs: got %d, want %d", len(got), fieldDeviceCount)
	}
	seen := make(map[uuid.UUID]struct{}, len(got))
	for i, id := range got {
		if id == excluded.ID {
			t.Fatalf("included ID %s from an unrelated system type", id)
		}
		if _, duplicate := seen[id]; duplicate {
			t.Fatalf("duplicate ID %s", id)
		}
		seen[id] = struct{}{}
		if i > 0 && got[i-1].String() >= id.String() {
			t.Fatalf("IDs are not strictly ordered at %d: %s then %s", i, got[i-1], id)
		}
	}
}

func TestBacnetSoftwareReferenceAssignmentUsesOneSetBasedUpdate(t *testing.T) {
	ctx := context.Background()
	db := newFieldDeviceRepoTestDB(t)
	repository := NewBacnetObjectRepository(db)

	first := &domainFacility.BacnetObject{
		TextFix:        "AI-1",
		SoftwareType:   domainFacility.BacnetSoftwareTypeAI,
		SoftwareNumber: 1,
	}
	second := &domainFacility.BacnetObject{
		TextFix:        "AO-1",
		SoftwareType:   domainFacility.BacnetSoftwareTypeAO,
		SoftwareNumber: 2,
	}
	third := &domainFacility.BacnetObject{
		TextFix:        "AO-2",
		SoftwareType:   domainFacility.BacnetSoftwareTypeAO,
		SoftwareNumber: 3,
	}
	for _, object := range []*domainFacility.BacnetObject{first, second, third} {
		if err := repository.Create(ctx, object); err != nil {
			t.Fatalf("create BACnet object: %v", err)
		}
	}

	updateStatements := registerUpdateCounter(t, db)
	if err := repository.AssignSoftwareReferenceIDs(ctx, map[uuid.UUID]uuid.UUID{
		second.ID: first.ID,
		third.ID:  first.ID,
	}); err != nil {
		t.Fatalf("assign software reference IDs: %v", err)
	}
	if got := *updateStatements; got != 1 {
		t.Fatalf("update statements: got %d, want 1", got)
	}

	items, err := repository.GetByIds(ctx, []uuid.UUID{second.ID, third.ID})
	if err != nil {
		t.Fatalf("load assigned BACnet objects: %v", err)
	}
	for _, item := range items {
		if item.SoftwareReferenceID == nil || *item.SoftwareReferenceID != first.ID {
			t.Fatalf("BACnet object %s reference: got %v, want %s", item.ID, item.SoftwareReferenceID, first.ID)
		}
	}
}

func TestBacnetAlarmValuesLoadForObjectSetInOneBatch(t *testing.T) {
	ctx := context.Background()
	db := newFieldDeviceRepoTestDB(t)
	if err := db.AutoMigrate(
		&domainFacility.AlarmField{},
		&domainFacility.AlarmTypeField{},
		&domainFacility.Unit{},
		&domainFacility.BacnetObjectAlarmValue{},
	); err != nil {
		t.Fatalf("migrate alarm-value tables: %v", err)
	}

	firstObjectID := uuid.New()
	secondObjectID := uuid.New()
	excludedObjectID := uuid.New()
	alarmField := seedFacilityRecord(t, db, &domainFacility.AlarmField{
		Key:      "limit",
		Label:    "Limit",
		DataType: "number",
	})
	alarmType := seedFacilityRecord(t, db, &domainFacility.AlarmType{Code: "high", Name: "High"})
	alarmTypeField := seedFacilityRecord(t, db, &domainFacility.AlarmTypeField{
		AlarmTypeID:  alarmType.ID,
		AlarmFieldID: alarmField.ID,
	})
	value := 12.5
	repository := NewBacnetObjectAlarmValueRepository(db)
	for _, item := range []*domainFacility.BacnetObjectAlarmValue{
		{BacnetObjectID: firstObjectID, AlarmTypeFieldID: alarmTypeField.ID, ValueNumber: &value},
		{BacnetObjectID: secondObjectID, AlarmTypeFieldID: alarmTypeField.ID, ValueNumber: &value},
		{BacnetObjectID: excludedObjectID, AlarmTypeFieldID: alarmTypeField.ID, ValueNumber: &value},
	} {
		if err := repository.Create(ctx, item); err != nil {
			t.Fatalf("create alarm value: %v", err)
		}
	}

	queryStatements := registerQueryCounter(t, db)
	values, err := repository.GetByBacnetObjectIDs(ctx, []uuid.UUID{firstObjectID, secondObjectID})
	if err != nil {
		t.Fatalf("load alarm values by BACnet object IDs: %v", err)
	}
	if len(values) != 2 {
		t.Fatalf("alarm values: got %d, want 2", len(values))
	}
	if got := *queryStatements; got != 1 {
		t.Fatalf("query statements: got %d, want 1", got)
	}
	parents := map[uuid.UUID]struct{}{}
	for _, item := range values {
		parents[item.BacnetObjectID] = struct{}{}
		if item.AlarmTypeFieldID != alarmTypeField.ID {
			t.Fatalf("alarm type field ID: got %s, want %s", item.AlarmTypeFieldID, alarmTypeField.ID)
		}
	}
	if _, ok := parents[firstObjectID]; !ok {
		t.Fatalf("missing first BACnet object's alarm value: %v", parents)
	}
	if _, ok := parents[secondObjectID]; !ok {
		t.Fatalf("missing second BACnet object's alarm value: %v", parents)
	}
	if _, ok := parents[excludedObjectID]; ok {
		t.Fatalf("included unrelated BACnet object's alarm value: %v", parents)
	}
}

func createCopyTestFieldDevice(
	t *testing.T,
	ctx context.Context,
	repository interface {
		Create(context.Context, *domainFacility.FieldDevice) error
	},
) *domainFacility.FieldDevice {
	t.Helper()
	item := &domainFacility.FieldDevice{
		ApparatNr:                 1,
		SPSControllerSystemTypeID: uuid.New(),
		SystemPartID:              uuid.New(),
		ApparatID:                 uuid.New(),
	}
	if err := repository.Create(ctx, item); err != nil {
		t.Fatalf("create field device: %v", err)
	}
	return item
}

func registerUpdateCounter(t *testing.T, db *gorm.DB) *int {
	t.Helper()
	count := 0
	callbackName := "test:count-updates:" + uuid.NewString()
	if err := db.Callback().Update().Before("gorm:update").Register(callbackName, func(*gorm.DB) {
		count++
	}); err != nil {
		t.Fatalf("register update counter: %v", err)
	}
	t.Cleanup(func() {
		_ = db.Callback().Update().Remove(callbackName)
	})
	return &count
}

func registerQueryCounter(t *testing.T, db *gorm.DB) *int {
	t.Helper()
	count := 0
	callbackName := "test:count-queries:" + uuid.NewString()
	if err := db.Callback().Query().Before("gorm:query").Register(callbackName, func(*gorm.DB) {
		count++
	}); err != nil {
		t.Fatalf("register query counter: %v", err)
	}
	t.Cleanup(func() {
		_ = db.Callback().Query().Remove(callbackName)
	})
	return &count
}
