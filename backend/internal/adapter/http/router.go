package http

import (
	"github.com/besart951/go_infra_link/backend/internal/adapter/http/handler"
	"github.com/gin-gonic/gin"
)

func NewRouter(userHandler *handler.UserHandler, projectHandler *handler.ProjectHandler, objectHandler *handler.ObjectHandler) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())

	r.GET("/health", func(c *gin.Context) {
		c.Status(200)
	})

	users := r.Group("/users")
	users.POST("/", userHandler.Create)
	users.GET("/:id", userHandler.GetByID)

	projects := r.Group("/projects")
	projects.POST("/", projectHandler.Create)
	projects.GET("/:id", projectHandler.GetByID)
	projects.POST("/:id/members", projectHandler.AddMember)

	objects := r.Group("/objects")
	objects.POST("/", objectHandler.Create)
	objects.GET("/:id", objectHandler.GetByID)
	objects.POST("/:id/permissions", objectHandler.GrantAccess)

	return r
}
