package exporting

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainExport "github.com/besart951/go_infra_link/backend/internal/domain/exporting"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	domainProject "github.com/besart951/go_infra_link/backend/internal/domain/project"
	facilitysql "github.com/besart951/go_infra_link/backend/internal/repository/facilitysql"
	projectsql "github.com/besart951/go_infra_link/backend/internal/repository/projectsql"
	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestDataProviderListFieldDevicesByControllerHydratesExportRelations(t *testing.T) {
	ctx := context.Background()
	db := newExportProviderTestDB(t)

	fieldDeviceRepo := facilitysql.NewFieldDeviceRepository(db)
	projectFieldDeviceRepo := projectsql.NewProjectFieldDeviceRepository(db)
	provider := NewDataProvider(
		fieldDeviceRepo,
		facilitysql.NewSpecificationRepository(db),
		facilitysql.NewBacnetObjectRepository(db),
		facilitysql.NewSPSControllerRepository(db),
		facilitysql.NewControlCabinetRepository(db),
	)

	systemType := seedExportProviderRecord(t, db, &domainFacility.SystemType{Name: "HVAC", NumberMin: 1, NumberMax: 99})
	controller := seedExportProviderRecord(t, db, &domainFacility.SPSController{
		ControlCabinetID: uuid.New(),
		DeviceName:       "SPS-A",
	})
	documentName := "DOC-7"
	number := 7
	spsSystemType := seedExportProviderRecord(t, db, &domainFacility.SPSControllerSystemType{
		Number:          &number,
		DocumentName:    &documentName,
		SPSControllerID: controller.ID,
		SystemTypeID:    systemType.ID,
	})
	systemPart := seedExportProviderRecord(t, db, &domainFacility.SystemPart{ShortName: "AIR", Name: "Air"})
	apparat := seedExportProviderRecord(t, db, &domainFacility.Apparat{ShortName: "PMP", Name: "Pump"})

	bmk := "FD-01"
	description := "Primary pump"
	fieldDevice := &domainFacility.FieldDevice{
		BMK:                       &bmk,
		Description:               &description,
		ApparatNr:                 11,
		SPSControllerSystemTypeID: spsSystemType.ID,
		SystemPartID:              systemPart.ID,
		ApparatID:                 apparat.ID,
	}
	if err := fieldDeviceRepo.Create(ctx, fieldDevice); err != nil {
		t.Fatalf("expected field device create to succeed, got %v", err)
	}

	supplier := "Siemens"
	specification := seedExportProviderRecord(t, db, &domainFacility.Specification{
		FieldDeviceID:         &fieldDevice.ID,
		SpecificationSupplier: &supplier,
	})
	fieldDevice.SpecificationID = &specification.ID
	if err := fieldDeviceRepo.Update(ctx, fieldDevice); err != nil {
		t.Fatalf("expected field device update to persist specification id, got %v", err)
	}

	stateText1 := "Off"
	stateText := seedExportProviderRecord(t, db, &domainFacility.StateText{RefNumber: 1, StateText1: &stateText1})
	notificationClass := seedExportProviderRecord(t, db, &domainFacility.NotificationClass{
		EventCategory:       "alarm",
		Nc:                  10,
		ObjectDescription:   "object",
		InternalDescription: "internal",
		Meaning:             "meaning",
	})
	alarmType := seedExportProviderRecord(t, db, &domainFacility.AlarmType{Code: "limit_high", Name: "Limit High"})
	seedExportProviderRecord(t, db, &domainFacility.BacnetObject{
		TextFix:             "AI1",
		SoftwareType:        domainFacility.BacnetSoftwareTypeAI,
		SoftwareNumber:      1,
		HardwareType:        domainFacility.BacnetHardwareTypeAI,
		HardwareQuantity:    1,
		FieldDeviceID:       &fieldDevice.ID,
		StateTextID:         &stateText.ID,
		NotificationClassID: &notificationClass.ID,
		AlarmTypeID:         &alarmType.ID,
	})

	projectID := uuid.New()
	if err := projectFieldDeviceRepo.Create(ctx, &domainProject.ProjectFieldDevice{ProjectID: projectID, FieldDeviceID: fieldDevice.ID}); err != nil {
		t.Fatalf("expected project field device link create to succeed, got %v", err)
	}

	items, total, err := provider.ListFieldDevicesByController(ctx, controller.ID, domainExport.Request{
		ProjectIDs: []uuid.UUID{projectID},
	}, 1, 10)
	if err != nil {
		t.Fatalf("expected export field device list to succeed, got %v", err)
	}
	if total != 1 || len(items) != 1 {
		t.Fatalf("expected one export field device, got total=%d items=%d", total, len(items))
	}

	item := items[0]
	if item.Specification == nil || item.Specification.SpecificationSupplier == nil || *item.Specification.SpecificationSupplier != supplier {
		t.Fatalf("expected export field device to include specification, got %+v", item.Specification)
	}
	if len(item.BacnetObjects) != 1 {
		t.Fatalf("expected export field device to include one bacnet object, got %+v", item.BacnetObjects)
	}
	bacnetObject := item.BacnetObjects[0]
	if bacnetObject.StateText == nil || bacnetObject.StateText.StateText1 == nil || *bacnetObject.StateText.StateText1 != stateText1 {
		t.Fatalf("expected bacnet object state text to be hydrated, got %+v", bacnetObject.StateText)
	}
	if bacnetObject.NotificationClass == nil || bacnetObject.NotificationClass.Nc != notificationClass.Nc {
		t.Fatalf("expected bacnet object notification class to be hydrated, got %+v", bacnetObject.NotificationClass)
	}
	if bacnetObject.AlarmType == nil || bacnetObject.AlarmType.Name != alarmType.Name {
		t.Fatalf("expected bacnet object alarm type to be hydrated, got %+v", bacnetObject.AlarmType)
	}
}

func newExportProviderTestDB(t *testing.T) *gorm.DB {
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
		&projectsql.ProjectFieldDeviceRecord{},
		&domainFacility.SystemType{},
		&domainFacility.SPSController{},
		&domainFacility.SPSControllerSystemType{},
		&domainFacility.SystemPart{},
		&domainFacility.Apparat{},
		&domainFacility.Specification{},
		&facilitysql.FieldDeviceRecord{},
		&domainFacility.StateText{},
		&domainFacility.NotificationClass{},
		&domainFacility.AlarmType{},
		&domainFacility.BacnetObject{},
	}
	if err := db.AutoMigrate(models...); err != nil {
		t.Fatalf("expected export provider test tables to migrate, got %v", err)
	}

	return db
}

func seedExportProviderRecord[T interface{ GetBase() *domain.Base }](t *testing.T, db *gorm.DB, entity T) T {
	t.Helper()

	if err := entity.GetBase().InitForCreate(time.Now().UTC()); err != nil {
		t.Fatalf("expected base init to succeed, got %v", err)
	}
	if err := db.Create(entity).Error; err != nil {
		t.Fatalf("expected record seed to succeed, got %v", err)
	}
	return entity
}
