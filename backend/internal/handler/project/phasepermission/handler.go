package phasepermission

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
	Create(ctx context.Context, rule *domainProject.PhasePermission) error
	GetByID(ctx context.Context, id uuid.UUID) (*domainProject.PhasePermission, error)
	List(ctx context.Context, phaseID *uuid.UUID) ([]domainProject.PhasePermission, error)
	Update(ctx context.Context, rule *domainProject.PhasePermission) error
	DeleteByID(ctx context.Context, id uuid.UUID) error
}

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) CreatePhasePermission(c *gin.Context) {
	var req dto.CreatePhasePermissionRequest
	if !handlerutil.BindJSON(c, &req) {
		return
	}

	rule := &domainProject.PhasePermission{
		PhaseID:     req.PhaseID,
		Role:        req.Role,
		Permissions: permissionsFromCreateRequest(req),
	}

	if err := h.service.Create(c.Request.Context(), rule); err != nil {
		respondPhasePermissionError(c, err, "creation_failed", "phase_permission.creation_failed")
		return
	}

	c.JSON(http.StatusCreated, toResponse(rule))
}

func (h *Handler) GetPhasePermission(c *gin.Context) {
	id, ok := handlerutil.ParseUUIDParam(c, "id")
	if !ok {
		return
	}

	rule, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		respondPhasePermissionError(c, err, "fetch_failed", "phase_permission.fetch_failed")
		return
	}

	c.JSON(http.StatusOK, toResponse(rule))
}

func (h *Handler) ListPhasePermissions(c *gin.Context) {
	var phaseID *uuid.UUID
	if raw := c.Query("phase_id"); raw != "" {
		parsed, err := uuid.Parse(raw)
		if err != nil {
			handlerutil.RespondLocalizedError(c, http.StatusBadRequest, "invalid_phase_id", "phase_permission.fetch_failed")
			return
		}
		phaseID = &parsed
	}

	rules, err := h.service.List(c.Request.Context(), phaseID)
	if err != nil {
		respondPhasePermissionError(c, err, "fetch_failed", "phase_permission.fetch_failed")
		return
	}

	c.JSON(http.StatusOK, dto.PhasePermissionListResponse{Items: toListResponse(rules)})
}

func (h *Handler) UpdatePhasePermission(c *gin.Context) {
	id, ok := handlerutil.ParseUUIDParam(c, "id")
	if !ok {
		return
	}

	var req dto.UpdatePhasePermissionRequest
	if !handlerutil.BindJSON(c, &req) {
		return
	}

	ctx := c.Request.Context()
	rule, err := h.service.GetByID(ctx, id)
	if err != nil {
		respondPhasePermissionError(c, err, "fetch_failed", "phase_permission.fetch_failed")
		return
	}

	applyUpdate(rule, req)
	if err := h.service.Update(ctx, rule); err != nil {
		respondPhasePermissionError(c, err, "update_failed", "phase_permission.update_failed")
		return
	}

	c.JSON(http.StatusOK, toResponse(rule))
}

func (h *Handler) DeletePhasePermission(c *gin.Context) {
	id, ok := handlerutil.ParseUUIDParam(c, "id")
	if !ok {
		return
	}

	if err := h.service.DeleteByID(c.Request.Context(), id); err != nil {
		respondPhasePermissionError(c, err, "deletion_failed", "phase_permission.deletion_failed")
		return
	}

	c.Status(http.StatusNoContent)
}

func permissionsFromCreateRequest(req dto.CreatePhasePermissionRequest) []string {
	if req.Permissions != nil {
		return req.Permissions
	}
	if req.Permission != nil {
		return []string{*req.Permission}
	}
	return []string{}
}

func applyUpdate(rule *domainProject.PhasePermission, req dto.UpdatePhasePermissionRequest) {
	if req.PhaseID != nil {
		rule.PhaseID = *req.PhaseID
	}
	if req.Role != nil {
		rule.Role = *req.Role
	}
	if req.Permissions != nil {
		rule.Permissions = *req.Permissions
	} else if req.Permission != nil {
		rule.Permissions = []string{*req.Permission}
	}
}

func toResponse(rule *domainProject.PhasePermission) dto.PhasePermissionResponse {
	return dto.PhasePermissionResponse{
		ID:          rule.ID,
		PhaseID:     rule.PhaseID,
		Role:        rule.Role,
		Permissions: append([]string{}, rule.Permissions...),
		CreatedAt:   rule.CreatedAt,
		UpdatedAt:   rule.UpdatedAt,
	}
}

func toListResponse(rules []domainProject.PhasePermission) []dto.PhasePermissionResponse {
	out := make([]dto.PhasePermissionResponse, len(rules))
	for i := range rules {
		out[i] = toResponse(&rules[i])
	}
	return out
}

func respondPhasePermissionError(c *gin.Context, err error, fallbackCode, fallbackKey string) {
	handlerutil.RespondDomainError(c, err,
		handlerutil.LocalizedError(http.StatusInternalServerError, fallbackCode, fallbackKey),
		handlerutil.MapError(domain.ErrNotFound, handlerutil.LocalizedError(http.StatusNotFound, "not_found", "phase_permission.permission_not_found")),
		handlerutil.MapError(domain.ErrConflict, handlerutil.LocalizedError(http.StatusConflict, "conflict", fallbackKey)),
		handlerutil.MapError(domain.ErrInvalidArgument, handlerutil.LocalizedError(http.StatusBadRequest, "validation_error", fallbackKey)),
	)
}
