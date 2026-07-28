package wire

import (
	"context"
	"os"
	"testing"

	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// This test is opt-in because it exercises PostgreSQL-specific history
// snapshots, constraints, and trigger behavior. Point it at a disposable,
// fully migrated database.
func TestHierarchyDeleteCleanerRemovesBoundedOwnedGraphAndPreservesSharedBacnet(
	t *testing.T,
) {
	dsn := os.Getenv("FACILITY_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("FACILITY_TEST_DATABASE_URL is not set")
	}

	database, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("connect test database: %v", err)
	}
	tx := database.Begin()
	if tx.Error != nil {
		t.Fatalf("begin test transaction: %v", tx.Error)
	}
	t.Cleanup(func() {
		_ = tx.Rollback().Error
	})

	ctx := context.Background()
	repositories, err := NewRepositories(tx)
	if err != nil {
		t.Fatalf("construct repositories: %v", err)
	}
	cleaner := &hierarchyDeleteCleaner{db: tx, repos: repositories}

	ids := newHierarchyDeleteFixtureIDs()
	mustExecHierarchyDeleteFixture(t, tx, `
		INSERT INTO buildings (id, iws_code, building_group)
		VALUES (?, ?, ?)
	`, ids.building, "DEL-"+ids.building.String(), 91)
	mustExecHierarchyDeleteFixture(t, tx, `
		INSERT INTO control_cabinets (id, building_id, control_cabinet_nr)
		VALUES (?, ?, ?)
	`, ids.cabinet, ids.building, "DEL-"+ids.cabinet.String())
	mustExecHierarchyDeleteFixture(t, tx, `
		INSERT INTO sps_controllers (id, control_cabinet_id, device_name, ga_device)
		VALUES (?, ?, ?, ?)
	`, ids.controller, ids.cabinet, "DEL-"+ids.controller.String(), "DEL-"+ids.controller.String())
	mustExecHierarchyDeleteFixture(t, tx, `
		INSERT INTO system_types (id, number_min, number_max, name)
		VALUES (?, 1, 99, ?)
	`, ids.systemType, "DEL-"+ids.systemType.String())
	mustExecHierarchyDeleteFixture(t, tx, `
		INSERT INTO sps_controller_system_types (id, sps_controller_id, system_type_id)
		VALUES (?, ?, ?)
	`, ids.assignment, ids.controller, ids.systemType)
	mustExecHierarchyDeleteFixture(t, tx, `
		INSERT INTO system_parts (id, short_name, name)
		VALUES (?, ?, ?)
	`, ids.systemPart, "DEL-"+ids.systemPart.String(), "Delete part "+ids.systemPart.String())
	mustExecHierarchyDeleteFixture(t, tx, `
		INSERT INTO apparats (id, short_name, name)
		VALUES (?, ?, ?)
	`, ids.apparat, "DEL-"+ids.apparat.String(), "Delete apparatus "+ids.apparat.String())
	mustExecHierarchyDeleteFixture(t, tx, `
		INSERT INTO field_devices (
			id,
			apparat_nr,
			sps_controller_system_type_id,
			system_part_id,
			apparat_id
		)
		VALUES (?, 1, ?, ?, ?)
	`, ids.fieldDevice, ids.assignment, ids.systemPart, ids.apparat)
	mustExecHierarchyDeleteFixture(t, tx, `
		INSERT INTO specifications (id, field_device_id, specification_brand)
		VALUES (?, ?, 'delete-fixture')
	`, ids.specification, ids.fieldDevice)
	mustExecHierarchyDeleteFixture(t, tx, `
		UPDATE field_devices
		SET specification_id = ?
		WHERE id = ?
	`, ids.specification, ids.fieldDevice)
	mustExecHierarchyDeleteFixture(t, tx, `
		INSERT INTO bacnet_objects (
			id,
			text_fix,
			software_type,
			software_number,
			field_device_id
		)
		VALUES (?, 'exclusive', 'ai', 1, ?)
	`, ids.exclusiveBacnet, ids.fieldDevice)
	mustExecHierarchyDeleteFixture(t, tx, `
		INSERT INTO bacnet_objects (
			id,
			text_fix,
			software_type,
			software_number,
			field_device_id
		)
		VALUES (?, 'shared', 'ai', 2, ?)
	`, ids.sharedBacnet, ids.fieldDevice)
	mustExecHierarchyDeleteFixture(t, tx, `
		INSERT INTO bacnet_objects (
			id,
			text_fix,
			software_type,
			software_number,
			software_reference_id
		)
		VALUES (?, 'external', 'ai', 3, ?)
	`, ids.externalBacnet, ids.exclusiveBacnet)

	// Deliberately seed association-only rows without unrelated project,
	// ObjectData, and AlarmType fixture graphs. PostgreSQL does not revalidate
	// existing rows when replication role returns to origin.
	mustExecHierarchyDeleteFixture(t, tx, `SET LOCAL session_replication_role = 'replica'`)
	mustExecHierarchyDeleteFixture(t, tx, `
		INSERT INTO project_control_cabinets (id, project_id, control_cabinet_id)
		VALUES (?, ?, ?)
	`, uuid.New(), ids.project, ids.cabinet)
	mustExecHierarchyDeleteFixture(t, tx, `
		INSERT INTO project_sps_controllers (id, project_id, sps_controller_id)
		VALUES (?, ?, ?)
	`, uuid.New(), ids.project, ids.controller)
	mustExecHierarchyDeleteFixture(t, tx, `
		INSERT INTO project_field_devices (id, project_id, field_device_id)
		VALUES (?, ?, ?)
	`, uuid.New(), ids.project, ids.fieldDevice)
	mustExecHierarchyDeleteFixture(t, tx, `
		INSERT INTO object_data_bacnet_objects (
			object_data_id,
			bacnet_object_id,
			software_type_normalized,
			software_number
		)
		VALUES (?, ?, 'ai', 2)
	`, ids.objectData, ids.sharedBacnet)
	mustExecHierarchyDeleteFixture(t, tx, `
		INSERT INTO bacnet_object_alarm_values (
			id,
			bacnet_object_id,
			alarm_type_field_id,
			source
		)
		VALUES (?, ?, ?, 'user')
	`, ids.alarmValue, ids.exclusiveBacnet, ids.alarmTypeField)
	mustExecHierarchyDeleteFixture(t, tx, `SET LOCAL session_replication_role = 'origin'`)

	if err := cleaner.deleteControlCabinet(ctx, ids.cabinet); err != nil {
		t.Fatalf("delete hierarchy: %v", err)
	}

	assertHierarchyDeleteCount(t, tx, "control_cabinets", "id", ids.cabinet, 0)
	assertHierarchyDeleteCount(t, tx, "sps_controllers", "id", ids.controller, 0)
	assertHierarchyDeleteCount(t, tx, "sps_controller_system_types", "id", ids.assignment, 0)
	assertHierarchyDeleteCount(t, tx, "field_devices", "id", ids.fieldDevice, 0)
	assertHierarchyDeleteCount(t, tx, "specifications", "id", ids.specification, 0)
	assertHierarchyDeleteCount(t, tx, "bacnet_objects", "id", ids.exclusiveBacnet, 0)
	assertHierarchyDeleteCount(t, tx, "bacnet_object_alarm_values", "id", ids.alarmValue, 0)
	assertHierarchyDeleteCount(t, tx, "project_control_cabinets", "project_id", ids.project, 0)
	assertHierarchyDeleteCount(t, tx, "project_sps_controllers", "project_id", ids.project, 0)
	assertHierarchyDeleteCount(t, tx, "project_field_devices", "project_id", ids.project, 0)

	assertHierarchyDeleteCount(t, tx, "bacnet_objects", "id", ids.sharedBacnet, 1)
	assertHierarchyDeleteCount(t, tx, "object_data_bacnet_objects", "bacnet_object_id", ids.sharedBacnet, 1)
	assertHierarchyDeleteNullUUID(t, tx, "bacnet_objects", "field_device_id", ids.sharedBacnet)
	assertHierarchyDeleteCount(t, tx, "bacnet_objects", "id", ids.externalBacnet, 1)
	assertHierarchyDeleteNullUUID(t, tx, "bacnet_objects", "software_reference_id", ids.externalBacnet)

	for _, id := range []uuid.UUID{
		ids.cabinet,
		ids.controller,
		ids.assignment,
		ids.fieldDevice,
		ids.specification,
		ids.exclusiveBacnet,
		ids.alarmValue,
	} {
		var count int64
		if err := tx.Raw(`
			SELECT COUNT(*)
			FROM change_events
			WHERE entity_id = ? AND action = 'delete'
		`, id).Scan(&count).Error; err != nil {
			t.Fatalf("count delete history for %s: %v", id, err)
		}
		if count != 1 {
			t.Fatalf("delete history for %s: got %d, want 1", id, count)
		}
	}
}

type hierarchyDeleteFixtureIDs struct {
	building        uuid.UUID
	cabinet         uuid.UUID
	controller      uuid.UUID
	systemType      uuid.UUID
	assignment      uuid.UUID
	systemPart      uuid.UUID
	apparat         uuid.UUID
	fieldDevice     uuid.UUID
	specification   uuid.UUID
	exclusiveBacnet uuid.UUID
	sharedBacnet    uuid.UUID
	externalBacnet  uuid.UUID
	objectData      uuid.UUID
	alarmValue      uuid.UUID
	alarmTypeField  uuid.UUID
	project         uuid.UUID
}

func newHierarchyDeleteFixtureIDs() hierarchyDeleteFixtureIDs {
	return hierarchyDeleteFixtureIDs{
		building:        uuid.New(),
		cabinet:         uuid.New(),
		controller:      uuid.New(),
		systemType:      uuid.New(),
		assignment:      uuid.New(),
		systemPart:      uuid.New(),
		apparat:         uuid.New(),
		fieldDevice:     uuid.New(),
		specification:   uuid.New(),
		exclusiveBacnet: uuid.New(),
		sharedBacnet:    uuid.New(),
		externalBacnet:  uuid.New(),
		objectData:      uuid.New(),
		alarmValue:      uuid.New(),
		alarmTypeField:  uuid.New(),
		project:         uuid.New(),
	}
}

func mustExecHierarchyDeleteFixture(
	t *testing.T,
	tx *gorm.DB,
	query string,
	args ...any,
) {
	t.Helper()
	if err := tx.Exec(query, args...).Error; err != nil {
		t.Fatalf("seed hierarchy delete fixture: %v", err)
	}
}

func assertHierarchyDeleteCount(
	t *testing.T,
	tx *gorm.DB,
	table string,
	column string,
	id uuid.UUID,
	want int64,
) {
	t.Helper()
	var count int64
	query := "SELECT COUNT(*) FROM " + table + " WHERE " + column + " = ?"
	if err := tx.Raw(query, id).Scan(&count).Error; err != nil {
		t.Fatalf("count %s.%s=%s: %v", table, column, id, err)
	}
	if count != want {
		t.Fatalf("count %s.%s=%s: got %d, want %d", table, column, id, count, want)
	}
}

func assertHierarchyDeleteNullUUID(
	t *testing.T,
	tx *gorm.DB,
	table string,
	column string,
	id uuid.UUID,
) {
	t.Helper()
	var count int64
	query := "SELECT COUNT(*) FROM " + table + " WHERE id = ? AND " + column + " IS NULL"
	if err := tx.Raw(query, id).Scan(&count).Error; err != nil {
		t.Fatalf("count null %s.%s for %s: %v", table, column, id, err)
	}
	if count != 1 {
		t.Fatalf("%s.%s for %s is not null", table, column, id)
	}
}
