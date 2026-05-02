package history

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainHistory "github.com/besart951/go_infra_link/backend/internal/domain/history"
	"github.com/besart951/go_infra_link/backend/internal/handlerutil"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Service interface {
	ListTimeline(ctx context.Context, filter domainHistory.TimelineFilter) (*domain.PaginatedList[domainHistory.ChangeEvent], error)
	GetEvent(ctx context.Context, id uuid.UUID) (*domainHistory.ChangeEvent, error)
	RestoreEntityToEvent(ctx context.Context, eventID uuid.UUID, mode domainHistory.RestoreMode) (*domainHistory.RestoreResult, error)
	RestoreControlCabinet(ctx context.Context, controlCabinetID uuid.UUID, req domainHistory.RestoreControlCabinetRequest) (*domainHistory.RestoreResult, error)
}

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) ListTimeline(c *gin.Context) {
	filter, ok := parseTimelineFilter(c)
	if !ok {
		return
	}
	result, err := h.service.ListTimeline(c.Request.Context(), filter)
	if err != nil {
		handlerutil.RespondDomainError(c, err,
			handlerutil.LocalizedError(http.StatusInternalServerError, "history_fetch_failed", "facility.fetch_failed"),
		)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *Handler) GetEvent(c *gin.Context) {
	id, ok := handlerutil.ParseUUIDParam(c, "id")
	if !ok {
		return
	}
	event, err := h.service.GetEvent(c.Request.Context(), id)
	if err != nil {
		handlerutil.RespondDomainError(c, err,
			handlerutil.LocalizedError(http.StatusInternalServerError, "history_fetch_failed", "facility.fetch_failed"),
			handlerutil.MapError(domain.ErrNotFound, handlerutil.LocalizedError(http.StatusNotFound, "not_found", "errors.not_found")),
		)
		return
	}
	c.JSON(http.StatusOK, event)
}

func (h *Handler) RestoreEntity(c *gin.Context) {
	eventID, ok := handlerutil.ParseUUIDParam(c, "id")
	if !ok {
		return
	}
	var req domainHistory.RestoreEntityRequest
	if err := c.ShouldBindJSON(&req); err != nil && c.Request.ContentLength > 0 {
		handlerutil.RespondLocalizedError(c, http.StatusBadRequest, "invalid_request", "validation.invalid_request")
		return
	}
	mode := req.Mode
	if mode == "" {
		mode = domainHistory.RestoreModeAfter
	}
	result, err := h.service.RestoreEntityToEvent(c.Request.Context(), eventID, mode)
	if err != nil {
		handlerutil.RespondDomainError(c, err,
			handlerutil.LocalizedError(http.StatusInternalServerError, "restore_failed", "facility.update_failed"),
			handlerutil.MapError(domain.ErrNotFound, handlerutil.LocalizedError(http.StatusNotFound, "not_found", "errors.not_found")),
		)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *Handler) RestoreControlCabinet(c *gin.Context) {
	controlCabinetID, ok := handlerutil.ParseUUIDParam(c, "id")
	if !ok {
		return
	}
	req, ok := parseControlCabinetRestoreRequest(c)
	if !ok {
		return
	}
	result, err := h.service.RestoreControlCabinet(c.Request.Context(), controlCabinetID, req)
	if err != nil {
		handlerutil.RespondDomainError(c, err,
			handlerutil.LocalizedError(http.StatusInternalServerError, "restore_failed", "facility.update_failed"),
			handlerutil.MapError(domain.ErrNotFound, handlerutil.LocalizedError(http.StatusNotFound, "not_found", "errors.not_found")),
		)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *Handler) ListProjectTimeline(c *gin.Context) {
	projectID, ok := handlerutil.ParseUUIDParam(c, "id")
	if !ok {
		return
	}
	filter, ok := parseTimelineFilter(c)
	if !ok {
		return
	}
	if filter.ScopeType != "" && filter.ScopeID != uuid.Nil {
		filter.SecondaryScopeType = filter.ScopeType
		filter.SecondaryScopeID = filter.ScopeID
	}
	filter.ScopeType = "project"
	filter.ScopeID = projectID
	result, err := h.service.ListTimeline(c.Request.Context(), filter)
	if err != nil {
		handlerutil.RespondLocalizedError(c, http.StatusInternalServerError, "history_fetch_failed", "facility.fetch_failed")
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *Handler) RestoreProjectControlCabinet(c *gin.Context) {
	projectID, ok := handlerutil.ParseUUIDParam(c, "id")
	if !ok {
		return
	}
	controlCabinetID, ok := handlerutil.ParseUUIDParam(c, "controlCabinetId")
	if !ok {
		return
	}
	req, ok := parseControlCabinetRestoreRequest(c)
	if !ok {
		return
	}
	req.ProjectID = &projectID
	result, err := h.service.RestoreControlCabinet(c.Request.Context(), controlCabinetID, req)
	if err != nil {
		handlerutil.RespondLocalizedError(c, http.StatusInternalServerError, "restore_failed", "facility.update_failed")
		return
	}
	c.JSON(http.StatusOK, result)
}

func parseTimelineFilter(c *gin.Context) (domainHistory.TimelineFilter, bool) {
	filter := domainHistory.TimelineFilter{
		ScopeType:   c.Query("scope_type"),
		EntityTable: c.Query("entity_table"),
	}
	if scopeID := c.Query("scope_id"); scopeID != "" {
		id, err := uuid.Parse(scopeID)
		if err != nil {
			handlerutil.RespondLocalizedError(c, http.StatusBadRequest, "invalid_scope_id", "validation.invalid_uuid_format")
			return filter, false
		}
		filter.ScopeID = id
	}
	if entityID := c.Query("entity_id"); entityID != "" {
		id, err := uuid.Parse(entityID)
		if err != nil {
			handlerutil.RespondLocalizedError(c, http.StatusBadRequest, "invalid_entity_id", "validation.invalid_uuid_format")
			return filter, false
		}
		filter.EntityID = id
	}
	page, ok := parseOptionalIntQuery(c, "page")
	if !ok {
		return filter, false
	}
	if page != nil {
		filter.Page = *page
	}
	limit, ok := parseOptionalIntQuery(c, "limit")
	if !ok {
		return filter, false
	}
	if limit != nil {
		filter.Limit = *limit
	}
	return filter, true
}

func parseControlCabinetRestoreRequest(c *gin.Context) (domainHistory.RestoreControlCabinetRequest, bool) {
	var req domainHistory.RestoreControlCabinetRequest
	if c.Request.ContentLength > 0 {
		if err := c.ShouldBindJSON(&req); err != nil {
			handlerutil.RespondLocalizedError(c, http.StatusBadRequest, "invalid_request", "validation.invalid_request")
			return req, false
		}
	}
	if asOf := c.Query("as_of"); asOf != "" {
		parsed, err := time.Parse(time.RFC3339, asOf)
		if err != nil {
			handlerutil.RespondLocalizedError(c, http.StatusBadRequest, "invalid_as_of", "validation.invalid_request")
			return req, false
		}
		req.AsOf = &parsed
	}
	if eventID := c.Query("event_id"); eventID != "" {
		id, err := uuid.Parse(eventID)
		if err != nil {
			handlerutil.RespondLocalizedError(c, http.StatusBadRequest, "invalid_event_id", "validation.invalid_uuid_format")
			return req, false
		}
		req.EventID = &id
	}
	return req, true
}

func parseOptionalIntQuery(c *gin.Context, key string) (*int, bool) {
	raw := c.Query(key)
	if raw == "" {
		return nil, true
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		handlerutil.RespondLocalizedError(c, http.StatusBadRequest, "invalid_"+key, "validation.numeric")
		return nil, false
	}
	return &value, true
}
