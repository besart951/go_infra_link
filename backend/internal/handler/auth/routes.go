package auth

import (
	"github.com/besart951/go_infra_link/backend/internal/handler/middleware"
	"github.com/gin-gonic/gin"
)

func RegisterPublicRoutes(publicV1 *gin.RouterGroup, handler *AuthHandler, registrationHandler *RegistrationHandler) {
	publicAuth := publicV1.Group("/auth")
	{
		publicAuth.GET("/session", handler.Session)
		publicAuth.POST("/login", middleware.LoginRateLimitMiddleware(), handler.Login)
	}

	registrations := publicV1.Group("/auth/registrations")
	registrations.Use(middleware.RegistrationRateLimitMiddleware())
	{
		registrations.GET("/:token", registrationsHandler(registrationHandler).GetRegistration)
		registrations.POST("/:token/complete", registrationsHandler(registrationHandler).CompleteRegistration)
	}

	authCsrf := publicV1.Group("/auth")
	authCsrf.Use(middleware.AuthSensitiveRateLimitMiddleware())
	authCsrf.Use(middleware.CSRFMiddleware())
	{
		authCsrf.POST("/refresh", handler.Refresh)
		authCsrf.POST("/logout", handler.Logout)
	}
}

func registrationsHandler(handler *RegistrationHandler) *RegistrationHandler {
	if handler == nil {
		return NewRegistrationHandler(nil)
	}
	return handler
}

func RegisterProtectedRoutes(protectedV1 *gin.RouterGroup, handler *AuthHandler) {
	authProtected := protectedV1.Group("/auth")
	{
		authProtected.GET("/me", handler.Me)
	}
}
