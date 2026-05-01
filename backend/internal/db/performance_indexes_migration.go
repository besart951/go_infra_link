package db

import (
	"fmt"

	"gorm.io/gorm"
)

type postgresIndex struct {
	Name       string
	Table      string
	Definition string
}

func migratePerformanceIndexes(db *gorm.DB) error {
	if db.Dialector == nil || db.Dialector.Name() != "postgres" {
		return nil
	}

	indexes := []postgresIndex{
		{Name: "idx_projects_created_id_desc", Table: "projects", Definition: "(created_at DESC, id DESC)"},
		{Name: "idx_project_users_user_project", Table: "project_users", Definition: "(user_id, project_id)"},

		{Name: "idx_control_cabinets_building_created_id_desc", Table: "control_cabinets", Definition: "(building_id, created_at DESC, id DESC)"},
		{Name: "idx_sps_controllers_cabinet_created_id_desc", Table: "sps_controllers", Definition: "(control_cabinet_id, created_at DESC, id DESC)"},
		{Name: "idx_sps_controller_system_types_sps_created_id_desc", Table: "sps_controller_system_types", Definition: "(sps_controller_id, created_at DESC, id DESC)"},
		{Name: "idx_field_devices_sct_created_id_desc", Table: "field_devices", Definition: "(sps_controller_system_type_id, created_at DESC, id DESC)"},
		{Name: "idx_field_devices_conflict_scope", Table: "field_devices", Definition: "(sps_controller_system_type_id, system_part_id, apparat_id, apparat_nr)"},

		{Name: "idx_project_control_cabinets_project_created_id_desc", Table: "project_control_cabinets", Definition: "(project_id, created_at DESC, id DESC)"},
		{Name: "idx_project_control_cabinets_control_cabinet", Table: "project_control_cabinets", Definition: "(control_cabinet_id)"},
		{Name: "idx_project_sps_controllers_project_created_id_desc", Table: "project_sps_controllers", Definition: "(project_id, created_at DESC, id DESC)"},
		{Name: "idx_project_sps_controllers_sps_controller", Table: "project_sps_controllers", Definition: "(sps_controller_id)"},
		{Name: "idx_project_field_devices_project_created_id_desc", Table: "project_field_devices", Definition: "(project_id, created_at DESC, id DESC)"},
		{Name: "idx_project_field_devices_field_device", Table: "project_field_devices", Definition: "(field_device_id)"},
	}

	for _, index := range indexes {
		if err := db.Exec(index.createConcurrentlyStatement()).Error; err != nil {
			return err
		}
	}
	return nil
}

func (i postgresIndex) createConcurrentlyStatement() string {
	return fmt.Sprintf("CREATE INDEX CONCURRENTLY IF NOT EXISTS %s ON %s %s", i.Name, i.Table, i.Definition)
}
