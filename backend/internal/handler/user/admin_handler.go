package user

import (
	"errors"
	"net/http"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	"github.com/besart951/go_infra_link/backend/internal/domain/user"
	dto "github.com/besart951/go_infra_link/backend/internal/handler/dto/user"
	"github.com/besart951/go_infra_link/backend/internal/handler/middleware"
	"github.com/besart951/go_infra_link/backend/internal/handlerutil"
	"github.com/gin-gonic/gin"
)

type AdminHandler struct {
	adminService AdminService
}

func NewAdminHandler(adminService AdminService) *AdminHandler {
	return &AdminHandler{adminService: adminService}
}

// DisableUser godoc
// @Summary Disable a user
// @Tags admin
// @Param id path string true "User ID"
// @Success 204
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/admin/users/{id}/disable [post]
func (h *AdminHandler) DisableUser(c *gin.Context) {
	actorID, ok := middleware.GetUserID(c)
	if !ok {
		handlerutil.RespondLocalizedError(c, http.StatusUnauthorized, "unauthorized", "errors.unauthorized")
		return
	}
	userID, ok := handlerutil.ParseUUIDParam(c, "id")
	if !ok {
		return
	}
	if err := h.adminService.DisableUser(c.Request.Context(), actorID, userID); err != nil {
		switch {
		case errors.Is(err, user.ErrRoleNotAssignable):
			handlerutil.RespondLocalizedError(c, http.StatusForbidden, "role_not_assignable", "user.role_not_assignable")
		case errors.Is(err, domain.ErrNotFound):
			handlerutil.RespondLocalizedError(c, http.StatusNotFound, "not_found", "user.user_not_found")
		default:
			handlerutil.RespondDomainError(c, err, handlerutil.LocalizedError(http.StatusInternalServerError, "update_failed", "admin.user_disabled"))
		}
		return
	}
	c.Status(http.StatusNoContent)
}

// EnableUser godoc
// @Summary Enable a user
// @Tags admin
// @Param id path string true "User ID"
// @Success 204
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/admin/users/{id}/enable [post]
func (h *AdminHandler) EnableUser(c *gin.Context) {
	actorID, ok := middleware.GetUserID(c)
	if !ok {
		handlerutil.RespondLocalizedError(c, http.StatusUnauthorized, "unauthorized", "errors.unauthorized")
		return
	}
	userID, ok := handlerutil.ParseUUIDParam(c, "id")
	if !ok {
		return
	}
	if err := h.adminService.EnableUser(c.Request.Context(), actorID, userID); err != nil {
		switch {
		case errors.Is(err, user.ErrRoleNotAssignable):
			handlerutil.RespondLocalizedError(c, http.StatusForbidden, "role_not_assignable", "user.role_not_assignable")
		case errors.Is(err, user.ErrRegistrationPending):
			handlerutil.RespondLocalizedError(c, http.StatusConflict, "registration_pending", "auth.registration_pending")
		case errors.Is(err, domain.ErrNotFound):
			handlerutil.RespondLocalizedError(c, http.StatusNotFound, "not_found", "user.user_not_found")
		default:
			handlerutil.RespondLocalizedError(c, http.StatusInternalServerError, "update_failed", "admin.user_enabled")
		}
		return
	}
	c.Status(http.StatusNoContent)
}

// SetUserRole godoc
// @Summary Set a user's role
// @Tags admin
// @Accept json
// @Param id path string true "User ID"
// @Param payload body dto.AdminSetUserRoleRequest true "Role"
// @Success 204
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/admin/users/{id}/role [post]
func (h *AdminHandler) SetUserRole(c *gin.Context) {
	actorID, ok := middleware.GetUserID(c)
	if !ok {
		handlerutil.RespondLocalizedError(c, http.StatusUnauthorized, "unauthorized", "errors.unauthorized")
		return
	}
	userID, ok := handlerutil.ParseUUIDParam(c, "id")
	if !ok {
		return
	}
	var req dto.AdminSetUserRoleRequest
	if !handlerutil.BindJSON(c, &req) {
		return
	}

	if err := h.adminService.SetUserRole(c.Request.Context(), actorID, userID, user.Role(req.Role)); err != nil {
		switch {
		case errors.Is(err, user.ErrRoleNotAssignable):
			handlerutil.RespondLocalizedError(c, http.StatusForbidden, "role_not_assignable", "user.role_not_assignable")
		case errors.Is(err, domain.ErrNotFound):
			handlerutil.RespondLocalizedError(c, http.StatusNotFound, "not_found", "user.user_not_found")
		default:
			handlerutil.RespondDomainError(c, err, handlerutil.LocalizedError(http.StatusInternalServerError, "update_failed", "admin.user_role_updated"))
		}
		return
	}
	c.Status(http.StatusNoContent)
}
