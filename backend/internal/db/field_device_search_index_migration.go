package db

import "gorm.io/gorm"

func migrateFieldDeviceCombinedSearchIndex(database *gorm.DB) error {
	if database.Dialector == nil || database.Dialector.Name() != "postgres" {
		return nil
	}
	statement := `CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_field_devices_search_trgm ON field_devices USING gin (
LOWER(COALESCE(bmk,'') || CHR(1) || COALESCE(description,'')) gin_trgm_ops
)`
	return database.Exec(statement).Error
}
