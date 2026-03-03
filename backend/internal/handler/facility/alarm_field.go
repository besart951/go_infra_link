package facility

import (
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/besart951/go_infra_link/backend/internal/handler/dto"
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

func (h *AlarmFieldHandler) CreateAlarmField(c *gin.Context) { h.crud.handleCreate(c) }
func (h *AlarmFieldHandler) GetAlarmField(c *gin.Context)    { h.crud.handleGetByID(c) }
func (h *AlarmFieldHandler) ListAlarmFields(c *gin.Context)  { h.crud.handleList(c) }
func (h *AlarmFieldHandler) UpdateAlarmField(c *gin.Context) { h.crud.handleUpdate(c) }
func (h *AlarmFieldHandler) DeleteAlarmField(c *gin.Context) { h.crud.handleDelete(c) }
