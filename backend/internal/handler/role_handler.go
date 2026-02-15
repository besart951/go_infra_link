package handler

import (
	"errors"
	"net/http"
	"time"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainUser "github.com/besart951/go_infra_link/backend/internal/domain/user"
	"github.com/besart951/go_infra_link/backend/internal/handler/dto"
	"github.com/besart951/go_infra_link/backend/internal/handlerutil"
	"github.com/gin-gonic/gin"
)

type RoleHandler struct {
	service RolePermissionService
}

func NewRoleHandler(service RolePermissionService) *RoleHandler {
	return &RoleHandler{service: service}
}

// ListRoles godoc
// @Summary List roles with permissions
// @Tags roles
// @Produce json
// @Success 200 {array} dto.RoleResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/roles [get]
func (h *RoleHandler) ListRoles(c *gin.Context) {
	roles, err := h.service.ListRolesWithPermissions()
	if err != nil {
		handlerutil.RespondLocalizedError(c, http.StatusInternalServerError, "fetch_failed", "roles.fetch_failed")
		return
	}

	now := time.Now().UTC()
	response := make([]dto.RoleResponse, len(roles))
	for i, role := range roles {
		response[i] = dto.RoleResponse{
			ID:          string(role.Name),
			Name:        role.Name,
			DisplayName: role.DisplayName,
			Description: role.Description,
			Level:       role.Level,
			Permissions: role.Permissions,
			CanManage:   role.CanManage,
			CreatedAt:   now,
			UpdatedAt:   now,
		}
	}

	c.JSON(http.StatusOK, response)
}

// UpdateRolePermissions godoc
// @Summary Replace permissions for a role
// @Tags roles
// @Accept json
// @Produce json
// @Param role path string true "Role"
// @Param payload body dto.UpdateRolePermissionsRequest true "Role permissions"
// @Success 200 {object} dto.RoleResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/roles/{role}/permissions [put]
func (h *RoleHandler) UpdateRolePermissions(c *gin.Context) {
	roleParam := domainUser.Role(c.Param("role"))
	if !domainUser.IsValidRole(roleParam) {
		handlerutil.RespondLocalizedError(c, http.StatusBadRequest, "invalid_role", "roles.invalid_role")
		return
	}

	var req dto.UpdateRolePermissionsRequest
	if !handlerutil.BindJSON(c, &req) {
		return
	}

	permissions, err := h.service.UpdateRolePermissions(roleParam, req.Permissions)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			handlerutil.RespondLocalizedError(c, http.StatusNotFound, "permission_not_found", "roles.permission_not_found")
			return
		}
		handlerutil.RespondLocalizedError(c, http.StatusInternalServerError, "update_failed", "roles.update_failed")
		return
	}
	if permissions == nil {
		permissions = []string{}
	}

	now := time.Now().UTC()
	c.JSON(http.StatusOK, dto.RoleResponse{
		ID:          string(roleParam),
		Name:        roleParam,
		DisplayName: domainUser.RoleDisplayName(roleParam),
		Description: domainUser.RoleDescription(roleParam),
		Level:       domainUser.RoleLevel(roleParam),
		Permissions: permissions,
		CanManage:   domainUser.GetAllowedRoles(roleParam),
		CreatedAt:   now,
		UpdatedAt:   now,
	})
}

// AddRolePermission godoc
// @Summary Assign a permission to a role
// @Tags roles
// @Accept json
// @Produce json
// @Param role path string true "Role"
// @Param payload body dto.AddRolePermissionRequest true "Permission"
// @Success 201 {object} dto.RolePermissionResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/roles/{role}/permissions [post]
func (h *RoleHandler) AddRolePermission(c *gin.Context) {
	roleParam := domainUser.Role(c.Param("role"))
	if !domainUser.IsValidRole(roleParam) {
		handlerutil.RespondLocalizedError(c, http.StatusBadRequest, "invalid_role", "roles.invalid_role")
		return
	}

	var req dto.AddRolePermissionRequest
	if !handlerutil.BindJSON(c, &req) {
		return
	}

	perm, err := h.service.AddRolePermission(roleParam, req.Permission)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			handlerutil.RespondLocalizedError(c, http.StatusNotFound, "permission_not_found", "roles.permission_not_found")
			return
		}
		handlerutil.RespondLocalizedError(c, http.StatusInternalServerError, "creation_failed", "roles.permission_assign_failed")
		return
	}

	c.JSON(http.StatusCreated, dto.RolePermissionResponse{
		ID:         perm.ID.String(),
		Role:       perm.Role,
		Permission: perm.Permission,
		CreatedAt:  perm.CreatedAt,
		UpdatedAt:  perm.UpdatedAt,
	})
}

// RemoveRolePermission godoc
// @Summary Remove a permission from a role
// @Tags roles
// @Produce json
// @Param role path string true "Role"
// @Param permission path string true "Permission name"
// @Success 204
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/roles/{role}/permissions/{permission} [delete]
func (h *RoleHandler) RemoveRolePermission(c *gin.Context) {
	roleParam := domainUser.Role(c.Param("role"))
	if !domainUser.IsValidRole(roleParam) {
		handlerutil.RespondLocalizedError(c, http.StatusBadRequest, "invalid_role", "roles.invalid_role")
		return
	}

	permissionName := c.Param("permission")
	if permissionName == "" {
		handlerutil.RespondLocalizedError(c, http.StatusBadRequest, "invalid_permission", "roles.invalid_permission")
		return
	}

	if err := h.service.RemoveRolePermission(roleParam, permissionName); err != nil {
		handlerutil.RespondLocalizedError(c, http.StatusInternalServerError, "deletion_failed", "roles.permission_remove_failed")
		return
	}

	c.Status(http.StatusNoContent)
}
