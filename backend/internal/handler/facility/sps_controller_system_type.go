package facility

import (
	"net/http"

	"github.com/besart951/go_infra_link/backend/internal/handler/dto"
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
// @Success 200 {object} dto.SPSControllerSystemTypeListResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/facility/sps-controller-system-types [get]
func (h *SPSControllerSystemTypeHandler) ListSPSControllerSystemTypes(c *gin.Context) {
	var query dto.PaginationQuery
	if !bindQuery(c, &query) {
		return
	}

	result, err := h.service.List(query.Page, query.Limit, query.Search)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "fetch_failed", err.Error())
		return
	}

	items := make([]dto.SPSControllerSystemTypeResponse, len(result.Items))
	for i, item := range result.Items {
		items[i] = dto.SPSControllerSystemTypeResponse{
			ID:                item.ID,
			SPSControllerID:   item.SPSControllerID,
			SystemTypeID:      item.SystemTypeID,
			SPSControllerName: item.SPSController.DeviceName, // This will be empty if not populated
			SystemTypeName:    item.SystemType.Name,          // This will be empty if not populated
			Number:            item.Number,
			DocumentName:      item.DocumentName,
			CreatedAt:         item.CreatedAt,
			UpdatedAt:         item.UpdatedAt,
		}
	}

	response := dto.SPSControllerSystemTypeListResponse{
		Items:      items,
		Total:      result.Total,
		Page:       result.Page,
		TotalPages: result.TotalPages,
	}

	c.JSON(http.StatusOK, response)
}
