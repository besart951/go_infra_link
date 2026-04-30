package shared

import (
	"context"
	"fmt"
	"net/http"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainUser "github.com/besart951/go_infra_link/backend/internal/domain/user"
	"github.com/besart951/go_infra_link/backend/internal/handler/middleware"
	"github.com/besart951/go_infra_link/backend/internal/handlerutil"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AccessPolicyService interface {
	CanAccessProject(ctx context.Context, requesterID, projectID uuid.UUID, requesterRole *domainUser.Role) (bool, error)
	CanUseProjectPermission(ctx context.Context, requesterID uuid.UUID, requesterRole *domainUser.Role, permission string) (bool, error)
	CanUseProjectPermissionForProject(ctx context.Context, requesterID, projectID uuid.UUID, requesterRole *domainUser.Role, permission string) (bool, error)
}

type ProjectChangeNotifier func(*gin.Context, uuid.UUID, string, ...string)

func EnsureProjectAccess(c *gin.Context, access AccessPolicyService, projectID uuid.UUID) bool {
	if access == nil {
		handlerutil.RespondLocalizedError(c, http.StatusInternalServerError, "authorization_failed", "authorization_failed")
		return false
	}

	userID, ok := middleware.GetUserID(c)
	if !ok {
		handlerutil.RespondLocalizedError(c, http.StatusUnauthorized, "unauthorized", "errors.unauthorized")
		return false
	}

	var requesterRole *domainUser.Role
	if role, ok := middleware.GetUserRole(c); ok {
		requesterRole = &role
	}

	cacheKey := projectAccessCacheKey(userID, projectID, requesterRole)
	if cached, ok := c.Get(cacheKey); ok {
		if hasAccess, ok := cached.(bool); ok {
			if !hasAccess {
				handlerutil.RespondLocalizedError(c, http.StatusForbidden, "forbidden", "errors.forbidden")
			}
			return hasAccess
		}
	}

	hasAccess, err := access.CanAccessProject(c.Request.Context(), userID, projectID, requesterRole)
	if err != nil {
		handlerutil.RespondDomainError(c, err,
			handlerutil.LocalizedError(http.StatusInternalServerError, "fetch_failed", "project.fetch_failed"),
			handlerutil.MapError(domain.ErrNotFound, handlerutil.LocalizedError(http.StatusNotFound, "not_found", "project.project_not_found")),
		)
		return false
	}

	if !hasAccess {
		c.Set(cacheKey, false)
		handlerutil.RespondLocalizedError(c, http.StatusForbidden, "forbidden", "errors.forbidden")
		return false
	}

	c.Set(cacheKey, true)
	return true
}

func EnsureProjectAccessAndPermission(c *gin.Context, access AccessPolicyService, projectID uuid.UUID, permission string) bool {
	if !EnsureProjectAccess(c, access, projectID) {
		return false
	}
	return EnsureProjectPermissionForProject(c, access, projectID, permission)
}

func EnsureProjectAccessAndAnyPermission(c *gin.Context, access AccessPolicyService, projectID uuid.UUID, permissions ...string) bool {
	if !EnsureProjectAccess(c, access, projectID) {
		return false
	}
	return EnsureProjectAnyPermissionForProject(c, access, projectID, permissions...)
}

func EnsureProjectPermission(c *gin.Context, access AccessPolicyService, permission string) bool {
	return EnsureProjectAnyPermission(c, access, permission)
}

func EnsureProjectAnyPermission(c *gin.Context, access AccessPolicyService, permissions ...string) bool {
	if access == nil {
		handlerutil.RespondLocalizedError(c, http.StatusInternalServerError, "authorization_failed", "authorization_failed")
		return false
	}

	userID, ok := middleware.GetUserID(c)
	if !ok {
		handlerutil.RespondLocalizedError(c, http.StatusUnauthorized, "unauthorized", "errors.unauthorized")
		return false
	}

	var requesterRole *domainUser.Role
	if role, ok := middleware.GetUserRole(c); ok {
		requesterRole = &role
	}

	for _, permission := range permissions {
		cacheKey := projectPermissionCacheKey(userID, requesterRole, permission)
		if cached, ok := c.Get(cacheKey); ok {
			if hasPermission, ok := cached.(bool); ok {
				if hasPermission {
					return true
				}
				continue
			}
		}

		hasPermission, err := access.CanUseProjectPermission(c.Request.Context(), userID, requesterRole, permission)
		if err != nil {
			handlerutil.RespondLocalizedError(c, http.StatusInternalServerError, "authorization_failed", "authorization_failed")
			return false
		}
		c.Set(cacheKey, hasPermission)
		if hasPermission {
			return true
		}
	}

	handlerutil.RespondLocalizedError(c, http.StatusForbidden, "forbidden", "errors.forbidden")
	return false
}

func EnsureProjectPermissionForProject(c *gin.Context, access AccessPolicyService, projectID uuid.UUID, permission string) bool {
	return EnsureProjectAnyPermissionForProject(c, access, projectID, permission)
}

func EnsureProjectAnyPermissionForProject(c *gin.Context, access AccessPolicyService, projectID uuid.UUID, permissions ...string) bool {
	if access == nil {
		handlerutil.RespondLocalizedError(c, http.StatusInternalServerError, "authorization_failed", "authorization_failed")
		return false
	}

	userID, ok := middleware.GetUserID(c)
	if !ok {
		handlerutil.RespondLocalizedError(c, http.StatusUnauthorized, "unauthorized", "errors.unauthorized")
		return false
	}

	var requesterRole *domainUser.Role
	if role, ok := middleware.GetUserRole(c); ok {
		requesterRole = &role
	}

	for _, permission := range permissions {
		cacheKey := projectScopedPermissionCacheKey(userID, projectID, requesterRole, permission)
		if cached, ok := c.Get(cacheKey); ok {
			if hasPermission, ok := cached.(bool); ok {
				if hasPermission {
					return true
				}
				continue
			}
		}

		hasPermission, err := access.CanUseProjectPermissionForProject(c.Request.Context(), userID, projectID, requesterRole, permission)
		if err != nil {
			handlerutil.RespondLocalizedError(c, http.StatusInternalServerError, "authorization_failed", "authorization_failed")
			return false
		}
		c.Set(cacheKey, hasPermission)
		if hasPermission {
			return true
		}
	}

	handlerutil.RespondLocalizedError(c, http.StatusForbidden, "forbidden", "errors.forbidden")
	return false
}

func projectAccessCacheKey(userID, projectID uuid.UUID, requesterRole *domainUser.Role) string {
	role := ""
	if requesterRole != nil {
		role = string(*requesterRole)
	}
	return fmt.Sprintf("project_access:%s:%s:%s", userID, projectID, role)
}

func projectPermissionCacheKey(userID uuid.UUID, requesterRole *domainUser.Role, permission string) string {
	role := ""
	if requesterRole != nil {
		role = string(*requesterRole)
	}
	return fmt.Sprintf("project_permission:%s:%s:%s", userID, role, permission)
}

func projectScopedPermissionCacheKey(userID, projectID uuid.UUID, requesterRole *domainUser.Role, permission string) string {
	role := ""
	if requesterRole != nil {
		role = string(*requesterRole)
	}
	return fmt.Sprintf("project_permission:%s:%s:%s:%s", userID, projectID, role, permission)
}
