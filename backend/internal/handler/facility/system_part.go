package facility

import (
	"net/http"
	"strings"

	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/besart951/go_infra_link/backend/internal/handler/dto"
	"github.com/gin-gonic/gin"
)

type SystemPartHandler struct {
	service        SystemPartService
	apparatService ApparatService
}

func NewSystemPartHandler(service SystemPartService, apparatService ApparatService) *SystemPartHandler {
	return &SystemPartHandler{
		service:        service,
		apparatService: apparatService,
	}
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

	systemPart := toSystemPartModel(req)

	if err := h.service.Create(systemPart); respondValidationOrError(c, err, "creation_failed") {
		return
	}

	c.JSON(http.StatusCreated, toSystemPartResponse(*systemPart))
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
		if respondNotFoundIf(c, err, "System Part not found") {
			return
		}
		respondError(c, http.StatusInternalServerError, "fetch_failed", err.Error())
		return
	}

	c.JSON(http.StatusOK, toSystemPartResponse(*systemPart))
}

// ListSystemParts godoc
// @Summary List system parts with pagination
// @Tags facility-system-parts
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Param search query string false "Search query"
// @Param apparat_id query string false "Filter by Apparat ID"
// @Success 200 {object} dto.SystemPartListResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/facility/system-parts [get]
func (h *SystemPartHandler) ListSystemParts(c *gin.Context) {
	query, ok := parsePaginationQuery(c)
	if !ok {
		return
	}

	apparatIDStr := c.Query("apparat_id")
	
	// If apparat_id is provided, filter system parts
	if apparatIDStr != "" {
		apparatID, err := parseUUIDString(apparatIDStr)
		if err != nil {
			respondError(c, http.StatusBadRequest, "invalid_apparat_id", "Invalid apparat_id format")
			return
		}
		
		systemPartIDs, err := h.apparatService.GetSystemPartIDs(apparatID)
		if err != nil {
			if respondNotFoundIf(c, err, "Apparat not found") {
				return
			}
			respondError(c, http.StatusInternalServerError, "fetch_failed", err.Error())
			return
		}
		
		if len(systemPartIDs) == 0 {
			c.JSON(http.StatusOK, dto.SystemPartListResponse{
				Items:      []dto.SystemPartResponse{},
				TotalPages: 0,
				Page:       1,
				Limit:      query.Limit,
				Total:      0,
			})
			return
		}
		
		// Get all system parts and filter by search if needed
		systemParts := make([]*domainFacility.SystemPart, 0, len(systemPartIDs))
		for _, id := range systemPartIDs {
			systemPart, err := h.service.GetByID(id)
			if err == nil && systemPart != nil {
				// Apply search filter if provided
				if query.Search == "" || 
					strings.Contains(strings.ToLower(systemPart.ShortName), strings.ToLower(query.Search)) || 
					strings.Contains(strings.ToLower(systemPart.Name), strings.ToLower(query.Search)) ||
					(systemPart.Description != nil && strings.Contains(strings.ToLower(*systemPart.Description), strings.ToLower(query.Search))) {
					systemParts = append(systemParts, systemPart)
				}
			}
		}
		
		// Simple pagination
		total := len(systemParts)
		totalPages := (total + query.Limit - 1) / query.Limit
		if totalPages == 0 {
			totalPages = 1
		}
		
		start := (query.Page - 1) * query.Limit
		end := start + query.Limit
		if start >= total {
			start = 0
			end = 0
		} else if end > total {
			end = total
		}
		
		pageSystemParts := systemParts[start:end]
		responses := make([]dto.SystemPartResponse, 0, len(pageSystemParts))
		for _, systemPart := range pageSystemParts {
			responses = append(responses, toSystemPartResponse(*systemPart))
		}
		
		c.JSON(http.StatusOK, dto.SystemPartListResponse{
			Items:      responses,
			TotalPages: totalPages,
			Page:       query.Page,
			Limit:      query.Limit,
			Total:      total,
		})
		return
	}

	result, err := h.service.List(query.Page, query.Limit, query.Search)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "fetch_failed", err.Error())
		return
	}

	c.JSON(http.StatusOK, toSystemPartListResponse(result))
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
		if respondNotFoundIf(c, err, "System Part not found") {
			return
		}
		respondError(c, http.StatusInternalServerError, "fetch_failed", err.Error())
		return
	}

	applySystemPartUpdate(systemPart, req)

	if err := h.service.Update(systemPart); respondValidationOrError(c, err, "update_failed") {
		return
	}

	c.JSON(http.StatusOK, toSystemPartResponse(*systemPart))
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

	if err := h.service.DeleteByID(id); err != nil {
		respondError(c, http.StatusInternalServerError, "deletion_failed", err.Error())
		return
	}

	c.Status(http.StatusNoContent)
}
