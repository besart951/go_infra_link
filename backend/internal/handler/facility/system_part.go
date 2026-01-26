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

type SystemPartHandler struct {
	service SystemPartService
}

func NewSystemPartHandler(service SystemPartService) *SystemPartHandler {
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
	if !bindJSON(c, &req) {
		return
	}

	systemPart := &domainFacility.SystemPart{
		ShortName:   req.ShortName,
		Name:        req.Name,
		Description: req.Description,
	}

	if err := h.service.Create(systemPart); err != nil {
		respondError(c, http.StatusInternalServerError, "creation_failed", err.Error())
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
	id, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}

	systemPart, err := h.service.GetByID(id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			respondNotFound(c, "System Part not found")
			return
		}
		respondError(c, http.StatusInternalServerError, "fetch_failed", err.Error())
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
	if !bindQuery(c, &query) {
		return
	}

	result, err := h.service.List(query.Page, query.Limit, query.Search)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "fetch_failed", err.Error())
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
	id, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}

	var req dto.UpdateSystemPartRequest
	if !bindJSON(c, &req) {
		return
	}

	systemPart, err := h.service.GetByID(id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			respondNotFound(c, "System Part not found")
			return
		}
		respondError(c, http.StatusInternalServerError, "fetch_failed", err.Error())
		return
	}

	if req.ShortName != "" {
		systemPart.ShortName = req.ShortName
	}
	if req.Name != "" {
		systemPart.Name = req.Name
	}
	if req.Description != nil {
		systemPart.Description = req.Description
	}

	if err := h.service.Update(systemPart); err != nil {
		respondError(c, http.StatusInternalServerError, "update_failed", err.Error())
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
