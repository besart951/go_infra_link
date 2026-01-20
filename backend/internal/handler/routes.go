package handler

import (
	"github.com/besart951/go_infra_link/backend/internal/handler/middleware"
	authsvc "github.com/besart951/go_infra_link/backend/internal/service/auth"
	"github.com/gin-gonic/gin"
)

type Handlers struct {
	ProjectHandler *ProjectHandler
	UserHandler    *UserHandler
	AuthHandler    *AuthHandler
}

// RegisterRoutes registers all API routes
func RegisterRoutes(r *gin.Engine, handlers *Handlers, jwtService authsvc.JWTService) {
	// Public API v1 group (login only)
	publicV1 := r.Group("/api/v1")
	publicAuth := publicV1.Group("/auth")
	{
		publicAuth.POST("/login", handlers.AuthHandler.Login)
		publicAuth.POST("/dev-login", handlers.AuthHandler.DevLogin)
	}

	// CSRF-protected auth endpoints (no access token required)
	authCsrf := publicV1.Group("/auth")
	authCsrf.Use(middleware.CSRFMiddleware())
	{
		authCsrf.POST("/refresh", handlers.AuthHandler.Refresh)
		authCsrf.POST("/logout", handlers.AuthHandler.Logout)
	}

	// Protected API v1 group (all other routes)
	protectedV1 := r.Group("/api/v1")
	protectedV1.Use(middleware.AuthGuard(jwtService))
	protectedV1.Use(middleware.CSRFMiddleware())

	// Project routes
	projects := protectedV1.Group("/projects")
	{
		projects.POST("", handlers.ProjectHandler.CreateProject)
		projects.GET("", handlers.ProjectHandler.ListProjects)
		projects.GET("/:id", handlers.ProjectHandler.GetProject)
		projects.PUT("/:id", handlers.ProjectHandler.UpdateProject)
		projects.DELETE("/:id", handlers.ProjectHandler.DeleteProject)
	}

	// User routes
	users := protectedV1.Group("/users")
	{
		users.POST("", handlers.UserHandler.CreateUser)
		users.GET("", handlers.UserHandler.ListUsers)
		users.GET("/:id", handlers.UserHandler.GetUser)
		users.PUT("/:id", handlers.UserHandler.UpdateUser)
		users.DELETE("/:id", handlers.UserHandler.DeleteUser)
	}

	// Auth routes (protected)
	authProtected := protectedV1.Group("/auth")
	{
		authProtected.GET("/me", handlers.AuthHandler.Me)
	}
}
