package handler

import (
	domainAuth "github.com/besart951/go_infra_link/backend/internal/domain/auth"
	authhandler "github.com/besart951/go_infra_link/backend/internal/handler/auth"
	dashboardhandler "github.com/besart951/go_infra_link/backend/internal/handler/dashboard"
	facilityhandler "github.com/besart951/go_infra_link/backend/internal/handler/facility"
	i18nhandler "github.com/besart951/go_infra_link/backend/internal/handler/i18n"
	"github.com/besart951/go_infra_link/backend/internal/handler/middleware"
	notificationhandler "github.com/besart951/go_infra_link/backend/internal/handler/notification"
	projecthandler "github.com/besart951/go_infra_link/backend/internal/handler/project"
	teamhandler "github.com/besart951/go_infra_link/backend/internal/handler/team"
	userhandler "github.com/besart951/go_infra_link/backend/internal/handler/user"
	"github.com/gin-gonic/gin"
)

// RegisterRoutes registers all API routes.
func RegisterRoutes(r *gin.Engine, handlers *Handlers, tokenValidator domainAuth.TokenValidator, authChecker middleware.AuthorizationChecker, userStatusSvc middleware.UserStatusService) {
	publicV1 := r.Group("/api/v1")
	i18nhandler.RegisterRoutes(publicV1, handlers.I18n)
	authhandler.RegisterPublicRoutes(publicV1, handlers.Auth)

	protectedV1 := r.Group("/api/v1")
	protectedV1.Use(middleware.AuthGuard(tokenValidator))
	protectedV1.Use(middleware.AccountStatusGuard(userStatusSvc))
	protectedV1.Use(middleware.CSRFMiddleware())

	dashboardhandler.RegisterRoutes(protectedV1, handlers.Dashboard)
	projecthandler.RegisterProjectRoutes(protectedV1, handlers.Project)
	projecthandler.RegisterPhaseRoutes(protectedV1, handlers.Project, authChecker)
	userhandler.RegisterUserRoutes(protectedV1, handlers.User, authChecker)
	userhandler.RegisterRoleRoutes(protectedV1, handlers.User, authChecker)
	teamhandler.RegisterRoutes(protectedV1, handlers.Team, authChecker)
	userhandler.RegisterAdminRoutes(protectedV1, handlers.User, authChecker)
	notificationhandler.RegisterRoutes(protectedV1, handlers.Notification, authChecker)
	authhandler.RegisterProtectedRoutes(protectedV1, handlers.Auth)
	facilityhandler.RegisterRoutes(protectedV1, handlers.Facility)
}
