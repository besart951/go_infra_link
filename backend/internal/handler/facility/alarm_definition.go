package facility

import (
	"errors"
	"net/http"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	"github.com/besart951/go_infra_link/backend/internal/handler/dto"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AlarmDefinitionHandler struct {
	service AlarmDefinitionService
}

func NewAlarmDefinitionHandler(service AlarmDefinitionService) *AlarmDefinitionHandler {
	return &AlarmDefinitionHandler{service: service}
}

// @Router /api/v1/facility/alarm-definitions/{id} [get]
func (h *AlarmDefinitionHandler) GetAlarmDefinition(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid UUID format",
		})
		return
	}

	alarmDef, err := h.service.GetByID(id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			c.JSON(http.StatusNotFound, dto.ErrorResponse{
				Error:   "not_found",
				Message: "Alarm definition not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "fetch_failed",
			Message: err.Error(),
		})
		return
	}

	response := dto.AlarmDefinitionResponse{
		ID:        alarmDef.ID,
		Name:      alarmDef.Name,
		AlarmNote: alarmDef.AlarmNote,
		CreatedAt: alarmDef.CreatedAt,
		UpdatedAt: alarmDef.UpdatedAt,
	}

	c.JSON(http.StatusOK, response)
}

// @Router /api/v1/facility/alarm-definitions [get]
func (h *AlarmDefinitionHandler) ListAlarmDefinitions(c *gin.Context) {
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

	items := make([]dto.AlarmDefinitionResponse, len(result.Items))
	for i, alarmDef := range result.Items {
		items[i] = dto.AlarmDefinitionResponse{
			ID:        alarmDef.ID,
			Name:      alarmDef.Name,
			AlarmNote: alarmDef.AlarmNote,
			CreatedAt: alarmDef.CreatedAt,
			UpdatedAt: alarmDef.UpdatedAt,
		}
	}

	response := dto.AlarmDefinitionListResponse{
		Items:      items,
		Total:      result.Total,
		Page:       result.Page,
		TotalPages: result.TotalPages,
	}

	c.JSON(http.StatusOK, response)
}
