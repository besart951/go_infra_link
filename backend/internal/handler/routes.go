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
	// API v1 group
	v1 := r.Group("/api/v1")

	// Project routes
	projects := v1.Group("/projects")
	{
		projects.POST("", handlers.ProjectHandler.CreateProject)
		projects.GET("", handlers.ProjectHandler.ListProjects)
		projects.GET("/:id", handlers.ProjectHandler.GetProject)
		projects.PUT("/:id", handlers.ProjectHandler.UpdateProject)
		projects.DELETE("/:id", handlers.ProjectHandler.DeleteProject)
	}

	// User routes
	users := v1.Group("/users")
	{
		users.POST("", handlers.UserHandler.CreateUser)
		users.GET("", handlers.UserHandler.ListUsers)
		users.GET("/:id", handlers.UserHandler.GetUser)
		users.PUT("/:id", handlers.UserHandler.UpdateUser)
		users.DELETE("/:id", handlers.UserHandler.DeleteUser)
	}

	// Auth routes
	auth := v1.Group("/auth")
	{
		auth.POST("/login", handlers.AuthHandler.Login)
	}

	authProtected := v1.Group("/auth")
	authProtected.Use(middleware.AuthGuard(jwtService))
	{
		authProtected.GET("/me", handlers.AuthHandler.Me)
	}

	authCsrf := v1.Group("/auth")
	authCsrf.Use(middleware.CSRFMiddleware())
	{
		authCsrf.POST("/refresh", handlers.AuthHandler.Refresh)
		authCsrf.POST("/logout", handlers.AuthHandler.Logout)
	}
}
