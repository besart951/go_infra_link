package history

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainHistory "github.com/besart951/go_infra_link/backend/internal/domain/history"
	dto "github.com/besart951/go_infra_link/backend/internal/handler/dto/history"
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

// ListTimeline godoc
// @Summary List global audit activities
// @Description Returns authoritative audit events with their actual before/after diff. Multiple action and field parameters are combined as OR filters within their category.
// @Tags history
// @Produce json
// @Param scope_type query string false "Scope type"
// @Param scope_id query string false "Scope UUID"
// @Param entity_table query string false "Entity table"
// @Param entity_id query string false "Entity UUID"
// @Param actor_id query string false "Actor UUID"
// @Param occurred_from query string false "Earliest ISO-8601 timestamp"
// @Param occurred_to query string false "Latest ISO-8601 timestamp"
// @Param action query []string false "Actions: create, update, delete, restore" collectionFormat(multi)
// @Param field query []string false "Changed field names" collectionFormat(multi)
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(50)
// @Success 200 {object} dto.TimelineResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/history/timeline [get]
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
	c.JSON(http.StatusOK, dto.TimelineResponseFrom(result))
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
	if c.Request.ContentLength > 0 && !handlerutil.BindJSON(c, &req) {
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

// ListProjectTimeline godoc
// @Summary List project audit activities
// @Description Returns authoritative audit events for one project. The optional entity, field, action, actor and date filters work like the global timeline.
// @Tags history, projects
// @Produce json
// @Param id path string true "Project UUID"
// @Param scope_type query string false "Secondary scope type"
// @Param scope_id query string false "Secondary scope UUID"
// @Param entity_table query string false "Entity table"
// @Param entity_id query string false "Entity UUID"
// @Param actor_id query string false "Actor UUID"
// @Param occurred_from query string false "Earliest ISO-8601 timestamp"
// @Param occurred_to query string false "Latest ISO-8601 timestamp"
// @Param action query []string false "Actions: create, update, delete, restore" collectionFormat(multi)
// @Param field query []string false "Changed field names" collectionFormat(multi)
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(50)
// @Success 200 {object} dto.TimelineResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/projects/{id}/history/timeline [get]
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
	c.JSON(http.StatusOK, dto.TimelineResponseFrom(result))
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
	if actorID := c.Query("actor_id"); actorID != "" {
		id, err := uuid.Parse(actorID)
		if err != nil {
			handlerutil.RespondLocalizedError(c, http.StatusBadRequest, "invalid_actor_id", "validation.invalid_uuid_format")
			return filter, false
		}
		filter.ActorID = id
	}
	if occurredFrom := c.Query("occurred_from"); occurredFrom != "" {
		parsed, err := parseTimelineTime(occurredFrom)
		if err != nil {
			handlerutil.RespondLocalizedError(c, http.StatusBadRequest, "invalid_occurred_from", "validation.invalid_request")
			return filter, false
		}
		filter.OccurredFrom = &parsed
	}
	if occurredTo := c.Query("occurred_to"); occurredTo != "" {
		parsed, err := parseTimelineTime(occurredTo)
		if err != nil {
			handlerutil.RespondLocalizedError(c, http.StatusBadRequest, "invalid_occurred_to", "validation.invalid_request")
			return filter, false
		}
		filter.OccurredTo = &parsed
	}
	actions, ok := parseTimelineActions(c)
	if !ok {
		return filter, false
	}
	filter.Actions = actions
	filter.Fields = parseStringListQuery(c, "field", 40)
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

func parseTimelineActions(c *gin.Context) ([]domainHistory.Action, bool) {
	values := parseStringListQuery(c, "action", 8)
	actions := make([]domainHistory.Action, 0, len(values))
	for _, value := range values {
		action := domainHistory.Action(value)
		if !domainHistory.IsAction(action) {
			handlerutil.RespondLocalizedError(c, http.StatusBadRequest, "invalid_action", "validation.invalid_request")
			return nil, false
		}
		actions = append(actions, action)
	}
	return actions, true
}

func parseControlCabinetRestoreRequest(c *gin.Context) (domainHistory.RestoreControlCabinetRequest, bool) {
	var req domainHistory.RestoreControlCabinetRequest
	if c.Request.ContentLength > 0 {
		if !handlerutil.BindJSON(c, &req) {
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

func parseTimelineTime(raw string) (time.Time, error) {
	if parsed, err := time.Parse(time.RFC3339Nano, raw); err == nil {
		return parsed.UTC(), nil
	}
	parsed, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		return time.Time{}, err
	}
	return parsed.UTC(), nil
}

func parseStringListQuery(c *gin.Context, key string, maxItems int) []string {
	values := c.QueryArray(key)
	if raw := c.Query(key); raw != "" && len(values) == 0 {
		values = []string{raw}
	}

	seen := map[string]struct{}{}
	out := make([]string, 0, len(values))
	for _, value := range values {
		for _, part := range strings.Split(value, ",") {
			trimmed := strings.TrimSpace(part)
			if trimmed == "" {
				continue
			}
			if len(trimmed) > 96 {
				continue
			}
			if _, ok := seen[trimmed]; ok {
				continue
			}
			seen[trimmed] = struct{}{}
			out = append(out, trimmed)
			if maxItems > 0 && len(out) >= maxItems {
				return out
			}
		}
	}
	return out
}
