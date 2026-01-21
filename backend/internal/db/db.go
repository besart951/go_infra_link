package db

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "github.com/glebarez/go-sqlite"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/besart951/go_infra_link/backend/internal/config"
)

func Open(cfg config.Config) (*sql.DB, error) {
	var driverName string
	switch cfg.DBDriver {
	case "postgres":
		driverName = "pgx"
	case "sqlite":
		if err := ensureSQLiteDir(cfg.DBDsn); err != nil {
			return nil, err
		}
		driverName = "sqlite"
	case "mysql":
		driverName = "mysql"
	default:
		return nil, fmt.Errorf("unsupported DB driver: %s", cfg.DBDriver)
	}

	db, err := sql.Open(driverName, cfg.DBDsn)
	if err != nil {
		return nil, err
	}

	// Connection pool settings
	if cfg.DBMaxOpenConns > 0 {
		db.SetMaxOpenConns(cfg.DBMaxOpenConns)
	}
	if cfg.DBMaxIdleConns > 0 {
		db.SetMaxIdleConns(cfg.DBMaxIdleConns)
	}
	if cfg.DBConnMaxLifetime > 0 {
		db.SetConnMaxLifetime(cfg.DBConnMaxLifetime)
	}

	// Ensure basic connectivity
	ctx := context.Background()
	if cfg.DBConnectTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, cfg.DBConnectTimeout)
		defer cancel()
	}
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}

	// SQLite-specific pragmas
	if cfg.DBDriver == "sqlite" {
		ctx2, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if _, err := db.ExecContext(ctx2, "PRAGMA foreign_keys = ON"); err != nil {
			_ = db.Close()
			return nil, err
		}
	}

	return db, nil
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
