package facilitysql

import (
	"sync"

	"gorm.io/gorm"
)

const lifecycleTable = "facility_aggregate_lifecycle"

var lifecycleAvailability sync.Map

func activeControlCabinets(query *gorm.DB) *gorm.DB {
	return activeRoots(query, "control_cabinet", "control_cabinets.id")
}

func activeObjectData(query *gorm.DB) *gorm.DB {
	return activeRoots(query, "object_data", "object_data.id")
}

func activeSPSControllers(query *gorm.DB) *gorm.DB {
	if !hasLifecycleTable(query) {
		return query
	}
	return query.Where(`NOT EXISTS (
		SELECT 1 FROM facility_aggregate_lifecycle lifecycle
		WHERE (lifecycle.kind = 'sps_controller' AND lifecycle.resource_id = sps_controllers.id)
		   OR (lifecycle.kind = 'control_cabinet' AND lifecycle.resource_id = sps_controllers.control_cabinet_id)
	)`)
}

func activeSPSControllerSystemTypes(query *gorm.DB) *gorm.DB {
	if !hasLifecycleTable(query) {
		return query
	}
	return query.Where(`NOT EXISTS (
		SELECT 1 FROM facility_aggregate_lifecycle lifecycle
		JOIN sps_controllers lifecycle_controller ON lifecycle_controller.id = sps_controller_system_types.sps_controller_id
		WHERE (lifecycle.kind = 'sps_controller_system_type' AND lifecycle.resource_id = sps_controller_system_types.id)
		   OR (lifecycle.kind = 'sps_controller' AND lifecycle.resource_id = lifecycle_controller.id)
		   OR (lifecycle.kind = 'control_cabinet' AND lifecycle.resource_id = lifecycle_controller.control_cabinet_id)
	)`)
}

func activeFieldDevices(query *gorm.DB) *gorm.DB {
	if !hasLifecycleTable(query) {
		return query
	}
	return query.Where(`NOT EXISTS (
		SELECT 1 FROM facility_aggregate_lifecycle lifecycle
		JOIN sps_controller_system_types lifecycle_system_type ON lifecycle_system_type.id = field_devices.sps_controller_system_type_id
		JOIN sps_controllers lifecycle_controller ON lifecycle_controller.id = lifecycle_system_type.sps_controller_id
		WHERE (lifecycle.kind = 'field_device' AND lifecycle.resource_id = field_devices.id)
		   OR (lifecycle.kind = 'sps_controller_system_type' AND lifecycle.resource_id = lifecycle_system_type.id)
		   OR (lifecycle.kind = 'sps_controller' AND lifecycle.resource_id = lifecycle_controller.id)
		   OR (lifecycle.kind = 'control_cabinet' AND lifecycle.resource_id = lifecycle_controller.control_cabinet_id)
	)`)
}

func activeRoots(query *gorm.DB, kind, idColumn string) *gorm.DB {
	if !hasLifecycleTable(query) {
		return query
	}
	return query.Where(
		"NOT EXISTS (SELECT 1 FROM facility_aggregate_lifecycle lifecycle WHERE lifecycle.kind = ? AND lifecycle.resource_id = "+idColumn+")",
		kind,
	)
}

func hasLifecycleTable(db *gorm.DB) bool {
	if db == nil {
		return false
	}
	if available, ok := lifecycleAvailability.Load(db.Config); ok {
		return available.(bool)
	}
	available := db.Migrator().HasTable(lifecycleTable)
	lifecycleAvailability.Store(db.Config, available)
	return available
}
