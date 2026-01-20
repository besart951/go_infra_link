package app

import (
	"fmt"

	"github.com/besart951/go_infra_link/backend/internal/config"
	"github.com/besart951/go_infra_link/backend/internal/db"
	"github.com/besart951/go_infra_link/backend/internal/handler"
	projectrepo "github.com/besart951/go_infra_link/backend/internal/repository/project"
	userrepo "github.com/besart951/go_infra_link/backend/internal/repository/user"
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

	// Initialize services
	projService := projectservice.New(projRepo)
	usrService := userservice.New(usrRepo)

	// Initialize handlers
	handlers := &handler.Handlers{
		ProjectHandler: handler.NewProjectHandler(projService),
		UserHandler:    handler.NewUserHandler(usrService),
	}

	// Setup Gin router
	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	// Register all routes
	handler.RegisterRoutes(router, handlers)

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
