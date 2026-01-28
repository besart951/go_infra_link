package handler

import (
	"errors"
	"net/http"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	"github.com/besart951/go_infra_link/backend/internal/domain/project"
	"github.com/besart951/go_infra_link/backend/internal/handler/dto"
	"github.com/besart951/go_infra_link/backend/internal/handler/mapper"
	"github.com/besart951/go_infra_link/backend/internal/handlerutil"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type PhaseService interface {
	Create(phase *project.Phase) error
	GetByID(id uuid.UUID) (*project.Phase, error)
	List(page, limit int, search string) (*domain.PaginatedList[project.Phase], error)
	Update(phase *project.Phase) error
	DeleteByID(id uuid.UUID) error
}

type PhaseHandler struct {
	service PhaseService
}

func NewPhaseHandler(service PhaseService) *PhaseHandler {
	return &PhaseHandler{service: service}
}

// CreatePhase godoc
// @Summary Create a new phase
// @Tags phases
// @Accept json
// @Produce json
// @Param phase body dto.CreatePhaseRequest true "Phase data"
// @Success 201 {object} dto.PhaseResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/phases [post]
func (h *PhaseHandler) CreatePhase(c *gin.Context) {
	var req dto.CreatePhaseRequest
	if !handlerutil.BindJSON(c, &req) {
		return
	}

	phase := mapper.ToPhaseModel(req)

	if err := h.service.Create(phase); err != nil {
		handlerutil.RespondError(c, http.StatusInternalServerError, "creation_failed", err.Error())
		return
	}

	c.JSON(http.StatusCreated, mapper.ToPhaseResponse(phase))
}

// GetPhase godoc
// @Summary Get a phase by ID
// @Tags phases
// @Produce json
// @Param id path string true "Phase ID"
// @Success 200 {object} dto.PhaseResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/phases/{id} [get]
func (h *PhaseHandler) GetPhase(c *gin.Context) {
	id, ok := handlerutil.ParseUUIDParam(c, "id")
	if !ok {
		return
	}

	phase, err := h.service.GetByID(id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			handlerutil.RespondNotFound(c, "Phase not found")
			return
		}
		handlerutil.RespondError(c, http.StatusInternalServerError, "fetch_failed", err.Error())
		return
	}

	c.JSON(http.StatusOK, mapper.ToPhaseResponse(phase))
}

// ListPhases godoc
// @Summary List phases with pagination
// @Tags phases
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Param search query string false "Search query"
// @Success 200 {object} dto.PhaseListResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/phases [get]
func (h *PhaseHandler) ListPhases(c *gin.Context) {
	var query dto.PaginationQuery
	if !handlerutil.BindQuery(c, &query) {
		return
	}

	result, err := h.service.List(query.Page, query.Limit, query.Search)
	if err != nil {
		handlerutil.RespondError(c, http.StatusInternalServerError, "fetch_failed", err.Error())
		return
	}

	response := dto.PhaseListResponse{
		Items:      mapper.ToPhaseListResponse(result.Items),
		Total:      result.Total,
		Page:       result.Page,
		TotalPages: result.TotalPages,
	}

	c.JSON(http.StatusOK, response)
}

// UpdatePhase godoc
// @Summary Update a phase
// @Tags phases
// @Accept json
// @Produce json
// @Param id path string true "Phase ID"
// @Param phase body dto.UpdatePhaseRequest true "Phase data"
// @Success 200 {object} dto.PhaseResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/phases/{id} [put]
func (h *PhaseHandler) UpdatePhase(c *gin.Context) {
	id, ok := handlerutil.ParseUUIDParam(c, "id")
	if !ok {
		return
	}

	var req dto.UpdatePhaseRequest
	if !handlerutil.BindJSON(c, &req) {
		return
	}

	phase, err := h.service.GetByID(id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			handlerutil.RespondNotFound(c, "Phase not found")
			return
		}
		handlerutil.RespondError(c, http.StatusInternalServerError, "fetch_failed", err.Error())
		return
	}

	mapper.ApplyPhaseUpdate(phase, req)

	if err := h.service.Update(phase); err != nil {
		handlerutil.RespondError(c, http.StatusInternalServerError, "update_failed", err.Error())
		return
	}

	c.JSON(http.StatusOK, mapper.ToPhaseResponse(phase))
}

// DeletePhase godoc
// @Summary Delete a phase
// @Tags phases
// @Produce json
// @Param id path string true "Phase ID"
// @Success 204
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/phases/{id} [delete]
func (h *PhaseHandler) DeletePhase(c *gin.Context) {
	id, ok := handlerutil.ParseUUIDParam(c, "id")
	if !ok {
		return
	}

	if err := h.service.DeleteByID(id); err != nil {
		handlerutil.RespondError(c, http.StatusInternalServerError, "deletion_failed", err.Error())
		return
	}

	c.Status(http.StatusNoContent)
}
