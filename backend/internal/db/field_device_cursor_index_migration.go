package db

import "gorm.io/gorm"

func migrateFieldDeviceCursorIndexes(database *gorm.DB) error {
	if database.Dialector == nil || database.Dialector.Name() != "postgres" {
		return nil
	}
	statements := []string{
		`CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_fd_apparat_nr_id ON field_devices (apparat_nr,id)`,
		`CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_fd_bmk_cursor_asc ON field_devices ((CASE WHEN bmk IS NULL THEN 1 ELSE 0 END),bmk,id)`,
		`CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_fd_bmk_cursor_desc ON field_devices ((CASE WHEN bmk IS NULL THEN 1 ELSE 0 END),bmk DESC,id DESC)`,
		`CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_fd_description_cursor_asc ON field_devices ((CASE WHEN description IS NULL THEN 1 ELSE 0 END),description,id)`,
		`CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_fd_description_cursor_desc ON field_devices ((CASE WHEN description IS NULL THEN 1 ELSE 0 END),description DESC,id DESC)`,
		`CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_fd_scts_id ON field_devices (sps_controller_system_type_id,id)`,
		`CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_scts_number_document_cursor ON sps_controller_system_types ((CASE WHEN number IS NULL THEN 1 ELSE 0 END),number,(CASE WHEN document_name IS NULL THEN 1 ELSE 0 END),document_name,id)`,
		`CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_specs_supplier_cursor ON specifications ((CASE WHEN specification_supplier IS NULL THEN 1 ELSE 0 END),specification_supplier,field_device_id)`,
		`CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_pfd_project_field_device ON project_field_devices (project_id,field_device_id)`,
	}
	for _, statement := range statements {
		if err := database.Exec(statement).Error; err != nil {
			return err
		}
	}
	return nil
}
