package facility

import (
	"net/http"

	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/besart951/go_infra_link/backend/internal/handler/dto"
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

func (h *AlarmTypeHandler) CreateAlarmType(c *gin.Context)  { h.crud.handleCreate(c) }
func (h *AlarmTypeHandler) GetAlarmType(c *gin.Context)     { h.crud.handleGetByID(c) }
func (h *AlarmTypeHandler) UpdateAlarmType(c *gin.Context)  { h.crud.handleUpdate(c) }
func (h *AlarmTypeHandler) DeleteAlarmType(c *gin.Context)  { h.crud.handleDelete(c) }

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
