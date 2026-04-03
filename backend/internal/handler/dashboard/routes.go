package dashboard

import "github.com/gin-gonic/gin"

func RegisterRoutes(protectedV1 *gin.RouterGroup, handler *DashboardHandler) {
	protectedV1.GET("/dashboard", handler.GetDashboard)
}
