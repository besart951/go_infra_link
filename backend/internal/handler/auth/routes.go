package auth

import (
	"github.com/besart951/go_infra_link/backend/internal/handler/middleware"
	"github.com/gin-gonic/gin"
)

func RegisterPublicRoutes(publicV1 *gin.RouterGroup, handler *AuthHandler) {
	publicAuth := publicV1.Group("/auth")
	{
		publicAuth.POST("/login", handler.Login)
	}

	authCsrf := publicV1.Group("/auth")
	authCsrf.Use(middleware.CSRFMiddleware())
	{
		authCsrf.POST("/refresh", handler.Refresh)
		authCsrf.POST("/logout", handler.Logout)
	}
}

func RegisterProtectedRoutes(protectedV1 *gin.RouterGroup, handler *AuthHandler) {
	authProtected := protectedV1.Group("/auth")
	{
		authProtected.GET("/me", handler.Me)
	}
}
