package facility

import (
	"net/http"

	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	dto "github.com/besart951/go_infra_link/backend/internal/handler/dto/facility"
	"github.com/gin-gonic/gin"
)

type AlarmTypeHandler struct {
	crud    crudHandler[domainFacility.AlarmType, dto.CreateAlarmTypeRequest, dto.UpdateAlarmTypeRequest]
	service AlarmTypeService
}

func NewAlarmTypeHandler(svc AlarmTypeService) *AlarmTypeHandler {
	return &AlarmTypeHandler{
		crud: newCRUD(
			svc,
			toAlarmTypeModel,
			applyAlarmTypeUpdate,
			respFn(toAlarmTypeResponse),
			listRespFn(toAlarmTypeListResponse),
			"facility.alarm_type_not_found",
		),
		service: svc,
	}
}

// ListAlarmTypes godoc
// @Summary List alarm types
// @Tags facility-alarm-types
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Param search query string false "Search query"
// @Router /api/v1/facility/alarm-types [get]
func (h *AlarmTypeHandler) ListAlarmTypes(c *gin.Context) { h.crud.handleList(c) }

// CreateAlarmType godoc
// @Summary Create an alarm type
// @Tags facility-alarm-types
// @Accept json
// @Produce json
// @Param type body dto.CreateAlarmTypeRequest true "Alarm type data"
// @Success 201 {object} dto.AlarmTypeResponse
// @Router /api/v1/facility/alarm-types [post]
func (h *AlarmTypeHandler) CreateAlarmType(c *gin.Context) { h.crud.handleCreate(c) }

// GetAlarmType godoc
// @Summary Get an alarm type
// @Tags facility-alarm-types
// @Produce json
// @Param id path string true "Alarm Type ID"
// @Success 200 {object} dto.AlarmTypeResponse
// @Router /api/v1/facility/alarm-types/{id} [get]
func (h *AlarmTypeHandler) GetAlarmType(c *gin.Context) { h.crud.handleGetByID(c) }

// UpdateAlarmType godoc
// @Summary Update an alarm type
// @Tags facility-alarm-types
// @Accept json
// @Produce json
// @Param id path string true "Alarm Type ID"
// @Param type body dto.UpdateAlarmTypeRequest true "Alarm type data"
// @Success 200 {object} dto.AlarmTypeResponse
// @Router /api/v1/facility/alarm-types/{id} [put]
func (h *AlarmTypeHandler) UpdateAlarmType(c *gin.Context) { h.crud.handleUpdate(c) }

// DeleteAlarmType godoc
// @Summary Delete an alarm type
// @Tags facility-alarm-types
// @Param id path string true "Alarm Type ID"
// @Success 204
// @Router /api/v1/facility/alarm-types/{id} [delete]
func (h *AlarmTypeHandler) DeleteAlarmType(c *gin.Context) { h.crud.handleDelete(c) }

// GetAlarmTypeFields godoc
// @Summary Get fields for an alarm type
// @Tags facility-alarm-types
// @Produce json
// @Param id path string true "Alarm Type ID"
// @Success 200 {object} dto.AlarmTypeResponse
// @Router /api/v1/facility/alarm-types/{id}/fields [get]
func (h *AlarmTypeHandler) GetAlarmTypeFields(c *gin.Context) {
	id, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}
	alarmType, err := h.service.GetWithFields(c.Request.Context(), id)
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
