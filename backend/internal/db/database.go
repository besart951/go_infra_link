package db

import (
	"context"
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/besart951/go_infra_link/backend/internal/config"
)

// Connect establishes a database connection and verifies reachability.
func Connect(cfg config.DBConfig) (*gorm.DB, error) {
	var dialector gorm.Dialector

	switch cfg.Type {
	case "postgres":
		dialector = postgres.Open(cfg.Dsn)

	default:
		return nil, fmt.Errorf("unsupported database type: %s (supported: postgres)", cfg.Type)
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

	if cfg.MaxOpenConns > 0 {
		sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	}
	if cfg.MaxIdleConns > 0 {
		sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	}
	if cfg.ConnMaxLifetime > 0 {
		sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	}

	if cfg.ConnectTimeout > 0 {
		ctx, cancel := context.WithTimeout(context.Background(), cfg.ConnectTimeout)
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
func Bootstrap(cfg config.DBConfig) error {
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
