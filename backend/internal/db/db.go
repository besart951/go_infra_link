package db

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/glebarez/sqlite"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/besart951/go_infra_link/backend/internal/config"
)

func Open(cfg config.Config) (*gorm.DB, error) {
	switch cfg.DBDriver {
	case "postgres":
		return gorm.Open(postgres.Open(cfg.DBDsn), &gorm.Config{})
	case "sqlite":
		if err := ensureSQLiteDir(cfg.DBDsn); err != nil {
			return nil, err
		}
		return gorm.Open(sqlite.Open(cfg.DBDsn), &gorm.Config{})
	default:
		return nil, fmt.Errorf("unsupported DB driver: %s", cfg.DBDriver)
	}
}

func ensureSQLiteDir(dsn string) error {
	// If the DSN is ':memory:' or a URI, don't try to create dirs.
	if dsn == ":memory:" {
		return nil
	}
	// For plain file paths like ./data/app.db, ensure the parent directory exists.
	dir := filepath.Dir(dsn)
	if dir == "." || dir == "" {
		return nil
	}
	return os.MkdirAll(dir, 0o755)
}
