package facilitysql

import (
	"testing"
	"time"

	hierarchyrestore "github.com/besart951/go_infra_link/backend/internal/application/hierarchyrestore"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	domainHistory "github.com/besart951/go_infra_link/backend/internal/domain/history"
	"github.com/besart951/go_infra_link/backend/internal/repository/historysql"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type restoreParents struct {
	cabinetID    uuid.UUID
	assignmentID uuid.UUID
	partID       uuid.UUID
	apparatID    uuid.UUID
}

type restoreVersionSeed struct {
	cabinetID uuid.UUID
	entityID  uuid.UUID
	snapshot  domainHistory.JSONB
	versionAt time.Time
}

func TestHierarchyRestoreChunksDeleteAndRestoreAgainstOneSnapshot(t *testing.T) {
	db := newFieldDeviceRepoTestDB(t)
	if err := db.AutoMigrate(&domainFacility.Building{}, &domainFacility.ControlCabinet{}); err != nil {
		t.Fatalf("migrate hierarchy: %v", err)
	}
	if err := historysql.AutoMigrate(db); err != nil {
		t.Fatalf("migrate history: %v", err)
	}
	parents := seedRestoreParents(t, db)
	asOf := time.Now().UTC()
	keep := seedFacilityRecord(t, db, &FieldDeviceRecord{
		SPSControllerSystemTypeID: parents.assignmentID, SystemPartID: parents.partID, ApparatID: parents.apparatID, ApparatNr: 1,
	})
	store := historysql.NewStore(db)
	keepSnapshot, _, err := store.LoadRow(t.Context(), "field_devices", keep.ID)
	if err != nil {
		t.Fatalf("load keep snapshot: %v", err)
	}
	changed := "changed after snapshot"
	if err := db.Model(&FieldDeviceRecord{}).Where("id = ?", keep.ID).Update("description", changed).Error; err != nil {
		t.Fatalf("change field device: %v", err)
	}
	remove := seedFacilityRecord(t, db, &FieldDeviceRecord{
		SPSControllerSystemTypeID: parents.assignmentID, SystemPartID: parents.partID, ApparatID: parents.apparatID, ApparatNr: 2,
	})
	removeSnapshot, _, err := store.LoadRow(t.Context(), "field_devices", remove.ID)
	if err != nil {
		t.Fatalf("load remove snapshot: %v", err)
	}
	seedRestoreVersion(t, db, restoreVersionSeed{
		cabinetID: parents.cabinetID, entityID: keep.ID, snapshot: keepSnapshot, versionAt: asOf.Add(-time.Minute),
	})
	seedRestoreVersion(t, db, restoreVersionSeed{
		cabinetID: parents.cabinetID, entityID: remove.ID, snapshot: removeSnapshot, versionAt: asOf.Add(time.Minute),
	})

	command := hierarchyrestore.Command{
		ControlCabinetID: parents.cabinetID, AsOf: asOf, Table: "field_devices",
		Limit: 1, ActorID: uuid.New(), BatchID: uuid.New(),
	}
	command.Phase = hierarchyrestore.PhaseDelete
	runRestoreChunks(t, store, command)
	assertEntityExists(t, db, keep.ID, true)
	assertEntityExists(t, db, remove.ID, false)

	command.Phase = hierarchyrestore.PhaseRestore
	command.AfterID = uuid.Nil
	runRestoreChunks(t, store, command)
	var restored FieldDeviceRecord
	if err := db.Where("id = ?", keep.ID).First(&restored).Error; err != nil {
		t.Fatalf("load restored field device: %v", err)
	}
	if restored.Description != nil {
		t.Fatalf("restored description = %q, want snapshot nil", *restored.Description)
	}
}

func seedRestoreParents(t *testing.T, db *gorm.DB) restoreParents {
	building := seedFacilityRecord(t, db, &domainFacility.Building{IWSCode: "RST", BuildingGroup: 1})
	cabinet := seedFacilityRecord(t, db, &domainFacility.ControlCabinet{BuildingID: building.ID})
	controller := seedFacilityRecord(t, db, &domainFacility.SPSController{ControlCabinetID: cabinet.ID, DeviceName: "Restore SPS"})
	typeDefinition := seedFacilityRecord(t, db, &domainFacility.SystemType{Name: "Restore Type", NumberMin: 1, NumberMax: 10})
	assignment := seedFacilityRecord(t, db, &domainFacility.SPSControllerSystemType{SPSControllerID: controller.ID, SystemTypeID: typeDefinition.ID})
	part := seedFacilityRecord(t, db, &domainFacility.SystemPart{ShortName: "RST-P", Name: "Restore Part"})
	apparat := seedFacilityRecord(t, db, &domainFacility.Apparat{ShortName: "RST-A", Name: "Restore Apparat"})
	return restoreParents{cabinetID: cabinet.ID, assignmentID: assignment.ID, partID: part.ID, apparatID: apparat.ID}
}

func seedRestoreVersion(t *testing.T, db *gorm.DB, seed restoreVersionSeed) {
	eventID := uuid.New()
	event := domainHistory.ChangeEvent{
		ID: eventID, OccurredAt: seed.versionAt, Action: domainHistory.ActionCreate,
		EntityTable: "field_devices", EntityID: seed.entityID, AfterJSON: seed.snapshot,
	}
	if err := db.Create(&event).Error; err != nil {
		t.Fatalf("create restore event: %v", err)
	}
	if err := db.Create(&domainHistory.ChangeEventScope{
		ID: uuid.New(), ChangeEventID: eventID, ScopeType: "control_cabinet",
		ScopeID: seed.cabinetID, OccurredAt: seed.versionAt,
	}).Error; err != nil {
		t.Fatalf("create restore scope: %v", err)
	}
	if err := db.Create(&domainHistory.EntityVersion{
		ID: uuid.New(), ChangeEventID: eventID, EntityTable: "field_devices",
		EntityID: seed.entityID, VersionAt: seed.versionAt, Action: domainHistory.ActionCreate, SnapshotJSON: seed.snapshot,
	}).Error; err != nil {
		t.Fatalf("create restore version: %v", err)
	}
}

func runRestoreChunks(t *testing.T, store *historysql.Store, command hierarchyrestore.Command) {
	t.Helper()
	for {
		result, err := store.RestoreChunk(t.Context(), command)
		if err != nil {
			t.Fatalf("restore chunk: %v", err)
		}
		if result.Done {
			return
		}
		command.AfterID = result.NextID
	}
}

func assertEntityExists(t *testing.T, db *gorm.DB, id uuid.UUID, want bool) {
	t.Helper()
	var count int64
	if err := db.Model(&FieldDeviceRecord{}).Where("id = ?", id).Count(&count).Error; err != nil {
		t.Fatalf("count field device: %v", err)
	}
	if (count == 1) != want {
		t.Fatalf("field device %s exists = %t, want %t", id, count == 1, want)
	}
}
