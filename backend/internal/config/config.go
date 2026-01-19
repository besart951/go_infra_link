package config

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv            string
	LogLevel          string
	HTTPAddr          string
	DBDriver          string
	DBDsn             string
	DBMaxOpenConns    int
	DBMaxIdleConns    int
	DBConnMaxLifetime time.Duration
	DBConnectTimeout  time.Duration
}

func Load() Config {
	_ = godotenv.Load("configs/.env")

	maxOpen, _ := strconv.Atoi(getEnv("DB_MAX_OPEN_CONNS", "25"))
	maxIdle, _ := strconv.Atoi(getEnv("DB_MAX_IDLE_CONNS", "5"))
	connMaxLifetime, _ := time.ParseDuration(getEnv("DB_CONN_MAX_LIFETIME", "1h"))
	connectTimeout, _ := time.ParseDuration(getEnv("DB_CONNECT_TIMEOUT", "5s"))

	dbDriver := normalizeDriver(getEnv("DB_DRIVER", "sqlite"))

	return Config{
		AppEnv:            getEnv("APP_ENV", "development"),
		LogLevel:          getEnv("APP_LOG_LEVEL", "info"),
		HTTPAddr:          getEnv("HTTP_ADDR", ":8080"),
		DBDriver:          dbDriver,
		DBDsn:             getEnv("DB_DSN", "./data/app.db"),
		DBMaxOpenConns:    maxOpen,
		DBMaxIdleConns:    maxIdle,
		DBConnMaxLifetime: connMaxLifetime,
		DBConnectTimeout:  connectTimeout,
	}
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}

func normalizeDriver(driver string) string {
	switch driver {
	case "sqlite3", "sqlite":
		return "sqlite"
	case "postgres", "pg", "pgx":
		return "pgx"
	case "mysql", "mariadb":
		return "mysql"
	default:
		return driver
	}
}
