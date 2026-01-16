package database

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/jackc/pgx/v5/stdlib"
	_ "modernc.org/sqlite"

	"github.com/rs/zerolog"
)

type Config struct {
	Driver          string
	DSN             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnectTimeout  time.Duration
	Logger          zerolog.Logger
}

func NewConfigFromEnv(l zerolog.Logger) Config {
	maxOpen, _ := strconv.Atoi(getEnv("DB_MAX_OPEN_CONNS", "25"))
	maxIdle, _ := strconv.Atoi(getEnv("DB_MAX_IDLE_CONNS", "5"))
	connMaxLifetimeStr := getEnv("DB_CONN_MAX_LIFETIME", "1h")
	connMaxLifetime, _ := time.ParseDuration(connMaxLifetimeStr)
	connectTimeoutStr := getEnv("DB_CONNECT_TIMEOUT", "5s")
	connectTimeout, _ := time.ParseDuration(connectTimeoutStr)

	return Config{
		Driver:          getEnv("DB_DRIVER", "sqlite"),
		DSN:             getEnv("DB_DSN", "./data/app.db"),
		MaxOpenConns:    maxOpen,
		MaxIdleConns:    maxIdle,
		ConnMaxLifetime: connMaxLifetime,
		ConnectTimeout:  connectTimeout,
		Logger:          l,
	}
}

func Open(ctx context.Context, cfg Config) (*sql.DB, error) {
	logger := cfg.Logger.With().Str("component", "db").Logger()
	logger.Info().Str("driver", cfg.Driver).Msg("Opening DB connection")

	driverName := cfg.Driver
	switch cfg.Driver {
	case "sqlite", "sqlite3":
		driverName = "sqlite"
	case "postgres", "pg", "pgx":
		driverName = "pgx"
	case "mysql", "mariadb":
		driverName = "mysql"
	}

	db, err := sql.Open(driverName, cfg.DSN)
	if err != nil {
		logger.Error().Err(err).Msg("sql.Open failed")
		return nil, fmt.Errorf("open db: %w", err)
	}

	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	ctxPing, cancel := context.WithTimeout(ctx, cfg.ConnectTimeout)
	defer cancel()
	if err := db.PingContext(ctxPing); err != nil {
		_ = db.Close()
		logger.Error().Err(err).Msg("db ping failed")
		return nil, fmt.Errorf("ping db: %w", err)
	}

	logger.Info().Msg("DB connection established")
	return db, nil
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}
