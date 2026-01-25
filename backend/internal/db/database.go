package db

import (
	"fmt"
	"os"
	"path/filepath"

	sqlite "github.com/glebarez/sqlite"
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
	fmt.Printf("DEBUG: DBType=%s DSN=%s\n", cfg.DBType, cfg.DBDsn)
	var dialector gorm.Dialector

	switch cfg.DBType {
	case "sqlite":
		sqliteDsn := resolveSQLiteDSN(cfg.DBDsn)
		if err := ensureSQLiteDir(sqliteDsn); err != nil {
			return nil, fmt.Errorf("ensure sqlite directory: %w", err)
		}
		dialector = sqlite.Open(sqliteDsn)

	case "postgres":
		dialector = postgres.Open(cfg.DBDsn)

	case "mysql", "mariadb":
		// MariaDB is MySQL-compatible and uses the same driver
		dialector = mysql.Open(cfg.DBDsn)

	default:
		return nil, fmt.Errorf("unsupported database type: %s (supported: sqlite, postgres, mysql, mariadb)", cfg.DBType)
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
	return db.AutoMigrate(
		// User domain
		&user.User{},
		&user.BusinessDetails{},

		// Auth domain
		&auth.RefreshToken{},
		&auth.PasswordResetToken{},
		&auth.LoginAttempt{},

		// Team domain
		&team.Team{},
		&team.TeamMember{},

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
	)
}

func resolveSQLiteDSN(dsn string) string {
	if dsn == ":memory:" {
		return dsn
	}
	if filepath.IsAbs(dsn) {
		return dsn
	}
	// If you run from repo root, ./data/app.db doesn't exist, but backend/data/app.db does.
	// If you run from backend/, ./data/app.db exists and will be used.
	if fileExists(dsn) {
		return dsn
	}
	alt := filepath.Join("backend", dsn)
	if fileExists(alt) {
		return alt
	}
	return dsn
}

func fileExists(path string) bool {
	if _, err := os.Stat(path); err != nil {
		return false
	}
	return true
}

// ensureSQLiteDir creates the directory for SQLite database file if needed
func ensureSQLiteDir(dsn string) error {
	if dsn == ":memory:" {
		return nil
	}
	dir := filepath.Dir(dsn)
	if dir == "." || dir == "" {
		return nil
	}
	return os.MkdirAll(dir, 0o755)
}

// GetModels returns a list of all domain models for reference
func GetModels() []interface{} {
	return []interface{}{
		&user.User{},
		&user.BusinessDetails{},
		&auth.RefreshToken{},
		&auth.PasswordResetToken{},
		&auth.LoginAttempt{},
		&team.Team{},
		&team.TeamMember{},
		&project.Phase{},
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
	}
}
