package db

import (
	"fmt"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/besart951/go_infra_link/backend/internal/config"
	"github.com/besart951/go_infra_link/backend/internal/domain/auth"
	"github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/besart951/go_infra_link/backend/internal/domain/project"
	"github.com/besart951/go_infra_link/backend/internal/domain/team"
	"github.com/besart951/go_infra_link/backend/internal/domain/user"
)

// Connect establishes a database connection based on configuration
// and runs AutoMigrate for all models
func Connect(cfg config.Config) (*gorm.DB, error) {
	var dialector gorm.Dialector

	switch cfg.DBType {
	case "postgres":
		dialector = postgres.Open(cfg.DBDsn)

	case "mysql", "mariadb":
		// MariaDB is MySQL-compatible and uses the same driver
		dialector = mysql.Open(cfg.DBDsn)

	default:
		return nil, fmt.Errorf("unsupported database type: %s (supported: postgres, mysql, mariadb)", cfg.DBType)
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

	// Run AutoMigrate for all models
	if err := autoMigrate(db); err != nil {
		return nil, fmt.Errorf("auto migration failed: %w", err)
	}

	return db, nil
}

// autoMigrate runs GORM's AutoMigrate for all domain models
func autoMigrate(db *gorm.DB) error {
	if err := db.AutoMigrate(
		// User domain
		&user.User{},
		&user.BusinessDetails{},
		&user.Permission{},
		&user.RolePermission{},

		// Auth domain
		&auth.RefreshToken{},
		&auth.PasswordResetToken{},
		&auth.LoginAttempt{},

		// Team domain
		&team.Team{},
		&team.TeamMember{},
		&user.UserTeam{},

		// Project domain
		&project.Phase{},
		&project.PhasePermission{},
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

	if err := ensureObjectDataApparatsCascade(db); err != nil {
		return err
	}

	// Cleanup legacy phases.project_id column (phases are independent)
	if db.Migrator().HasColumn(&project.Phase{}, "project_id") {
		if err := db.Migrator().DropColumn(&project.Phase{}, "project_id"); err != nil {
			return err
		}
	}

	return nil
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

	case "mysql", "mariadb":
		var count int64
		if err := db.Raw(
			"SELECT COUNT(*) FROM information_schema.TABLE_CONSTRAINTS "+
				"WHERE CONSTRAINT_SCHEMA = DATABASE() "+
				"AND TABLE_NAME = ? AND CONSTRAINT_NAME = ? AND CONSTRAINT_TYPE = 'FOREIGN KEY'",
			"object_data_apparats",
			constraintName,
		).Scan(&count).Error; err != nil {
			return err
		}

		if count > 0 {
			var deleteRule string
			if err := db.Raw(
				"SELECT DELETE_RULE FROM information_schema.REFERENTIAL_CONSTRAINTS "+
					"WHERE CONSTRAINT_SCHEMA = DATABASE() AND CONSTRAINT_NAME = ? LIMIT 1",
				constraintName,
			).Scan(&deleteRule).Error; err != nil {
				return err
			}
			if deleteRule == "CASCADE" {
				return nil
			}
		}

		if count > 0 {
			if err := db.Exec(
				"ALTER TABLE object_data_apparats DROP FOREIGN KEY " + constraintName,
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

// GetModels returns a list of all domain models for reference
func GetModels() []interface{} {
	return []interface{}{
		&user.User{},
		&user.BusinessDetails{},
		&user.UserTeam{},
		&user.Permission{},
		&user.RolePermission{},
		&auth.RefreshToken{},
		&auth.PasswordResetToken{},
		&auth.LoginAttempt{},
		&team.Team{},
		&team.TeamMember{},
		&project.Phase{},
		&project.PhasePermission{},
		&project.Project{},
		&project.ProjectFieldDevice{},
		&project.ProjectControlCabinet{},
		&project.ProjectSPSController{},
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
	}
}
