package handler

import (
	"errors"
	"net/http"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainUser "github.com/besart951/go_infra_link/backend/internal/domain/user"
	"github.com/besart951/go_infra_link/backend/internal/handler/dto"
	"github.com/besart951/go_infra_link/backend/internal/handlerutil"
	"github.com/gin-gonic/gin"
)

type PermissionHandler struct {
	service PermissionService
}

func NewPermissionHandler(service PermissionService) *PermissionHandler {
	return &PermissionHandler{service: service}
}

// ListPermissions godoc
// @Summary List permission types
// @Tags permissions
// @Produce json
// @Success 200 {array} dto.PermissionResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/permissions [get]
func (h *PermissionHandler) ListPermissions(c *gin.Context) {
	perms, err := h.service.ListPermissions()
	if err != nil {
		handlerutil.RespondLocalizedError(c, http.StatusInternalServerError, "fetch_failed", "permissions.fetch_failed")
		return
	}

	response := make([]dto.PermissionResponse, len(perms))
	for i, perm := range perms {
		response[i] = dto.PermissionResponse{
			ID:          perm.ID,
			Name:        perm.Name,
			Description: perm.Description,
			Resource:    perm.Resource,
			Action:      perm.Action,
			CreatedAt:   perm.CreatedAt,
			UpdatedAt:   perm.UpdatedAt,
		}
	}

	c.JSON(http.StatusOK, response)
}

// CreatePermission godoc
// @Summary Create a permission type
// @Tags permissions
// @Accept json
// @Produce json
// @Param permission body dto.CreatePermissionRequest true "Permission data"
// @Success 201 {object} dto.PermissionResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/permissions [post]
func (h *PermissionHandler) CreatePermission(c *gin.Context) {
	var req dto.CreatePermissionRequest
	if !handlerutil.BindJSON(c, &req) {
		return
	}

	perm := &domainUser.Permission{
		Name:        req.Name,
		Description: req.Description,
		Resource:    req.Resource,
		Action:      req.Action,
	}

	if err := h.service.CreatePermission(perm); err != nil {
		handlerutil.RespondLocalizedError(c, http.StatusInternalServerError, "creation_failed", "permissions.creation_failed")
		return
	}

	c.JSON(http.StatusCreated, dto.PermissionResponse{
		ID:          perm.ID,
		Name:        perm.Name,
		Description: perm.Description,
		Resource:    perm.Resource,
		Action:      perm.Action,
		CreatedAt:   perm.CreatedAt,
		UpdatedAt:   perm.UpdatedAt,
	})
}

// UpdatePermission godoc
// @Summary Update a permission type
// @Tags permissions
// @Accept json
// @Produce json
// @Param id path string true "Permission ID"
// @Param permission body dto.UpdatePermissionRequest true "Permission data"
// @Success 200 {object} dto.PermissionResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/permissions/{id} [put]
func (h *PermissionHandler) UpdatePermission(c *gin.Context) {
	id, ok := handlerutil.ParseUUIDParam(c, "id")
	if !ok {
		return
	}

	var req dto.UpdatePermissionRequest
	if !handlerutil.BindJSON(c, &req) {
		return
	}

	perm, err := h.service.GetPermissionByID(id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			handlerutil.RespondLocalizedError(c, http.StatusNotFound, "not_found", "permissions.permission_not_found")
			return
		}
		handlerutil.RespondLocalizedError(c, http.StatusInternalServerError, "fetch_failed", "permissions.fetch_failed")
		return
	}

	if req.Description != nil {
		perm.Description = *req.Description
	}

	if err := h.service.UpdatePermission(perm); err != nil {
		handlerutil.RespondLocalizedError(c, http.StatusInternalServerError, "update_failed", "permissions.update_failed")
		return
	}

	c.JSON(http.StatusOK, dto.PermissionResponse{
		ID:          perm.ID,
		Name:        perm.Name,
		Description: perm.Description,
		Resource:    perm.Resource,
		Action:      perm.Action,
		CreatedAt:   perm.CreatedAt,
		UpdatedAt:   perm.UpdatedAt,
	})
}

// DeletePermission godoc
// @Summary Delete a permission type
// @Tags permissions
// @Produce json
// @Param id path string true "Permission ID"
// @Success 204
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/permissions/{id} [delete]
func (h *PermissionHandler) DeletePermission(c *gin.Context) {
	id, ok := handlerutil.ParseUUIDParam(c, "id")
	if !ok {
		return
	}

	if err := h.service.DeletePermission(id); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			handlerutil.RespondLocalizedError(c, http.StatusNotFound, "not_found", "permissions.permission_not_found")
			return
		}
		handlerutil.RespondLocalizedError(c, http.StatusInternalServerError, "deletion_failed", "permissions.deletion_failed")
		return
	}

	c.Status(http.StatusNoContent)
}
