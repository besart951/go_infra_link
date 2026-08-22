package history

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/besart951/go_infra_link/backend/internal/cursor"
	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainHistory "github.com/besart951/go_infra_link/backend/internal/domain/history"
	facilitydto "github.com/besart951/go_infra_link/backend/internal/handler/dto/facility"
	dto "github.com/besart951/go_infra_link/backend/internal/handler/dto/history"
	"github.com/besart951/go_infra_link/backend/internal/handler/middleware"
	sharedpresenter "github.com/besart951/go_infra_link/backend/internal/handler/presenter/shared"
	"github.com/besart951/go_infra_link/backend/internal/handlerutil"
	facilityservice "github.com/besart951/go_infra_link/backend/internal/service/facility"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Service interface {
	ListTimeline(ctx context.Context, filter domainHistory.TimelineFilter) (*domain.PaginatedList[domainHistory.ChangeEvent], error)
	ListTimelineCursor(ctx context.Context, filter domainHistory.TimelineFilter) (*domainHistory.TimelineCursorPage, error)
	GetEvent(ctx context.Context, id uuid.UUID) (*domainHistory.ChangeEvent, error)
	RestoreEntityToEvent(ctx context.Context, eventID uuid.UUID, mode domainHistory.RestoreMode) (*domainHistory.RestoreResult, error)
	UndoBatch(ctx context.Context, batchID uuid.UUID) (*domainHistory.RestoreResult, error)
	RestoreControlCabinet(ctx context.Context, controlCabinetID uuid.UUID, req domainHistory.RestoreControlCabinetRequest) (*domainHistory.RestoreResult, error)
}

type Handler struct {
	service Service
	jobs    *facilityservice.CopyJobManager
}

func NewHandler(service Service, jobs ...*facilityservice.CopyJobManager) *Handler {
	handler := &Handler{service: service}
	if len(jobs) > 0 {
		handler.jobs = jobs[0]
	}
	return handler
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
// @Param page query int false "Legacy page number" default(1)
// @Param cursor query string false "Opaque keyset cursor; omit page to use cursor pagination"
// @Param limit query int false "Items per page" default(50)
// @Success 200 {object} dto.TimelineCursorResponse
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
	h.listTimeline(c, filter)
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
	h.restoreEntity(c, "")
}

func (h *Handler) UndoEntity(c *gin.Context) {
	h.restoreEntity(c, domainHistory.RestoreModeBefore)
}

func (h *Handler) UndoBatch(c *gin.Context) {
	batchID, ok := handlerutil.ParseUUIDParam(c, "id")
	if !ok {
		return
	}
	result, err := h.service.UndoBatch(c.Request.Context(), batchID)
	if err != nil {
		var conflict *domainHistory.UndoConflictError
		if errors.As(err, &conflict) {
			c.JSON(http.StatusConflict, dto.UndoConflictResponseFrom(conflict.Conflict))
			return
		}
		handlerutil.RespondDomainError(c, err,
			handlerutil.LocalizedError(http.StatusInternalServerError, "undo_failed", "facility.update_failed"),
			handlerutil.MapError(domain.ErrNotFound, handlerutil.LocalizedError(http.StatusNotFound, "not_found", "errors.not_found")),
		)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *Handler) restoreEntity(c *gin.Context, requiredMode domainHistory.RestoreMode) {
	eventID, ok := handlerutil.ParseUUIDParam(c, "id")
	if !ok {
		return
	}
	var req domainHistory.RestoreEntityRequest
	if c.Request.ContentLength > 0 && !handlerutil.BindJSON(c, &req) {
		return
	}
	mode := requiredMode
	if mode == "" {
		mode = req.Mode
	}
	if mode == "" {
		mode = domainHistory.RestoreModeAfter
	}
	result, err := h.service.RestoreEntityToEvent(c.Request.Context(), eventID, mode)
	if err != nil {
		var conflict *domainHistory.UndoConflictError
		if errors.As(err, &conflict) {
			c.JSON(http.StatusConflict, dto.UndoConflictResponseFrom(conflict.Conflict))
			return
		}
		handlerutil.RespondDomainError(c, err,
			handlerutil.LocalizedError(http.StatusInternalServerError, "restore_failed", "facility.update_failed"),
			handlerutil.MapError(domain.ErrNotFound, handlerutil.LocalizedError(http.StatusNotFound, "not_found", "errors.not_found")),
		)
		return
	}
	c.JSON(http.StatusOK, result)
}

// RestoreControlCabinet godoc
// @Summary Restore a control-cabinet hierarchy asynchronously
// @Tags history
// @Accept json
// @Produce json
// @Param id path string true "Control cabinet UUID"
// @Param Idempotency-Key header string false "Stable operation UUID"
// @Param body body domainHistory.RestoreControlCabinetRequest false "Restore checkpoint"
// @Success 202 {object} facilitydto.FacilityJobResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 409 {object} dto.ErrorResponse
// @Router /api/v1/history/control-cabinets/{id}/restore [post]
func (h *Handler) RestoreControlCabinet(c *gin.Context) {
	controlCabinetID, ok := handlerutil.ParseUUIDParam(c, "id")
	if !ok {
		return
	}
	req, ok := parseControlCabinetRestoreRequest(c)
	if !ok {
		return
	}
	h.submitControlCabinetRestore(c, controlCabinetID, req)
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
// @Param page query int false "Legacy page number" default(1)
// @Param cursor query string false "Opaque keyset cursor; omit page to use cursor pagination"
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
	h.listTimeline(c, filter)
}

func (h *Handler) listTimeline(c *gin.Context, filter domainHistory.TimelineFilter) {
	if _, legacy := c.GetQuery("page"); legacy {
		result, err := h.service.ListTimeline(c.Request.Context(), filter)
		if h.respondTimelineError(c, err) {
			return
		}
		c.JSON(http.StatusOK, dto.TimelineResponseFrom(result))
		return
	}
	result, err := h.service.ListTimelineCursor(c.Request.Context(), filter)
	if h.respondTimelineError(c, err) {
		return
	}
	c.JSON(http.StatusOK, dto.TimelineCursorResponseFrom(result))
}

func (h *Handler) respondTimelineError(c *gin.Context, err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, cursor.ErrInvalid) {
		handlerutil.RespondLocalizedError(c, http.StatusBadRequest, "invalid_cursor", "validation.invalid_request")
		return true
	}
	handlerutil.RespondLocalizedError(c, http.StatusInternalServerError, "history_fetch_failed", "facility.fetch_failed")
	return true
}

// RestoreProjectControlCabinet godoc
// @Summary Restore a project-scoped control-cabinet hierarchy asynchronously
// @Tags history, projects
// @Accept json
// @Produce json
// @Param id path string true "Project UUID"
// @Param controlCabinetId path string true "Control cabinet UUID"
// @Param Idempotency-Key header string false "Stable operation UUID"
// @Param body body domainHistory.RestoreControlCabinetRequest false "Restore checkpoint"
// @Success 202 {object} facilitydto.FacilityJobResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 409 {object} dto.ErrorResponse
// @Router /api/v1/projects/{id}/history/control-cabinets/{controlCabinetId}/restore [post]
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
	h.submitControlCabinetRestore(c, controlCabinetID, req)
}

// submitControlCabinetRestore starts the canonical durable hierarchy restore.
func (h *Handler) submitControlCabinetRestore(c *gin.Context, cabinetID uuid.UUID, req domainHistory.RestoreControlCabinetRequest) {
	if h.jobs == nil || !h.jobs.SupportsDurableTasks() {
		handlerutil.RespondLocalizedError(c, http.StatusServiceUnavailable, "durable_jobs_unavailable", "errors.service_unavailable")
		return
	}
	actorID, ok := middleware.GetUserID(c)
	if !ok {
		handlerutil.RespondLocalizedError(c, http.StatusUnauthorized, "unauthorized", "errors.unauthorized")
		return
	}
	jobID, ok := historyOperationID(c)
	if !ok {
		return
	}
	asOf := req.AsOf
	if asOf == nil && (req.EventID == nil || *req.EventID == uuid.Nil) {
		now := time.Now().UTC()
		asOf = &now
	}
	payload, err := json.Marshal(domainHistory.RestoreControlCabinetJobPayload{
		ControlCabinetID: cabinetID, ProjectID: req.ProjectID, AsOf: asOf, EventID: req.EventID,
	})
	if err != nil {
		handlerutil.RespondLocalizedError(c, http.StatusInternalServerError, "restore_failed", "facility.update_failed")
		return
	}
	job, err := h.jobs.SubmitTask(c.Request.Context(), facilityservice.CopyJob{
		ID: jobID, OwnerID: actorID, Kind: facilityservice.CopyJobKindControlCabinet,
		Class: facilityservice.FacilityJobClassMutation, Type: facilityservice.FacilityJobTypeRestore,
		Task: domainHistory.TaskRestoreControlCabinet, Payload: payload,
		Admission: &facilityservice.FacilityAggregateAdmission{
			ResourceID: cabinetID, State: facilityservice.FacilityAggregateStateRestoreStaging, AllowMissing: true,
		},
	})
	if err != nil {
		status := http.StatusServiceUnavailable
		code := "service_unavailable"
		if errors.Is(err, facilityservice.ErrAggregateLocked) || errors.Is(err, facilityservice.ErrFacilityJobLimit) {
			status, code = http.StatusConflict, "aggregate_locked"
		}
		handlerutil.RespondLocalizedError(c, status, code, "facility.aggregate_locked")
		return
	}
	c.JSON(http.StatusAccepted, sharedpresenter.ToCopyJobResponse(job))
}

func historyOperationID(c *gin.Context) (uuid.UUID, bool) {
	raw := strings.TrimSpace(c.GetHeader("Idempotency-Key"))
	if raw == "" {
		raw = strings.TrimSpace(c.GetHeader("X-Copy-Operation-ID"))
	}
	if raw == "" {
		return uuid.New(), true
	}
	id, err := uuid.Parse(raw)
	if err != nil {
		handlerutil.RespondLocalizedError(c, http.StatusBadRequest, "invalid_idempotency_key", "validation.invalid_uuid_format")
		return uuid.Nil, false
	}
	return id, true
}

var _ facilitydto.FacilityJobResponse

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
	filter.Cursor = c.Query("cursor")
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
