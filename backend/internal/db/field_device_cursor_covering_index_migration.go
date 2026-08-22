package db

import "gorm.io/gorm"

func migrateFieldDeviceCursorCoveringIndexes(database *gorm.DB) error {
	if database.Dialector == nil || database.Dialector.Name() != "postgres" {
		return nil
	}
	for _, statement := range fieldDeviceCursorCoveringIndexes() {
		if err := database.Exec(statement).Error; err != nil {
			return err
		}
	}
	return nil
}

func fieldDeviceCursorCoveringIndexes() []string {
	indexes := []string{
		`CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_fd_apparat_nr_cursor_asc ON field_devices ((CASE WHEN apparat_nr IS NULL THEN 1 ELSE 0 END),apparat_nr,id)`,
		`CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_fd_apparat_nr_cursor_desc ON field_devices ((CASE WHEN apparat_nr IS NULL THEN 1 ELSE 0 END),apparat_nr DESC,id DESC)`,
		`CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_scts_number_document_cursor_desc ON sps_controller_system_types ((CASE WHEN number IS NULL THEN 1 ELSE 0 END),number DESC,(CASE WHEN document_name IS NULL THEN 1 ELSE 0 END),document_name DESC,id DESC)`,
		`CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_specs_supplier_cursor_desc ON specifications ((CASE WHEN specification_supplier IS NULL THEN 1 ELSE 0 END),specification_supplier DESC,field_device_id DESC)`,
	}
	for _, column := range specificationCursorColumns() {
		indexes = append(indexes, specificationCursorIndex(column, "ASC"), specificationCursorIndex(column, "DESC"))
	}
	return indexes
}

func specificationCursorColumns() []string {
	return []string{
		"specification_brand", "specification_type", "additional_info_motor_valve", "additional_info_size",
		"additional_information_installation_location", "electrical_connection_ph", "electrical_connection_acdc",
		"electrical_connection_amperage", "electrical_connection_power", "electrical_connection_rotation",
	}
}

func specificationCursorIndex(column, direction string) string {
	suffix := "asc"
	if direction == "DESC" {
		suffix = "desc"
	}
	return "CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_specs_" + compactIndexColumn(column) + "_cursor_" + suffix +
		" ON specifications ((CASE WHEN " + column + " IS NULL THEN 1 ELSE 0 END)," + column + " " + direction + ",field_device_id " + direction + ")"
}

func compactIndexColumn(column string) string {
	names := map[string]string{
		"specification_brand": "brand", "specification_type": "type", "additional_info_motor_valve": "motor_valve",
		"additional_info_size": "size", "additional_information_installation_location": "install_loc",
		"electrical_connection_ph": "ph", "electrical_connection_acdc": "acdc",
		"electrical_connection_amperage": "amperage", "electrical_connection_power": "power",
		"electrical_connection_rotation": "rotation",
	}
	return names[column]
}
