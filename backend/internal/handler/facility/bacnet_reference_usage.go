package facility

import (
	"net/http"

	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	dto "github.com/besart951/go_infra_link/backend/internal/handler/dto/facility"
	"github.com/gin-gonic/gin"
)

type BacnetReferenceUsageHandler struct {
	service BacnetReferenceUsageService
}

func NewBacnetReferenceUsageHandler(service BacnetReferenceUsageService) *BacnetReferenceUsageHandler {
	return &BacnetReferenceUsageHandler{service: service}
}

// GetBacnetReferenceUsages godoc
// @Summary Count BACnet object usage for reference data
// @Tags facility-bacnet-reference-usages
// @Produce json
// @Param resource query string true "Reference resource"
// @Param ids query []string true "Reference IDs"
// @Success 200 {object} dto.BacnetReferenceUsageListResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/facility/bacnet-reference-usages [get]
func (h *BacnetReferenceUsageHandler) GetBacnetReferenceUsages(c *gin.Context) {
	resource, ok := domainFacility.ParseBacnetReferenceResource(c.Query("resource"))
	if !ok {
		respondLocalizedInvalidArgument(c, "facility.invalid_bacnet_reference_resource")
		return
	}

	ids, ok := parseUUIDListQueryParam(c, "ids")
	if !ok {
		return
	}

	items, err := h.service.CountByResource(c.Request.Context(), resource, ids)
	if err != nil {
		respondLocalizedDomainError(c, err, "fetch_failed", "facility.fetch_failed")
		return
	}

	responses := make([]dto.BacnetReferenceUsageResponse, 0, len(items))
	for _, item := range items {
		responses = append(responses, dto.BacnetReferenceUsageResponse{
			Resource:          string(item.Resource),
			ID:                item.ID,
			BacnetObjectCount: item.BacnetObjectCount,
		})
	}
	c.JSON(http.StatusOK, dto.BacnetReferenceUsageListResponse{Items: responses})
}
