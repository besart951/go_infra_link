package bootstrap

import (
	"context"
	"fmt"

	"github.com/besart951/go_infra_link/backend/internal/adapter/http"
	"github.com/besart951/go_infra_link/backend/internal/adapter/http/handler"
	"github.com/besart951/go_infra_link/backend/internal/adapter/repo"
	"github.com/besart951/go_infra_link/backend/internal/app/object"
	"github.com/besart951/go_infra_link/backend/internal/app/project"
	"github.com/besart951/go_infra_link/backend/internal/app/user"
	"github.com/besart951/go_infra_link/backend/internal/domain"
	"github.com/besart951/go_infra_link/backend/internal/infrastructure/config"
	"github.com/besart951/go_infra_link/backend/internal/infrastructure/database"
	"github.com/besart951/go_infra_link/backend/internal/infrastructure/logger"
	"gorm.io/gorm"
)

type Application struct {
	HTTPServer *http.Server
	DB         *gorm.DB
}

func NewApplication() (*Application, error) {
	cfg := config.Load()

	log := logger.New(cfg.LogLevel)
	log.Info().Str("env", cfg.AppEnv).Msg("app")

	db, err := database.New(cfg.DBDriver, cfg.DBDsn)
	if err != nil {
		return nil, fmt.Errorf("database init: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("database sql handle: %w", err)
	}
	sqlDB.SetMaxOpenConns(cfg.DBMaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.DBMaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.DBConnMaxLifetime)

	if err := db.AutoMigrate(
		&domain.User{},
		&domain.Project{},
		&domain.ProjectMember{},
		&domain.Object{},
		&domain.ObjectPermission{},
	); err != nil {
		return nil, fmt.Errorf("auto migrate: %w", err)
	}

	userRepo := repo.NewUserRepository(db)
	projectRepo := repo.NewProjectRepository(db)
	objectRepo := repo.NewObjectRepository(db)

	userService := user.NewService(userRepo)
	projectService := project.NewService(projectRepo)
	objectService := object.NewService(objectRepo)

	userHandler := handler.NewUserHandler(userService)
	projectHandler := handler.NewProjectHandler(projectService)
	objectHandler := handler.NewObjectHandler(objectService)

	router := http.NewRouter(userHandler, projectHandler, objectHandler)
	server := http.NewServer(cfg.HTTPAddr, router)

	return &Application{
		HTTPServer: server,
		DB:         db,
	}, nil
}

func (a *Application) Shutdown(ctx context.Context) {
	_ = a.HTTPServer.Shutdown(ctx)
	sqlDB, err := a.DB.DB()
	if err == nil {
		_ = sqlDB.Close()
	}
}
