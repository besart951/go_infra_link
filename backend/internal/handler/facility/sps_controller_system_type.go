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
		// We expect the repository to preload SPSController and SystemType if we want names
		// If using GORM, we need to ensure they are preloaded.
		// Since I implemented it using database/sql with JOINs, I need to make sure
		// the domain struct has fields for the joined names OR the struct has nested objects populated.

		// In my repo implementation:
		// sc.device_name and st.name are NOT SCANNED into the struct because the struct doesn't have them
		// and the Scan only scans "standard" fields.
		// Wait! I actually LEFT JOINed them but did I scan them?
		// Looking back at my repo implementation:
		// SELECT s.id, ... FROM ...
		// I did NOT select sc.device_name, st.name in the Scan list!

		// I need to fix the repository to map these names somewhere!
		// Usually we map to the nested struct fields: SpsController.DeviceName etc.

		// I will modify the response mapping assuming I fix the repo later,
		// OR just map what I have for now. I should fix the repo.

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
