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
	version             string
	description         string
	blueGreenCompatible bool
	apply               func(db *gorm.DB) error
}

var migrations = []migration{
	{
		version:             "202604030001",
		description:         "baseline_schema",
		blueGreenCompatible: true,
		apply:               autoMigrateCurrentSchema,
	},
	{
		version:             "202604030007",
		description:         "migrate_project_permissions",
		blueGreenCompatible: true,
		apply:               migrateProjectPermissions,
	},
	{
		version:             "202604030008",
		description:         "ensure_notification_smtp_manage_permission",
		blueGreenCompatible: true,
		apply:               ensureNotificationSMTPManagePermission,
	},
	{
		version:             "202604030009",
		description:         "expand_project_subresource_permissions",
		blueGreenCompatible: true,
		apply:               expandProjectSubresourcePermissions,
	},
	{
		version:             "202604300001",
		description:         "phase_permission_rules_and_remove_project_edit_permissions",
		blueGreenCompatible: true,
		apply:               migratePhasePermissions,
	},
	{
		version:             "202604300002",
		description:         "user_notification_preferences_and_system_notifications",
		blueGreenCompatible: true,
		apply:               migrateUserNotificationPreferences,
	},
	{
		version:             "202604300003",
		description:         "notification_email_verification",
		blueGreenCompatible: true,
		apply:               migrateUserNotificationPreferences,
	},
	{
		version:             "202604300004",
		description:         "notification_outbox_and_rules",
		blueGreenCompatible: true,
		apply:               migrateNotificationDispatch,
	},
	{
		version:             "202604300005",
		description:         "enable_pg_trgm_search_indexes",
		blueGreenCompatible: true,
		apply:               migratePGTrgmSearch,
	},
	{
		version:             "202605010001",
		description:         "performance_indexes_for_large_project_data",
		blueGreenCompatible: true,
		apply:               migratePerformanceIndexes,
	},
	{
		version:             "202605010002",
		description:         "control_cabinet_search_and_list_indexes",
		blueGreenCompatible: true,
		apply:               migrateControlCabinetPerformance,
	},
	{
		version:             "202605010003",
		description:         "system_notification_importance",
		blueGreenCompatible: true,
		apply:               migrateSystemNotificationImportance,
	},
	{
		version:             "202605020001",
		description:         "facility_project_change_history",
		blueGreenCompatible: true,
		apply:               migrateHistory,
	},
	{
		version:             "202605030001",
		description:         "timeline_permissions",
		blueGreenCompatible: true,
		apply:               ensureTimelinePermissions,
	},
	{
		version:             "202605070001",
		description:         "realtime_events_for_cross_instance_fanout",
		blueGreenCompatible: true,
		apply:               migrateRealtimeEvents,
	},
	{
		version:             "202605080001",
		description:         "unify_permission_model",
		blueGreenCompatible: true,
		apply:               unifyPermissionModel,
	},
	{
		version:             "202605080002",
		description:         "user_registration_invitations",
		blueGreenCompatible: true,
		apply:               migrateUserRegistrations,
	},
	{
		version:             "202605080003",
		description:         "project_phases_for_existing_databases",
		blueGreenCompatible: true,
		apply:               migrateProjectPhases,
	},
	{
		version:             "202605110001",
		description:         "user_soft_delete_and_anonymization",
		blueGreenCompatible: true,
		apply:               migrateUserSoftDelete,
	},
	{
		version:             "202605110002",
		description:         "ensure_user_read_deleted_permission",
		blueGreenCompatible: true,
		apply:               ensureDeletedUserReadPermission,
	},
}

type MigrationOptions struct {
	RequireBlueGreenCompatible bool
}

func ApplyMigrations(db *gorm.DB, options ...MigrationOptions) error {
	opts := MigrationOptions{}
	if len(options) > 0 {
		opts = options[0]
	}

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

		if opts.RequireBlueGreenCompatible && !migration.blueGreenCompatible {
			return fmt.Errorf(
				"migration %s (%s) is not marked blue-green compatible; use a maintenance release plan",
				migration.version,
				migration.description,
			)
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
