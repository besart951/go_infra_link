package facility

import (
	"errors"
	"net/http"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	"github.com/besart951/go_infra_link/backend/internal/handler/dto"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ObjectDataHandler struct {
	service ObjectDataService
}

func NewObjectDataHandler(service ObjectDataService) *ObjectDataHandler {
	return &ObjectDataHandler{service: service}
}

// @Router /api/v1/facility/object-data/{id} [get]
func (h *ObjectDataHandler) GetObjectData(c *gin.Context) {
	id, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}

	obj, err := h.service.GetByID(id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			respondNotFound(c, "Object data not found")
			return
		}
		respondError(c, http.StatusInternalServerError, "fetch_failed", err.Error())
		return
	}

	response := dto.ObjectDataResponse{
		ID:          obj.ID,
		Description: obj.Description,
		Version:     obj.Version,
		IsActive:    obj.IsActive,
		ProjectID:   obj.ProjectID,
		CreatedAt:   obj.CreatedAt,
		UpdatedAt:   obj.UpdatedAt,
	}

	c.JSON(http.StatusOK, response)
}

// @Router /api/v1/facility/object-data [get]
func (h *ObjectDataHandler) ListObjectData(c *gin.Context) {
	var query dto.PaginationQuery
	if !bindQuery(c, &query) {
		return
	}

	result, err := h.service.List(query.Page, query.Limit, query.Search)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "fetch_failed", err.Error())
		return
	}

	items := make([]dto.ObjectDataResponse, len(result.Items))
	for i, obj := range result.Items {
		items[i] = dto.ObjectDataResponse{
			ID:          obj.ID,
			Description: obj.Description,
			Version:     obj.Version,
			IsActive:    obj.IsActive,
			ProjectID:   obj.ProjectID,
			CreatedAt:   obj.CreatedAt,
			UpdatedAt:   obj.UpdatedAt,
		}
	}

	response := dto.ObjectDataListResponse{
		Items:      items,
		Total:      result.Total,
		Page:       result.Page,
		TotalPages: result.TotalPages,
	}

	c.JSON(http.StatusOK, response)
}
