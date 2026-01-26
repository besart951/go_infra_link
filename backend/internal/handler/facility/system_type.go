package facility

import (
	"net/http"

	"github.com/besart951/go_infra_link/backend/internal/handler/dto"
	"github.com/gin-gonic/gin"
)

type SystemTypeHandler struct {
	service SystemTypeService
}

func NewSystemTypeHandler(service SystemTypeService) *SystemTypeHandler {
	return &SystemTypeHandler{service: service}
}

// CreateSystemType godoc
// @Summary Create a new system type
// @Tags facility-system-types
// @Accept json
// @Produce json
// @Param system_type body dto.CreateSystemTypeRequest true "System Type data"
// @Success 201 {object} dto.SystemTypeResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/facility/system-types [post]
func (h *SystemTypeHandler) CreateSystemType(c *gin.Context) {
	var req dto.CreateSystemTypeRequest
	if !bindJSON(c, &req) {
		return
	}

	systemType := toSystemTypeModel(req)

	if err := h.service.Create(systemType); respondValidationOrError(c, err, "creation_failed") {
		return
	}

	c.JSON(http.StatusCreated, toSystemTypeResponse(*systemType))
}

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
func (h *SystemTypeHandler) GetSystemType(c *gin.Context) {
	id, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}

	systemType, err := h.service.GetByID(id)
	if err != nil {
		if respondNotFoundIf(c, err, "System Type not found") {
			return
		}
		respondError(c, http.StatusInternalServerError, "fetch_failed", err.Error())
		return
	}

	c.JSON(http.StatusOK, toSystemTypeResponse(*systemType))
}

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
func (h *SystemTypeHandler) ListSystemTypes(c *gin.Context) {
	query, ok := parsePaginationQuery(c)
	if !ok {
		return
	}

	result, err := h.service.List(query.Page, query.Limit, query.Search)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "fetch_failed", err.Error())
		return
	}

	c.JSON(http.StatusOK, toSystemTypeListResponse(result))
}

// UpdateSystemType godoc
// @Summary Update a system type
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
func (h *SystemTypeHandler) UpdateSystemType(c *gin.Context) {
	id, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}

	var req dto.UpdateSystemTypeRequest
	if !bindJSON(c, &req) {
		return
	}

	systemType, err := h.service.GetByID(id)
	if err != nil {
		if respondNotFoundIf(c, err, "System Type not found") {
			return
		}
		respondError(c, http.StatusInternalServerError, "fetch_failed", err.Error())
		return
	}

	applySystemTypeUpdate(systemType, req)

	if err := h.service.Update(systemType); respondValidationOrError(c, err, "update_failed") {
		return
	}

	c.JSON(http.StatusOK, toSystemTypeResponse(*systemType))
}

// DeleteSystemType godoc
// @Summary Delete a system type
// @Tags facility-system-types
// @Produce json
// @Param id path string true "System Type ID"
// @Success 204
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/facility/system-types/{id} [delete]
func (h *SystemTypeHandler) DeleteSystemType(c *gin.Context) {
	id, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}

	if err := h.service.DeleteByID(id); err != nil {
		respondError(c, http.StatusInternalServerError, "deletion_failed", err.Error())
		return
	}

	c.Status(http.StatusNoContent)
}
