package facilitysql

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestBacnetReferenceUsageRepo_CountByResource(t *testing.T) {
	ctx := context.Background()
	db := newBacnetReferenceUsageRepoTestDB(t)
	repo := NewBacnetReferenceUsageRepository(db)

	systemType := seedBacnetReferenceUsageRecord(t, db, &domainFacility.SystemType{Name: "HVAC", NumberMin: 1, NumberMax: 99})
	controller := seedBacnetReferenceUsageRecord(t, db, &domainFacility.SPSController{ControlCabinetID: uuid.New(), DeviceName: "SPS-A"})
	number := 1
	spsSystemType := seedBacnetReferenceUsageRecord(t, db, &domainFacility.SPSControllerSystemType{
		Number:          &number,
		SPSControllerID: controller.ID,
		SystemTypeID:    systemType.ID,
	})
	systemPart := seedBacnetReferenceUsageRecord(t, db, &domainFacility.SystemPart{ShortName: "AIR", Name: "Air"})
	apparat := seedBacnetReferenceUsageRecord(t, db, &domainFacility.Apparat{ShortName: "PMP", Name: "Pump"})
	if err := db.Model(systemPart).Association("Apparats").Append(apparat); err != nil {
		t.Fatalf("expected system part apparat association to succeed, got %v", err)
	}

	fieldDevice := seedBacnetReferenceUsageRecord(t, db, &FieldDeviceRecord{
		SPSControllerSystemTypeID: spsSystemType.ID,
		SystemPartID:              systemPart.ID,
		ApparatID:                 apparat.ID,
		ApparatNr:                 1,
	})
	stateText := seedBacnetReferenceUsageRecord(t, db, &domainFacility.StateText{RefNumber: 10})
	notificationClass := seedBacnetReferenceUsageRecord(t, db, &domainFacility.NotificationClass{
		EventCategory:       "alarm",
		Nc:                  1,
		ObjectDescription:   "object",
		InternalDescription: "internal",
		Meaning:             "meaning",
	})
	alarmType := seedBacnetReferenceUsageRecord(t, db, &domainFacility.AlarmType{Code: "limit", Name: "Limit"})
	alarmDefinition := seedBacnetReferenceUsageRecord(t, db, &domainFacility.AlarmDefinition{Name: "Limit alarm", AlarmTypeID: &alarmType.ID})

	seedBacnetReferenceUsageRecord(t, db, &domainFacility.BacnetObject{
		TextFix:             "AI1",
		SoftwareType:        domainFacility.BacnetSoftwareTypeAI,
		SoftwareNumber:      1,
		FieldDeviceID:       &fieldDevice.ID,
		StateTextID:         &stateText.ID,
		NotificationClassID: &notificationClass.ID,
		AlarmTypeID:         &alarmType.ID,
	})
	templateObject := seedBacnetReferenceUsageRecord(t, db, &domainFacility.BacnetObject{
		TextFix:        "AI2",
		SoftwareType:   domainFacility.BacnetSoftwareTypeAI,
		SoftwareNumber: 2,
		StateTextID:    &stateText.ID,
		AlarmTypeID:    &alarmType.ID,
	})
	objectData := seedBacnetReferenceUsageRecord(t, db, &domainFacility.ObjectData{
		Description: "Template",
		Version:     "1.0",
		IsActive:    true,
	})
	if err := db.Model(objectData).Association("Apparats").Append(apparat); err != nil {
		t.Fatalf("expected object data apparat association to succeed, got %v", err)
	}
	if err := db.Model(objectData).Association("BacnetObjects").Append(templateObject); err != nil {
		t.Fatalf("expected object data bacnet association to succeed, got %v", err)
	}

	tests := []struct {
		name     string
		resource domainFacility.BacnetReferenceResource
		id       uuid.UUID
		want     int64
	}{
		{"apparat counts field device and object data paths", domainFacility.BacnetReferenceResourceApparat, apparat.ID, 2},
		{"system part counts field device and object data paths", domainFacility.BacnetReferenceResourceSystemPart, systemPart.ID, 2},
		{"system type counts field device path", domainFacility.BacnetReferenceResourceSystemType, systemType.ID, 1},
		{"state text counts direct bacnet objects", domainFacility.BacnetReferenceResourceStateText, stateText.ID, 2},
		{"notification class counts direct bacnet objects", domainFacility.BacnetReferenceResourceNotificationClass, notificationClass.ID, 1},
		{"alarm type counts direct bacnet objects", domainFacility.BacnetReferenceResourceAlarmType, alarmType.ID, 2},
		{"alarm definition counts objects by alarm type", domainFacility.BacnetReferenceResourceAlarmDefinition, alarmDefinition.ID, 2},
		{"object data counts linked template objects", domainFacility.BacnetReferenceResourceObjectData, objectData.ID, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			counts, err := repo.CountByResource(ctx, tt.resource, []uuid.UUID{tt.id, uuid.New()})
			if err != nil {
				t.Fatalf("expected usage count to succeed, got %v", err)
			}
			if counts[tt.id] != tt.want {
				t.Fatalf("expected count %d, got %d", tt.want, counts[tt.id])
			}
		})
	}
}

func newBacnetReferenceUsageRepoTestDB(t *testing.T) *gorm.DB {
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
		&domainFacility.SystemType{},
		&domainFacility.SPSController{},
		&domainFacility.SPSControllerSystemType{},
		&domainFacility.SystemPart{},
		&domainFacility.Apparat{},
		&FieldDeviceRecord{},
		&domainFacility.StateText{},
		&domainFacility.NotificationClass{},
		&domainFacility.AlarmType{},
		&domainFacility.AlarmDefinition{},
		&domainFacility.BacnetObject{},
		&domainFacility.ObjectData{},
	}
	if err := db.AutoMigrate(models...); err != nil {
		t.Fatalf("expected usage repo tables to migrate, got %v", err)
	}

	return db
}

func seedBacnetReferenceUsageRecord[T interface{ GetBase() *domain.Base }](t *testing.T, db *gorm.DB, entity T) T {
	t.Helper()

	if err := entity.GetBase().InitForCreate(time.Now().UTC()); err != nil {
		t.Fatalf("expected base init to succeed, got %v", err)
	}
	if err := db.Create(entity).Error; err != nil {
		t.Fatalf("expected record seed to succeed, got %v", err)
	}
	return entity
}
