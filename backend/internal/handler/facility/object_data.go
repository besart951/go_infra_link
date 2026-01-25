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
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid UUID format",
		})
		return
	}

	obj, err := h.service.GetByID(id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			c.JSON(http.StatusNotFound, dto.ErrorResponse{
				Error:   "not_found",
				Message: "Object data not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "fetch_failed",
			Message: err.Error(),
		})
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
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "validation_error",
			Message: err.Error(),
		})
		return
	}

	result, err := h.service.List(query.Page, query.Limit, query.Search)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "fetch_failed",
			Message: err.Error(),
		})
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
