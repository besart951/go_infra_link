package config

import (
	"fmt"
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
	DBType            string
	DBDsn             string
	DBMaxOpenConns    int
	DBMaxIdleConns    int
	DBConnMaxLifetime time.Duration
	DBConnectTimeout  time.Duration
}

const DefaultIssuer = "go_infra_link"

func Load() (Config, error) {
	loadEnvFiles()

	appEnv := getEnvFirst("development", "APP_ENV", "ENV")
	logLevel := getEnvFirst("info", "APP_LOG_LEVEL", "LOG_LEVEL")
	jwtSecret := getEnv("JWT_SECRET", "change-me")

	if IsProduction(appEnv) {
		if jwtSecret == "change-me" {
			return Config{}, fmt.Errorf("missing JWT_SECRET in production environment")
		}
	}

	maxOpen := getEnvInt("DB_MAX_OPEN_CONNS", 25)
	maxIdle := getEnvInt("DB_MAX_IDLE_CONNS", 5)
	connMaxLifetime := getEnvDuration("DB_CONN_MAX_LIFETIME", time.Hour)
	connectTimeout := getEnvDuration("DB_CONNECT_TIMEOUT", 5*time.Second)
	accessTokenTTL := getEnvDuration("ACCESS_TOKEN_TTL", 15*time.Minute)
	refreshTokenTTL := getEnvDuration("REFRESH_TOKEN_TTL", 720*time.Hour)
	cookieSecure := getEnvBool("COOKIE_SECURE", false)

	seedUserEnabled := getEnvBool("SEED_USER_ENABLED", !IsProduction(appEnv))
	devAuthEnabled := getEnvBool("DEV_AUTH_ENABLED", false)

	dbType := normalizeDBType(getEnvFirst("sqlite", "DB_TYPE", "DB_DRIVER"))
	dbDsnFallback := "host=localhost user=postgres password=postgres dbname=mydb port=5432 sslmode=disable"
	if dbType == "sqlite" {
		dbDsnFallback = "./data/app.db"
	}
	dbDsn := getEnvFirst(dbDsnFallback, "DATABASE_URL", "DB_DSN")

	return Config{
		AppEnv:            appEnv,
		LogLevel:          logLevel,
		HTTPAddr:          getEnv("HTTP_ADDR", ":8080"),
		JWTSecret:         jwtSecret,
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
		DBType:            dbType,
		DBDsn:             dbDsn,
		DBMaxOpenConns:    maxOpen,
		DBMaxIdleConns:    maxIdle,
		DBConnMaxLifetime: connMaxLifetime,
		DBConnectTimeout:  connectTimeout,
	}, nil
}

func loadEnvFiles() {
	// .env files are optional; ignore missing/unreadable files.
	discardErr(godotenv.Load())
	discardErr(godotenv.Load("configs/.env"))
	discardErr(godotenv.Load("../.env"))
}

func discardErr(err error) {
	_ = err
}

func getEnvInt(key string, fallback int) int {
	v, ok := os.LookupEnv(key)
	if !ok || v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}

func getEnvBool(key string, fallback bool) bool {
	v, ok := os.LookupEnv(key)
	if !ok || v == "" {
		return fallback
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return fallback
	}
	return b
}

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	v, ok := os.LookupEnv(key)
	if !ok || v == "" {
		return fallback
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return fallback
	}
	return d
}

func IsProduction(env string) bool {
	return strings.EqualFold(env, "production") || strings.EqualFold(env, "prod")
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

func normalizeDBType(dbType string) string {
	dbType = strings.ToLower(strings.TrimSpace(dbType))
	switch dbType {
	case "sqlite3", "sqlite":
		return "sqlite"
	case "postgres", "pg", "postgresql", "pgx":
		return "postgres"
	case "mysql", "mariadb":
		return "mysql"
	default:
		return dbType
	}
}
