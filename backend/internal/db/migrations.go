package db

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

type schemaMigration struct {
	Version     string    `gorm:"primaryKey;size:32"`
	Description string    `gorm:"size:255;not null"`
	AppliedAt   time.Time `gorm:"not null"`
}

func (schemaMigration) TableName() string {
	return "schema_migrations"
}

type migration struct {
	version     string
	description string
	apply       func(db *gorm.DB) error
}

var migrations = []migration{
	{
		version:     "202604030001",
		description: "baseline_schema",
		apply:       autoMigrateCurrentSchema,
	},
	{
		version:     "202604030002",
		description: "cleanup_facility_delete_orphans",
		apply:       cleanupFacilityDeleteOrphans,
	},
	{
		version:     "202604030003",
		description: "ensure_facility_delete_cascades",
		apply:       ensureFacilityDeleteCascades,
	},
	{
		version:     "202604030004",
		description: "ensure_object_data_apparats_cascade",
		apply:       ensureObjectDataApparatsCascade,
	},
	{
		version:     "202604030005",
		description: "ensure_bacnet_object_textfix_index_non_unique",
		apply:       ensureBacnetObjectTextFixIndexNonUnique,
	},
	{
		version:     "202604030006",
		description: "drop_legacy_phases_project_id",
		apply:       dropLegacyPhaseProjectID,
	},
	{
		version:     "202604030007",
		description: "migrate_project_permissions",
		apply:       migrateProjectPermissions,
	},
	{
		version:     "202604030008",
		description: "ensure_notification_smtp_manage_permission",
		apply:       ensureNotificationSMTPManagePermission,
	},
	{
		version:     "202604030009",
		description: "expand_project_subresource_permissions",
		apply:       expandProjectSubresourcePermissions,
	},
}

func ApplyMigrations(db *gorm.DB) error {
	if err := db.AutoMigrate(&schemaMigration{}); err != nil {
		return fmt.Errorf("migrations table: %w", err)
	}

	applied := make(map[string]struct{}, len(migrations))
	var rows []schemaMigration
	if err := db.Order("version ASC").Find(&rows).Error; err != nil {
		return fmt.Errorf("load applied migrations: %w", err)
	}
	for _, row := range rows {
		applied[row.Version] = struct{}{}
	}

	for _, migration := range migrations {
		if _, ok := applied[migration.version]; ok {
			continue
		}

		if err := migration.apply(db); err != nil {
			return fmt.Errorf("apply migration %s (%s): %w", migration.version, migration.description, err)
		}

		if err := db.Create(&schemaMigration{
			Version:     migration.version,
			Description: migration.description,
			AppliedAt:   time.Now().UTC(),
		}).Error; err != nil {
			return fmt.Errorf("record migration %s: %w", migration.version, err)
		}
	}

	return nil
}
