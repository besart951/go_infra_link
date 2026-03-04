package handler

import (
	"net/http"

	"github.com/besart951/go_infra_link/backend/internal/handler/middleware"
	"github.com/besart951/go_infra_link/backend/internal/handlerutil"
	"github.com/gin-gonic/gin"
)

type DashboardHandler struct {
	service DashboardService
}

func NewDashboardHandler(service DashboardService) *DashboardHandler {
	return &DashboardHandler{service: service}
}

func (h *DashboardHandler) GetDashboard(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		handlerutil.RespondLocalizedError(c, http.StatusUnauthorized, "unauthorized", "errors.unauthorized")
		return
	}

	dashboard, err := h.service.GetUserDashboard(userID)
	if err != nil {
		handlerutil.RespondLocalizedError(c, http.StatusInternalServerError, "fetch_failed", "errors.fetch_failed")
		return
	}

	c.JSON(http.StatusOK, dashboard)
}
