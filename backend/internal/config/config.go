package config

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv                        string
	LogLevel                      string
	AppPublicURL                  string
	HTTPAddr                      string
	SwaggerEnabled                bool
	JWTSecret                     string
	AccessTokenTTL                time.Duration
	RefreshTokenTTL               time.Duration
	CookieDomain                  string
	CookieSecure                  bool
	CookieSameSite                string
	CORSAllowedOrigins            []string
	TrustedProxies                []string
	Realtime                      RealtimeConfig
	SeedUserEnabled               bool
	SeedUserFirstName             string
	SeedUserLastName              string
	SeedUserEmail                 string
	SeedUserPassword              string
	SeedDummyNotificationsEnabled bool
	SeedDummyNotificationsEmail   string
	DBConfig                      DBConfig
}

type DBConfig struct {
	Type               string
	Dsn                string
	MaxOpenConns       int
	MaxIdleConns       int
	ConnMaxLifetime    time.Duration
	ConnectTimeout     time.Duration
	AllowUnsafeSSLMode bool
}

type RealtimeConfig struct {
	Bus              string
	NodeID           string
	PostgresChannel  string
	SubscriberBuffer int
	EventTTL         time.Duration
}

const DefaultIssuer = "go_infra_link"
const defaultJWTSecret = "change-me"

func Load() (Config, error) {
	loadEnvFiles()
	env := newEnvParser(os.LookupEnv)
	appEnv := env.First("development", "APP_ENV", "ENV")

	cfg := Config{
		AppEnv:             appEnv,
		LogLevel:           env.First("info", "APP_LOG_LEVEL", "LOG_LEVEL"),
		AppPublicURL:       normalizePublicURL(env.First("http://localhost:5173", "APP_PUBLIC_URL", "PUBLIC_APP_URL", "FRONTEND_PUBLIC_URL")),
		HTTPAddr:           resolveHTTPAddr(env),
		SwaggerEnabled:     env.Bool("SWAGGER_ENABLED", !IsProduction(appEnv)),
		JWTSecret:          env.String("JWT_SECRET", defaultJWTSecret),
		AccessTokenTTL:     env.Duration("ACCESS_TOKEN_TTL", 15*time.Minute),
		RefreshTokenTTL:    env.Duration("REFRESH_TOKEN_TTL", 720*time.Hour),
		CookieDomain:       env.String("COOKIE_DOMAIN", ""),
		CookieSecure:       env.Bool("COOKIE_SECURE", false),
		CookieSameSite:     normalizeSameSite(env.String("COOKIE_SAME_SITE", "strict")),
		CORSAllowedOrigins: env.List("CORS_ALLOWED_ORIGINS"),
		TrustedProxies:     env.List("TRUSTED_PROXIES"),
		Realtime: RealtimeConfig{
			Bus:              normalizeRealtimeBus(env.First("memory", "REALTIME_BUS", "REALTIME_ADAPTER")),
			NodeID:           env.String("REALTIME_NODE_ID", ""),
			PostgresChannel:  normalizeRealtimePostgresChannel(env.String("REALTIME_POSTGRES_CHANNEL", "go_infra_link_realtime")),
			SubscriberBuffer: env.Int("REALTIME_SUBSCRIBER_BUFFER", 64),
			EventTTL:         env.Duration("REALTIME_EVENT_TTL", 10*time.Minute),
		},
		DBConfig: DBConfig{
			Type:               normalizeDBType(env.First("postgres", "DB_TYPE", "DB_DRIVER")),
			MaxOpenConns:       env.Int("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:       env.Int("DB_MAX_IDLE_CONNS", 5),
			ConnMaxLifetime:    env.Duration("DB_CONN_MAX_LIFETIME", time.Hour),
			ConnectTimeout:     env.Duration("DB_CONNECT_TIMEOUT", 5*time.Second),
			AllowUnsafeSSLMode: env.Bool("DB_ALLOW_UNSAFE_SSLMODE", env.Bool("DB_ALLOW_UNSAFE_SSL", false)),
		},
	}

	applySeedUserConfig(&cfg, env)
	applySeedDummyNotificationConfig(&cfg, env)
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

func (p envParser) List(key string) []string {
	value, ok := p.lookup(key)
	if !ok || strings.TrimSpace(value) == "" {
		return nil
	}

	parts := strings.Split(value, ",")
	values := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			values = append(values, trimmed)
		}
	}
	return values
}

func applySeedUserConfig(cfg *Config, env envParser) {
	firstNameDefault, lastNameDefault, emailDefault, passwordDefault := seedUserDefaults(cfg.AppEnv)
	cfg.SeedUserEnabled = env.Bool("SEED_USER_ENABLED", !IsProduction(cfg.AppEnv))
	cfg.SeedUserFirstName = env.String("SEED_USER_FIRST_NAME", firstNameDefault)
	cfg.SeedUserLastName = env.String("SEED_USER_LAST_NAME", lastNameDefault)
	cfg.SeedUserEmail = env.String("SEED_USER_EMAIL", emailDefault)
	cfg.SeedUserPassword = env.String("SEED_USER_PASSWORD", passwordDefault)
}

func applySeedDummyNotificationConfig(cfg *Config, env envParser) {
	cfg.SeedDummyNotificationsEnabled = env.Bool("SEED_DUMMY_NOTIFICATIONS", !IsProduction(cfg.AppEnv))
	cfg.SeedDummyNotificationsEmail = env.String(
		"SEED_DUMMY_NOTIFICATIONS_EMAIL",
		cfg.SeedUserEmail,
	)
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
	pgSSLMode := env.First("disable", "DB_SSLMODE", "POSTGRES_SSLMODE")

	fallback := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		pgHost,
		pgUser,
		pgPassword,
		pgDatabase,
		pgPort,
		pgSSLMode,
	)

	return env.First(fallback, "DATABASE_URL", "DB_DSN")
}

func validateConfig(cfg Config) error {
	var errs []error
	if IsProduction(cfg.AppEnv) {
		if isInsecureJWTSecret(cfg.JWTSecret) {
			errs = append(errs, fmt.Errorf("JWT_SECRET must be set to a strong non-default value in production"))
		}
		if !cfg.CookieSecure {
			errs = append(errs, fmt.Errorf("COOKIE_SECURE must be true in production because auth cookies require HTTPS"))
		}
		if len(cfg.TrustedProxies) == 0 {
			errs = append(errs, fmt.Errorf("TRUSTED_PROXIES must be set explicitly in production to the reverse proxy IP/CIDR"))
		}
	}

	if err := validateCookieDomain(cfg.CookieDomain, IsProduction(cfg.AppEnv)); err != nil {
		errs = append(errs, err)
	}
	if err := validateCookieSameSite(cfg.CookieSameSite, IsProduction(cfg.AppEnv)); err != nil {
		errs = append(errs, err)
	}
	if err := validateCORSAllowedOrigins(cfg.CORSAllowedOrigins, IsProduction(cfg.AppEnv)); err != nil {
		errs = append(errs, err)
	}
	if err := validateAppPublicURL(cfg.AppPublicURL, IsProduction(cfg.AppEnv)); err != nil {
		errs = append(errs, err)
	}
	if err := validateDatabaseSSLMode(cfg); err != nil {
		errs = append(errs, err)
	}
	if err := validateRealtimeConfig(cfg.Realtime); err != nil {
		errs = append(errs, err)
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
	if IsProduction(cfg.AppEnv) && cfg.SeedDummyNotificationsEnabled {
		if strings.TrimSpace(cfg.SeedDummyNotificationsEmail) == "" {
			errs = append(errs, fmt.Errorf("SEED_DUMMY_NOTIFICATIONS_EMAIL is required when SEED_DUMMY_NOTIFICATIONS=true in production"))
		}
	}

	for _, proxy := range cfg.TrustedProxies {
		if !isValidTrustedProxy(proxy) {
			errs = append(errs, fmt.Errorf("TRUSTED_PROXIES contains invalid IP/CIDR %q", proxy))
			continue
		}
		if IsProduction(cfg.AppEnv) && isWildcardTrustedProxy(proxy) {
			errs = append(errs, fmt.Errorf("TRUSTED_PROXIES must not trust every proxy in production: %q", proxy))
		}
	}

	return errors.Join(errs...)
}

func isInsecureJWTSecret(secret string) bool {
	value := strings.TrimSpace(secret)
	lower := strings.ToLower(value)
	if len(value) < 32 {
		return true
	}
	switch lower {
	case "", defaultJWTSecret, "changeme", "change_me", "secret", "jwt-secret", "super-long-secret-change-me-in-production":
		return true
	default:
		return strings.Contains(lower, "change-me") || strings.Contains(lower, "change_me") || strings.Contains(lower, "change me")
	}
}

func validateCookieDomain(domain string, production bool) error {
	domain = strings.TrimSpace(domain)
	if domain == "" {
		return nil
	}

	normalized := strings.TrimPrefix(domain, ".")
	switch {
	case strings.Contains(domain, "://"):
		return fmt.Errorf("COOKIE_DOMAIN must be a domain only, not a URL")
	case strings.Contains(domain, "*"):
		return fmt.Errorf("COOKIE_DOMAIN must not contain wildcards")
	case strings.ContainsAny(domain, `/\:_`) || strings.ContainsAny(domain, " \t\r\n"):
		return fmt.Errorf("COOKIE_DOMAIN must not include a port, path, underscore, or whitespace")
	case normalized == "":
		return fmt.Errorf("COOKIE_DOMAIN must include a hostname")
	case net.ParseIP(normalized) != nil:
		return fmt.Errorf("COOKIE_DOMAIN must be a DNS domain, not an IP address")
	}

	if production {
		lower := strings.ToLower(normalized)
		if lower == "localhost" || !strings.Contains(lower, ".") || strings.Contains(lower, "change-me") || strings.Contains(lower, "change_me") {
			return fmt.Errorf("COOKIE_DOMAIN must be a real production domain when set")
		}
	}
	return nil
}

func validateCookieSameSite(value string, production bool) error {
	switch normalizeSameSite(value) {
	case "strict", "lax":
		return nil
	case "none":
		if production {
			return fmt.Errorf("COOKIE_SAME_SITE=none is not allowed in production for the same-origin SPA; use strict or lax")
		}
		return nil
	default:
		return fmt.Errorf("COOKIE_SAME_SITE must be one of strict, lax, or none")
	}
}

func validateCORSAllowedOrigins(origins []string, production bool) error {
	var errs []error
	for _, origin := range origins {
		trimmed := strings.TrimSpace(origin)
		if trimmed == "*" {
			errs = append(errs, fmt.Errorf("CORS_ALLOWED_ORIGINS must not contain wildcard '*'"))
			continue
		}
		parsed, err := url.Parse(trimmed)
		if err != nil || parsed.Scheme == "" || parsed.Host == "" || parsed.Path != "" || parsed.RawQuery != "" || parsed.Fragment != "" {
			errs = append(errs, fmt.Errorf("CORS_ALLOWED_ORIGINS contains invalid origin %q", origin))
			continue
		}
		if parsed.Scheme != "http" && parsed.Scheme != "https" {
			errs = append(errs, fmt.Errorf("CORS_ALLOWED_ORIGINS contains unsupported scheme %q", parsed.Scheme))
			continue
		}
		if production && parsed.Scheme != "https" {
			errs = append(errs, fmt.Errorf("CORS_ALLOWED_ORIGINS must use https in production: %q", origin))
		}
	}
	return errors.Join(errs...)
}

func validateAppPublicURL(value string, production bool) error {
	trimmed := strings.TrimSpace(value)
	parsed, err := url.Parse(trimmed)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" || parsed.RawQuery != "" || parsed.Fragment != "" {
		return fmt.Errorf("APP_PUBLIC_URL must be an absolute URL without query or fragment")
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return fmt.Errorf("APP_PUBLIC_URL contains unsupported scheme %q", parsed.Scheme)
	}
	if production && parsed.Scheme != "https" {
		return fmt.Errorf("APP_PUBLIC_URL must use https in production")
	}
	return nil
}

func validateDatabaseSSLMode(cfg Config) error {
	if !IsProduction(cfg.AppEnv) || cfg.DBConfig.Type != "postgres" {
		return nil
	}

	sslMode := extractPostgresSSLMode(cfg.DBConfig.Dsn)
	if isUnsafePostgresSSLMode(sslMode) && !cfg.DBConfig.AllowUnsafeSSLMode {
		if sslMode == "" {
			return fmt.Errorf("DATABASE_URL/DB_DSN must include a safe sslmode in production, or set DB_ALLOW_UNSAFE_SSLMODE=true for an intentionally private database link")
		}
		return fmt.Errorf("database sslmode=%s is unsafe in production; use require, verify-ca, or verify-full, or set DB_ALLOW_UNSAFE_SSLMODE=true if this is intentional", sslMode)
	}
	return nil
}

func extractPostgresSSLMode(dsn string) string {
	dsn = strings.TrimSpace(dsn)
	if dsn == "" {
		return ""
	}

	if strings.Contains(dsn, "://") {
		parsed, err := url.Parse(dsn)
		if err != nil {
			return ""
		}
		return strings.ToLower(strings.TrimSpace(parsed.Query().Get("sslmode")))
	}

	for _, field := range strings.Fields(dsn) {
		key, value, ok := strings.Cut(field, "=")
		if !ok || !strings.EqualFold(key, "sslmode") {
			continue
		}
		return strings.ToLower(strings.Trim(strings.TrimSpace(value), `'"`))
	}
	return ""
}

func isUnsafePostgresSSLMode(sslMode string) bool {
	switch strings.ToLower(strings.TrimSpace(sslMode)) {
	case "require", "verify-ca", "verify-full":
		return false
	default:
		return true
	}
}

func isValidTrustedProxy(proxy string) bool {
	if net.ParseIP(proxy) != nil {
		return true
	}

	_, _, err := net.ParseCIDR(proxy)
	return err == nil
}

func isWildcardTrustedProxy(proxy string) bool {
	if strings.TrimSpace(proxy) == "0.0.0.0" || strings.TrimSpace(proxy) == "::" {
		return true
	}
	_, network, err := net.ParseCIDR(proxy)
	if err != nil {
		return false
	}
	ones, bits := network.Mask.Size()
	return ones == 0 && (bits == 32 || bits == 128)
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

func normalizeRealtimeBus(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	switch value {
	case "", "memory", "inmemory", "in_memory", "local":
		return "memory"
	case "postgres", "postgresql", "pg", "listen_notify", "listen-notify":
		return "postgres"
	default:
		return value
	}
}

func validateRealtimeConfig(cfg RealtimeConfig) error {
	var errs []error
	switch normalizeRealtimeBus(cfg.Bus) {
	case "memory", "postgres":
	default:
		errs = append(errs, fmt.Errorf("REALTIME_BUS must be one of memory or postgres"))
	}
	if cfg.SubscriberBuffer < 0 {
		errs = append(errs, fmt.Errorf("REALTIME_SUBSCRIBER_BUFFER must be >= 0"))
	}
	if cfg.EventTTL < 0 {
		errs = append(errs, fmt.Errorf("REALTIME_EVENT_TTL must be >= 0"))
	}
	if normalizeRealtimeBus(cfg.Bus) == "postgres" && !isSafeRealtimePostgresChannel(cfg.PostgresChannel) {
		errs = append(errs, fmt.Errorf("REALTIME_POSTGRES_CHANNEL contains an invalid PostgreSQL channel name"))
	}
	return errors.Join(errs...)
}

func normalizeRealtimePostgresChannel(channel string) string {
	channel = strings.TrimSpace(channel)
	if isSafeRealtimePostgresChannel(channel) {
		return channel
	}
	return "go_infra_link_realtime"
}

func isSafeRealtimePostgresChannel(channel string) bool {
	if channel == "" || len(channel) > 63 {
		return false
	}
	for index, r := range channel {
		switch {
		case r == '_':
		case r >= 'a' && r <= 'z':
		case r >= 'A' && r <= 'Z':
		case index > 0 && r >= '0' && r <= '9':
		default:
			return false
		}
	}
	return true
}

func normalizeSameSite(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func normalizePublicURL(value string) string {
	return strings.TrimRight(strings.TrimSpace(value), "/")
}
