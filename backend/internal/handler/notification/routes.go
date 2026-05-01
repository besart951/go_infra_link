package notification

import (
	domainUser "github.com/besart951/go_infra_link/backend/internal/domain/user"
	"github.com/besart951/go_infra_link/backend/internal/handler/middleware"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(protectedV1 *gin.RouterGroup, handler *NotificationSettingsHandler, authChecker middleware.AuthorizationChecker) {
	accountNotifications := protectedV1.Group("/account/notifications")
	{
		accountNotifications.GET("", handler.ListSystemNotifications)
		accountNotifications.POST("/read-all", handler.MarkAllSystemNotificationsRead)
		accountNotifications.POST("/:id/read", handler.MarkSystemNotificationRead)
		accountNotifications.POST("/:id/read-toggle", handler.ToggleSystemNotificationRead)
		accountNotifications.POST("/:id/important", handler.ToggleSystemNotificationImportant)
		accountNotifications.DELETE("/:id", handler.DeleteSystemNotification)
		accountNotifications.GET("/preferences", handler.GetUserPreference)
		accountNotifications.PUT("/preferences", handler.UpsertUserPreference)
		accountNotifications.POST("/preferences/email-verification", handler.SendUserPreferenceVerificationCode)
		accountNotifications.POST("/preferences/email-verification/verify", handler.VerifyUserPreferenceEmail)
	}

	notificationsAdmin := protectedV1.Group("/admin/notifications")
	notificationsAdmin.Use(middleware.RequirePermission(authChecker, domainUser.PermissionNotificationSMTPManage))
	{
		notificationsAdmin.GET("/smtp", handler.GetSMTPSettings)
		notificationsAdmin.PUT("/smtp", handler.UpsertSMTPSettings)
		notificationsAdmin.POST("/smtp/test", handler.SendSMTPTestEmail)
		notificationsAdmin.GET("/rules", handler.ListNotificationRules)
		notificationsAdmin.POST("/rules", handler.CreateNotificationRule)
		notificationsAdmin.PUT("/rules/:id", handler.UpdateNotificationRule)
		notificationsAdmin.DELETE("/rules/:id", handler.DeleteNotificationRule)
	}
}
