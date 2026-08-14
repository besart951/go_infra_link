package db

import (
	infrarealtime "github.com/besart951/go_infra_link/backend/internal/infrastructure/realtime"
	projectchange "github.com/besart951/go_infra_link/backend/internal/repository/projectchange"
	"gorm.io/gorm"
)

func migrateProjectChanges(db *gorm.DB) error {
	if err := projectchange.AutoMigrate(db); err != nil {
		return err
	}
	if err := infrarealtime.AutoMigrateProjectCollaboration(db); err != nil {
		return err
	}

	// Existing installations predate aggregate versioning. AddColumn is used
	// instead of AutoMigrate here so relationship graphs are not traversed.
	for _, model := range currentSchemaModels() {
		statement := &gorm.Statement{DB: db}
		if err := statement.Parse(model); err != nil {
			return err
		}
		if statement.Schema.LookUpField("Version") == nil {
			continue
		}
		if db.Migrator().HasColumn(model, "Version") {
			continue
		}
		if err := db.Migrator().AddColumn(model, "Version"); err != nil {
			return err
		}
	}
	return nil
}
