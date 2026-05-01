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
	{
		version:     "202604300001",
		description: "phase_permission_rules_and_remove_project_edit_permissions",
		apply:       migratePhasePermissions,
	},
	{
		version:     "202604300002",
		description: "user_notification_preferences_and_system_notifications",
		apply:       migrateUserNotificationPreferences,
	},
	{
		version:     "202604300003",
		description: "notification_email_verification",
		apply:       migrateUserNotificationPreferences,
	},
	{
		version:     "202604300004",
		description: "notification_outbox_and_rules",
		apply:       migrateNotificationDispatch,
	},
	{
		version:     "202604300005",
		description: "enable_pg_trgm_search_indexes",
		apply:       migratePGTrgmSearch,
	},
	{
		version:     "202605010001",
		description: "performance_indexes_for_large_project_data",
		apply:       migratePerformanceIndexes,
	},
	{
		version:     "202605010002",
		description: "control_cabinet_search_and_list_indexes",
		apply:       migrateControlCabinetPerformance,
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
