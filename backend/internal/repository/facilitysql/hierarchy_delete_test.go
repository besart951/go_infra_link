package facilitysql

import (
	"testing"
	"time"

	hierarchydelete "github.com/besart951/go_infra_link/backend/internal/application/hierarchydelete"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	domainHistory "github.com/besart951/go_infra_link/backend/internal/domain/history"
	"github.com/besart951/go_infra_link/backend/internal/repository/historysql"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func TestHierarchyDeleteStoreDeletesCabinetInRestartableChunks(t *testing.T) {
	db := newFieldDeviceRepoTestDB(t)
	if err := db.AutoMigrate(
		&domainFacility.Building{}, &domainFacility.ControlCabinet{}, &domainFacility.BacnetObjectAlarmValue{},
	); err != nil {
		t.Fatalf("migrate hierarchy roots: %v", err)
	}
	if err := historysql.AutoMigrate(db); err != nil {
		t.Fatalf("migrate history: %v", err)
	}

	root := seedDeleteHierarchy(t, db)
	store := NewHierarchyDeleteStore(db)
	command := hierarchydelete.Command{
		RootKind: hierarchydelete.RootControlCabinet, RootID: root,
		Limit: 1, ActorID: uuid.New(), BatchID: uuid.New(),
	}
	for _, stage := range hierarchydelete.Stages(command.RootKind) {
		command.Stage = stage
		for {
			result, err := store.DeleteChunk(t.Context(), command)
			if err != nil {
				t.Fatalf("delete stage %s: %v", stage, err)
			}
			if result.Done {
				break
			}
		}
	}

	assertDeleteTableCount(t, db, "field_devices", 0)
	assertDeleteTableCount(t, db, "sps_controller_system_types", 0)
	assertDeleteTableCount(t, db, "sps_controllers", 0)
	assertDeleteTableCount(t, db, "control_cabinets", 0)
	assertDeleteTableCount(t, db, lifecycleTable, 0)
	var fieldDeviceEvents int64
	db.Model(&domainHistory.ChangeEvent{}).
		Where("entity_table = ? AND action = ?", "field_devices", domainHistory.ActionDelete).
		Count(&fieldDeviceEvents)
	if fieldDeviceEvents != 2 {
		t.Fatalf("field device delete events = %d, want 2", fieldDeviceEvents)
	}
}

func seedDeleteHierarchy(t *testing.T, db *gorm.DB) uuid.UUID {
	t.Helper()
	building := seedFacilityRecord(t, db, &domainFacility.Building{IWSCode: "DEL", BuildingGroup: 1})
	number := "DEL-1"
	cabinet := seedFacilityRecord(t, db, &domainFacility.ControlCabinet{
		BuildingID: building.ID, ControlCabinetNr: &number,
	})
	controller := seedFacilityRecord(t, db, &domainFacility.SPSController{
		ControlCabinetID: cabinet.ID, DeviceName: "Delete Controller",
	})
	systemType := seedFacilityRecord(t, db, &domainFacility.SystemType{
		Name: "Delete Type", NumberMin: 1, NumberMax: 99,
	})
	assignment := seedFacilityRecord(t, db, &domainFacility.SPSControllerSystemType{
		SPSControllerID: controller.ID, SystemTypeID: systemType.ID,
	})
	part := seedFacilityRecord(t, db, &domainFacility.SystemPart{ShortName: "DEL-P", Name: "Delete Part"})
	apparat := seedFacilityRecord(t, db, &domainFacility.Apparat{ShortName: "DEL-A", Name: "Delete Apparat"})
	for number := 1; number <= 2; number++ {
		seedFacilityRecord(t, db, &FieldDeviceRecord{
			SPSControllerSystemTypeID: assignment.ID, SystemPartID: part.ID,
			ApparatID: apparat.ID, ApparatNr: number,
		})
	}
	now := time.Now().UTC()
	if err := db.Create(&facilityLifecycleTestRecord{
		Kind: string(hierarchydelete.RootControlCabinet), ResourceID: cabinet.ID,
		State: "deleting", OwnerID: uuid.New(), JobID: uuid.New(), CreatedAt: now, UpdatedAt: now,
	}).Error; err != nil {
		t.Fatalf("create lifecycle lock: %v", err)
	}
	return cabinet.ID
}

func assertDeleteTableCount(t *testing.T, db *gorm.DB, table string, want int64) {
	t.Helper()
	var count int64
	if err := db.Table(table).Count(&count).Error; err != nil {
		t.Fatalf("count %s: %v", table, err)
	}
	if count != want {
		t.Fatalf("%s count = %d, want %d", table, count, want)
	}
}
