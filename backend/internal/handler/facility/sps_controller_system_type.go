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

	var result *domain.PaginatedList[domainFacility.SPSControllerSystemType]
	var err error
	if spsControllerID != nil {
		result, err = h.service.ListBySPSControllerID(*spsControllerID, query.Page, query.Limit, query.Search)
	} else {
		result, err = h.service.List(query.Page, query.Limit, query.Search)
	}
	if err != nil {
		respondLocalizedError(c, http.StatusInternalServerError, "fetch_failed", "facility.fetch_failed")
		return
	}

	c.JSON(http.StatusOK, toSPSControllerSystemTypeListResponse(result))
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

	copyEntity, err := h.service.CopyByID(id)
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
