package app

import (
	"fmt"
	"net/http"

	"github.com/besart951/go_infra_link/backend/internal/config"
	"github.com/besart951/go_infra_link/backend/internal/db"
	domainUser "github.com/besart951/go_infra_link/backend/internal/domain/user"
	"github.com/besart951/go_infra_link/backend/internal/handler"
	"github.com/besart951/go_infra_link/backend/internal/repository/auth"
	projectrepo "github.com/besart951/go_infra_link/backend/internal/repository/project"
	userrepo "github.com/besart951/go_infra_link/backend/internal/repository/user"
	authservice "github.com/besart951/go_infra_link/backend/internal/service/auth"
	passwordsvc "github.com/besart951/go_infra_link/backend/internal/service/password"
	projectservice "github.com/besart951/go_infra_link/backend/internal/service/project"
	userservice "github.com/besart951/go_infra_link/backend/internal/service/user"
	applogger "github.com/besart951/go_infra_link/backend/pkg/logger"
	"github.com/gin-gonic/gin"
)

func Run() error {
	cfg := config.Load()
	log := applogger.Setup(cfg.AppEnv, cfg.LogLevel)

	database, err := db.Open(cfg)
	if err != nil {
		log.Error("Failed to connect to database", "err", err)
		return fmt.Errorf("db open: %w", err)
	}

	log.Info("Migrating database...")
	if err := db.Migrate(database); err != nil {
		log.Error("Database migration failed", "err", err)
		return fmt.Errorf("db migrate: %w", err)
	}

	// Initialize repositories
	projRepo := projectrepo.NewProjectRepository(database)
	usrRepo := userrepo.NewUserRepository(database)
	refreshTokenRepo := auth.NewRefreshTokenRepository(database)
	userEmailRepo, ok := usrRepo.(domainUser.UserEmailRepository)
	if !ok {
		log.Error("User repository does not implement email lookup")
		return fmt.Errorf("user repository missing GetByEmail")
	}

	// Initialize services
	projService := projectservice.New(projRepo)
	passwordService := passwordsvc.New()
	usrService := userservice.NewWithPasswordService(usrRepo, passwordService)
	jwtService := authservice.NewJWTService(cfg.JWTSecret, "go_infra_link")
	authService := authservice.NewService(
		jwtService,
		usrRepo,
		userEmailRepo,
		refreshTokenRepo,
		passwordService,
		cfg.AccessTokenTTL,
		cfg.RefreshTokenTTL,
		"go_infra_link",
	)

	cookieSecure := cfg.CookieSecure
	if cfg.AppEnv == "production" {
		cookieSecure = true
	}
	cookieSettings := handler.CookieSettings{
		Domain:   cfg.CookieDomain,
		Secure:   cookieSecure,
		SameSite: http.SameSiteStrictMode,
	}

	// Initialize handlers
	handlers := &handler.Handlers{
		ProjectHandler: handler.NewProjectHandler(projService),
		UserHandler:    handler.NewUserHandler(usrService),
		AuthHandler:    handler.NewAuthHandler(authService, usrService, cfg.AccessTokenTTL, cfg.RefreshTokenTTL, cookieSettings, cfg.DevAuthEnabled, cfg.DevAuthEmail, cfg.DevAuthPassword),
	}

	// Setup Gin router
	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	// Register all routes
	handler.RegisterRoutes(router, handlers, jwtService)

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	log.Info("Starting server on :8080...")
	if err := router.Run(":8080"); err != nil {
		log.Error("Failed to start server", "err", err)
		return fmt.Errorf("server start: %w", err)
	}

	return nil
}
