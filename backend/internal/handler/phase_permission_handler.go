package handler

import (
	"errors"
	"net/http"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	"github.com/besart951/go_infra_link/backend/internal/domain/project"
	"github.com/besart951/go_infra_link/backend/internal/domain/user"
	"github.com/besart951/go_infra_link/backend/internal/handler/dto"
	"github.com/besart951/go_infra_link/backend/internal/handler/mapper"
	"github.com/besart951/go_infra_link/backend/internal/handlerutil"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type PhasePermissionService interface {
	Create(perm *project.PhasePermission) error
	GetByID(id uuid.UUID) (*project.PhasePermission, error)
	GetByPhaseAndRole(phaseID uuid.UUID, role user.Role) (*project.PhasePermission, error)
	ListByPhase(phaseID uuid.UUID) ([]project.PhasePermission, error)
	Update(perm *project.PhasePermission) error
	DeleteByID(id uuid.UUID) error
	DeleteByPhaseAndRole(phaseID uuid.UUID, role user.Role) error
}

type PhasePermissionHandler struct {
	service PhasePermissionService
}

func NewPhasePermissionHandler(service PhasePermissionService) *PhasePermissionHandler {
	return &PhasePermissionHandler{service: service}
}

// CreatePhasePermission godoc
// @Summary Create a new phase permission
// @Tags phase-permissions
// @Accept json
// @Produce json
// @Param permission body dto.CreatePhasePermissionRequest true "Permission data"
// @Success 201 {object} dto.PhasePermissionResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/phase-permissions [post]
func (h *PhasePermissionHandler) CreatePhasePermission(c *gin.Context) {
	var req dto.CreatePhasePermissionRequest
	if !handlerutil.BindJSON(c, &req) {
		return
	}

	perm := mapper.ToPhasePermissionModel(req)

	if err := h.service.Create(perm); err != nil {
		handlerutil.RespondLocalizedError(c, http.StatusInternalServerError, "creation_failed", "phase_permission.creation_failed")
		return
	}

	c.JSON(http.StatusCreated, mapper.ToPhasePermissionResponse(perm))
}

// GetPhasePermission godoc
// @Summary Get a phase permission by ID
// @Tags phase-permissions
// @Produce json
// @Param id path string true "Permission ID"
// @Success 200 {object} dto.PhasePermissionResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/phase-permissions/{id} [get]
func (h *PhasePermissionHandler) GetPhasePermission(c *gin.Context) {
	id, ok := handlerutil.ParseUUIDParam(c, "id")
	if !ok {
		return
	}

	perm, err := h.service.GetByID(id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			handlerutil.RespondLocalizedError(c, http.StatusNotFound, "not_found", "phase_permission.permission_not_found")
			return
		}
		handlerutil.RespondLocalizedError(c, http.StatusInternalServerError, "fetch_failed", "phase_permission.fetch_failed")
		return
	}

	c.JSON(http.StatusOK, mapper.ToPhasePermissionResponse(perm))
}

// ListPhasePermissions godoc
// @Summary List permissions for a specific phase
// @Tags phase-permissions
// @Produce json
// @Param phase_id query string true "Phase ID"
// @Success 200 {object} dto.PhasePermissionListResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/phase-permissions [get]
func (h *PhasePermissionHandler) ListPhasePermissions(c *gin.Context) {
	phaseIDStr := c.Query("phase_id")
	if phaseIDStr == "" {
		handlerutil.RespondLocalizedError(c, http.StatusBadRequest, "invalid_input", "validation.phase_id_required")
		return
	}

	phaseID, err := uuid.Parse(phaseIDStr)
	if err != nil {
		handlerutil.RespondLocalizedError(c, http.StatusBadRequest, "invalid_input", "validation.invalid_uuid_format")
		return
	}

	perms, err := h.service.ListByPhase(phaseID)
	if err != nil {
		handlerutil.RespondLocalizedError(c, http.StatusInternalServerError, "fetch_failed", "phase_permission.fetch_failed")
		return
	}

	response := dto.PhasePermissionListResponse{
		Items:      mapper.ToPhasePermissionListResponse(perms),
		Total:      int64(len(perms)),
		Page:       1,
		TotalPages: 1,
	}

	c.JSON(http.StatusOK, response)
}

// UpdatePhasePermission godoc
// @Summary Update a phase permission
// @Tags phase-permissions
// @Accept json
// @Produce json
// @Param id path string true "Permission ID"
// @Param permission body dto.UpdatePhasePermissionRequest true "Permission data"
// @Success 200 {object} dto.PhasePermissionResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/phase-permissions/{id} [put]
func (h *PhasePermissionHandler) UpdatePhasePermission(c *gin.Context) {
	id, ok := handlerutil.ParseUUIDParam(c, "id")
	if !ok {
		return
	}

	var req dto.UpdatePhasePermissionRequest
	if !handlerutil.BindJSON(c, &req) {
		return
	}

	perm, err := h.service.GetByID(id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			handlerutil.RespondLocalizedError(c, http.StatusNotFound, "not_found", "phase_permission.permission_not_found")
			return
		}
		handlerutil.RespondLocalizedError(c, http.StatusInternalServerError, "fetch_failed", "phase_permission.fetch_failed")
		return
	}

	mapper.ApplyPhasePermissionUpdate(perm, req)

	if err := h.service.Update(perm); err != nil {
		handlerutil.RespondLocalizedError(c, http.StatusInternalServerError, "update_failed", "phase_permission.update_failed")
		return
	}

	c.JSON(http.StatusOK, mapper.ToPhasePermissionResponse(perm))
}

// DeletePhasePermission godoc
// @Summary Delete a phase permission
// @Tags phase-permissions
// @Produce json
// @Param id path string true "Permission ID"
// @Success 204
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/phase-permissions/{id} [delete]
func (h *PhasePermissionHandler) DeletePhasePermission(c *gin.Context) {
	id, ok := handlerutil.ParseUUIDParam(c, "id")
	if !ok {
		return
	}

	if err := h.service.DeleteByID(id); err != nil {
		handlerutil.RespondLocalizedError(c, http.StatusInternalServerError, "deletion_failed", "phase_permission.deletion_failed")
		return
	}

	c.Status(http.StatusNoContent)
}
