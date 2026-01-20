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

type SystemPartHandler struct {
	service *facilityService.Service
}

func NewSystemPartHandler(service *facilityService.Service) *SystemPartHandler {
	return &SystemPartHandler{service: service}
}

// CreateSystemPart godoc
// @Summary Create a new system part
// @Tags facility-system-parts
// @Accept json
// @Produce json
// @Param system_part body dto.CreateSystemPartRequest true "System Part data"
// @Success 201 {object} dto.SystemPartResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/facility/system-parts [post]
func (h *SystemPartHandler) CreateSystemPart(c *gin.Context) {
	var req dto.CreateSystemPartRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "validation_error",
			Message: err.Error(),
		})
		return
	}

	systemPart := &domainFacility.SystemPart{
		ShortName:   req.ShortName,
		Name:        req.Name,
		Description: req.Description,
	}

	if err := h.service.SystemParts.Create(systemPart); err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "creation_failed",
			Message: err.Error(),
		})
		return
	}

	response := dto.SystemPartResponse{
		ID:          systemPart.ID,
		ShortName:   systemPart.ShortName,
		Name:        systemPart.Name,
		Description: systemPart.Description,
		CreatedAt:   systemPart.CreatedAt,
		UpdatedAt:   systemPart.UpdatedAt,
	}

	c.JSON(http.StatusCreated, response)
}

// GetSystemPart godoc
// @Summary Get a system part by ID
// @Tags facility-system-parts
// @Produce json
// @Param id path string true "System Part ID"
// @Success 200 {object} dto.SystemPartResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/facility/system-parts/{id} [get]
func (h *SystemPartHandler) GetSystemPart(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid UUID format",
		})
		return
	}

	systemParts, err := h.service.SystemParts.GetByIds([]uuid.UUID{id})
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "fetch_failed",
			Message: err.Error(),
		})
		return
	}

	if len(systemParts) == 0 {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Error:   "not_found",
			Message: "System Part not found",
		})
		return
	}

	systemPart := systemParts[0]
	response := dto.SystemPartResponse{
		ID:          systemPart.ID,
		ShortName:   systemPart.ShortName,
		Name:        systemPart.Name,
		Description: systemPart.Description,
		CreatedAt:   systemPart.CreatedAt,
		UpdatedAt:   systemPart.UpdatedAt,
	}

	c.JSON(http.StatusOK, response)
}

// ListSystemParts godoc
// @Summary List system parts with pagination
// @Tags facility-system-parts
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Param search query string false "Search query"
// @Success 200 {object} dto.SystemPartListResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/facility/system-parts [get]
func (h *SystemPartHandler) ListSystemParts(c *gin.Context) {
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

	result, err := h.service.SystemParts.GetPaginatedList(params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "fetch_failed",
			Message: err.Error(),
		})
		return
	}

	items := make([]dto.SystemPartResponse, len(result.Items))
	for i, systemPart := range result.Items {
		items[i] = dto.SystemPartResponse{
			ID:          systemPart.ID,
			ShortName:   systemPart.ShortName,
			Name:        systemPart.Name,
			Description: systemPart.Description,
			CreatedAt:   systemPart.CreatedAt,
			UpdatedAt:   systemPart.UpdatedAt,
		}
	}

	response := dto.SystemPartListResponse{
		Items:      items,
		Total:      result.Total,
		Page:       result.Page,
		TotalPages: result.TotalPages,
	}

	c.JSON(http.StatusOK, response)
}

// UpdateSystemPart godoc
// @Summary Update a system part
// @Tags facility-system-parts
// @Accept json
// @Produce json
// @Param id path string true "System Part ID"
// @Param system_part body dto.UpdateSystemPartRequest true "System Part data"
// @Success 200 {object} dto.SystemPartResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/facility/system-parts/{id} [put]
func (h *SystemPartHandler) UpdateSystemPart(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid UUID format",
		})
		return
	}

	var req dto.UpdateSystemPartRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "validation_error",
			Message: err.Error(),
		})
		return
	}

	systemParts, err := h.service.SystemParts.GetByIds([]uuid.UUID{id})
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "fetch_failed",
			Message: err.Error(),
		})
		return
	}

	if len(systemParts) == 0 {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Error:   "not_found",
			Message: "System Part not found",
		})
		return
	}

	systemPart := systemParts[0]
	if req.ShortName != "" {
		systemPart.ShortName = req.ShortName
	}
	if req.Name != "" {
		systemPart.Name = req.Name
	}
	if req.Description != nil {
		systemPart.Description = req.Description
	}

	if err := h.service.SystemParts.Update(systemPart); err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "update_failed",
			Message: err.Error(),
		})
		return
	}

	response := dto.SystemPartResponse{
		ID:          systemPart.ID,
		ShortName:   systemPart.ShortName,
		Name:        systemPart.Name,
		Description: systemPart.Description,
		CreatedAt:   systemPart.CreatedAt,
		UpdatedAt:   systemPart.UpdatedAt,
	}

	c.JSON(http.StatusOK, response)
}

// DeleteSystemPart godoc
// @Summary Delete a system part
// @Tags facility-system-parts
// @Produce json
// @Param id path string true "System Part ID"
// @Success 204
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/facility/system-parts/{id} [delete]
func (h *SystemPartHandler) DeleteSystemPart(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid UUID format",
		})
		return
	}

	if err := h.service.SystemParts.DeleteByIds([]uuid.UUID{id}); err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "deletion_failed",
			Message: err.Error(),
		})
		return
	}

	c.Status(http.StatusNoContent)
}
