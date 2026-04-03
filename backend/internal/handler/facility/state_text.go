package facility

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	dto "github.com/besart951/go_infra_link/backend/internal/handler/dto/facility"
	"github.com/gin-gonic/gin"
)

type StateTextHandler struct {
	crud crudHandler[domainFacility.StateText, dto.CreateStateTextRequest, dto.UpdateStateTextRequest]
}

func NewStateTextHandler(svc StateTextService) *StateTextHandler {
	return &StateTextHandler{crud: newCRUD(
		svc,
		toStateTextModel,
		applyStateTextUpdate,
		respFn(toStateTextResponse),
		listRespFn(toStateTextListResponse),
		"facility.state_text_not_found",
	)}
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
func (h *StateTextHandler) CreateStateText(c *gin.Context) { h.crud.handleCreate(c) }

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
func (h *StateTextHandler) GetStateText(c *gin.Context) { h.crud.handleGetByID(c) }

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
func (h *StateTextHandler) ListStateTexts(c *gin.Context) { h.crud.handleList(c) }

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
func (h *StateTextHandler) UpdateStateText(c *gin.Context) { h.crud.handleUpdate(c) }

// DeleteStateText godoc
// @Summary Delete a state text
// @Tags facility-state-texts
// @Produce json
// @Param id path string true "State Text ID"
// @Success 204
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/facility/state-texts/{id} [delete]
func (h *StateTextHandler) DeleteStateText(c *gin.Context) { h.crud.handleDelete(c) }

// ensure domain and dto imports are used (via type parameters above)
var _ = (*domain.PaginatedList[domainFacility.StateText])(nil)
