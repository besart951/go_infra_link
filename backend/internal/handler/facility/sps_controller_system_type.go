package facility

import (
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
		respondError(c, http.StatusInternalServerError, "fetch_failed", err.Error())
		return
	}

	c.JSON(http.StatusOK, toSPSControllerSystemTypeListResponse(result))
}
