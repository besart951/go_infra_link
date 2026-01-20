package handler

import (
	"github.com/gin-gonic/gin"
)

type Handlers struct {
	ProjectHandler *ProjectHandler
	UserHandler    *UserHandler
}

// RegisterRoutes registers all API routes
func RegisterRoutes(r *gin.Engine, handlers *Handlers) {
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
}
