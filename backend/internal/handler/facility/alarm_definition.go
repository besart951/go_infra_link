package facility

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type AlarmDefinitionHandler struct {
	service AlarmDefinitionService
}

func NewAlarmDefinitionHandler(service AlarmDefinitionService) *AlarmDefinitionHandler {
	return &AlarmDefinitionHandler{service: service}
}

// GetAlarmDefinition godoc
// @Summary Get an alarm definition by ID
// @Tags facility-alarm-definitions
// @Produce json
// @Param id path string true "Alarm Definition ID"
// @Success 200 {object} dto.AlarmDefinitionResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/facility/alarm-definitions/{id} [get]
func (h *AlarmDefinitionHandler) GetAlarmDefinition(c *gin.Context) {
	id, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}

	alarmDef, err := h.service.GetByID(id)
	if err != nil {
		if respondNotFoundIf(c, err, "Alarm definition not found") {
			return
		}
		respondError(c, http.StatusInternalServerError, "fetch_failed", err.Error())
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
// @Success 200 {object} dto.AlarmDefinitionListResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/facility/alarm-definitions [get]
func (h *AlarmDefinitionHandler) ListAlarmDefinitions(c *gin.Context) {
	query, ok := parsePaginationQuery(c)
	if !ok {
		return
	}

	result, err := h.service.List(query.Page, query.Limit, query.Search)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "fetch_failed", err.Error())
		return
	}

	c.JSON(http.StatusOK, toAlarmDefinitionListResponse(result))
}
