package facility

import (
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	dto "github.com/besart951/go_infra_link/backend/internal/handler/dto/facility"
	"github.com/gin-gonic/gin"
)

type UnitHandler struct {
	crud crudHandler[domainFacility.Unit, dto.CreateUnitRequest, dto.UpdateUnitRequest]
}

func NewUnitHandler(svc UnitService) *UnitHandler {
	return &UnitHandler{crud: newCRUD(
		svc,
		toUnitModel,
		applyUnitUpdate,
		respFn(toUnitResponse),
		listRespFn(toUnitListResponse),
		"alarm_unit",
		"facility.not_found",
	)}
}

// CreateUnit godoc
// @Summary Create an alarm unit
// @Tags facility-alarm-units
// @Accept json
// @Produce json
// @Param unit body dto.CreateUnitRequest true "Unit data"
// @Success 201 {object} dto.UnitResponse
// @Router /api/v1/facility/alarm-units [post]
func (h *UnitHandler) CreateUnit(c *gin.Context) { h.crud.handleCreate(c) }

// GetUnit godoc
// @Summary Get an alarm unit
// @Tags facility-alarm-units
// @Produce json
// @Param id path string true "Unit ID"
// @Success 200 {object} dto.UnitResponse
// @Router /api/v1/facility/alarm-units/{id} [get]
func (h *UnitHandler) GetUnit(c *gin.Context) { h.crud.handleGetByID(c) }

// ListUnits godoc
// @Summary List alarm units
// @Tags facility-alarm-units
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Param search query string false "Search query"
// @Success 200 {object} dto.UnitListResponse
// @Router /api/v1/facility/alarm-units [get]
func (h *UnitHandler) ListUnits(c *gin.Context) { h.crud.handleList(c) }

// UpdateUnit godoc
// @Summary Update an alarm unit
// @Tags facility-alarm-units
// @Accept json
// @Produce json
// @Param id path string true "Unit ID"
// @Param unit body dto.UpdateUnitRequest true "Unit data"
// @Success 200 {object} dto.UnitResponse
// @Router /api/v1/facility/alarm-units/{id} [put]
func (h *UnitHandler) UpdateUnit(c *gin.Context) { h.crud.handleUpdate(c) }

// DeleteUnit godoc
// @Summary Delete an alarm unit
// @Tags facility-alarm-units
// @Param id path string true "Unit ID"
// @Param base_version query integer true "Expected aggregate version" minimum(1)
// @Success 204
// @Router /api/v1/facility/alarm-units/{id} [delete]
func (h *UnitHandler) DeleteUnit(c *gin.Context) { h.crud.handleDelete(c) }
