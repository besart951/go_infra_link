package facility

import (
	"net/http"

	"github.com/besart951/go_infra_link/backend/internal/handler/dto"
	"github.com/gin-gonic/gin"
)

type StateTextHandler struct {
	service StateTextService
}

func NewStateTextHandler(service StateTextService) *StateTextHandler {
	return &StateTextHandler{service: service}
}

// CreateStateText godoc
// @Summary Create a new state text
// @Tags facility-state-texts
// @Accept json
// @Produce json
// @Param state_text body dto.CreateStateTextRequest true "State Text data"
// @Success 201 {object} dto.StateTextResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/facility/state-texts [post]
func (h *StateTextHandler) CreateStateText(c *gin.Context) {
	var req dto.CreateStateTextRequest
	if !bindJSON(c, &req) {
		return
	}

	stateText := toStateTextModel(req)

	if err := h.service.Create(stateText); respondLocalizedValidationOrError(c, err, "facility.creation_failed") {
		return
	}

	c.JSON(http.StatusCreated, toStateTextResponse(*stateText))
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
		if respondLocalizedNotFoundIf(c, err, "facility.state_text_not_found") {
			return
		}
		respondLocalizedError(c, http.StatusInternalServerError, "fetch_failed", "facility.fetch_failed")
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
		respondLocalizedError(c, http.StatusInternalServerError, "fetch_failed", "facility.fetch_failed")
		return
	}

	c.JSON(http.StatusOK, toStateTextListResponse(result))
}

// UpdateStateText godoc
// @Summary Update a state text
// @Tags facility-state-texts
// @Accept json
// @Produce json
// @Param id path string true "State Text ID"
// @Param state_text body dto.UpdateStateTextRequest true "State Text data"
// @Success 200 {object} dto.StateTextResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/facility/state-texts/{id} [put]
func (h *StateTextHandler) UpdateStateText(c *gin.Context) {
	id, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}

	var req dto.UpdateStateTextRequest
	if !bindJSON(c, &req) {
		return
	}

	stateText, err := h.service.GetByID(id)
	if err != nil {
		if respondLocalizedNotFoundIf(c, err, "facility.state_text_not_found") {
			return
		}
		respondLocalizedError(c, http.StatusInternalServerError, "fetch_failed", "facility.fetch_failed")
		return
	}

	applyStateTextUpdate(stateText, req)

	if err := h.service.Update(stateText); respondLocalizedValidationOrError(c, err, "facility.update_failed") {
		return
	}

	c.JSON(http.StatusOK, toStateTextResponse(*stateText))
}

// DeleteStateText godoc
// @Summary Delete a state text
// @Tags facility-state-texts
// @Produce json
// @Param id path string true "State Text ID"
// @Success 204
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/facility/state-texts/{id} [delete]
func (h *StateTextHandler) DeleteStateText(c *gin.Context) {
	id, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}

	if err := h.service.DeleteByID(id); err != nil {
		respondLocalizedError(c, http.StatusInternalServerError, "deletion_failed", "facility.deletion_failed")
		return
	}

	c.Status(http.StatusNoContent)
}
