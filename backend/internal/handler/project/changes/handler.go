package changes

import (
	"context"
	"net/http"

	domainProject "github.com/besart951/go_infra_link/backend/internal/domain/project"
	dto "github.com/besart951/go_infra_link/backend/internal/handler/dto/project"
	projectshared "github.com/besart951/go_infra_link/backend/internal/handler/project/shared"
	"github.com/besart951/go_infra_link/backend/internal/handlerutil"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Service interface {
	ListAfter(ctx context.Context, projectID uuid.UUID, afterRevision uint64, limit int) (*domainProject.ChangePage, error)
}

type Handler struct {
	access  projectshared.AccessPolicyService
	service Service
}

func NewHandler(access projectshared.AccessPolicyService, service Service) *Handler {
	return &Handler{access: access, service: service}
}

func (h *Handler) List(c *gin.Context) {
	projectID, ok := handlerutil.ParseUUIDParam(c, "id")
	if !ok {
		return
	}
	if !projectshared.EnsureProjectAccess(c, h.access, projectID) {
		return
	}

	var query dto.ListProjectChangesQuery
	if !handlerutil.BindQuery(c, &query) {
		return
	}
	page, err := h.service.ListAfter(c.Request.Context(), projectID, query.AfterRevision, query.Limit)
	if err != nil {
		handlerutil.RespondLocalizedError(c, http.StatusInternalServerError, "fetch_failed", "project.fetch_failed")
		return
	}
	c.JSON(http.StatusOK, toResponse(page))
}

func toResponse(page *domainProject.ChangePage) dto.ProjectChangesResponse {
	events := make([]dto.ProjectChangeResponse, len(page.Events))
	for i, event := range page.Events {
		events[i] = dto.ProjectChangeResponse{
			EventID: event.EventID, Revision: event.Revision,
			AggregateType: event.AggregateType, AggregateID: event.AggregateID,
			Action: string(event.Action), ActorID: event.ActorID,
			ChangedFields: event.ChangedFields, ParentRefs: event.ParentRefs,
			OccurredAt: event.OccurredAt,
		}
	}
	return dto.ProjectChangesResponse{
		ProjectID: page.ProjectID, CurrentRevision: page.CurrentRevision,
		Events: events, HasMore: page.HasMore, ResetRequired: page.ResetRequired,
	}
}
