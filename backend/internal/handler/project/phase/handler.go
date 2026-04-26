package phase

import (
	"context"
	"net/http"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainProject "github.com/besart951/go_infra_link/backend/internal/domain/project"
	dto "github.com/besart951/go_infra_link/backend/internal/handler/dto/project"
	"github.com/besart951/go_infra_link/backend/internal/handlerutil"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Service interface {
	Create(ctx context.Context, phase *domainProject.Phase) error
	GetByID(ctx context.Context, id uuid.UUID) (*domainProject.Phase, error)
	List(ctx context.Context, page, limit int, search string) (*domain.PaginatedList[domainProject.Phase], error)
	Update(ctx context.Context, phase *domainProject.Phase) error
	DeleteByID(ctx context.Context, id uuid.UUID) error
}

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
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
func (h *Handler) CreatePhase(c *gin.Context) {
	var req dto.CreatePhaseRequest
	if !handlerutil.BindJSON(c, &req) {
		return
	}

	phase := toModel(req)

	if err := h.service.Create(c.Request.Context(), phase); err != nil {
		handlerutil.RespondDomainError(c, err,
			handlerutil.LocalizedError(http.StatusInternalServerError, "creation_failed", "phase.creation_failed"),
			handlerutil.MapError(domain.ErrConflict, handlerutil.LocalizedError(http.StatusConflict, "conflict", "phase.creation_failed")),
			handlerutil.MapError(domain.ErrInvalidArgument, handlerutil.LocalizedError(http.StatusBadRequest, "validation_error", "phase.creation_failed")),
		)
		return
	}

	c.JSON(http.StatusCreated, toResponse(phase))
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
func (h *Handler) GetPhase(c *gin.Context) {
	id, ok := handlerutil.ParseUUIDParam(c, "id")
	if !ok {
		return
	}

	phase, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		handlerutil.RespondDomainError(c, err,
			handlerutil.LocalizedError(http.StatusInternalServerError, "fetch_failed", "phase.fetch_failed"),
			handlerutil.MapError(domain.ErrNotFound, handlerutil.LocalizedError(http.StatusNotFound, "not_found", "phase.phase_not_found")),
		)
		return
	}

	c.JSON(http.StatusOK, toResponse(phase))
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
func (h *Handler) ListPhases(c *gin.Context) {
	var query dto.PaginationQuery
	if !handlerutil.BindQuery(c, &query) {
		return
	}

	result, err := h.service.List(c.Request.Context(), query.Page, query.Limit, query.Search)
	if err != nil {
		handlerutil.RespondLocalizedError(c, http.StatusInternalServerError, "fetch_failed", "phase.fetch_failed")
		return
	}

	response := dto.PhaseListResponse{
		Items:      toListResponse(result.Items),
		Total:      result.Total,
		Page:       result.Page,
		TotalPages: result.TotalPages,
	}

	c.JSON(http.StatusOK, response)
}

// UpdatePhase godoc
// @Summary Update a phase
// @Description PATCH-like update: omitted fields remain unchanged and present string fields are applied even when empty. PUT is kept for compatibility; PATCH is the preferred method.
// @Tags phases
// @Accept json
// @Produce json
// @Param id path string true "Phase ID"
// @Param phase body dto.UpdatePhaseRequest true "Partial phase data"
// @Success 200 {object} dto.PhaseResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/phases/{id} [patch]
// @Router /api/v1/phases/{id} [put]
func (h *Handler) UpdatePhase(c *gin.Context) {
	id, ok := handlerutil.ParseUUIDParam(c, "id")
	if !ok {
		return
	}

	var req dto.UpdatePhaseRequest
	if !handlerutil.BindJSON(c, &req) {
		return
	}

	ctx := c.Request.Context()

	phase, err := h.service.GetByID(ctx, id)
	if err != nil {
		handlerutil.RespondDomainError(c, err,
			handlerutil.LocalizedError(http.StatusInternalServerError, "fetch_failed", "phase.fetch_failed"),
			handlerutil.MapError(domain.ErrNotFound, handlerutil.LocalizedError(http.StatusNotFound, "not_found", "phase.phase_not_found")),
		)
		return
	}

	applyUpdate(phase, req)

	if err := h.service.Update(ctx, phase); err != nil {
		handlerutil.RespondDomainError(c, err,
			handlerutil.LocalizedError(http.StatusInternalServerError, "update_failed", "phase.update_failed"),
			handlerutil.MapError(domain.ErrNotFound, handlerutil.LocalizedError(http.StatusNotFound, "not_found", "phase.phase_not_found")),
			handlerutil.MapError(domain.ErrConflict, handlerutil.LocalizedError(http.StatusConflict, "conflict", "phase.update_failed")),
			handlerutil.MapError(domain.ErrInvalidArgument, handlerutil.LocalizedError(http.StatusBadRequest, "validation_error", "phase.update_failed")),
		)
		return
	}

	c.JSON(http.StatusOK, toResponse(phase))
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
func (h *Handler) DeletePhase(c *gin.Context) {
	id, ok := handlerutil.ParseUUIDParam(c, "id")
	if !ok {
		return
	}

	if err := h.service.DeleteByID(c.Request.Context(), id); err != nil {
		handlerutil.RespondDomainError(c, err,
			handlerutil.LocalizedError(http.StatusInternalServerError, "deletion_failed", "phase.deletion_failed"),
			handlerutil.MapError(domain.ErrNotFound, handlerutil.LocalizedError(http.StatusNotFound, "not_found", "phase.phase_not_found")),
		)
		return
	}

	c.Status(http.StatusNoContent)
}

func toModel(req dto.CreatePhaseRequest) *domainProject.Phase {
	return &domainProject.Phase{Name: req.Name}
}

func toResponse(phase *domainProject.Phase) dto.PhaseResponse {
	return dto.PhaseResponse{
		ID:        phase.ID,
		Name:      phase.Name,
		CreatedAt: phase.CreatedAt,
		UpdatedAt: phase.UpdatedAt,
	}
}

func toListResponse(phases []domainProject.Phase) []dto.PhaseResponse {
	result := make([]dto.PhaseResponse, len(phases))
	for i, phase := range phases {
		result[i] = dto.PhaseResponse{
			ID:        phase.ID,
			Name:      phase.Name,
			CreatedAt: phase.CreatedAt,
			UpdatedAt: phase.UpdatedAt,
		}
	}
	return result
}

func applyUpdate(phase *domainProject.Phase, req dto.UpdatePhaseRequest) {
	if req.Name != nil {
		phase.Name = *req.Name
	}
}