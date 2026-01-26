package facility

import (
	"errors"
	"net/http"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	"github.com/besart951/go_infra_link/backend/internal/handler/dto"
	"github.com/gin-gonic/gin"
)

type StateTextHandler struct {
	service StateTextService
}

func NewStateTextHandler(service StateTextService) *StateTextHandler {
	return &StateTextHandler{service: service}
}

// GetStateText godoc
// @Summary Get a state text by ID
// @Tags facility-state-texts
// @Produce json
// @Param id path string true "State Text ID"
// @Success 200 {object} dto.StateTextResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/facility/state-texts/{id} [get]
func (h *StateTextHandler) GetStateText(c *gin.Context) {
	id, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}

	stateText, err := h.service.GetByID(id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			respondNotFound(c, "State text not found")
			return
		}
		respondError(c, http.StatusInternalServerError, "fetch_failed", err.Error())
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

// ListStateTexts godoc
// @Summary List state texts with pagination
// @Tags facility-state-texts
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Param search query string false "Search query"
// @Success 200 {object} dto.StateTextListResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/facility/state-texts [get]
func (h *StateTextHandler) ListStateTexts(c *gin.Context) {
	var query dto.PaginationQuery
	if !bindQuery(c, &query) {
		return
	}

	result, err := h.service.List(query.Page, query.Limit, query.Search)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "fetch_failed", err.Error())
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
