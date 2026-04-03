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
	SwaggerEnabled    bool
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
	swaggerEnabled := getEnvBool("SWAGGER_ENABLED", !IsProduction(appEnv))

	seedUserEnabled := getEnvBool("SEED_USER_ENABLED", !IsProduction(appEnv))
	seedUserFirstNameDefault := "Besart"
	seedUserLastNameDefault := "Morina"
	seedUserEmailDefault := "besart_morina@hotmail.com"
	seedUserPasswordDefault := "password"
	if IsProduction(appEnv) {
		seedUserFirstNameDefault = ""
		seedUserLastNameDefault = ""
		seedUserEmailDefault = ""
		seedUserPasswordDefault = ""
	}
	seedUserFirstName := getEnv("SEED_USER_FIRST_NAME", seedUserFirstNameDefault)
	seedUserLastName := getEnv("SEED_USER_LAST_NAME", seedUserLastNameDefault)
	seedUserEmail := getEnv("SEED_USER_EMAIL", seedUserEmailDefault)
	seedUserPassword := getEnv("SEED_USER_PASSWORD", seedUserPasswordDefault)
	if IsProduction(appEnv) && seedUserEnabled {
		switch {
		case strings.TrimSpace(seedUserEmail) == "":
			return Config{}, fmt.Errorf("SEED_USER_EMAIL is required when SEED_USER_ENABLED=true in production")
		case strings.TrimSpace(seedUserPassword) == "":
			return Config{}, fmt.Errorf("SEED_USER_PASSWORD is required when SEED_USER_ENABLED=true in production")
		case seedUserPassword == "password":
			return Config{}, fmt.Errorf("SEED_USER_PASSWORD must not use the default development password in production")
		}
	}
	dbType := normalizeDBType(getEnvFirst("postgres", "DB_TYPE", "DB_DRIVER"))
	pgHost := getEnv("POSTGRES_HOST", "localhost")
	pgPort := getEnv("POSTGRES_PORT", "5432")
	pgUser := getEnv("POSTGRES_USER", "postgres")
	pgPassword := getEnv("POSTGRES_PASSWORD", "postgres")
	pgDatabase := getEnv("POSTGRES_DB", "go_infra_link")
	dbDsnFallback := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		pgHost,
		pgUser,
		pgPassword,
		pgDatabase,
		pgPort,
	)
	dbDsn := getEnvFirst(dbDsnFallback, "DATABASE_URL", "DB_DSN")

	return Config{
		AppEnv:            appEnv,
		LogLevel:          logLevel,
		HTTPAddr:          resolveHTTPAddr(),
		SwaggerEnabled:    swaggerEnabled,
		JWTSecret:         jwtSecret,
		AccessTokenTTL:    accessTokenTTL,
		RefreshTokenTTL:   refreshTokenTTL,
		CookieDomain:      getEnv("COOKIE_DOMAIN", ""),
		CookieSecure:      cookieSecure,
		SeedUserEnabled:   seedUserEnabled,
		SeedUserFirstName: seedUserFirstName,
		SeedUserLastName:  seedUserLastName,
		SeedUserEmail:     seedUserEmail,
		SeedUserPassword:  seedUserPassword,
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
	// Prefer a repo-root .env so frontend and backend can share one file.
	// If we're running from the backend/ directory, repo root is ../.env.
	if fileExists("../.env") {
		discardErr(godotenv.Load("../.env"))
	} else {
		discardErr(godotenv.Load(".env"))
	}
	// Optional additional env locations.
	discardErr(godotenv.Load("configs/.env"))
}

func fileExists(path string) bool {
	if _, err := os.Stat(path); err != nil {
		return false
	}
	return true
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

func resolveHTTPAddr() string {
	if addr := getEnv("HTTP_ADDR", ""); addr != "" {
		return addr
	}
	if port := getEnv("BACKEND_PORT", ""); port != "" {
		if strings.HasPrefix(port, ":") {
			return port
		}
		return ":" + port
	}
	return ":8080"
}

func normalizeDBType(dbType string) string {
	dbType = strings.ToLower(strings.TrimSpace(dbType))
	switch dbType {
	case "postgres", "pg", "postgresql", "pgx":
		return "postgres"
	case "mysql", "mariadb":
		return "mysql"
	default:
		return dbType
	}
}
