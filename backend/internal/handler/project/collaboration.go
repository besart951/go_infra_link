package project

import (
	"net/http"

	"github.com/besart951/go_infra_link/backend/internal/handler/middleware"
	"github.com/besart951/go_infra_link/backend/internal/handlerutil"
	"github.com/besart951/go_infra_link/backend/internal/infrastructure/realtime"
	"github.com/gin-gonic/gin"
)

type ProjectCollaborationHub = realtime.ProjectCollaborationHub
type ProjectChange = realtime.ProjectChange

func NewProjectCollaborationHub() *ProjectCollaborationHub {
	return realtime.NewProjectCollaborationHub()
}

// StreamProjectCollaboration godoc
// @Summary Stream project collaboration events
// @Description Upgrades the authenticated request to a WebSocket. Messages use the frontend's ProjectSyncInboundMessage contract.
// @Tags projects
// @Param id path string true "Project ID"
// @Success 101
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Router /api/v1/projects/{id}/collaboration [get]
func (h *ProjectHandler) StreamProjectCollaboration(c *gin.Context) {
	projectID, ok := handlerutil.ParseUUIDParam(c, "id")
	if !ok {
		return
	}

	if !h.ensureProjectAccess(c, projectID) {
		return
	}

	userID, ok := middleware.GetUserID(c)
	if !ok {
		handlerutil.RespondLocalizedError(c, http.StatusUnauthorized, "unauthorized", "errors.unauthorized")
		return
	}

	if err := h.collaboration.Stream(c.Writer, c.Request, projectID, userID); err != nil {
		return
	}
}
