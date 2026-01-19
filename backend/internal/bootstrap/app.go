package bootstrap

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/besart951/go_infra_link/backend/internal/infrastructure/config"

	facility "github.com/besart951/go_infra_link/backend/internal/core/domain/facility"

	"github.com/besart951/go_infra_link/backend/internal/core/domain/user"
	userservice "github.com/besart951/go_infra_link/backend/internal/core/service/user"

	"github.com/besart951/go_infra_link/backend/internal/core/domain/project"
	projectservice "github.com/besart951/go_infra_link/backend/internal/core/service/project"

	objectrepo "github.com/besart951/go_infra_link/backend/internal/adapters/storage/sqlite/object"
	projectrepo "github.com/besart951/go_infra_link/backend/internal/adapters/storage/sqlite/repository/project"
	userrepo "github.com/besart951/go_infra_link/backend/internal/adapters/storage/sqlite/repository/user"
)

type Application struct {
	DB         *gorm.DB
	HTTPServer *http.Server
}

func NewApplication() (*Application, error) {
	cfg := config.Load()

	db, err := gorm.Open(sqlite.Open(cfg.DBDsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("database connection failed: %w", err)
	}

	err = db.AutoMigrate(
		&facility.Building{},
		&facility.ControlCabinet{},
		&user.User{},
		&project.Project{},
	)
	if err != nil {
		return nil, fmt.Errorf("migration failed: %w", err)
	}

	buildingRepo := objectrepo.NewBuildingStorage(db)
	cabinetRepo := objectrepo.NewCabinetStorage(db)
	userRepo := userrepo.NewUserStorage(db)
	projectRepo := projectrepo.NewProjectStorage(db)

	_ = objectservice.NewBuildingService(buildingRepo)
	_ = objectservice.NewCabinetService(cabinetRepo)
	_ = userservice.NewUserService(userRepo)
	_ = projectservice.NewProjectService(projectRepo)

	router := gin.Default()

	// Hier würdest du normalerweise Handlers initialisieren:
	// buildingHandler := handlers.NewBuildingHandler(buildingService)
	// buildingHandler.RegisterRoutes(router)

	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "services_loaded": true})
	})

	// 7. HTTP Server konfigurieren
	srv := &http.Server{
		Addr:    cfg.HTTPAddr,
		Handler: router,
	}

	// 8. Die fertige App zurückgeben
	return &Application{
		DB:         db,
		HTTPServer: srv,
	}, nil
}

func (app *Application) Start() error {
	log.Printf("Server is running on %s...", app.HTTPServer.Addr)
	return app.HTTPServer.ListenAndServe()
}

func (app *Application) Shutdown(ctx context.Context) error {
	sqlDB, err := app.DB.DB()
	if err == nil {
		sqlDB.Close()
	}
	return app.HTTPServer.Shutdown(ctx)
}
