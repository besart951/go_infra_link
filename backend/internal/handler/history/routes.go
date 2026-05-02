package history

import "github.com/gin-gonic/gin"

func RegisterRoutes(protectedV1 *gin.RouterGroup, handler *Handler) {
	if handler == nil {
		return
	}
	history := protectedV1.Group("/history")
	history.GET("/timeline", handler.ListTimeline)
	history.GET("/events/:id", handler.GetEvent)
	history.POST("/events/:id/restore", handler.RestoreEntity)
	history.POST("/control-cabinets/:id/restore", handler.RestoreControlCabinet)

	projects := protectedV1.Group("/projects/:id/history")
	projects.GET("/timeline", handler.ListProjectTimeline)
	projects.POST("/control-cabinets/:controlCabinetId/restore", handler.RestoreProjectControlCabinet)
}
