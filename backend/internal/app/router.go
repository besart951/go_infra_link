package app

import (
	"net/http"

	docs "github.com/besart951/go_infra_link/backend/docs"
	"github.com/besart951/go_infra_link/backend/internal/config"
	"github.com/besart951/go_infra_link/backend/internal/handler"
	"github.com/besart951/go_infra_link/backend/internal/handler/middleware"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func newRouter(appRuntime *runtime) *gin.Engine {
	configureGinMode(appRuntime.cfg.AppEnv)

	router := gin.Default()
	registerSwaggerRoute(router, appRuntime.cfg)
	router.Use(middleware.LocaleMiddleware(appRuntime.translator, defaultLocale))

	handler.RegisterRoutes(
		router,
		appRuntime.handlers,
		appRuntime.services.JWT,
		appRuntime.services.RBAC,
		appRuntime.services.User,
	)

	registerHealthRoute(router)
	logRegisteredRoutes(appRuntime.log, appRuntime.cfg, router)
	return router
}

func registerSwaggerRoute(router *gin.Engine, cfg config.Config) {
	if !cfg.SwaggerEnabled {
		return
	}

	docs.SwaggerInfo.BasePath = "/api/v1"
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
}

func configureGinMode(appEnv string) {
	if config.IsProduction(appEnv) {
		gin.SetMode(gin.ReleaseMode)
		return
	}

	gin.DebugPrintRouteFunc = func(httpMethod, absolutePath, handlerName string, numHandlers int) {
		// Suppress per-route debug output in dev; we log a summary instead.
	}
}

func registerHealthRoute(router *gin.Engine) {
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
}

func logRegisteredRoutes(log interface {
	Info(msg string, args ...any)
}, cfg config.Config, router *gin.Engine) {
	if config.IsProduction(cfg.AppEnv) {
		return
	}

	routes := router.Routes()
	log.Info(
		"Routes registered",
		"count",
		len(routes),
		"health",
		localURL(cfg.HTTPAddr, "/health"),
	)

	if cfg.SwaggerEnabled {
		log.Info("Swagger UI enabled", "url", localURL(cfg.HTTPAddr, "/swagger/index.html"))
	}
}
