package project

import (
	"context"
	"net/http"
	"time"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	"github.com/besart951/go_infra_link/backend/internal/handler/middleware"
	"github.com/besart951/go_infra_link/backend/internal/handlerutil"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ProjectHandler struct {
	lifecycle     ProjectLifecycleService
	access        ProjectAccessPolicyService
	membership    ProjectMembershipService
	facilityLink  ProjectFacilityLinkService
	events        *ProjectEventHub
	collaboration *ProjectCollaborationHub
}

func NewProjectHandler(lifecycle ProjectLifecycleService, access ProjectAccessPolicyService, membership ProjectMembershipService, facilityLink ProjectFacilityLinkService) *ProjectHandler {
	return &ProjectHandler{
		lifecycle:     lifecycle,
		access:        access,
		membership:    membership,
		facilityLink:  facilityLink,
		events:        NewProjectEventHub(),
		collaboration: NewProjectCollaborationHub(),
	}
}

func (h *ProjectHandler) BroadcastRefreshForControlCabinet(ctx context.Context, controlCabinetID uuid.UUID, scope string) {
	if h.collaboration == nil {
		return
	}

	projectIDs, err := h.facilityLink.ListProjectIDsByControlCabinetID(ctx, controlCabinetID)
	if err != nil {
		return
	}

	for _, projectID := range projectIDs {
		h.collaboration.BroadcastRefreshRequest(projectID, nil, scope, nil)
	}
}

func (h *ProjectHandler) BroadcastRefreshForSPSController(ctx context.Context, spsControllerID uuid.UUID, scope string) {
	if h.collaboration == nil {
		return
	}

	projectIDs, err := h.facilityLink.ListProjectIDsBySPSControllerID(ctx, spsControllerID)
	if err != nil {
		return
	}

	for _, projectID := range projectIDs {
		h.collaboration.BroadcastRefreshRequest(projectID, nil, scope, nil)
	}
}

func (h *ProjectHandler) notifyProjectChange(c *gin.Context, projectID uuid.UUID, eventType string) {
	if h.events == nil {
		return
	}

	var actorID *uuid.UUID
	if userID, ok := middleware.GetUserID(c); ok {
		actorID = &userID
	}

	h.events.Publish(projectID, eventType, actorID)
}

func (h *ProjectHandler) StreamProjectEvents(c *gin.Context) {
	projectID, ok := handlerutil.ParseUUIDParam(c, "id")
	if !ok {
		return
	}

	if !h.ensureProjectAccess(c, projectID) {
		return
	}

	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		handlerutil.RespondLocalizedError(c, http.StatusInternalServerError, "stream_unsupported", "project.fetch_failed")
		return
	}

	events, unsubscribe := h.events.Subscribe(projectID)
	defer unsubscribe()

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")
	c.Status(http.StatusOK)

	readyPayload := map[string]any{
		"type":       "ready",
		"project_id": projectID,
		"at":         time.Now().UTC(),
	}
	if msg, err := formatSSE(projectEventName, readyPayload); err == nil {
		_, _ = c.Writer.WriteString(msg)
		flusher.Flush()
	}

	heartbeat := time.NewTicker(25 * time.Second)
	defer heartbeat.Stop()

	for {
		select {
		case <-c.Request.Context().Done():
			return
		case <-heartbeat.C:
			_, _ = c.Writer.WriteString(": ping\n\n")
			flusher.Flush()
		case event, ok := <-events:
			if !ok {
				return
			}
			msg, err := formatSSE(projectEventName, event)
			if err != nil {
				continue
			}
			_, _ = c.Writer.WriteString(msg)
			flusher.Flush()
		}
	}
}

func (h *ProjectHandler) ensureProjectAccess(c *gin.Context, projectID uuid.UUID) bool {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		handlerutil.RespondLocalizedError(c, http.StatusUnauthorized, "unauthorized", "errors.unauthorized")
		return false
	}

	hasAccess, err := h.access.CanAccessProject(c.Request.Context(), userID, projectID)
	if err != nil {
		handlerutil.RespondDomainError(c, err,
			handlerutil.LocalizedError(http.StatusInternalServerError, "fetch_failed", "project.fetch_failed"),
			handlerutil.MapError(domain.ErrNotFound, handlerutil.LocalizedError(http.StatusNotFound, "not_found", "project.project_not_found")),
		)
		return false
	}

	if !hasAccess {
		handlerutil.RespondLocalizedError(c, http.StatusForbidden, "forbidden", "errors.forbidden")
		return false
	}

	return true
}
