package app

import (
	"fmt"
	"time"

	"github.com/besart951/go_infra_link/backend/internal/config"
	applogger "github.com/besart951/go_infra_link/backend/pkg/logger"
)

func Run() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("config load: %w", err)
	}
	log := applogger.Setup(cfg.AppEnv, cfg.LogLevel)
	runtimeDeps, cleanup, err := bootstrapRuntime(cfg, log)
	if err != nil {
		return err
	}
	defer cleanup()
	stopNotificationWorker := runtimeDeps.services.Notification.StartEmailOutboxWorker(time.Minute, 100)
	defer stopNotificationWorker()

	router := newRouter(runtimeDeps)
	return serveHTTP(runtimeDeps.cfg, runtimeDeps.log, router)
}
