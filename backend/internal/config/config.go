package config

import (
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv            string
	LogLevel          string
	HTTPAddr          string
	JWTSecret         string
	AccessTokenTTL    time.Duration
	RefreshTokenTTL   time.Duration
	CookieDomain      string
	CookieSecure      bool
	SeedUserEnabled   bool
	SeedUserFirstName string
	SeedUserLastName  string
	SeedUserEmail     string
	SeedUserPassword  string
	DevAuthEnabled    bool
	DevAuthEmail      string
	DevAuthPassword   string
	DBDriver          string
	DBDsn             string
	DBMaxOpenConns    int
	DBMaxIdleConns    int
	DBConnMaxLifetime time.Duration
	DBConnectTimeout  time.Duration
}

func Load() Config {
	// Load .env (if present). We try a few common locations so it works whether
	// you run from `backend/` or from a subdir.
	_ = godotenv.Load()
	_ = godotenv.Load("configs/.env")
	_ = godotenv.Load("../.env")

	appEnv := getEnvFirst("development", "APP_ENV", "ENV")
	logLevel := getEnvFirst("info", "APP_LOG_LEVEL", "LOG_LEVEL")

	maxOpen, _ := strconv.Atoi(getEnv("DB_MAX_OPEN_CONNS", "25"))
	maxIdle, _ := strconv.Atoi(getEnv("DB_MAX_IDLE_CONNS", "5"))
	connMaxLifetime, _ := time.ParseDuration(getEnv("DB_CONN_MAX_LIFETIME", "1h"))
	connectTimeout, _ := time.ParseDuration(getEnv("DB_CONNECT_TIMEOUT", "5s"))
	accessTokenTTL, _ := time.ParseDuration(getEnv("ACCESS_TOKEN_TTL", "15m"))
	refreshTokenTTL, _ := time.ParseDuration(getEnv("REFRESH_TOKEN_TTL", "720h"))
	cookieSecure, _ := strconv.ParseBool(getEnv("COOKIE_SECURE", "false"))

	seedEnabledDefault := "true"
	if strings.EqualFold(appEnv, "production") || strings.EqualFold(appEnv, "prod") {
		seedEnabledDefault = "false"
	}
	seedUserEnabled, _ := strconv.ParseBool(getEnv("SEED_USER_ENABLED", seedEnabledDefault))
	devAuthEnabled, _ := strconv.ParseBool(getEnv("DEV_AUTH_ENABLED", "false"))

	dbDriver := normalizeDriver(getEnvFirst("sqlite", "DB_DRIVER"))
	// Prefer DATABASE_URL if present (common convention), else DB_DSN.
	// DSN fallback depends on selected driver.
	dbDsnFallback := "host=localhost user=postgres password=postgres dbname=mydb port=5432 sslmode=disable"
	if dbDriver == "sqlite" {
		dbDsnFallback = "./data/app.db"
	}
	dbDsn := getEnvFirst(dbDsnFallback, "DATABASE_URL", "DB_DSN")

	return Config{
		AppEnv:            appEnv,
		LogLevel:          logLevel,
		HTTPAddr:          getEnv("HTTP_ADDR", ":8080"),
		JWTSecret:         getEnv("JWT_SECRET", "change-me"),
		AccessTokenTTL:    accessTokenTTL,
		RefreshTokenTTL:   refreshTokenTTL,
		CookieDomain:      getEnv("COOKIE_DOMAIN", ""),
		CookieSecure:      cookieSecure,
		SeedUserEnabled:   seedUserEnabled,
		SeedUserFirstName: getEnv("SEED_USER_FIRST_NAME", "Besart"),
		SeedUserLastName:  getEnv("SEED_USER_LAST_NAME", "Morina"),
		SeedUserEmail:     getEnv("SEED_USER_EMAIL", "besart_morina@hotmail.com"),
		SeedUserPassword:  getEnv("SEED_USER_PASSWORD", "password"),
		DevAuthEnabled:    devAuthEnabled,
		DevAuthEmail:      getEnv("DEV_AUTH_EMAIL", ""),
		DevAuthPassword:   getEnv("DEV_AUTH_PASSWORD", ""),
		DBDriver:          dbDriver,
		DBDsn:             dbDsn,
		DBMaxOpenConns:    maxOpen,
		DBMaxIdleConns:    maxIdle,
		DBConnMaxLifetime: connMaxLifetime,
		DBConnectTimeout:  connectTimeout,
	}
}

func getEnvFirst(fallback string, keys ...string) string {
	for _, key := range keys {
		if v, ok := os.LookupEnv(key); ok && v != "" {
			return v
		}
	}
	return fallback
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}

func normalizeDriver(driver string) string {
	driver = strings.ToLower(strings.TrimSpace(driver))
	switch driver {
	case "sqlite3", "sqlite":
		return "sqlite"
	case "postgres", "pg", "postgresql", "pgx":
		// gorm's driver package is named postgres; it uses pgx internally.
		return "postgres"
	case "mysql", "mariadb":
		return "mysql"
	default:
		return driver
	}
}
