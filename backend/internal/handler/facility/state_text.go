package facility

import (
	"errors"
	"net/http"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	"github.com/besart951/go_infra_link/backend/internal/handler/dto"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type StateTextHandler struct {
	service StateTextService
}

func NewStateTextHandler(service StateTextService) *StateTextHandler {
	return &StateTextHandler{service: service}
}

// @Router /api/v1/facility/state-texts/{id} [get]
func (h *StateTextHandler) GetStateText(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid UUID format",
		})
		return
	}

	stateText, err := h.service.GetByID(id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			c.JSON(http.StatusNotFound, dto.ErrorResponse{
				Error:   "not_found",
				Message: "State text not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "fetch_failed",
			Message: err.Error(),
		})
		return
	}

	response := dto.StateTextResponse{
		ID:          stateText.ID,
		RefNumber:   stateText.RefNumber,
		StateText1:  stateText.StateText1,
		StateText2:  stateText.StateText2,
		StateText3:  stateText.StateText3,
		StateText4:  stateText.StateText4,
		StateText5:  stateText.StateText5,
		StateText6:  stateText.StateText6,
		StateText7:  stateText.StateText7,
		StateText8:  stateText.StateText8,
		StateText9:  stateText.StateText9,
		StateText10: stateText.StateText10,
		StateText11: stateText.StateText11,
		StateText12: stateText.StateText12,
		StateText13: stateText.StateText13,
		StateText14: stateText.StateText14,
		StateText15: stateText.StateText15,
		StateText16: stateText.StateText16,
		CreatedAt:   stateText.CreatedAt,
		UpdatedAt:   stateText.UpdatedAt,
	}

	c.JSON(http.StatusOK, response)
}

// @Router /api/v1/facility/state-texts [get]
func (h *StateTextHandler) ListStateTexts(c *gin.Context) {
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

	items := make([]dto.StateTextResponse, len(result.Items))
	for i, stateText := range result.Items {
		items[i] = dto.StateTextResponse{
			ID:          stateText.ID,
			RefNumber:   stateText.RefNumber,
			StateText1:  stateText.StateText1,
			StateText2:  stateText.StateText2,
			StateText3:  stateText.StateText3,
			StateText4:  stateText.StateText4,
			StateText5:  stateText.StateText5,
			StateText6:  stateText.StateText6,
			StateText7:  stateText.StateText7,
			StateText8:  stateText.StateText8,
			StateText9:  stateText.StateText9,
			StateText10: stateText.StateText10,
			StateText11: stateText.StateText11,
			StateText12: stateText.StateText12,
			StateText13: stateText.StateText13,
			StateText14: stateText.StateText14,
			StateText15: stateText.StateText15,
			StateText16: stateText.StateText16,
			CreatedAt:   stateText.CreatedAt,
			UpdatedAt:   stateText.UpdatedAt,
		}
	}

	response := dto.StateTextListResponse{
		Items:      items,
		Total:      result.Total,
		Page:       result.Page,
		TotalPages: result.TotalPages,
	}

	c.JSON(http.StatusOK, response)
}
