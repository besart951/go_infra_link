package db

import (
	"context"
	"fmt"
	"strings"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/besart951/go_infra_link/backend/internal/config"
	"github.com/besart951/go_infra_link/backend/internal/domain/auth"
	"github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/besart951/go_infra_link/backend/internal/domain/notification"
	"github.com/besart951/go_infra_link/backend/internal/domain/project"
	"github.com/besart951/go_infra_link/backend/internal/domain/team"
	"github.com/besart951/go_infra_link/backend/internal/domain/user"
)

// Connect establishes a database connection and verifies reachability.
func Connect(cfg config.Config) (*gorm.DB, error) {
	var dialector gorm.Dialector

	switch cfg.DBType {
	case "postgres":
		dialector = postgres.Open(cfg.DBDsn)

	default:
		return nil, fmt.Errorf("unsupported database type: %s (supported: postgres)", cfg.DBType)
	}

	db, err := gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	if cfg.DBMaxOpenConns > 0 {
		sqlDB.SetMaxOpenConns(cfg.DBMaxOpenConns)
	}
	if cfg.DBMaxIdleConns > 0 {
		sqlDB.SetMaxIdleConns(cfg.DBMaxIdleConns)
	}
	if cfg.DBConnMaxLifetime > 0 {
		sqlDB.SetConnMaxLifetime(cfg.DBConnMaxLifetime)
	}

	if cfg.DBConnectTimeout > 0 {
		ctx, cancel := context.WithTimeout(context.Background(), cfg.DBConnectTimeout)
		defer cancel()
		if err := sqlDB.PingContext(ctx); err != nil {
			return nil, fmt.Errorf("failed to ping database: %w", err)
		}
	} else {
		if err := sqlDB.Ping(); err != nil {
			return nil, fmt.Errorf("failed to ping database: %w", err)
		}
	}

	return db, nil
}

// Bootstrap prepares the database schema explicitly.
// Use this from a dedicated bootstrap or migration command, not normal app startup.
func Bootstrap(cfg config.Config) error {
	db, err := Connect(cfg)
	if err != nil {
		return err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get database instance: %w", err)
	}
	defer func() {
		_ = sqlDB.Close()
	}()

	return ApplyMigrations(db)
}

// autoMigrateCurrentSchema creates the current schema baseline.
// Keep this limited to schema creation only; all repairs and follow-up changes belong in versioned migrations.
func autoMigrateCurrentSchema(db *gorm.DB) error {
	if err := db.AutoMigrate(
		// User domain
		&user.User{},
		&user.BusinessDetails{},
		&user.Permission{},
		&user.RolePermission{},

		// Auth domain
		&auth.RefreshToken{},
		&notification.SMTPSettings{},

		// Team domain
		&team.Team{},
		&team.TeamMember{},
		&user.UserTeam{},

		// Project domain
		&project.Phase{},
		&project.Project{},
		&project.ProjectFieldDevice{},
		&project.ProjectControlCabinet{},
		&project.ProjectSPSController{},

		// Facility domain
		&facility.Building{},
		&facility.ControlCabinet{},
		&facility.SPSController{},
		&facility.SystemType{},
		&facility.SPSControllerSystemType{},
		&facility.SystemPart{},
		&facility.Apparat{},
		&facility.Specification{},
		&facility.FieldDevice{},
		&facility.StateText{},
		&facility.NotificationClass{},
		&facility.AlarmDefinition{},
		&facility.BacnetObject{},
		&facility.ObjectData{},
		&facility.Unit{},
		&facility.AlarmField{},
		&facility.AlarmType{},
		&facility.AlarmTypeField{},
		&facility.AlarmDefinitionFieldOverride{},
		&facility.BacnetObjectAlarmValue{},
	); err != nil {
		return err
	}

	return nil
}

func dropLegacyPhaseProjectID(db *gorm.DB) error {
	if !db.Migrator().HasColumn(&project.Phase{}, "project_id") {
		return nil
	}
	return db.Migrator().DropColumn(&project.Phase{}, "project_id")
}

type foreignKeySpec struct {
	childTable     string
	childColumn    string
	parentTable    string
	parentColumn   string
	constraintName string
	onDelete       string
}

func cleanupFacilityDeleteOrphans(db *gorm.DB) error {
	queries := []string{
		`UPDATE field_devices
		SET specification_id = NULL
		WHERE specification_id IS NOT NULL
		  AND NOT EXISTS (
			SELECT 1
			FROM specifications
			WHERE specifications.id = field_devices.specification_id
		  )`,
		`DELETE FROM specifications
		WHERE field_device_id IS NOT NULL
		  AND NOT EXISTS (
			SELECT 1
			FROM field_devices
			WHERE field_devices.id = specifications.field_device_id
		  )`,
	}

	return db.Transaction(func(tx *gorm.DB) error {
		for _, query := range queries {
			if err := tx.Exec(query).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func ensureFacilityDeleteCascades(db *gorm.DB) error {
	specs := []foreignKeySpec{
		{
			childTable:     "control_cabinets",
			childColumn:    "building_id",
			parentTable:    "buildings",
			parentColumn:   "id",
			constraintName: "fk_control_cabinets_building",
			onDelete:       "CASCADE",
		},
		{
			childTable:     "sps_controllers",
			childColumn:    "control_cabinet_id",
			parentTable:    "control_cabinets",
			parentColumn:   "id",
			constraintName: "fk_sps_controllers_control_cabinet",
			onDelete:       "CASCADE",
		},
		{
			childTable:     "sps_controller_system_types",
			childColumn:    "sps_controller_id",
			parentTable:    "sps_controllers",
			parentColumn:   "id",
			constraintName: "fk_sps_controller_system_types_sps_controller",
			onDelete:       "CASCADE",
		},
		{
			childTable:     "field_devices",
			childColumn:    "sps_controller_system_type_id",
			parentTable:    "sps_controller_system_types",
			parentColumn:   "id",
			constraintName: "fk_field_devices_sps_controller_system_type",
			onDelete:       "CASCADE",
		},
		{
			childTable:     "specifications",
			childColumn:    "field_device_id",
			parentTable:    "field_devices",
			parentColumn:   "id",
			constraintName: "fk_specifications_field_device",
			onDelete:       "CASCADE",
		},
		{
			childTable:     "field_devices",
			childColumn:    "specification_id",
			parentTable:    "specifications",
			parentColumn:   "id",
			constraintName: "fk_field_devices_specification",
			onDelete:       "SET NULL",
		},
		{
			childTable:     "bacnet_objects",
			childColumn:    "field_device_id",
			parentTable:    "field_devices",
			parentColumn:   "id",
			constraintName: "fk_bacnet_objects_field_device",
			onDelete:       "CASCADE",
		},
		{
			childTable:     "bacnet_object_alarm_values",
			childColumn:    "bacnet_object_id",
			parentTable:    "bacnet_objects",
			parentColumn:   "id",
			constraintName: "fk_bacnet_object_alarm_values_bacnet_object",
			onDelete:       "CASCADE",
		},
		{
			childTable:     "project_control_cabinets",
			childColumn:    "control_cabinet_id",
			parentTable:    "control_cabinets",
			parentColumn:   "id",
			constraintName: "fk_project_control_cabinets_control_cabinet",
			onDelete:       "CASCADE",
		},
		{
			childTable:     "project_sps_controllers",
			childColumn:    "sps_controller_id",
			parentTable:    "sps_controllers",
			parentColumn:   "id",
			constraintName: "fk_project_sps_controllers_sps_controller",
			onDelete:       "CASCADE",
		},
		{
			childTable:     "project_field_devices",
			childColumn:    "field_device_id",
			parentTable:    "field_devices",
			parentColumn:   "id",
			constraintName: "fk_project_field_devices_field_device",
			onDelete:       "CASCADE",
		},
		{
			childTable:     "object_data_bacnet_objects",
			childColumn:    "object_data_id",
			parentTable:    "object_data",
			parentColumn:   "id",
			constraintName: "fk_object_data_bacnet_objects_object_data",
			onDelete:       "CASCADE",
		},
		{
			childTable:     "object_data_bacnet_objects",
			childColumn:    "bacnet_object_id",
			parentTable:    "bacnet_objects",
			parentColumn:   "id",
			constraintName: "fk_object_data_bacnet_objects_bacnet_object",
			onDelete:       "CASCADE",
		},
	}

	for _, spec := range specs {
		if err := ensureForeignKey(db, spec); err != nil {
			return fmt.Errorf("ensure %s: %w", spec.constraintName, err)
		}
	}

	return nil
}

func ensureForeignKey(db *gorm.DB, spec foreignKeySpec) error {
	switch db.Dialector.Name() {
	case "postgres":
		type existingForeignKey struct {
			ConstraintName string
			Definition     string
		}

		var existing []existingForeignKey
		if err := db.Raw(
			`SELECT
				tc.constraint_name AS constraint_name,
				pg_get_constraintdef(c.oid) AS definition
			FROM information_schema.table_constraints tc
			JOIN information_schema.key_column_usage kcu
			  ON tc.constraint_name = kcu.constraint_name
			 AND tc.table_schema = kcu.table_schema
			JOIN information_schema.constraint_column_usage ccu
			  ON ccu.constraint_name = tc.constraint_name
			 AND ccu.table_schema = tc.table_schema
			JOIN pg_constraint c
			  ON c.conname = tc.constraint_name
			JOIN pg_namespace n
			  ON n.oid = c.connamespace
			 AND n.nspname = tc.table_schema
			WHERE tc.constraint_type = 'FOREIGN KEY'
			  AND tc.table_schema = current_schema()
			  AND tc.table_name = ?
			  AND kcu.column_name = ?
			  AND ccu.table_name = ?
			  AND ccu.column_name = ?`,
			spec.childTable,
			spec.childColumn,
			spec.parentTable,
			spec.parentColumn,
		).Scan(&existing).Error; err != nil {
			return err
		}

		for _, fk := range existing {
			definition := normalizeRule(fk.Definition)
			if strings.Contains(definition, "ON UPDATE CASCADE") &&
				strings.Contains(definition, "ON DELETE "+spec.onDelete) {
				return nil
			}
		}

		for _, fk := range existing {
			if err := db.Exec(
				"ALTER TABLE " + spec.childTable + " DROP CONSTRAINT " + fk.ConstraintName,
			).Error; err != nil {
				return err
			}
		}

		return db.Exec(
			"ALTER TABLE " + spec.childTable +
				" ADD CONSTRAINT " + spec.constraintName +
				" FOREIGN KEY (" + spec.childColumn + ") REFERENCES " + spec.parentTable + "(" + spec.parentColumn + ")" +
				" ON UPDATE CASCADE ON DELETE " + spec.onDelete,
		).Error
	}

	return nil
}

func normalizeRule(value string) string {
	return strings.ToUpper(strings.TrimSpace(value))
}

func ensureObjectDataApparatsCascade(db *gorm.DB) error {
	constraintName := "fk_object_data_apparats_object_data"

	switch db.Dialector.Name() {
	case "postgres":
		var exists bool
		if err := db.Raw(
			"SELECT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = ?)",
			constraintName,
		).Scan(&exists).Error; err != nil {
			return err
		}

		if exists {
			var isCascade bool
			if err := db.Raw(
				"SELECT COALESCE(position('ON DELETE CASCADE' IN pg_get_constraintdef(oid)) > 0, false) "+
					"FROM pg_constraint WHERE conname = ? LIMIT 1",
				constraintName,
			).Scan(&isCascade).Error; err != nil {
				return err
			}
			if isCascade {
				return nil
			}
		}

		if exists {
			if err := db.Exec(
				"ALTER TABLE object_data_apparats DROP CONSTRAINT " + constraintName,
			).Error; err != nil {
				return err
			}
		}

		return db.Exec(
			"ALTER TABLE object_data_apparats " +
				"ADD CONSTRAINT " + constraintName + " " +
				"FOREIGN KEY (object_data_id) REFERENCES object_data(id) " +
				"ON UPDATE CASCADE ON DELETE CASCADE",
		).Error
	}

	return nil
}

func ensureBacnetObjectTextFixIndexNonUnique(db *gorm.DB) error {
	const indexName = "idx_field_device_textfix"

	switch db.Dialector.Name() {
	case "postgres":
		if err := db.Exec("DROP INDEX IF EXISTS " + indexName).Error; err != nil {
			return err
		}
		return db.Exec(
			"CREATE INDEX IF NOT EXISTS " + indexName + " ON bacnet_objects (field_device_id, text_fix)",
		).Error
	}

	return nil
}
