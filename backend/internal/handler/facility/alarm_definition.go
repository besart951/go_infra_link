package facility

import (
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/besart951/go_infra_link/backend/internal/handler/dto"
	"github.com/gin-gonic/gin"
)

type AlarmDefinitionHandler struct {
	crud crudHandler[domainFacility.AlarmDefinition, dto.CreateAlarmDefinitionRequest, dto.UpdateAlarmDefinitionRequest]
}

func NewAlarmDefinitionHandler(svc AlarmDefinitionService) *AlarmDefinitionHandler {
	return &AlarmDefinitionHandler{crud: newCRUD(
		svc,
		toAlarmDefinitionModel,
		applyAlarmDefinitionUpdate,
		respFn(toAlarmDefinitionResponse),
		listRespFn(toAlarmDefinitionListResponse),
		"facility.alarm_definition_not_found",
	)}
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
func (h *AlarmDefinitionHandler) CreateAlarmDefinition(c *gin.Context) { h.crud.handleCreate(c) }

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
func (h *AlarmDefinitionHandler) GetAlarmDefinition(c *gin.Context) { h.crud.handleGetByID(c) }

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
func (h *AlarmDefinitionHandler) ListAlarmDefinitions(c *gin.Context) { h.crud.handleList(c) }

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
func (h *AlarmDefinitionHandler) UpdateAlarmDefinition(c *gin.Context) { h.crud.handleUpdate(c) }

// DeleteAlarmDefinition godoc
// @Summary Delete an alarm definition
// @Tags facility-alarm-definitions
// @Produce json
// @Param id path string true "Alarm Definition ID"
// @Success 204
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/facility/alarm-definitions/{id} [delete]
func (h *AlarmDefinitionHandler) DeleteAlarmDefinition(c *gin.Context) { h.crud.handleDelete(c) }
