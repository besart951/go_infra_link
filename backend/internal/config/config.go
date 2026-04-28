package config

import (
	"errors"
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
	DBConfig          DBConfig
}

type DBConfig struct {
	Type            string
	Dsn             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnectTimeout  time.Duration
}

const DefaultIssuer = "go_infra_link"

func Load() (Config, error) {
	loadEnvFiles()
	env := newEnvParser(os.LookupEnv)
	appEnv := env.First("development", "APP_ENV", "ENV")

	cfg := Config{
		AppEnv:            appEnv,
		LogLevel:          env.First("info", "APP_LOG_LEVEL", "LOG_LEVEL"),
		HTTPAddr:          resolveHTTPAddr(env),
		SwaggerEnabled:    env.Bool("SWAGGER_ENABLED", !IsProduction(appEnv)),
		JWTSecret:         env.String("JWT_SECRET", "change-me"),
		AccessTokenTTL:    env.Duration("ACCESS_TOKEN_TTL", 15*time.Minute),
		RefreshTokenTTL:   env.Duration("REFRESH_TOKEN_TTL", 720*time.Hour),
		CookieDomain:      env.String("COOKIE_DOMAIN", ""),
		CookieSecure:      env.Bool("COOKIE_SECURE", false),
		DBConfig: DBConfig{
			Type:            normalizeDBType(env.First("postgres", "DB_TYPE", "DB_DRIVER")),
			MaxOpenConns:    env.Int("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    env.Int("DB_MAX_IDLE_CONNS", 5),
			ConnMaxLifetime: env.Duration("DB_CONN_MAX_LIFETIME", time.Hour),
			ConnectTimeout:  env.Duration("DB_CONNECT_TIMEOUT", 5*time.Second),
		},
	}

	applySeedUserConfig(&cfg, env)
	cfg.DBConfig.Dsn = resolveDatabaseDSN(env)

	if err := validateConfig(cfg); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

type envParser struct {
	lookup func(string) (string, bool)
}

func newEnvParser(lookup func(string) (string, bool)) envParser {
	return envParser{lookup: lookup}
}

func (p envParser) String(key, fallback string) string {
	if value, ok := p.lookup(key); ok && value != "" {
		return value
	}
	return fallback
}

func (p envParser) First(fallback string, keys ...string) string {
	for _, key := range keys {
		if value, ok := p.lookup(key); ok && value != "" {
			return value
		}
	}
	return fallback
}

func (p envParser) Int(key string, fallback int) int {
	value, ok := p.lookup(key)
	if !ok || value == "" {
		return fallback
	}
	n, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return n
}

func (p envParser) Bool(key string, fallback bool) bool {
	value, ok := p.lookup(key)
	if !ok || value == "" {
		return fallback
	}
	b, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}
	return b
}

func (p envParser) Duration(key string, fallback time.Duration) time.Duration {
	value, ok := p.lookup(key)
	if !ok || value == "" {
		return fallback
	}
	d, err := time.ParseDuration(value)
	if err != nil {
		return fallback
	}
	return d
}

func applySeedUserConfig(cfg *Config, env envParser) {
	firstNameDefault, lastNameDefault, emailDefault, passwordDefault := seedUserDefaults(cfg.AppEnv)
	cfg.SeedUserEnabled = env.Bool("SEED_USER_ENABLED", !IsProduction(cfg.AppEnv))
	cfg.SeedUserFirstName = env.String("SEED_USER_FIRST_NAME", firstNameDefault)
	cfg.SeedUserLastName = env.String("SEED_USER_LAST_NAME", lastNameDefault)
	cfg.SeedUserEmail = env.String("SEED_USER_EMAIL", emailDefault)
	cfg.SeedUserPassword = env.String("SEED_USER_PASSWORD", passwordDefault)
}

func seedUserDefaults(appEnv string) (firstName, lastName, email, password string) {
	if IsProduction(appEnv) {
		return "", "", "", ""
	}

	return "Besart", "Morina", "besart_morina@hotmail.com", "password"
}

func resolveDatabaseDSN(env envParser) string {
	pgHost := env.String("POSTGRES_HOST", "localhost")
	pgPort := env.String("POSTGRES_PORT", "5432")
	pgUser := env.String("POSTGRES_USER", "postgres")
	pgPassword := env.String("POSTGRES_PASSWORD", "postgres")
	pgDatabase := env.String("POSTGRES_DB", "go_infra_link")

	fallback := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		pgHost,
		pgUser,
		pgPassword,
		pgDatabase,
		pgPort,
	)

	return env.First(fallback, "DATABASE_URL", "DB_DSN")
}

func validateConfig(cfg Config) error {
	var errs []error
	if IsProduction(cfg.AppEnv) && cfg.JWTSecret == "change-me" {
		errs = append(errs, fmt.Errorf("missing JWT_SECRET in production environment"))
	}

	if IsProduction(cfg.AppEnv) && cfg.SeedUserEnabled {
		switch {
		case strings.TrimSpace(cfg.SeedUserEmail) == "":
			errs = append(errs, fmt.Errorf("SEED_USER_EMAIL is required when SEED_USER_ENABLED=true in production"))
		case strings.TrimSpace(cfg.SeedUserPassword) == "":
			errs = append(errs, fmt.Errorf("SEED_USER_PASSWORD is required when SEED_USER_ENABLED=true in production"))
		case cfg.SeedUserPassword == "password":
			errs = append(errs, fmt.Errorf("SEED_USER_PASSWORD must not use the default development password in production"))
		}
	}

	return errors.Join(errs...)
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

func IsProduction(env string) bool {
	return strings.EqualFold(env, "production") || strings.EqualFold(env, "prod")
}

func resolveHTTPAddr(env envParser) string {
	if addr := env.String("HTTP_ADDR", ""); addr != "" {
		return addr
	}
	if port := env.String("BACKEND_PORT", ""); port != "" {
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
	default:
		return dbType
	}
}
