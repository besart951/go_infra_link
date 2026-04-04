package facility

import (
	"errors"
	"net/http"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/gin-gonic/gin"
)

type SPSControllerSystemTypeHandler struct {
	service SPSControllerSystemTypeService
}

func NewSPSControllerSystemTypeHandler(service SPSControllerSystemTypeService) *SPSControllerSystemTypeHandler {
	return &SPSControllerSystemTypeHandler{service: service}
}

// ListSPSControllerSystemTypes godoc
// @Summary List SPS controller system types with pagination
// @Tags facility-sps-controller-system-types
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Param search query string false "Search query"
// @Param sps_controller_id query string false "SPS Controller ID"
// @Param project_id query string false "Project ID (filter by project)"
// @Success 200 {object} SPSControllerSystemTypeListResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/facility/sps-controller-system-types [get]
func (h *SPSControllerSystemTypeHandler) ListSPSControllerSystemTypes(c *gin.Context) {
	query, ok := parsePaginationQuery(c)
	if !ok {
		return
	}

	spsControllerID, ok := parseUUIDQueryParam(c, "sps_controller_id")
	if !ok {
		return
	}

	projectID, ok := parseUUIDQueryParam(c, "project_id")
	if !ok {
		return
	}

	ctx := c.Request.Context()

	var result *domain.PaginatedList[domainFacility.SPSControllerSystemType]
	var err error
	if spsControllerID != nil {
		result, err = h.service.ListBySPSControllerID(ctx, *spsControllerID, query.Page, query.Limit, query.Search)
	} else if projectID != nil {
		result, err = h.service.ListByProjectID(ctx, *projectID, query.Page, query.Limit, query.Search)
	} else {
		result, err = h.service.List(ctx, query.Page, query.Limit, query.Search)
	}
	if err != nil {
		respondLocalizedError(c, http.StatusInternalServerError, "fetch_failed", "facility.fetch_failed")
		return
	}

	c.JSON(http.StatusOK, toSPSControllerSystemTypeListResponse(result))
}

// GetSPSControllerSystemType godoc
// @Summary Get an SPS controller system type by ID
// @Tags facility-sps-controller-system-types
// @Produce json
// @Param id path string true "SPS Controller System Type ID"
// @Success 200 {object} SPSControllerSystemTypeResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/facility/sps-controller-system-types/{id} [get]
func (h *SPSControllerSystemTypeHandler) GetSPSControllerSystemType(c *gin.Context) {
	id, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}

	item, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		if respondLocalizedNotFoundIf(c, err, "facility.sps_controller_system_type_not_found") {
			return
		}
		respondLocalizedError(c, http.StatusInternalServerError, "fetch_failed", "facility.fetch_failed")
		return
	}

	c.JSON(http.StatusOK, toSPSControllerSystemTypeResponse(*item))
}

// CopySPSControllerSystemType godoc
// @Summary Copy an SPS controller system type
// @Tags facility-sps-controller-system-types
// @Produce json
// @Param id path string true "SPS Controller System Type ID"
// @Success 201 {object} SPSControllerSystemTypeResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/facility/sps-controller-system-types/{id}/copy [post]
func (h *SPSControllerSystemTypeHandler) CopySPSControllerSystemType(c *gin.Context) {
	id, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}

	copyEntity, err := h.service.CopyByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			respondLocalizedNotFoundError(c, "facility.sps_controller_system_type_not_found")
			return
		}
		if ve, ok := domain.AsValidationError(err); ok {
			respondValidationError(c, ve.Fields)
			return
		}
		respondLocalizedError(c, http.StatusInternalServerError, "creation_failed", "facility.creation_failed")
		return
	}

	c.JSON(http.StatusCreated, toSPSControllerSystemTypeResponse(*copyEntity))
}

// DeleteSPSControllerSystemType godoc
// @Summary Delete an SPS controller system type
// @Tags facility-sps-controller-system-types
// @Produce json
// @Param id path string true "SPS Controller System Type ID"
// @Success 204
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/facility/sps-controller-system-types/{id} [delete]
func (h *SPSControllerSystemTypeHandler) DeleteSPSControllerSystemType(c *gin.Context) {
	id, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}

	if err := h.service.DeleteByID(c.Request.Context(), id); err != nil {
		if respondLocalizedNotFoundIf(c, err, "facility.sps_controller_system_type_not_found") {
			return
		}
		respondLocalizedError(c, http.StatusInternalServerError, "deletion_failed", "facility.deletion_failed")
		return
	}

	c.Status(http.StatusNoContent)
}
