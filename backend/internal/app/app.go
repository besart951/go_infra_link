package app

import (
	"fmt"

	"github.com/besart951/go_infra_link/backend/internal/config"
	"github.com/besart951/go_infra_link/backend/internal/db"
	projectrepo "github.com/besart951/go_infra_link/backend/internal/repository/project"
	projectservice "github.com/besart951/go_infra_link/backend/internal/service/project"
	applogger "github.com/besart951/go_infra_link/backend/pkg/logger"
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

	projRepo := projectrepo.NewProjectRepository(database)
	_ = projectservice.New(projRepo)

	log.Info("Server ready to start...")
	return nil
}
