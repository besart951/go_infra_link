package facility

import (
	"errors"
	"net/http"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/besart951/go_infra_link/backend/internal/handler/dto"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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

	systemType := &domainFacility.SystemType{
		NumberMin: req.NumberMin,
		NumberMax: req.NumberMax,
		Name:      req.Name,
	}

	if err := h.service.Create(systemType); err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "creation_failed",
			Message: err.Error(),
		})
		return
	}

	response := dto.SystemTypeResponse{
		ID:        systemType.ID,
		NumberMin: systemType.NumberMin,
		NumberMax: systemType.NumberMax,
		Name:      systemType.Name,
		CreatedAt: systemType.CreatedAt,
		UpdatedAt: systemType.UpdatedAt,
	}

	c.JSON(http.StatusCreated, response)
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
		if errors.Is(err, domain.ErrNotFound) {
			respondNotFound(c, "System Type not found")
			return
		}
		respondError(c, http.StatusInternalServerError, "fetch_failed", err.Error())
		return
	}

	response := dto.SystemTypeResponse{
		ID:        systemType.ID,
		NumberMin: systemType.NumberMin,
		NumberMax: systemType.NumberMax,
		Name:      systemType.Name,
		CreatedAt: systemType.CreatedAt,
		UpdatedAt: systemType.UpdatedAt,
	}

	c.JSON(http.StatusOK, response)
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
	var query dto.PaginationQuery
	if !bindQuery(c, &query) {
		return
	}

	result, err := h.service.List(query.Page, query.Limit, query.Search)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "fetch_failed", err.Error())
		return
	}

	items := make([]dto.SystemTypeResponse, len(result.Items))
	for i, systemType := range result.Items {
		items[i] = dto.SystemTypeResponse{
			ID:        systemType.ID,
			NumberMin: systemType.NumberMin,
			NumberMax: systemType.NumberMax,
			Name:      systemType.Name,
			CreatedAt: systemType.CreatedAt,
			UpdatedAt: systemType.UpdatedAt,
		}
	}

	response := dto.SystemTypeListResponse{
		Items:      items,
		Total:      result.Total,
		Page:       result.Page,
		TotalPages: result.TotalPages,
	}

	c.JSON(http.StatusOK, response)
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
		if errors.Is(err, domain.ErrNotFound) {
			respondNotFound(c, "System Type not found")
			return
		}
		respondError(c, http.StatusInternalServerError, "fetch_failed", err.Error())
		return
	}

	if req.NumberMin != 0 {
		systemType.NumberMin = req.NumberMin
	}
	if req.NumberMax != 0 {
		systemType.NumberMax = req.NumberMax
	}
	if req.Name != "" {
		systemType.Name = req.Name
	}

	if err := h.service.Update(systemType); err != nil {
		respondError(c, http.StatusInternalServerError, "update_failed", err.Error())
		return
	}

	response := dto.SystemTypeResponse{
		ID:        systemType.ID,
		NumberMin: systemType.NumberMin,
		NumberMax: systemType.NumberMax,
		Name:      systemType.Name,
		CreatedAt: systemType.CreatedAt,
		UpdatedAt: systemType.UpdatedAt,
	}

	c.JSON(http.StatusOK, response)
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

	if err := h.service.DeleteByIds([]uuid.UUID{id}); err != nil {
		respondError(c, http.StatusInternalServerError, "deletion_failed", err.Error())
		return
	}

	c.Status(http.StatusNoContent)
}
