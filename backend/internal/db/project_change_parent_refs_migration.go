package db

import "gorm.io/gorm"

func removeProjectChangeParentRefs(database *gorm.DB) error {
	const table = "project_changes"
	if !database.Migrator().HasTable(table) || !database.Migrator().HasColumn(table, "parent_refs") {
		return nil
	}
	return database.Exec("ALTER TABLE project_changes DROP COLUMN parent_refs").Error
}
