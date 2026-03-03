package facility

import (
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/besart951/go_infra_link/backend/internal/handler/dto"
	"github.com/gin-gonic/gin"
)

type SystemTypeHandler struct {
	crud crudHandler[domainFacility.SystemType, dto.CreateSystemTypeRequest, dto.UpdateSystemTypeRequest]
}

func NewSystemTypeHandler(svc SystemTypeService) *SystemTypeHandler {
	return &SystemTypeHandler{crud: newCRUD(
		svc,
		toSystemTypeModel,
		applySystemTypeUpdate,
		respFn(toSystemTypeResponse),
		listRespFn(toSystemTypeListResponse),
		"facility.system_type_not_found",
	)}
}

// CreateSystemType godoc
// @Summary Create a new system type
// @Description number_min and number_max must not overlap existing ranges. number_min may equal number_max.
// @Tags facility-system-types
// @Accept json
// @Produce json
// @Param system_type body dto.CreateSystemTypeRequest true "System Type data"
// @Success 201 {object} dto.SystemTypeResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/facility/system-types [post]
func (h *SystemTypeHandler) CreateSystemType(c *gin.Context) { h.crud.handleCreate(c) }

// GetSystemType godoc
// @Summary Get a system type by ID
// @Tags facility-system-types
// @Produce json
// @Param id path string true "System Type ID"
// @Success 200 {object} dto.SystemTypeResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/facility/system-types/{id} [get]
func (h *SystemTypeHandler) GetSystemType(c *gin.Context) { h.crud.handleGetByID(c) }

// ListSystemTypes godoc
// @Summary List system types with pagination
// @Tags facility-system-types
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Param search query string false "Search query"
// @Success 200 {object} dto.SystemTypeListResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/facility/system-types [get]
func (h *SystemTypeHandler) ListSystemTypes(c *gin.Context) { h.crud.handleList(c) }

// UpdateSystemType godoc
// @Summary Update a system type
// @Description number_min and number_max must not overlap existing ranges. number_min may equal number_max.
// @Tags facility-system-types
// @Accept json
// @Produce json
// @Param id path string true "System Type ID"
// @Param system_type body dto.UpdateSystemTypeRequest true "System Type data"
// @Success 200 {object} dto.SystemTypeResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/facility/system-types/{id} [put]
func (h *SystemTypeHandler) UpdateSystemType(c *gin.Context) { h.crud.handleUpdate(c) }

// DeleteSystemType godoc
// @Summary Delete a system type
// @Tags facility-system-types
// @Produce json
// @Param id path string true "System Type ID"
// @Success 204
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/facility/system-types/{id} [delete]
func (h *SystemTypeHandler) DeleteSystemType(c *gin.Context) { h.crud.handleDelete(c) }
