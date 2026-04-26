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
}

type ProjectChangeNotifier func(*gin.Context, uuid.UUID, string, ...string)

func EnsureProjectAccess(c *gin.Context, access AccessPolicyService, projectID uuid.UUID) bool {
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

func projectAccessCacheKey(userID, projectID uuid.UUID, requesterRole *domainUser.Role) string {
	role := ""
	if requesterRole != nil {
		role = string(*requesterRole)
	}
	return fmt.Sprintf("project_access:%s:%s:%s", userID, projectID, role)
}
