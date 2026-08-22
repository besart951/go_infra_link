package facilitysql

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/besart951/go_infra_link/backend/internal/cursor"
	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	domainFieldDevice "github.com/besart951/go_infra_link/backend/internal/domain/facility/fielddevice"
	domainProject "github.com/besart951/go_infra_link/backend/internal/domain/project"
	projectsql "github.com/besart951/go_infra_link/backend/internal/repository/projectsql"
	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestFieldDeviceRepo_CursorTraversesBothDirectionsWithoutDuplicates(t *testing.T) {
	ctx := context.Background()
	db := newFieldDeviceRepoTestDB(t)
	repo := NewFieldDeviceRepository(db)
	reader := repo.(domainFieldDevice.CursorReader)
	devices := seedCursorFieldDevices(t, db, repo, 5)

	first, err := reader.GetCursorPage(ctx, domainFacility.FieldDeviceCursorQuery{Limit: 2})
	if err != nil {
		t.Fatalf("first cursor page: %v", err)
	}
	second, err := reader.GetCursorPage(ctx, domainFacility.FieldDeviceCursorQuery{Limit: 2, Cursor: first.NextCursor})
	if err != nil {
		t.Fatalf("second cursor page: %v", err)
	}
	back, err := reader.GetCursorPage(ctx, domainFacility.FieldDeviceCursorQuery{Limit: 2, Cursor: second.PreviousCursor})
	if err != nil {
		t.Fatalf("previous cursor page: %v", err)
	}

	wantFirst := []uuid.UUID{devices[4].ID, devices[3].ID}
	assertFieldDeviceIDs(t, first.Items, wantFirst)
	assertFieldDeviceIDs(t, back.Items, wantFirst)
	assertFieldDeviceIDs(t, second.Items, []uuid.UUID{devices[2].ID, devices[1].ID})
	if first.PreviousCursor != "" || first.NextCursor == "" || second.PreviousCursor == "" || second.NextCursor == "" {
		t.Fatalf("unexpected cursors: first=%+v second=%+v", first, second)
	}
}

func TestFieldDeviceRepo_CursorRejectsChangedQuery(t *testing.T) {
	db := newFieldDeviceRepoTestDB(t)
	repo := NewFieldDeviceRepository(db)
	reader := repo.(domainFieldDevice.CursorReader)
	seedCursorFieldDevices(t, db, repo, 3)

	page, err := reader.GetCursorPage(t.Context(), domainFacility.FieldDeviceCursorQuery{Limit: 1, Search: "FD"})
	if err != nil {
		t.Fatalf("first cursor page: %v", err)
	}
	_, err = reader.GetCursorPage(t.Context(), domainFacility.FieldDeviceCursorQuery{Limit: 1, Search: "other", Cursor: page.NextCursor})
	if !errors.Is(err, cursor.ErrInvalid) {
		t.Fatalf("changed-query error = %v, want %v", err, cursor.ErrInvalid)
	}
}

func TestFieldDeviceRepo_DeleteAtVersionIsConditional(t *testing.T) {
	db := newFieldDeviceRepoTestDB(t)
	repo := NewFieldDeviceRepository(db)
	device := seedCursorFieldDevices(t, db, repo, 1)[0]
	deleter := repo.(interface {
		DeleteAtVersion(context.Context, domainFacility.FieldDeviceDeleteCommand) error
	})
	stale := device.Version + 1

	err := deleter.DeleteAtVersion(t.Context(), domainFacility.FieldDeviceDeleteCommand{ID: device.ID, BaseVersion: domain.AggregateVersion(stale)})
	if !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("stale delete error = %v", err)
	}
	if items, _ := repo.GetByIds(t.Context(), []uuid.UUID{device.ID}); len(items) != 1 {
		t.Fatal("stale delete removed the field device")
	}
	if err := deleter.DeleteAtVersion(t.Context(), domainFacility.FieldDeviceDeleteCommand{ID: device.ID, BaseVersion: domain.AggregateVersion(device.Version)}); err != nil {
		t.Fatal(err)
	}
}

func assertFieldDeviceIDs(t *testing.T, items []domainFacility.FieldDevice, want []uuid.UUID) {
	t.Helper()
	got := make([]uuid.UUID, len(items))
	for index := range items {
		got[index] = items[index].ID
	}
	if fmt.Sprint(got) != fmt.Sprint(want) {
		t.Fatalf("field device ids = %v, want %v", got, want)
	}
}

func seedCursorFieldDevices(t *testing.T, db *gorm.DB, repo domainFieldDevice.FieldDeviceStore, count int) []*domainFacility.FieldDevice {
	t.Helper()
	systemType := seedFacilityRecord(t, db, &domainFacility.SystemType{Name: "Cursor", NumberMin: 1, NumberMax: 99})
	controller := seedFacilityRecord(t, db, &domainFacility.SPSController{ControlCabinetID: uuid.New(), DeviceName: "Cursor-SPS"})
	number := 1
	controllerType := seedFacilityRecord(t, db, &domainFacility.SPSControllerSystemType{Number: &number, SPSControllerID: controller.ID, SystemTypeID: systemType.ID})
	part := seedFacilityRecord(t, db, &domainFacility.SystemPart{ShortName: "CUR", Name: "Cursor Part"})
	apparat := seedFacilityRecord(t, db, &domainFacility.Apparat{ShortName: "CUR", Name: "Cursor Apparat"})
	items := make([]*domainFacility.FieldDevice, count)
	for index := range count {
		bmk := fmt.Sprintf("FD-%02d", index)
		item := &domainFacility.FieldDevice{BMK: &bmk, ApparatNr: index + 1, SPSControllerSystemTypeID: controllerType.ID, SystemPartID: part.ID, ApparatID: apparat.ID}
		if err := repo.Create(t.Context(), item); err != nil {
			t.Fatalf("create cursor field device: %v", err)
		}
		createdAt := time.Date(2026, 1, 1, 0, index, 0, 0, time.UTC)
		if err := db.Model(&FieldDeviceRecord{}).Where("id = ?", item.ID).Updates(map[string]any{"created_at": createdAt, "updated_at": createdAt}).Error; err != nil {
			t.Fatalf("set cursor timestamp: %v", err)
		}
		item.CreatedAt = createdAt
		items[index] = item
	}
	return items
}

func TestFieldDeviceRepo_ProjectFilteredListMapsAggregateRelations(t *testing.T) {
	ctx := context.Background()
	db := newFieldDeviceRepoTestDB(t)
	repo := NewFieldDeviceRepository(db)
	projectLinkRepo := projectsql.NewProjectFieldDeviceRepository(db)

	systemType := seedFacilityRecord(t, db, &domainFacility.SystemType{Name: "HVAC", NumberMin: 1, NumberMax: 99})
	controller := seedFacilityRecord(t, db, &domainFacility.SPSController{ControlCabinetID: uuid.New(), DeviceName: "SPS-A"})
	documentName := "DOC-7"
	number := 7
	spsSystemType := seedFacilityRecord(t, db, &domainFacility.SPSControllerSystemType{
		Number:          &number,
		DocumentName:    &documentName,
		SPSControllerID: controller.ID,
		SystemTypeID:    systemType.ID,
	})
	systemPart := seedFacilityRecord(t, db, &domainFacility.SystemPart{ShortName: "AIR", Name: "Air"})
	apparat := seedFacilityRecord(t, db, &domainFacility.Apparat{ShortName: "PMP", Name: "Pump"})

	bmk := "FD-01"
	description := "Primary pump"
	textIndividuell := "custom text"
	fieldDevice := &domainFacility.FieldDevice{
		BMK:                       &bmk,
		Description:               &description,
		ApparatNr:                 11,
		TextIndividuell:           &textIndividuell,
		SPSControllerSystemTypeID: spsSystemType.ID,
		SystemPartID:              systemPart.ID,
		ApparatID:                 apparat.ID,
	}
	if err := repo.Create(ctx, fieldDevice); err != nil {
		t.Fatalf("expected field device create to succeed, got %v", err)
	}

	supplier := "Siemens"
	specification := seedFacilityRecord(t, db, &domainFacility.Specification{FieldDeviceID: &fieldDevice.ID, SpecificationSupplier: &supplier})
	fieldDevice.SpecificationID = &specification.ID
	if err := repo.Update(ctx, fieldDevice); err != nil {
		t.Fatalf("expected field device update to persist specification id, got %v", err)
	}

	stateText := seedFacilityRecord(t, db, &domainFacility.StateText{RefNumber: 1})
	notificationClass := seedFacilityRecord(t, db, &domainFacility.NotificationClass{
		EventCategory:       "alarm",
		Nc:                  10,
		ObjectDescription:   "object",
		InternalDescription: "internal",
		Meaning:             "meaning",
	})
	alarmType := seedFacilityRecord(t, db, &domainFacility.AlarmType{Code: "limit_high", Name: "Limit High"})
	seedFacilityRecord(t, db, &domainFacility.BacnetObject{
		TextFix:             "AI1",
		SoftwareType:        domainFacility.BacnetSoftwareTypeAI,
		SoftwareNumber:      1,
		FieldDeviceID:       &fieldDevice.ID,
		StateTextID:         &stateText.ID,
		NotificationClassID: &notificationClass.ID,
		AlarmTypeID:         &alarmType.ID,
	})

	projectID := uuid.New()
	if err := projectLinkRepo.Create(ctx, &domainProject.ProjectFieldDevice{ProjectID: projectID, FieldDeviceID: fieldDevice.ID}); err != nil {
		t.Fatalf("expected project field device link create to succeed, got %v", err)
	}

	list, err := repo.GetPaginatedListWithFilters(ctx, domain.PaginationParams{Page: 1, Limit: 10}, domainFacility.FieldDeviceFilterParams{ProjectID: &projectID})
	if err != nil {
		t.Fatalf("expected project-filtered field device listing to succeed, got %v", err)
	}
	if len(list.Items) != 1 {
		t.Fatalf("expected one mapped field device, got %+v", list.Items)
	}

	item := list.Items[0]
	if item.ID != fieldDevice.ID || item.SPSControllerSystemType.ID != spsSystemType.ID || item.SPSControllerSystemType.SPSController.ID != controller.ID {
		t.Fatalf("expected mapped controller hierarchy, got %+v", item)
	}
	if item.SystemPart.ID != systemPart.ID || item.Apparat.ID != apparat.ID {
		t.Fatalf("expected mapped system part and apparat, got %+v", item)
	}
	if item.SpecificationID == nil || *item.SpecificationID != specification.ID {
		t.Fatalf("expected specification id to be preserved, got %+v", item.SpecificationID)
	}
	if item.Specification != nil {
		t.Fatalf("expected list rows to omit specification details, got %+v", item.Specification)
	}
	if len(item.BacnetObjects) != 0 {
		t.Fatalf("expected list rows to omit bacnet objects, got %+v", item.BacnetObjects)
	}
}

func TestFieldDeviceRepoUpdatePersistsEveryScalarRecordField(t *testing.T) {
	ctx := context.Background()
	db := newFieldDeviceRepoTestDB(t)
	repo := NewFieldDeviceRepository(db)

	bmk, description, text := "FD-01", "original description", "original text"
	fieldDevice := &domainFacility.FieldDevice{
		BMK:                       &bmk,
		Description:               &description,
		TextIndividuell:           &text,
		ApparatNr:                 1,
		SPSControllerSystemTypeID: uuid.New(),
		SystemPartID:              uuid.New(),
		ApparatID:                 uuid.New(),
	}
	if err := repo.Create(ctx, fieldDevice); err != nil {
		t.Fatalf("create field device: %v", err)
	}
	initialVersion := fieldDevice.Version

	updatedBMK := "FD-02"
	fieldDevice.BMK = &updatedBMK
	fieldDevice.Description = nil
	fieldDevice.TextIndividuell = nil
	fieldDevice.ApparatNr = 2
	fieldDevice.SPSControllerSystemTypeID = uuid.New()
	fieldDevice.SystemPartID = uuid.New()
	fieldDevice.ApparatID = uuid.New()
	if err := repo.Update(ctx, fieldDevice); err != nil {
		t.Fatalf("update field device: %v", err)
	}
	if fieldDevice.Version != initialVersion+1 {
		t.Fatalf("version = %d, want %d", fieldDevice.Version, initialVersion+1)
	}

	items, err := repo.GetByIds(ctx, []uuid.UUID{fieldDevice.ID})
	if err != nil {
		t.Fatalf("reload field device: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("reloaded %d field devices, want 1", len(items))
	}
	stored := items[0]
	if stored.BMK == nil || *stored.BMK != updatedBMK || stored.Description != nil || stored.TextIndividuell != nil || stored.ApparatNr != 2 {
		t.Fatalf("scalar fields were not persisted: %+v", stored)
	}
	if stored.SPSControllerSystemTypeID != fieldDevice.SPSControllerSystemTypeID || stored.SystemPartID != fieldDevice.SystemPartID || stored.ApparatID != fieldDevice.ApparatID {
		t.Fatalf("reference ids were not persisted: %+v", stored)
	}
}

func TestFieldDeviceRepo_ShortSearchFindsBMKSubstring(t *testing.T) {
	ctx := context.Background()
	db := newFieldDeviceRepoTestDB(t)
	repo := NewFieldDeviceRepository(db)

	systemType := seedFacilityRecord(t, db, &domainFacility.SystemType{Name: "HVAC", NumberMin: 1, NumberMax: 99})
	controller := seedFacilityRecord(t, db, &domainFacility.SPSController{ControlCabinetID: uuid.New(), DeviceName: "SPS-A"})
	number := 7
	spsSystemType := seedFacilityRecord(t, db, &domainFacility.SPSControllerSystemType{
		Number:          &number,
		SPSControllerID: controller.ID,
		SystemTypeID:    systemType.ID,
	})
	systemPart := seedFacilityRecord(t, db, &domainFacility.SystemPart{ShortName: "AIR", Name: "Air"})
	apparat := seedFacilityRecord(t, db, &domainFacility.Apparat{ShortName: "PMP", Name: "Pump"})

	bmk := "PERF-FD-064753"
	fieldDevice := &domainFacility.FieldDevice{
		BMK:                       &bmk,
		ApparatNr:                 11,
		SPSControllerSystemTypeID: spsSystemType.ID,
		SystemPartID:              systemPart.ID,
		ApparatID:                 apparat.ID,
	}
	if err := repo.Create(ctx, fieldDevice); err != nil {
		t.Fatalf("expected field device create to succeed, got %v", err)
	}

	list, err := repo.GetPaginatedList(ctx, domain.PaginationParams{Page: 1, Limit: 10, Search: "53"})
	if err != nil {
		t.Fatalf("expected short bmk search to succeed, got %v", err)
	}
	if len(list.Items) != 1 || list.Items[0].ID != fieldDevice.ID {
		t.Fatalf("expected short search to find bmk suffix, got %+v", list.Items)
	}
}

func TestFieldDeviceRepoHidesDevicesBelowDeletingController(t *testing.T) {
	db := newFieldDeviceRepoTestDB(t)
	controller := seedFacilityRecord(t, db, &domainFacility.SPSController{
		ControlCabinetID: uuid.New(), DeviceName: "Deleting SPS",
	})
	systemType := seedFacilityRecord(t, db, &domainFacility.SPSControllerSystemType{
		SPSControllerID: controller.ID, SystemTypeID: uuid.New(),
	})
	device := seedFacilityRecord(t, db, &FieldDeviceRecord{
		SPSControllerSystemTypeID: systemType.ID, SystemPartID: uuid.New(),
		ApparatID: uuid.New(), ApparatNr: 1,
	})
	if err := db.Create(&facilityLifecycleTestRecord{
		Kind: "sps_controller", ResourceID: controller.ID, State: "deleting",
		OwnerID: uuid.New(), JobID: uuid.New(), CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}).Error; err != nil {
		t.Fatalf("create lifecycle lock: %v", err)
	}

	repo := NewFieldDeviceRepository(db)
	items, err := repo.GetByIds(t.Context(), []uuid.UUID{device.ID})
	if err != nil {
		t.Fatalf("GetByIds() error = %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("visible items = %d, want zero", len(items))
	}
	changed := "blocked"
	lockedDevice := toFieldDeviceDomain(device)
	lockedDevice.Description = &changed
	if err := repo.Update(t.Context(), lockedDevice); !errors.Is(err, domainFacility.ErrAggregateLocked) {
		t.Fatalf("Update() error = %v, want %v", err, domainFacility.ErrAggregateLocked)
	}
}

type facilityLifecycleTestRecord struct {
	Kind       string    `gorm:"primaryKey"`
	ResourceID uuid.UUID `gorm:"primaryKey"`
	State      string
	OwnerID    uuid.UUID
	JobID      uuid.UUID
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func (facilityLifecycleTestRecord) TableName() string {
	return lifecycleTable
}

func newFieldDeviceRepoTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", strings.NewReplacer("/", "_", " ", "_", "#", "_").Replace(t.Name()))
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{DisableForeignKeyConstraintWhenMigrating: true})
	if err != nil {
		t.Fatalf("expected sqlite db to open, got %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("expected sql db handle, got %v", err)
	}
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() {
		_ = sqlDB.Close()
	})

	models := []any{
		&facilityLifecycleTestRecord{},
		&projectsql.ProjectControlCabinetRecord{},
		&projectsql.ProjectSPSControllerRecord{},
		&projectsql.ProjectFieldDeviceRecord{},
		&domainFacility.SystemType{},
		&domainFacility.SPSController{},
		&domainFacility.SPSControllerSystemType{},
		&domainFacility.SystemPart{},
		&domainFacility.Apparat{},
		&domainFacility.Specification{},
		&FieldDeviceRecord{},
		&domainFacility.StateText{},
		&domainFacility.NotificationClass{},
		&domainFacility.AlarmType{},
		&domainFacility.BacnetObject{},
	}
	if err := db.AutoMigrate(models...); err != nil {
		t.Fatalf("expected field device repo tables to migrate, got %v", err)
	}

	return db
}

func seedFacilityRecord[T interface{ GetBase() *domain.Base }](t *testing.T, db *gorm.DB, entity T) T {
	t.Helper()

	if err := entity.GetBase().InitForCreate(time.Now().UTC()); err != nil {
		t.Fatalf("expected base init to succeed, got %v", err)
	}
	if err := db.Create(entity).Error; err != nil {
		t.Fatalf("expected record seed to succeed, got %v", err)
	}
	return entity
}
