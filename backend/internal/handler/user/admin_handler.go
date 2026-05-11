package user

import (
	"net/http"

	"github.com/besart951/go_infra_link/backend/internal/domain/user"
	dto "github.com/besart951/go_infra_link/backend/internal/handler/dto/user"
	"github.com/besart951/go_infra_link/backend/internal/handler/middleware"
	"github.com/besart951/go_infra_link/backend/internal/handlerutil"
	"github.com/gin-gonic/gin"
)

type AdminHandler struct {
	adminService   AdminService
	userService    UserService
	errorResponder *handlerutil.ErrorResponder
}

func NewAdminHandler(adminService AdminService, userService UserService) *AdminHandler {
	return &AdminHandler{
		adminService:   adminService,
		userService:    userService,
		errorResponder: handlerutil.NewErrorResponder(),
	}
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
		h.errorResponder.RespondUserError(c, err)
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
		h.errorResponder.RespondUserError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// RestoreUser godoc
// @Summary Restore a deleted user
// @Tags admin
// @Param id path string true "User ID"
// @Success 204
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/admin/users/{id}/restore [post]
func (h *AdminHandler) RestoreUser(c *gin.Context) {
	actorID, ok := middleware.GetUserID(c)
	if !ok {
		handlerutil.RespondLocalizedError(c, http.StatusUnauthorized, "unauthorized", "errors.unauthorized")
		return
	}
	userID, ok := handlerutil.ParseUUIDParam(c, "id")
	if !ok {
		return
	}
	if err := h.userService.RestoreByIDForActor(c.Request.Context(), actorID, userID); err != nil {
		h.errorResponder.RespondUserError(c, err)
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
		h.errorResponder.RespondUserError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
