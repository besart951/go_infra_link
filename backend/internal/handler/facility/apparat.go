package facility

import (
	"net/http"
	"strings"

	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/besart951/go_infra_link/backend/internal/handler/dto"
	"github.com/gin-gonic/gin"
)

type ApparatHandler struct {
	service           ApparatService
	systemPartService SystemPartService
	objectDataService ObjectDataService
}

func NewApparatHandler(service ApparatService, systemPartService SystemPartService, objectDataService ObjectDataService) *ApparatHandler {
	return &ApparatHandler{
		service:           service,
		systemPartService: systemPartService,
		objectDataService: objectDataService,
	}
}

// CreateApparat godoc
// @Summary Create a new apparat
// @Tags facility-apparats
// @Accept json
// @Produce json
// @Param apparat body dto.CreateApparatRequest true "Apparat data"
// @Success 201 {object} dto.ApparatResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/facility/apparats [post]
func (h *ApparatHandler) CreateApparat(c *gin.Context) {
	var req dto.CreateApparatRequest
	if !bindJSON(c, &req) {
		return
	}

	// Load system parts if IDs are provided
	var systemParts []*domainFacility.SystemPart
	if len(req.SystemPartIDs) > 0 {
		loadedParts, err := h.systemPartService.GetByIDs(req.SystemPartIDs)
		if err != nil {
			respondError(c, http.StatusBadRequest, "invalid_system_parts", "Failed to load system parts")
			return
		}
		systemParts = loadedParts
	}

	apparat := toApparatModel(req, systemParts)

	if err := h.service.Create(apparat); respondValidationOrError(c, err, "creation_failed") {
		return
	}

	c.JSON(http.StatusCreated, toApparatResponse(*apparat))
}

// GetApparat godoc
// @Summary Get an apparat by ID
// @Tags facility-apparats
// @Produce json
// @Param id path string true "Apparat ID"
// @Success 200 {object} dto.ApparatResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/facility/apparats/{id} [get]
func (h *ApparatHandler) GetApparat(c *gin.Context) {
	id, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}

	apparat, err := h.service.GetByID(id)
	if err != nil {
		if respondNotFoundIf(c, err, "Apparat not found") {
			return
		}
		respondError(c, http.StatusInternalServerError, "fetch_failed", err.Error())
		return
	}

	c.JSON(http.StatusOK, toApparatResponse(*apparat))
}

// ListApparats godoc
// @Summary List apparats with pagination
// @Tags facility-apparats
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Param search query string false "Search query"
// @Param object_data_id query string false "Filter by Object Data ID"
// @Success 200 {object} dto.ApparatListResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/facility/apparats [get]
func (h *ApparatHandler) ListApparats(c *gin.Context) {
	query, ok := parsePaginationQuery(c)
	if !ok {
		return
	}

	objectDataIDStr := c.Query("object_data_id")
	
	// If object_data_id is provided, filter apparats
	if objectDataIDStr != "" {
		objectDataID, err := parseUUIDString(objectDataIDStr)
		if err != nil {
			respondError(c, http.StatusBadRequest, "invalid_object_data_id", "Invalid object_data_id format")
			return
		}
		
		apparatIDs, err := h.getApparatsForObjectData(objectDataID)
		if err != nil {
			if respondNotFoundIf(c, err, "Object data not found") {
				return
			}
			respondError(c, http.StatusInternalServerError, "fetch_failed", err.Error())
			return
		}
		
		if len(apparatIDs) == 0 {
			c.JSON(http.StatusOK, dto.ApparatListResponse{
				Items:      []dto.ApparatResponse{},
				TotalPages: 0,
				Page:       1,
				Limit:      query.Limit,
				Total:      0,
			})
			return
		}
		
		// Get all apparats and filter by search if needed
		apparats := make([]*domainFacility.Apparat, 0, len(apparatIDs))
		for _, id := range apparatIDs {
			apparat, err := h.service.GetByID(id)
			if err == nil && apparat != nil {
				// Apply search filter if provided
				if query.Search == "" || 
					containsIgnoreCase(apparat.ShortName, query.Search) || 
					containsIgnoreCase(apparat.Name, query.Search) ||
					(apparat.Description != nil && containsIgnoreCase(*apparat.Description, query.Search)) {
					apparats = append(apparats, apparat)
				}
			}
		}
		
		// Simple pagination
		total := len(apparats)
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
		
		pageApparats := apparats[start:end]
		responses := make([]dto.ApparatResponse, 0, len(pageApparats))
		for _, apparat := range pageApparats {
			responses = append(responses, toApparatResponse(*apparat))
		}
		
		c.JSON(http.StatusOK, dto.ApparatListResponse{
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

	c.JSON(http.StatusOK, toApparatListResponse(result))
}

// UpdateApparat godoc
// @Summary Update an apparat
// @Tags facility-apparats
// @Accept json
// @Produce json
// @Param id path string true "Apparat ID"
// @Param apparat body dto.UpdateApparatRequest true "Apparat data"
// @Success 200 {object} dto.ApparatResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/facility/apparats/{id} [put]
func (h *ApparatHandler) UpdateApparat(c *gin.Context) {
	id, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}

	var req dto.UpdateApparatRequest
	if !bindJSON(c, &req) {
		return
	}

	apparat, err := h.service.GetByID(id)
	if err != nil {
		if respondNotFoundIf(c, err, "Apparat not found") {
			return
		}
		respondError(c, http.StatusInternalServerError, "fetch_failed", err.Error())
		return
	}

	// Load system parts if IDs are provided
	var systemParts *[]*domainFacility.SystemPart
	if req.SystemPartIDs != nil {
		if len(*req.SystemPartIDs) > 0 {
			loadedParts, err := h.systemPartService.GetByIDs(*req.SystemPartIDs)
			if err != nil {
				respondError(c, http.StatusBadRequest, "invalid_system_parts", "Failed to load system parts")
				return
			}
			systemParts = &loadedParts
		} else {
			// Empty array means clear all system parts
			emptyParts := []*domainFacility.SystemPart{}
			systemParts = &emptyParts
		}
	}

	applyApparatUpdate(apparat, req, systemParts)

	if err := h.service.Update(apparat); respondValidationOrError(c, err, "update_failed") {
		return
	}

	c.JSON(http.StatusOK, toApparatResponse(*apparat))
}

// DeleteApparat godoc
// @Summary Delete an apparat
// @Tags facility-apparats
// @Produce json
// @Param id path string true "Apparat ID"
// @Success 204
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/facility/apparats/{id} [delete]
func (h *ApparatHandler) DeleteApparat(c *gin.Context) {
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

func (h *ApparatHandler) getApparatsForObjectData(objectDataID uuid.UUID) ([]uuid.UUID, error) {
	return h.objectDataService.GetApparatIDs(objectDataID)
}

func containsIgnoreCase(s, substr string) bool {
	return len(s) >= len(substr) && 
		   (s == substr || 
		    len(substr) == 0 || 
		    (len(s) > 0 && len(substr) > 0 && 
			 toLower(s) == toLower(substr) || 
			 strings.Contains(toLower(s), toLower(substr))))
}

func toLower(s string) string {
	return strings.ToLower(s)
}
