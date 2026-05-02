package app

import (
	"fmt"
	"net/http"

	"github.com/besart951/go_infra_link/backend/internal/config"
	"github.com/besart951/go_infra_link/backend/internal/db"
	"github.com/besart951/go_infra_link/backend/internal/handler"
	authhandler "github.com/besart951/go_infra_link/backend/internal/handler/auth"
	"github.com/besart951/go_infra_link/backend/internal/wire"
	"github.com/besart951/go_infra_link/backend/pkg/i18n"
	applogger "github.com/besart951/go_infra_link/backend/pkg/logger"
	"gorm.io/gorm"
)

type runtime struct {
	cfg        config.Config
	log        applogger.Logger
	services   *wire.Services
	handlers   *handler.Handlers
	translator *i18n.Translator
}

func bootstrapRuntime(cfg config.Config, log applogger.Logger) (*runtime, func(), error) {
	logDatabaseConfig(cfg, log)

	gormDB, cleanup, err := openDatabase(cfg.DBConfig, log)
	if err != nil {
		return nil, cleanup, err
	}

	repos, err := wire.NewRepositories(gormDB)
	if err != nil {
		log.Error("Failed to initialize repositories", "err", err)
		cleanup()
		return nil, func() {}, fmt.Errorf("repositories: %w", err)
	}

	runtimeAdapters := wire.NewRuntimeAdapters()
	services, err := wire.NewServices(gormDB, repos, wire.ServiceConfig{
		JWTSecret:       cfg.JWTSecret,
		Issuer:          config.DefaultIssuer,
		AccessTokenTTL:  cfg.AccessTokenTTL,
		RefreshTokenTTL: cfg.RefreshTokenTTL,
		Runtime:         runtimeAdapters,
	})
	if err != nil {
		log.Error("Failed to initialize services", "err", err)
		cleanup()
		return nil, func() {}, fmt.Errorf("services: %w", err)
	}

	if err := ensureSeedUser(cfg, log, services.User, repos.UserEmail); err != nil {
		log.Error("Failed seeding initial user", "err", err)
		cleanup()
		return nil, func() {}, fmt.Errorf("seed user: %w", err)
	}
	if err := ensureSeedSystemNotifications(cfg, log, repos.UserEmail, repos.SystemNotifications); err != nil {
		log.Error("Failed seeding dummy notifications", "err", err)
		cleanup()
		return nil, func() {}, fmt.Errorf("seed notifications: %w", err)
	}

	translator, loader := initializeTranslator(log)
	handlers := wire.NewHandlers(
		services,
		runtimeAdapters,
		cookieSettingsFromConfig(cfg),
		loader,
		cfg.AccessTokenTTL,
		cfg.RefreshTokenTTL,
	)

	return &runtime{
		cfg:        cfg,
		log:        log,
		services:   services,
		handlers:   handlers,
		translator: translator,
	}, cleanup, nil
}

func openDatabase(cfg config.DBConfig, log applogger.Logger) (*gorm.DB, func(), error) {
	gormDB, err := db.Connect(cfg)
	if err != nil {
		log.Error("Failed to connect to database", "err", err)
		return nil, func() {}, fmt.Errorf("db connect: %w", err)
	}

	sqlDB, err := gormDB.DB()
	if err != nil {
		log.Error("Failed to get sql.DB instance", "err", err)
		return nil, func() {}, fmt.Errorf("db instance: %w", err)
	}

	cleanup := func() {
		if err := sqlDB.Close(); err != nil {
			log.Error("Failed to close database", "err", err)
		}
	}

	return gormDB, cleanup, nil
}

func logDatabaseConfig(cfg config.Config, log applogger.Logger) {
	if config.IsProduction(cfg.AppEnv) {
		return
	}

	log.Info("Database config", "type", cfg.DBConfig.Type, "dsn", formatDSNForLog(cfg.DBConfig.Type, cfg.DBConfig.Dsn))
}

func cookieSettingsFromConfig(cfg config.Config) authhandler.CookieSettings {
	cookieSecure := cfg.CookieSecure
	if config.IsProduction(cfg.AppEnv) {
		cookieSecure = true
	}

	return authhandler.CookieSettings{
		Domain:   cfg.CookieDomain,
		Secure:   cookieSecure,
		SameSite: http.SameSiteStrictMode,
	}
}
