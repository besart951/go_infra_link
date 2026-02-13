package facility

import (
	"net/http"

	"github.com/besart951/go_infra_link/backend/internal/handler/dto"
	"github.com/gin-gonic/gin"
)

type AlarmDefinitionHandler struct {
	service AlarmDefinitionService
}

func NewAlarmDefinitionHandler(service AlarmDefinitionService) *AlarmDefinitionHandler {
	return &AlarmDefinitionHandler{service: service}
}

// CreateAlarmDefinition godoc
// @Summary Create a new alarm definition
// @Tags facility-alarm-definitions
// @Accept json
// @Produce json
// @Param alarm_definition body dto.CreateAlarmDefinitionRequest true "Alarm Definition data"
// @Success 201 {object} dto.AlarmDefinitionResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/facility/alarm-definitions [post]
func (h *AlarmDefinitionHandler) CreateAlarmDefinition(c *gin.Context) {
	var req dto.CreateAlarmDefinitionRequest
	if !bindJSON(c, &req) {
		return
	}

	alarmDef := toAlarmDefinitionModel(req)

	if err := h.service.Create(alarmDef); respondLocalizedValidationOrError(c, err, "facility.creation_failed") {
		return
	}

	c.JSON(http.StatusCreated, toAlarmDefinitionResponse(*alarmDef))
}

// GetAlarmDefinition godoc
// @Summary Get an alarm definition by ID
// @Tags facility-alarm-definitions
// @Produce json
// @Param id path string true "Alarm Definition ID"
// @Success 200 {object} AlarmDefinitionResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/facility/alarm-definitions/{id} [get]
func (h *AlarmDefinitionHandler) GetAlarmDefinition(c *gin.Context) {
	id, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}

	alarmDef, err := h.service.GetByID(id)
	if err != nil {
		if respondLocalizedNotFoundIf(c, err, "facility.alarm_definition_not_found") {
			return
		}
		respondLocalizedError(c, http.StatusInternalServerError, "fetch_failed", "facility.fetch_failed")
		return
	}

	c.JSON(http.StatusOK, toAlarmDefinitionResponse(*alarmDef))
}

// ListAlarmDefinitions godoc
// @Summary List alarm definitions with pagination
// @Tags facility-alarm-definitions
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Param search query string false "Search query"
// @Success 200 {object} AlarmDefinitionListResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/facility/alarm-definitions [get]
func (h *AlarmDefinitionHandler) ListAlarmDefinitions(c *gin.Context) {
	query, ok := parsePaginationQuery(c)
	if !ok {
		return
	}

	result, err := h.service.List(query.Page, query.Limit, query.Search)
	if err != nil {
		respondLocalizedError(c, http.StatusInternalServerError, "fetch_failed", "facility.fetch_failed")
		return
	}

	c.JSON(http.StatusOK, toAlarmDefinitionListResponse(result))
}

// UpdateAlarmDefinition godoc
// @Summary Update an alarm definition
// @Tags facility-alarm-definitions
// @Accept json
// @Produce json
// @Param id path string true "Alarm Definition ID"
// @Param alarm_definition body dto.UpdateAlarmDefinitionRequest true "Alarm Definition data"
// @Success 200 {object} dto.AlarmDefinitionResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/facility/alarm-definitions/{id} [put]
func (h *AlarmDefinitionHandler) UpdateAlarmDefinition(c *gin.Context) {
	id, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}

	var req dto.UpdateAlarmDefinitionRequest
	if !bindJSON(c, &req) {
		return
	}

	alarmDef, err := h.service.GetByID(id)
	if err != nil {
		if respondLocalizedNotFoundIf(c, err, "facility.alarm_definition_not_found") {
			return
		}
		respondLocalizedError(c, http.StatusInternalServerError, "fetch_failed", "facility.fetch_failed")
		return
	}

	applyAlarmDefinitionUpdate(alarmDef, req)

	if err := h.service.Update(alarmDef); respondLocalizedValidationOrError(c, err, "facility.update_failed") {
		return
	}

	c.JSON(http.StatusOK, toAlarmDefinitionResponse(*alarmDef))
}

// DeleteAlarmDefinition godoc
// @Summary Delete an alarm definition
// @Tags facility-alarm-definitions
// @Produce json
// @Param id path string true "Alarm Definition ID"
// @Success 204
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/facility/alarm-definitions/{id} [delete]
func (h *AlarmDefinitionHandler) DeleteAlarmDefinition(c *gin.Context) {
	id, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}

	if err := h.service.DeleteByID(id); err != nil {
		respondLocalizedError(c, http.StatusInternalServerError, "deletion_failed", "facility.deletion_failed")
		return
	}

	c.Status(http.StatusNoContent)
}
