package facility

import (
	"net/http"
	"strings"

	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/besart951/go_infra_link/backend/internal/handler/dto"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type SystemPartHandler struct {
	service           SystemPartService
	apparatService    ApparatService
	objectDataService ObjectDataService
}

func NewSystemPartHandler(service SystemPartService, apparatService ApparatService, objectDataService ObjectDataService) *SystemPartHandler {
	return &SystemPartHandler{
		service:           service,
		apparatService:    apparatService,
		objectDataService: objectDataService,
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

	if err := h.service.Create(systemPart); respondLocalizedValidationOrError(c, err, "facility.creation_failed") {
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
		if respondLocalizedNotFoundIf(c, err, "facility.system_part_not_found") {
			return
		}
		respondLocalizedError(c, http.StatusInternalServerError, "fetch_failed", "facility.fetch_failed")
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
// @Param object_data_id query string false "Filter by Object Data ID"
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
	objectDataIDStr := c.Query("object_data_id")

	filtering := apparatIDStr != "" || objectDataIDStr != ""
	if filtering {
		var byApparat []uuid.UUID
		var byObjectData []uuid.UUID

		if apparatIDStr != "" {
			apparatID, err := parseUUIDString(apparatIDStr)
			if err != nil {
				respondLocalizedError(c, http.StatusBadRequest, "invalid_apparat_id", "facility.invalid_apparat_id")
				return
			}
			ids, err := h.apparatService.GetSystemPartIDs(apparatID)
			if err != nil {
				if respondLocalizedNotFoundIf(c, err, "facility.system_part_not_found") {
					return
				}
				respondLocalizedError(c, http.StatusInternalServerError, "fetch_failed", "facility.fetch_failed")
				return
			}
			byApparat = ids
		}

		if objectDataIDStr != "" {
			objectDataID, err := parseUUIDString(objectDataIDStr)
			if err != nil {
				respondLocalizedError(c, http.StatusBadRequest, "invalid_object_data_id", "facility.invalid_object_data_id")
				return
			}

			apparatIDs, err := h.objectDataService.GetApparatIDs(objectDataID)
			if err != nil {
				if respondLocalizedNotFoundIf(c, err, "facility.system_part_not_found") {
					return
				}
				respondLocalizedError(c, http.StatusInternalServerError, "fetch_failed", "facility.fetch_failed")
				return
			}

			idSet := map[uuid.UUID]struct{}{}
			for _, aid := range apparatIDs {
				sysIDs, err := h.apparatService.GetSystemPartIDs(aid)
				if err != nil {
					continue
				}
				for _, sid := range sysIDs {
					idSet[sid] = struct{}{}
				}
			}
			byObjectData = make([]uuid.UUID, 0, len(idSet))
			for id := range idSet {
				byObjectData = append(byObjectData, id)
			}
		}

		// Combine filters: intersection if both present, else whichever exists
		finalIDs := byApparat
		if apparatIDStr == "" {
			finalIDs = byObjectData
		} else if objectDataIDStr != "" {
			set := map[uuid.UUID]struct{}{}
			for _, id := range byObjectData {
				set[id] = struct{}{}
			}
			out := make([]uuid.UUID, 0, len(byApparat))
			for _, id := range byApparat {
				if _, ok := set[id]; ok {
					out = append(out, id)
				}
			}
			finalIDs = out
		}

		if len(finalIDs) == 0 {
			c.JSON(http.StatusOK, dto.SystemPartListResponse{Items: []dto.SystemPartResponse{}, TotalPages: 0, Page: 1, Total: 0})
			return
		}

		systemParts, err := h.service.GetByIDs(finalIDs)
		if err != nil {
			respondLocalizedError(c, http.StatusInternalServerError, "fetch_failed", "facility.fetch_failed")
			return
		}

		filtered := make([]*domainFacility.SystemPart, 0, len(systemParts))
		for _, sp := range systemParts {
			if sp == nil {
				continue
			}
			if query.Search == "" ||
				strings.Contains(strings.ToLower(sp.ShortName), strings.ToLower(query.Search)) ||
				strings.Contains(strings.ToLower(sp.Name), strings.ToLower(query.Search)) ||
				(sp.Description != nil && strings.Contains(strings.ToLower(*sp.Description), strings.ToLower(query.Search))) {
				filtered = append(filtered, sp)
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
		responses := make([]dto.SystemPartResponse, 0, len(pageItems))
		for _, sp := range pageItems {
			responses = append(responses, toSystemPartResponse(*sp))
		}

		c.JSON(http.StatusOK, dto.SystemPartListResponse{Items: responses, TotalPages: totalPages, Page: query.Page, Total: int64(total)})
		return
	}

	result, err := h.service.List(query.Page, query.Limit, query.Search)
	if err != nil {
		respondLocalizedError(c, http.StatusInternalServerError, "fetch_failed", "facility.fetch_failed")
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
		if respondLocalizedNotFoundIf(c, err, "facility.system_part_not_found") {
			return
		}
		respondLocalizedError(c, http.StatusInternalServerError, "fetch_failed", "facility.fetch_failed")
		return
	}

	applySystemPartUpdate(systemPart, req)

	if err := h.service.Update(systemPart); respondLocalizedValidationOrError(c, err, "facility.update_failed") {
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
		respondLocalizedError(c, http.StatusInternalServerError, "deletion_failed", "facility.deletion_failed")
		return
	}

	c.Status(http.StatusNoContent)
}
