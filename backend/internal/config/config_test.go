package config

import (
	"strings"
	"testing"
	"time"
)

func TestLoadProductionValidation(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	t.Setenv("JWT_SECRET", "change-me")
	t.Setenv("COOKIE_SECURE", "false")
	t.Setenv("TRUSTED_PROXIES", "")
	t.Setenv("APP_PUBLIC_URL", "http://app.example.com")
	t.Setenv("DATABASE_URL", "host=postgres user=postgres password=postgres dbname=go_infra_link port=5432 sslmode=disable")
	t.Setenv("SEED_USER_ENABLED", "true")
	t.Setenv("SEED_USER_EMAIL", "")
	t.Setenv("SEED_USER_PASSWORD", "password")

	_, err := Load()
	if err == nil {
		t.Fatal("expected Load to fail for invalid production configuration")
	}

	message := err.Error()
	if !strings.Contains(message, "JWT_SECRET must be set to a strong non-default value in production") {
		t.Fatalf("expected JWT validation error, got %q", message)
	}
	if !strings.Contains(message, "COOKIE_SECURE must be true in production") {
		t.Fatalf("expected cookie secure validation error, got %q", message)
	}
	if !strings.Contains(message, "TRUSTED_PROXIES must be set explicitly in production") {
		t.Fatalf("expected trusted proxy validation error, got %q", message)
	}
	if !strings.Contains(message, "database sslmode=disable is unsafe in production") {
		t.Fatalf("expected database sslmode validation error, got %q", message)
	}
	if !strings.Contains(message, "APP_PUBLIC_URL must use https in production") {
		t.Fatalf("expected public app URL validation error, got %q", message)
	}
	if !strings.Contains(message, "SEED_USER_EMAIL is required") {
		t.Fatalf("expected seed email validation error, got %q", message)
	}
}

func TestLoadAcceptsHardenedProductionConfig(t *testing.T) {
	setValidProductionEnv(t)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if cfg.CookieSameSite != "strict" {
		t.Fatalf("expected strict same-site cookie setting, got %q", cfg.CookieSameSite)
	}
	if got, want := strings.Join(cfg.CORSAllowedOrigins, ","), "https://app.example.com"; got != want {
		t.Fatalf("expected CORS origins %q, got %q", want, got)
	}
	if !cfg.CookieSecure {
		t.Fatal("expected secure cookies in production config")
	}
}

func TestLoadAllowsUnsafeProductionDatabaseSSLModeWhenExplicit(t *testing.T) {
	setValidProductionEnv(t)
	t.Setenv("DATABASE_URL", "host=postgres user=app password=secret dbname=go_infra_link port=5432 sslmode=disable")
	t.Setenv("DB_ALLOW_UNSAFE_SSLMODE", "true")

	if _, err := Load(); err != nil {
		t.Fatalf("expected unsafe DB SSL mode to be allowed when explicitly configured, got %v", err)
	}
}

func TestLoadRejectsWildcardCORSOrigin(t *testing.T) {
	setValidProductionEnv(t)
	t.Setenv("CORS_ALLOWED_ORIGINS", "*")

	_, err := Load()
	if err == nil {
		t.Fatal("expected Load to fail for wildcard CORS origin")
	}
	if !strings.Contains(err.Error(), "CORS_ALLOWED_ORIGINS must not contain wildcard") {
		t.Fatalf("expected CORS wildcard validation error, got %q", err.Error())
	}
}

func TestLoadRejectsUnsafeProductionCookieSettings(t *testing.T) {
	setValidProductionEnv(t)
	t.Setenv("COOKIE_SAME_SITE", "none")
	t.Setenv("COOKIE_DOMAIN", "http://example.com")

	_, err := Load()
	if err == nil {
		t.Fatal("expected Load to fail for unsafe cookie settings")
	}

	message := err.Error()
	if !strings.Contains(message, "COOKIE_SAME_SITE=none is not allowed in production") {
		t.Fatalf("expected same-site validation error, got %q", message)
	}
	if !strings.Contains(message, "COOKIE_DOMAIN must be a domain only") {
		t.Fatalf("expected cookie domain validation error, got %q", message)
	}
}

func TestLoadUsesTypedEnvParsing(t *testing.T) {
	t.Setenv("APP_ENV", "development")
	t.Setenv("JWT_SECRET", "super-secret")
	t.Setenv("DB_TYPE", "pgx")
	t.Setenv("DB_MAX_OPEN_CONNS", "42")
	t.Setenv("DB_MAX_IDLE_CONNS", "7")
	t.Setenv("DB_CONN_MAX_LIFETIME", "30m")
	t.Setenv("DB_CONNECT_TIMEOUT", "12s")
	t.Setenv("ACCESS_TOKEN_TTL", "20m")
	t.Setenv("REFRESH_TOKEN_TTL", "48h")
	t.Setenv("REALTIME_BUS", "pg")
	t.Setenv("REALTIME_NODE_ID", "backend-1")
	t.Setenv("REALTIME_POSTGRES_CHANNEL", "infra_realtime")
	t.Setenv("REALTIME_SUBSCRIBER_BUFFER", "128")
	t.Setenv("REALTIME_EVENT_TTL", "2m")
	t.Setenv("BACKEND_PORT", "9090")
	t.Setenv("TRUSTED_PROXIES", "127.0.0.1, 10.0.0.0/8")
	t.Setenv("COOKIE_SAME_SITE", "lax")
	t.Setenv("CORS_ALLOWED_ORIGINS", "http://localhost:5173,http://127.0.0.1:5173")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if cfg.DBConfig.Type != "postgres" {
		t.Fatalf("expected normalized DB type postgres, got %q", cfg.DBConfig.Type)
	}
	if cfg.DBConfig.MaxOpenConns != 42 {
		t.Fatalf("expected max open conns 42, got %d", cfg.DBConfig.MaxOpenConns)
	}
	if cfg.DBConfig.MaxIdleConns != 7 {
		t.Fatalf("expected max idle conns 7, got %d", cfg.DBConfig.MaxIdleConns)
	}
	if cfg.DBConfig.ConnMaxLifetime != 30*time.Minute {
		t.Fatalf("expected conn lifetime 30m, got %s", cfg.DBConfig.ConnMaxLifetime)
	}
	if cfg.DBConfig.ConnectTimeout != 12*time.Second {
		t.Fatalf("expected connect timeout 12s, got %s", cfg.DBConfig.ConnectTimeout)
	}
	if cfg.AccessTokenTTL != 20*time.Minute {
		t.Fatalf("expected access token TTL 20m, got %s", cfg.AccessTokenTTL)
	}
	if cfg.RefreshTokenTTL != 48*time.Hour {
		t.Fatalf("expected refresh token TTL 48h, got %s", cfg.RefreshTokenTTL)
	}
	if cfg.Realtime.Bus != "postgres" {
		t.Fatalf("expected realtime bus postgres, got %q", cfg.Realtime.Bus)
	}
	if cfg.Realtime.NodeID != "backend-1" {
		t.Fatalf("expected realtime node id backend-1, got %q", cfg.Realtime.NodeID)
	}
	if cfg.Realtime.PostgresChannel != "infra_realtime" {
		t.Fatalf("expected realtime channel infra_realtime, got %q", cfg.Realtime.PostgresChannel)
	}
	if cfg.Realtime.SubscriberBuffer != 128 {
		t.Fatalf("expected realtime subscriber buffer 128, got %d", cfg.Realtime.SubscriberBuffer)
	}
	if cfg.Realtime.EventTTL != 2*time.Minute {
		t.Fatalf("expected realtime event TTL 2m, got %s", cfg.Realtime.EventTTL)
	}
	if cfg.HTTPAddr != ":9090" {
		t.Fatalf("expected HTTP addr :9090, got %q", cfg.HTTPAddr)
	}
	if got, want := strings.Join(cfg.TrustedProxies, ","), "127.0.0.1,10.0.0.0/8"; got != want {
		t.Fatalf("expected trusted proxies %q, got %q", want, got)
	}
	if cfg.CookieSameSite != "lax" {
		t.Fatalf("expected same-site lax, got %q", cfg.CookieSameSite)
	}
	if got, want := strings.Join(cfg.CORSAllowedOrigins, ","), "http://localhost:5173,http://127.0.0.1:5173"; got != want {
		t.Fatalf("expected CORS origins %q, got %q", want, got)
	}
}

func TestLoadRejectsInvalidTrustedProxy(t *testing.T) {
	t.Setenv("APP_ENV", "development")
	t.Setenv("TRUSTED_PROXIES", "not-a-proxy")

	_, err := Load()
	if err == nil {
		t.Fatal("expected Load to fail for invalid trusted proxy")
	}
	if !strings.Contains(err.Error(), "TRUSTED_PROXIES contains invalid IP/CIDR") {
		t.Fatalf("expected trusted proxy validation error, got %q", err.Error())
	}
}

func setValidProductionEnv(t *testing.T) {
	t.Helper()

	t.Setenv("APP_ENV", "production")
	t.Setenv("JWT_SECRET", "0123456789abcdefghijklmnopqrstuvwxyz0123456789")
	t.Setenv("COOKIE_SECURE", "true")
	t.Setenv("COOKIE_SAME_SITE", "strict")
	t.Setenv("COOKIE_DOMAIN", "")
	t.Setenv("CORS_ALLOWED_ORIGINS", "https://app.example.com")
	t.Setenv("APP_PUBLIC_URL", "https://app.example.com")
	t.Setenv("TRUSTED_PROXIES", "10.0.0.10")
	t.Setenv("DATABASE_URL", "host=postgres user=app password=secret dbname=go_infra_link port=5432 sslmode=require")
	t.Setenv("SEED_USER_ENABLED", "false")
	t.Setenv("SEED_DUMMY_NOTIFICATIONS", "false")
}
