package db

import "gorm.io/gorm"

func migrateControlCabinetPerformance(db *gorm.DB) error {
	if db.Dialector == nil || db.Dialector.Name() != "postgres" {
		return nil
	}

	statements := []string{
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_control_cabinets_created_id_desc ON control_cabinets (created_at DESC, id DESC)",
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_control_cabinets_nr_lower_pattern ON control_cabinets (LOWER(control_cabinet_nr) text_pattern_ops)",
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_field_devices_created_id_desc ON field_devices (created_at DESC, id DESC)",
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_sps_controllers_created_id_desc ON sps_controllers (created_at DESC, id DESC)",
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_sps_controller_system_types_created_id_desc ON sps_controller_system_types (created_at DESC, id DESC)",
	}
	for _, statement := range statements {
		if err := db.Exec(statement).Error; err != nil {
			return err
		}
	}
	return nil
}
