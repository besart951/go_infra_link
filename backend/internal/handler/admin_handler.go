package handler

import (
	"net/http"
	"time"

	"github.com/besart951/go_infra_link/backend/internal/domain/user"
	"github.com/besart951/go_infra_link/backend/internal/handler/dto"
	"github.com/besart951/go_infra_link/backend/internal/handler/middleware"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AdminHandler struct {
	adminService AdminService
	authService  AuthService
}

func NewAdminHandler(adminService AdminService, authService AuthService) *AdminHandler {
	return &AdminHandler{adminService: adminService, authService: authService}
}

func (h *AdminHandler) ResetUserPassword(c *gin.Context) {
	adminID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "unauthorized"})
		return
	}

	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid_id", Message: "Invalid UUID format"})
		return
	}

	resetToken, expiresAt, err := h.authService.CreatePasswordResetToken(adminID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "reset_failed", Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.AdminPasswordResetResponse{ResetToken: resetToken, ExpiresAt: expiresAt})
}

func (h *AdminHandler) DisableUser(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid_id", Message: "Invalid UUID format"})
		return
	}
	if err := h.adminService.DisableUser(userID); err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "update_failed", Message: err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *AdminHandler) EnableUser(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid_id", Message: "Invalid UUID format"})
		return
	}
	if err := h.adminService.EnableUser(userID); err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "update_failed", Message: err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *AdminHandler) LockUser(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid_id", Message: "Invalid UUID format"})
		return
	}
	var req dto.AdminLockUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "validation_error", Message: err.Error()})
		return
	}
	until := req.Until.UTC()
	if until.Before(time.Now().UTC()) {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "validation_error", Message: "until must be in the future"})
		return
	}
	if err := h.adminService.LockUserUntil(userID, until); err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "update_failed", Message: err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *AdminHandler) UnlockUser(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid_id", Message: "Invalid UUID format"})
		return
	}
	if err := h.adminService.UnlockUser(userID); err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "update_failed", Message: err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *AdminHandler) SetUserRole(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid_id", Message: "Invalid UUID format"})
		return
	}
	var req dto.AdminSetUserRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "validation_error", Message: err.Error()})
		return
	}

	if err := h.adminService.SetUserRole(userID, user.Role(req.Role)); err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "update_failed", Message: err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *AdminHandler) ListLoginAttempts(c *gin.Context) {
	var query dto.PaginationQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "validation_error", Message: err.Error()})
		return
	}

	res, err := h.authService.ListLoginAttempts(query.Page, query.Limit, query.Search)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "fetch_failed", Message: err.Error()})
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
