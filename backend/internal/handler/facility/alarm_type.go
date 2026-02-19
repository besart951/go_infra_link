package facility

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type AlarmTypeHandler struct {
	service AlarmTypeService
}

func NewAlarmTypeHandler(service AlarmTypeService) *AlarmTypeHandler {
	return &AlarmTypeHandler{service: service}
}

// ListAlarmTypes godoc
// @Summary List alarm types
// @Tags facility-alarm-types
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Param search query string false "Search query"
// @Router /api/v1/facility/alarm-types [get]
func (h *AlarmTypeHandler) ListAlarmTypes(c *gin.Context) {
	query, ok := parsePaginationQuery(c)
	if !ok {
		return
	}

	result, err := h.service.List(query.Page, query.Limit, query.Search)
	if err != nil {
		respondLocalizedError(c, http.StatusInternalServerError, "fetch_failed", "facility.fetch_failed")
		return
	}

	c.JSON(http.StatusOK, toAlarmTypeListResponse(result))
}

// GetAlarmTypeFields godoc
// @Summary Get fields for an alarm type
// @Tags facility-alarm-types
// @Produce json
// @Param id path string true "Alarm Type ID"
// @Router /api/v1/facility/alarm-types/{id}/fields [get]
func (h *AlarmTypeHandler) GetAlarmTypeFields(c *gin.Context) {
	id, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}

	alarmType, err := h.service.GetWithFields(id)
	if err != nil {
		if respondLocalizedNotFoundIf(c, err, "facility.alarm_type_not_found") {
			return
		}
		respondLocalizedError(c, http.StatusInternalServerError, "fetch_failed", "facility.fetch_failed")
		return
	}
	if alarmType == nil {
		respondLocalizedError(c, http.StatusNotFound, "not_found", "facility.alarm_type_not_found")
		return
	}

	c.JSON(http.StatusOK, toAlarmTypeResponse(*alarmType))
}

// GetAlarmType godoc
// @Summary Get an alarm type by ID
// @Tags facility-alarm-types
// @Produce json
// @Param id path string true "Alarm Type ID"
// @Router /api/v1/facility/alarm-types/{id} [get]
func (h *AlarmTypeHandler) GetAlarmType(c *gin.Context) {
	id, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}

	alarmType, err := h.service.GetByID(id)
	if err != nil {
		if respondLocalizedNotFoundIf(c, err, "facility.alarm_type_not_found") {
			return
		}
		respondLocalizedError(c, http.StatusInternalServerError, "fetch_failed", "facility.fetch_failed")
		return
	}

	c.JSON(http.StatusOK, toAlarmTypeResponse(*alarmType))
}
