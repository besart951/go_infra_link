package db

import (
	"strings"
	"testing"
	"time"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	facilitysql "github.com/besart951/go_infra_link/backend/internal/repository/facilitysql"
	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestFacilityOwnershipMigrationBackfillsTemplates(t *testing.T) {
	db := ownershipMigrationDB(t)
	objectDataID := createMigrationObjectData(t, db, "Template A")
	objectID := createMigrationBacnetObject(t, db)
	linkMigrationTemplate(t, db, objectDataID, objectID)

	if err := migrateFacilityOwnership(db); err != nil {
		t.Fatal(err)
	}
	var record facilitysql.BacnetObjectTemplateRecord
	if err := db.First(&record, "id = ?", objectID).Error; err != nil {
		t.Fatal(err)
	}
	if record.ObjectDataID != objectDataID {
		t.Fatalf("owner = %s, want %s", record.ObjectDataID, objectDataID)
	}
}

func TestFacilityOwnershipMigrationReportsConflictingIDs(t *testing.T) {
	db := ownershipMigrationDB(t)
	firstOwner := createMigrationObjectData(t, db, "Template A")
	secondOwner := createMigrationObjectData(t, db, "Template B")
	objectID := createMigrationBacnetObject(t, db)
	linkMigrationTemplate(t, db, firstOwner, objectID)
	linkMigrationTemplate(t, db, secondOwner, objectID)

	err := migrateFacilityOwnership(db)
	if err == nil || !strings.Contains(err.Error(), objectID.String()) {
		t.Fatalf("expected concrete conflicting ID, got %v", err)
	}
	if db.Migrator().HasTable(&facilitysql.BacnetObjectTemplateRecord{}) {
		t.Fatal("migration changed schema despite failed preflight")
	}
}

func ownershipMigrationDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:"+uuid.NewString()+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := autoMigrateCurrentSchema(db); err != nil {
		t.Fatal(err)
	}
	return db
}

func createMigrationObjectData(t *testing.T, db *gorm.DB, description string) uuid.UUID {
	t.Helper()
	id := uuid.New()
	item := domainFacility.ObjectData{
		Base:        domain.Base{ID: id, CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC(), Version: 1},
		Description: description, Version: "1", IsActive: true,
	}
	if err := db.Omit("BacnetObjects", "Apparats").Create(&item).Error; err != nil {
		t.Fatal(err)
	}
	return id
}

func createMigrationBacnetObject(t *testing.T, db *gorm.DB) uuid.UUID {
	t.Helper()
	id := uuid.New()
	item := domainFacility.BacnetObject{
		Base:    domain.Base{ID: id, CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC(), Version: 1},
		TextFix: "AI01", SoftwareType: domainFacility.BacnetSoftwareTypeAI, SoftwareNumber: 1,
	}
	if err := db.Omit("AlarmValues").Create(&item).Error; err != nil {
		t.Fatal(err)
	}
	return id
}

func linkMigrationTemplate(t *testing.T, db *gorm.DB, objectDataID, objectID uuid.UUID) {
	t.Helper()
	err := db.Table("object_data_bacnet_objects").Create(map[string]any{
		"object_data_id": objectDataID, "bacnet_object_id": objectID,
	}).Error
	if err != nil {
		t.Fatal(err)
	}
}
