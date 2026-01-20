package facility

import (
	"net/http"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/besart951/go_infra_link/backend/internal/handler/dto"
	facilityService "github.com/besart951/go_infra_link/backend/internal/service/facility"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type SystemTypeHandler struct {
	service *facilityService.Service
}

func NewSystemTypeHandler(service *facilityService.Service) *SystemTypeHandler {
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
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "validation_error",
			Message: err.Error(),
		})
		return
	}

	systemType := &domainFacility.SystemType{
		NumberMin: req.NumberMin,
		NumberMax: req.NumberMax,
		Name:      req.Name,
	}

	if err := h.service.SystemTypes.Create(systemType); err != nil {
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
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid UUID format",
		})
		return
	}

	systemTypes, err := h.service.SystemTypes.GetByIds([]uuid.UUID{id})
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "fetch_failed",
			Message: err.Error(),
		})
		return
	}

	if len(systemTypes) == 0 {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Error:   "not_found",
			Message: "System Type not found",
		})
		return
	}

	systemType := systemTypes[0]
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
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "validation_error",
			Message: err.Error(),
		})
		return
	}

	if query.Page == 0 {
		query.Page = 1
	}
	if query.Limit == 0 {
		query.Limit = 10
	}

	params := domain.PaginationParams{
		Page:   query.Page,
		Limit:  query.Limit,
		Search: query.Search,
	}

	result, err := h.service.SystemTypes.GetPaginatedList(params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "fetch_failed",
			Message: err.Error(),
		})
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
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid UUID format",
		})
		return
	}

	var req dto.UpdateSystemTypeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "validation_error",
			Message: err.Error(),
		})
		return
	}

	systemTypes, err := h.service.SystemTypes.GetByIds([]uuid.UUID{id})
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "fetch_failed",
			Message: err.Error(),
		})
		return
	}

	if len(systemTypes) == 0 {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Error:   "not_found",
			Message: "System Type not found",
		})
		return
	}

	systemType := systemTypes[0]
	if req.NumberMin != 0 {
		systemType.NumberMin = req.NumberMin
	}
	if req.NumberMax != 0 {
		systemType.NumberMax = req.NumberMax
	}
	if req.Name != "" {
		systemType.Name = req.Name
	}

	if err := h.service.SystemTypes.Update(systemType); err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "update_failed",
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
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid UUID format",
		})
		return
	}

	if err := h.service.SystemTypes.DeleteByIds([]uuid.UUID{id}); err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "deletion_failed",
			Message: err.Error(),
		})
		return
	}

	c.Status(http.StatusNoContent)
}
