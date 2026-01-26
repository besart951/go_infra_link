package facility

import (
	"errors"
	"net/http"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	"github.com/besart951/go_infra_link/backend/internal/handler/dto"
	"github.com/gin-gonic/gin"
)

type AlarmDefinitionHandler struct {
	service AlarmDefinitionService
}

func NewAlarmDefinitionHandler(service AlarmDefinitionService) *AlarmDefinitionHandler {
	return &AlarmDefinitionHandler{service: service}
}

// @Router /api/v1/facility/alarm-definitions/{id} [get]
func (h *AlarmDefinitionHandler) GetAlarmDefinition(c *gin.Context) {
	id, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}

	alarmDef, err := h.service.GetByID(id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			respondNotFound(c, "Alarm definition not found")
			return
		}
		respondError(c, http.StatusInternalServerError, "fetch_failed", err.Error())
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
	if !bindQuery(c, &query) {
		return
	}

	result, err := h.service.List(query.Page, query.Limit, query.Search)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "fetch_failed", err.Error())
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
