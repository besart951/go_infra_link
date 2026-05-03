package history

import (
	"github.com/besart951/go_infra_link/backend/internal/domain/user"
	"github.com/besart951/go_infra_link/backend/internal/handler/middleware"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(protectedV1 *gin.RouterGroup, handler *Handler, authChecker middleware.AuthorizationChecker) {
	if handler == nil {
		return
	}
	history := protectedV1.Group("/history")
	history.GET("/timeline", middleware.RequirePermission(authChecker, user.PermissionTimelineRead), handler.ListTimeline)
	history.GET("/events/:id", middleware.RequirePermission(authChecker, user.PermissionTimelineRead), handler.GetEvent)
	history.POST("/events/:id/restore", middleware.RequirePermission(authChecker, user.PermissionTimelineRestore), handler.RestoreEntity)
	history.POST("/control-cabinets/:id/restore", middleware.RequirePermission(authChecker, user.PermissionTimelineRestore), handler.RestoreControlCabinet)

	projects := protectedV1.Group("/projects/:id/history")
	projects.GET("/timeline", middleware.RequirePermission(authChecker, user.PermissionTimelineRead), handler.ListProjectTimeline)
	projects.POST("/control-cabinets/:controlCabinetId/restore", middleware.RequirePermission(authChecker, user.PermissionTimelineRestore), handler.RestoreProjectControlCabinet)
}
