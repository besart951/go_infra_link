package facility

import (
	"net/http"
	"strings"

	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/besart951/go_infra_link/backend/internal/handler/dto"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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
			respondLocalizedError(c, http.StatusBadRequest, "invalid_system_parts", "facility.invalid_system_parts")
			return
		}
		systemParts = loadedParts
	}

	apparat := toApparatModel(req, systemParts)

	if err := h.service.Create(apparat); respondLocalizedValidationOrError(c, err, "facility.creation_failed") {
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
		if respondLocalizedNotFoundIf(c, err, "facility.apparat_not_found") {
			return
		}
		respondLocalizedError(c, http.StatusInternalServerError, "fetch_failed", "facility.fetch_failed")
		return
	}

	c.JSON(http.StatusOK, toApparatResponse(*apparat))
}

// GetApparatsByIDs godoc
// @Summary Get multiple apparats by IDs
// @Tags facility-apparats
// @Accept json
// @Produce json
// @Param request body dto.ApparatBulkRequest true "Apparat IDs"
// @Success 200 {object} dto.ApparatBulkResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/facility/apparats/bulk [post]
func (h *ApparatHandler) GetApparatsByIDs(c *gin.Context) {
	var req dto.ApparatBulkRequest
	if !bindJSON(c, &req) {
		return
	}
	if len(req.Ids) == 0 {
		respondLocalizedInvalidArgument(c, "facility.ids_required")
		return
	}

	apparats, err := h.service.GetByIDs(req.Ids)
	if err != nil {
		respondLocalizedError(c, http.StatusInternalServerError, "fetch_failed", "facility.fetch_failed")
		return
	}

	items := make([]domainFacility.Apparat, 0, len(apparats))
	for _, apparat := range apparats {
		if apparat == nil {
			continue
		}
		items = append(items, *apparat)
	}

	c.JSON(http.StatusOK, dto.ApparatBulkResponse{Items: toApparatResponses(items)})
}

// ListApparats godoc
// @Summary List apparats with pagination
// @Tags facility-apparats
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Param search query string false "Search query"
// @Param object_data_id query string false "Filter by Object Data ID"
// @Param system_part_id query string false "Filter by System Part ID"
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
	systemPartIDStr := c.Query("system_part_id")

	filtering := objectDataIDStr != "" || systemPartIDStr != ""
	if filtering {
		var byObjectData []uuid.UUID
		var bySystemPart []uuid.UUID

		if objectDataIDStr != "" {
			objectDataID, err := parseUUIDString(objectDataIDStr)
			if err != nil {
				respondLocalizedError(c, http.StatusBadRequest, "invalid_object_data_id", "facility.invalid_object_data_id")
				return
			}

			ids, err := h.getApparatsForObjectData(objectDataID)
			if err != nil {
				if respondLocalizedNotFoundIf(c, err, "facility.apparat_not_found") {
					return
				}
				respondLocalizedError(c, http.StatusInternalServerError, "fetch_failed", "facility.fetch_failed")
				return
			}
			byObjectData = ids
		}

		if systemPartIDStr != "" {
			systemPartID, err := parseUUIDString(systemPartIDStr)
			if err != nil {
				respondLocalizedError(c, http.StatusBadRequest, "invalid_system_part_id", "facility.invalid_system_part_id")
				return
			}

			ids, err := h.systemPartService.GetApparatIDs(systemPartID)
			if err != nil {
				if respondLocalizedNotFoundIf(c, err, "facility.apparat_not_found") {
					return
				}
				respondLocalizedError(c, http.StatusInternalServerError, "fetch_failed", "facility.fetch_failed")
				return
			}
			bySystemPart = ids
		}

		finalIDs := byObjectData
		if objectDataIDStr == "" {
			finalIDs = bySystemPart
		} else if systemPartIDStr != "" {
			set := map[uuid.UUID]struct{}{}
			for _, id := range bySystemPart {
				set[id] = struct{}{}
			}
			out := make([]uuid.UUID, 0, len(byObjectData))
			for _, id := range byObjectData {
				if _, ok := set[id]; ok {
					out = append(out, id)
				}
			}
			finalIDs = out
		}

		if len(finalIDs) == 0 {
			c.JSON(http.StatusOK, dto.ApparatListResponse{Items: []dto.ApparatResponse{}, TotalPages: 0, Page: 1, Total: 0})
			return
		}

		apparats, err := h.service.GetByIDs(finalIDs)
		if err != nil {
			respondLocalizedError(c, http.StatusInternalServerError, "fetch_failed", "facility.fetch_failed")
			return
		}

		filtered := make([]*domainFacility.Apparat, 0, len(apparats))
		for _, apparat := range apparats {
			if apparat == nil {
				continue
			}
			if query.Search == "" ||
				containsIgnoreCase(apparat.ShortName, query.Search) ||
				containsIgnoreCase(apparat.Name, query.Search) ||
				(apparat.Description != nil && containsIgnoreCase(*apparat.Description, query.Search)) {
				filtered = append(filtered, apparat)
			}
		}

		total := len(filtered)
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

		pageItems := filtered[start:end]
		responses := make([]dto.ApparatResponse, 0, len(pageItems))
		for _, a := range pageItems {
			responses = append(responses, toApparatResponse(*a))
		}

		c.JSON(http.StatusOK, dto.ApparatListResponse{Items: responses, TotalPages: totalPages, Page: query.Page, Total: int64(total)})
		return
	}

	result, err := h.service.List(query.Page, query.Limit, query.Search)
	if err != nil {
		respondLocalizedError(c, http.StatusInternalServerError, "fetch_failed", "facility.fetch_failed")
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
		if respondLocalizedNotFoundIf(c, err, "facility.apparat_not_found") {
			return
		}
		respondLocalizedError(c, http.StatusInternalServerError, "fetch_failed", "facility.fetch_failed")
		return
	}

	// Load system parts if IDs are provided
	var systemParts *[]*domainFacility.SystemPart
	if req.SystemPartIDs != nil {
		if len(*req.SystemPartIDs) > 0 {
			loadedParts, err := h.systemPartService.GetByIDs(*req.SystemPartIDs)
			if err != nil {
				respondLocalizedError(c, http.StatusBadRequest, "invalid_system_parts", "facility.invalid_system_parts")
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

	if err := h.service.Update(apparat); respondLocalizedValidationOrError(c, err, "facility.update_failed") {
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
		respondLocalizedError(c, http.StatusInternalServerError, "deletion_failed", "facility.deletion_failed")
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
