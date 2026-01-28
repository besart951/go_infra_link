package facility

import (
	"net/http"

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
// @Success 200 {object} StateTextResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/facility/state-texts/{id} [get]
func (h *StateTextHandler) GetStateText(c *gin.Context) {
	id, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}

	stateText, err := h.service.GetByID(id)
	if err != nil {
		if respondNotFoundIf(c, err, "State text not found") {
			return
		}
		respondError(c, http.StatusInternalServerError, "fetch_failed", err.Error())
		return
	}

	c.JSON(http.StatusOK, toStateTextResponse(*stateText))
}

// ListStateTexts godoc
// @Summary List state texts with pagination
// @Tags facility-state-texts
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Param search query string false "Search query"
// @Success 200 {object} StateTextListResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/facility/state-texts [get]
func (h *StateTextHandler) ListStateTexts(c *gin.Context) {
	query, ok := parsePaginationQuery(c)
	if !ok {
		return
	}

	result, err := h.service.List(query.Page, query.Limit, query.Search)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "fetch_failed", err.Error())
		return
	}

	c.JSON(http.StatusOK, toStateTextListResponse(result))
}
