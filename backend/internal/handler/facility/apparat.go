package facility

import (
	"net/http"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	dto "github.com/besart951/go_infra_link/backend/internal/handler/dto/facility"
	"github.com/gin-gonic/gin"
)

type ApparatHandler struct {
	service ApparatService
}

func NewApparatHandler(service ApparatService) *ApparatHandler {
	return &ApparatHandler{service: service}
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

	apparat := toApparatModel(req)
	if err := h.service.CreateWithSystemPartIDs(c.Request.Context(), apparat, req.SystemPartIDs); respondLocalizedDomainError(c, err, "creation_failed", "facility.creation_failed", localizedInvalidSystemParts()) {
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

	apparat, err := h.service.GetByID(c.Request.Context(), id)
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

	apparats, err := h.service.GetByIDs(c.Request.Context(), req.Ids)
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

	filters := domainFacility.ApparatFilterParams{}
	objectDataIDStr := c.Query("object_data_id")
	systemPartIDStr := c.Query("system_part_id")

	if objectDataIDStr != "" {
		objectDataID, err := parseUUIDString(objectDataIDStr)
		if err != nil {
			respondLocalizedError(c, http.StatusBadRequest, "invalid_object_data_id", "facility.invalid_object_data_id")
			return
		}
		filters.ObjectDataID = &objectDataID
	}

	if systemPartIDStr != "" {
		systemPartID, err := parseUUIDString(systemPartIDStr)
		if err != nil {
			respondLocalizedError(c, http.StatusBadRequest, "invalid_system_part_id", "facility.invalid_system_part_id")
			return
		}
		filters.SystemPartID = &systemPartID
	}

	result, err := h.service.ListWithFilters(c.Request.Context(), domain.PaginationParams{
		Page:   query.Page,
		Limit:  query.Limit,
		Search: query.Search,
	}, filters)
	if err != nil {
		if respondLocalizedDomainError(c, err, "fetch_failed", "facility.fetch_failed", localizedNotFound("facility.apparat_not_found")) {
			return
		}
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

	ctx := c.Request.Context()

	apparat, err := h.service.GetByID(ctx, id)
	if err != nil {
		if respondLocalizedNotFoundIf(c, err, "facility.apparat_not_found") {
			return
		}
		respondLocalizedError(c, http.StatusInternalServerError, "fetch_failed", "facility.fetch_failed")
		return
	}

	applyApparatUpdate(apparat, req)

	if err := h.service.UpdateWithSystemPartIDs(ctx, apparat, req.SystemPartIDs); respondLocalizedDomainError(c, err, "update_failed", "facility.update_failed", localizedInvalidSystemParts()) {
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

	if err := h.service.DeleteByID(c.Request.Context(), id); err != nil {
		respondLocalizedDomainError(c, err, "deletion_failed", "facility.deletion_failed",
			localizedNotFound("facility.apparat_not_found"),
		)
		return
	}

	c.Status(http.StatusNoContent)
}
