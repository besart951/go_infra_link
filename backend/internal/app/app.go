package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/besart951/go_infra_link/backend/internal/config"
	"github.com/besart951/go_infra_link/backend/internal/db"
	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainUser "github.com/besart951/go_infra_link/backend/internal/domain/user"
	"github.com/besart951/go_infra_link/backend/internal/handler"
	userservice "github.com/besart951/go_infra_link/backend/internal/service/user"
	"github.com/besart951/go_infra_link/backend/internal/wire"
	applogger "github.com/besart951/go_infra_link/backend/pkg/logger"
	"github.com/gin-gonic/gin"
)

func Run() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("config load: %w", err)
	}
	log := applogger.Setup(cfg.AppEnv, cfg.LogLevel)

	gormDB, err := db.Connect(cfg)
	if err != nil {
		log.Error("Failed to connect to database", "err", err)
		return fmt.Errorf("db connect: %w", err)
	}

	// Get underlying sql.DB for existing repositories
	sqlDB, err := gormDB.DB()
	if err != nil {
		log.Error("Failed to get sql.DB instance", "err", err)
		return fmt.Errorf("db instance: %w", err)
	}
	defer func() {
		if err := sqlDB.Close(); err != nil {
			log.Error("Failed to close database", "err", err)
		}
	}()

	// Initialize dependencies via wire package
	repos, err := wire.NewRepositories(sqlDB, cfg.DBType)
	if err != nil {
		log.Error("Failed to initialize repositories", "err", err)
		return fmt.Errorf("repositories: %w", err)
	}

	services := wire.NewServices(repos, wire.ServiceConfig{
		JWTSecret:       cfg.JWTSecret,
		Issuer:          config.DefaultIssuer,
		AccessTokenTTL:  cfg.AccessTokenTTL,
		RefreshTokenTTL: cfg.RefreshTokenTTL,
	})

	if !config.IsProduction(cfg.AppEnv) {
		if err := ensureSeedUser(cfg, log, services.User, repos.UserEmail); err != nil {
			log.Error("Failed seeding initial user", "err", err)
			return fmt.Errorf("seed user: %w", err)
		}
	}

	cookieSecure := cfg.CookieSecure
	if cfg.AppEnv == "production" {
		cookieSecure = true
	}
	cookieSettings := handler.CookieSettings{
		Domain:   cfg.CookieDomain,
		Secure:   cookieSecure,
		SameSite: http.SameSiteStrictMode,
	}

	handlers := wire.NewHandlers(services, cookieSettings, wire.DevAuthConfig{
		Enabled:         cfg.DevAuthEnabled,
		Email:           cfg.DevAuthEmail,
		Password:        cfg.DevAuthPassword,
		AccessTokenTTL:  cfg.AccessTokenTTL,
		RefreshTokenTTL: cfg.RefreshTokenTTL,
	})

	// Setup Gin router
	if config.IsProduction(cfg.AppEnv) {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	// Register all routes
	handler.RegisterRoutes(router, handlers, services.JWT, services.RBAC, services.User)

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	httpAddr := cfg.HTTPAddr
	log.Info("Starting server", "addr", httpAddr, "env", cfg.AppEnv)

	srv := &http.Server{
		Addr:              httpAddr,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	serverErr := make(chan error, 1)
	go func() {
		serverErr <- srv.ListenAndServe()
	}()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	select {
	case err := <-serverErr:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("Server stopped unexpectedly", "err", err)
			return fmt.Errorf("server listen: %w", err)
		}
		return nil
	case <-ctx.Done():
		log.Info("Shutdown signal received")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error("Graceful shutdown failed", "err", err)
		_ = srv.Close()
		return fmt.Errorf("server shutdown: %w", err)
	}

	if err := <-serverErr; err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Error("Server stopped with error", "err", err)
		return fmt.Errorf("server stop: %w", err)
	}

	log.Info("Server stopped")
	return nil
}

func ensureSeedUser(cfg config.Config, log applogger.Logger, userService *userservice.Service, userEmailRepo domainUser.UserEmailRepository) error {
	if !cfg.SeedUserEnabled {
		return nil
	}

	if cfg.SeedUserEmail == "" || cfg.SeedUserPassword == "" {
		return nil
	}
	_, err := userEmailRepo.GetByEmail(cfg.SeedUserEmail)
	if err == nil {
		// User already exists
		return nil
	}
	if !errors.Is(err, domain.ErrNotFound) {
		return err
	}

	usr := &domainUser.User{
		FirstName: cfg.SeedUserFirstName,
		LastName:  cfg.SeedUserLastName,
		Email:     cfg.SeedUserEmail,
		IsActive:  true,
	}

	if err := userService.CreateWithPassword(usr, cfg.SeedUserPassword); err != nil {
		return err
	}

	log.Info("Seeded initial user", "email", cfg.SeedUserEmail)
	return nil
}
