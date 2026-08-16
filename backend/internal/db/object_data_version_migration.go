package db

import (
	"strings"

	"gorm.io/gorm"
)

// migrateObjectDataOptimisticLocking adds the Base.Version column to databases
// that were bootstrapped before ObjectData started using optimistic locking.
//
// ObjectData also has a domain-specific Version field mapped to obj_version.
// GORM therefore cannot address the embedded Base.Version field by name for
// this legacy migration, so the column is added explicitly.
func migrateObjectDataOptimisticLocking(db *gorm.DB) error {
	if !db.Migrator().HasTable("object_data") {
		return nil
	}

	if db.Dialector != nil && db.Dialector.Name() == "postgres" {
		return db.Exec(
			"ALTER TABLE object_data ADD COLUMN IF NOT EXISTS version bigint NOT NULL DEFAULT 1",
		).Error
	}

	var columns []struct {
		Name string
	}
	if err := db.Raw("PRAGMA table_info(object_data)").Scan(&columns).Error; err != nil {
		return err
	}
	for _, column := range columns {
		if strings.EqualFold(column.Name, "version") {
			return nil
		}
	}

	return db.Exec("ALTER TABLE object_data ADD COLUMN version bigint NOT NULL DEFAULT 1").Error
}
