package config

import (
	"strings"
	"testing"
	"time"
)

func TestLoadProductionValidation(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	t.Setenv("JWT_SECRET", "change-me")
	t.Setenv("SEED_USER_ENABLED", "true")
	t.Setenv("SEED_USER_EMAIL", "")
	t.Setenv("SEED_USER_PASSWORD", "password")

	_, err := Load()
	if err == nil {
		t.Fatal("expected Load to fail for invalid production configuration")
	}

	message := err.Error()
	if !strings.Contains(message, "missing JWT_SECRET in production environment") {
		t.Fatalf("expected JWT validation error, got %q", message)
	}
	if !strings.Contains(message, "SEED_USER_EMAIL is required") {
		t.Fatalf("expected seed email validation error, got %q", message)
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
	t.Setenv("BACKEND_PORT", "9090")

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
	if cfg.HTTPAddr != ":9090" {
		t.Fatalf("expected HTTP addr :9090, got %q", cfg.HTTPAddr)
	}
}
