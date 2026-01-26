package handler

import (
	"net/http"
	"time"

	"github.com/besart951/go_infra_link/backend/internal/domain/user"
	"github.com/besart951/go_infra_link/backend/internal/handler/dto"
	"github.com/besart951/go_infra_link/backend/internal/handler/middleware"
	"github.com/besart951/go_infra_link/backend/internal/handlerutil"
	"github.com/gin-gonic/gin"
)

type AdminHandler struct {
	adminService AdminService
	authService  AuthService
}

func NewAdminHandler(adminService AdminService, authService AuthService) *AdminHandler {
	return &AdminHandler{adminService: adminService, authService: authService}
}

// ResetUserPassword godoc
// @Summary Create a password reset token for a user
// @Tags admin
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} dto.AdminPasswordResetResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/admin/users/{id}/password-reset [post]
func (h *AdminHandler) ResetUserPassword(c *gin.Context) {
	adminID, ok := middleware.GetUserID(c)
	if !ok {
		handlerutil.RespondError(c, http.StatusUnauthorized, "unauthorized", "")
		return
	}

	userID, ok := handlerutil.ParseUUIDParam(c, "id")
	if !ok {
		return
	}

	resetToken, expiresAt, err := h.authService.CreatePasswordResetToken(adminID, userID)
	if err != nil {
		handlerutil.RespondError(c, http.StatusInternalServerError, "reset_failed", err.Error())
		return
	}

	c.JSON(http.StatusOK, dto.AdminPasswordResetResponse{ResetToken: resetToken, ExpiresAt: expiresAt})
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
	userID, ok := handlerutil.ParseUUIDParam(c, "id")
	if !ok {
		return
	}
	if err := h.adminService.DisableUser(userID); err != nil {
		handlerutil.RespondError(c, http.StatusInternalServerError, "update_failed", err.Error())
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
	userID, ok := handlerutil.ParseUUIDParam(c, "id")
	if !ok {
		return
	}
	if err := h.adminService.EnableUser(userID); err != nil {
		handlerutil.RespondError(c, http.StatusInternalServerError, "update_failed", err.Error())
		return
	}
	c.Status(http.StatusNoContent)
}

// LockUser godoc
// @Summary Lock a user until a given time
// @Tags admin
// @Accept json
// @Param id path string true "User ID"
// @Param payload body dto.AdminLockUserRequest true "Lock details"
// @Success 204
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/admin/users/{id}/lock [post]
func (h *AdminHandler) LockUser(c *gin.Context) {
	userID, ok := handlerutil.ParseUUIDParam(c, "id")
	if !ok {
		return
	}
	var req dto.AdminLockUserRequest
	if !handlerutil.BindJSON(c, &req) {
		return
	}
	until := req.Until.UTC()
	if until.Before(time.Now().UTC()) {
		handlerutil.RespondError(c, http.StatusBadRequest, "validation_error", "until must be in the future")
		return
	}
	if err := h.adminService.LockUserUntil(userID, until); err != nil {
		handlerutil.RespondError(c, http.StatusInternalServerError, "update_failed", err.Error())
		return
	}
	c.Status(http.StatusNoContent)
}

// UnlockUser godoc
// @Summary Unlock a user
// @Tags admin
// @Param id path string true "User ID"
// @Success 204
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/admin/users/{id}/unlock [post]
func (h *AdminHandler) UnlockUser(c *gin.Context) {
	userID, ok := handlerutil.ParseUUIDParam(c, "id")
	if !ok {
		return
	}
	if err := h.adminService.UnlockUser(userID); err != nil {
		handlerutil.RespondError(c, http.StatusInternalServerError, "update_failed", err.Error())
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
	userID, ok := handlerutil.ParseUUIDParam(c, "id")
	if !ok {
		return
	}
	var req dto.AdminSetUserRoleRequest
	if !handlerutil.BindJSON(c, &req) {
		return
	}

	if err := h.adminService.SetUserRole(userID, user.Role(req.Role)); err != nil {
		handlerutil.RespondError(c, http.StatusInternalServerError, "update_failed", err.Error())
		return
	}
	c.Status(http.StatusNoContent)
}

// ListLoginAttempts godoc
// @Summary List login attempts
// @Tags admin
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Param search query string false "Search query"
// @Success 200 {object} dto.LoginAttemptListResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/admin/login-attempts [get]
func (h *AdminHandler) ListLoginAttempts(c *gin.Context) {
	var query dto.PaginationQuery
	if !handlerutil.BindQuery(c, &query) {
		return
	}

	res, err := h.authService.ListLoginAttempts(query.Page, query.Limit, query.Search)
	if err != nil {
		handlerutil.RespondError(c, http.StatusInternalServerError, "fetch_failed", err.Error())
		return
	}

	items := make([]dto.LoginAttemptResponse, len(res.Items))
	for i, a := range res.Items {
		items[i] = dto.LoginAttemptResponse{
			ID:            a.ID,
			CreatedAt:     a.CreatedAt,
			UserID:        a.UserID,
			Email:         a.Email,
			IP:            a.IP,
			UserAgent:     a.UserAgent,
			Success:       a.Success,
			FailureReason: a.FailureReason,
		}
	}

	c.JSON(http.StatusOK, dto.LoginAttemptListResponse{Items: items, Total: res.Total, Page: res.Page, TotalPages: res.TotalPages})
}
