package shared

import (
	"context"
	"net/http"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	"github.com/besart951/go_infra_link/backend/internal/handler/middleware"
	"github.com/besart951/go_infra_link/backend/internal/handlerutil"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AccessPolicyService interface {
	CanAccessProject(ctx context.Context, requesterID, projectID uuid.UUID) (bool, error)
}

type ProjectChangeNotifier func(*gin.Context, uuid.UUID, string)

func EnsureProjectAccess(c *gin.Context, access AccessPolicyService, projectID uuid.UUID) bool {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		handlerutil.RespondLocalizedError(c, http.StatusUnauthorized, "unauthorized", "errors.unauthorized")
		return false
	}

	hasAccess, err := access.CanAccessProject(c.Request.Context(), userID, projectID)
	if err != nil {
		handlerutil.RespondDomainError(c, err,
			handlerutil.LocalizedError(http.StatusInternalServerError, "fetch_failed", "project.fetch_failed"),
			handlerutil.MapError(domain.ErrNotFound, handlerutil.LocalizedError(http.StatusNotFound, "not_found", "project.project_not_found")),
		)
		return false
	}

	if !hasAccess {
		handlerutil.RespondLocalizedError(c, http.StatusForbidden, "forbidden", "errors.forbidden")
		return false
	}

	return true
}
