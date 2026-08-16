package facility

import (
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	dto "github.com/besart951/go_infra_link/backend/internal/handler/dto/facility"
	"github.com/gin-gonic/gin"
)

type AlarmFieldHandler struct {
	crud crudHandler[domainFacility.AlarmField, dto.CreateAlarmFieldRequest, dto.UpdateAlarmFieldRequest]
}

func NewAlarmFieldHandler(svc AlarmFieldService) *AlarmFieldHandler {
	return &AlarmFieldHandler{crud: newCRUD(
		svc,
		toAlarmFieldModel,
		applyAlarmFieldUpdate,
		respFn(toAlarmFieldResponse),
		listRespFn(toAlarmFieldListResponse),
		"facility.not_found",
	)}
}

// CreateAlarmField godoc
// @Summary Create an alarm field
// @Tags facility-alarm-fields
// @Accept json
// @Produce json
// @Param field body dto.CreateAlarmFieldRequest true "Alarm field data"
// @Success 201 {object} dto.AlarmFieldResponse
// @Router /api/v1/facility/alarm-fields [post]
func (h *AlarmFieldHandler) CreateAlarmField(c *gin.Context) { h.crud.handleCreate(c) }

// GetAlarmField godoc
// @Summary Get an alarm field
// @Tags facility-alarm-fields
// @Produce json
// @Param id path string true "Alarm Field ID"
// @Success 200 {object} dto.AlarmFieldResponse
// @Router /api/v1/facility/alarm-fields/{id} [get]
func (h *AlarmFieldHandler) GetAlarmField(c *gin.Context) { h.crud.handleGetByID(c) }

// ListAlarmFields godoc
// @Summary List alarm fields
// @Tags facility-alarm-fields
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Param search query string false "Search query"
// @Success 200 {object} dto.AlarmFieldListResponse
// @Router /api/v1/facility/alarm-fields [get]
func (h *AlarmFieldHandler) ListAlarmFields(c *gin.Context) { h.crud.handleList(c) }

// UpdateAlarmField godoc
// @Summary Update an alarm field
// @Tags facility-alarm-fields
// @Accept json
// @Produce json
// @Param id path string true "Alarm Field ID"
// @Param field body dto.UpdateAlarmFieldRequest true "Alarm field data"
// @Success 200 {object} dto.AlarmFieldResponse
// @Router /api/v1/facility/alarm-fields/{id} [put]
func (h *AlarmFieldHandler) UpdateAlarmField(c *gin.Context) { h.crud.handleUpdate(c) }

// DeleteAlarmField godoc
// @Summary Delete an alarm field
// @Tags facility-alarm-fields
// @Param id path string true "Alarm Field ID"
// @Success 204
// @Router /api/v1/facility/alarm-fields/{id} [delete]
func (h *AlarmFieldHandler) DeleteAlarmField(c *gin.Context) { h.crud.handleDelete(c) }
