package facilitysql

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	domainProject "github.com/besart951/go_infra_link/backend/internal/domain/project"
	projectsql "github.com/besart951/go_infra_link/backend/internal/repository/projectsql"
	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

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
