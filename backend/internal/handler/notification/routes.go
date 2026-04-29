package notification

import (
	domainUser "github.com/besart951/go_infra_link/backend/internal/domain/user"
	"github.com/besart951/go_infra_link/backend/internal/handler/middleware"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(protectedV1 *gin.RouterGroup, handler *NotificationSettingsHandler, authChecker middleware.AuthorizationChecker) {
	notificationsAdmin := protectedV1.Group("/admin/notifications")
	notificationsAdmin.Use(middleware.RequirePermission(authChecker, domainUser.PermissionNotificationSMTPManage))
	{
		notificationsAdmin.GET("/smtp", handler.GetSMTPSettings)
		notificationsAdmin.PUT("/smtp", handler.UpsertSMTPSettings)
		notificationsAdmin.POST("/smtp/test", handler.SendSMTPTestEmail)
	}
}
