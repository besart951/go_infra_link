package user

import (
	"net/http"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainUser "github.com/besart951/go_infra_link/backend/internal/domain/user"
	dto "github.com/besart951/go_infra_link/backend/internal/handler/dto/user"
	"github.com/besart951/go_infra_link/backend/internal/handler/middleware"
	"github.com/besart951/go_infra_link/backend/internal/handlerutil"
	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	service     UserService
	roleService RoleQueryService
	directory   UserDirectoryService
}

func NewUserHandler(service UserService, roleService RoleQueryService, directory UserDirectoryService) *UserHandler {
	return &UserHandler{
		service:     service,
		roleService: roleService,
		directory:   directory,
	}
}

// CreateUser godoc
// @Summary Create a new user
// @Tags users
// @Accept json
// @Produce json
// @Param user body dto.CreateUserRequest true "User data"
// @Success 201 {object} dto.UserResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/users [post]
func (h *UserHandler) CreateUser(c *gin.Context) {
	var req dto.CreateUserRequest
	if !handlerutil.BindJSON(c, &req) {
		return
	}

	usr := ToUserModel(req)

	if err := h.service.CreateWithPassword(c.Request.Context(), usr, req.Password); err != nil {
		handlerutil.RespondLocalizedError(c, http.StatusInternalServerError, "creation_failed", "user.creation_failed")
		return
	}

	c.JSON(http.StatusCreated, ToUserResponse(usr))
}

// GetUser godoc
// @Summary Get a user by ID
// @Tags users
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} dto.UserResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/users/{id} [get]
func (h *UserHandler) GetUser(c *gin.Context) {
	id, ok := handlerutil.ParseUUIDParam(c, "id")
	if !ok {
		return
	}

	usr, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		handlerutil.RespondDomainError(
			c,
			err,
			handlerutil.LocalizedError(http.StatusInternalServerError, "fetch_failed", "user.fetch_failed"),
			handlerutil.MapError(domain.ErrNotFound, handlerutil.LocalizedError(http.StatusNotFound, "not_found", "user.user_not_found")),
		)
		return
	}

	c.JSON(http.StatusOK, ToUserResponse(usr))
}

// ListUsers godoc
// @Summary List users with pagination
// @Tags users
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Param search query string false "Search query"
// @Success 200 {object} dto.UserListResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/users [get]
func (h *UserHandler) ListUsers(c *gin.Context) {
	var query dto.PaginationQuery
	if !handlerutil.BindQuery(c, &query) {
		return
	}

	result, err := h.service.List(c.Request.Context(), query.Page, query.Limit, query.Search, query.OrderBy, query.Order)
	if err != nil {
		handlerutil.RespondLocalizedError(c, http.StatusInternalServerError, "fetch_failed", "user.fetch_failed")
		return
	}

	response := dto.UserListResponse{
		Items:      ToUserListResponse(result.Items),
		Total:      result.Total,
		Page:       result.Page,
		TotalPages: result.TotalPages,
	}

	c.JSON(http.StatusOK, response)
}

// ListDirectory godoc
// @Summary List visible users for the user directory
// @Tags users
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Param search query string false "Search query"
// @Param team_id query string false "Visible team filter"
// @Success 200 {object} dto.UserDirectoryListResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/users/directory [get]
func (h *UserHandler) ListDirectory(c *gin.Context) {
	requesterID, ok := middleware.GetUserID(c)
	if !ok {
		handlerutil.RespondLocalizedError(c, http.StatusUnauthorized, "unauthorized", "errors.unauthorized")
		return
	}

	var query struct {
		dto.PaginationQuery
		TeamID string `form:"team_id"`
	}
	if !handlerutil.BindQuery(c, &query) {
		return
	}

	result, err := h.directory.List(
		c.Request.Context(),
		requesterID,
		query.Page,
		query.Limit,
		query.Search,
		query.TeamID,
		query.OrderBy,
		query.Order,
	)
	if err != nil {
		if err == domainUser.ErrForbiddenUserDirectory {
			handlerutil.RespondLocalizedError(c, http.StatusForbidden, "forbidden", "errors.forbidden")
			return
		}
		handlerutil.RespondLocalizedError(c, http.StatusInternalServerError, "fetch_failed", "user.fetch_failed")
		return
	}

	c.JSON(http.StatusOK, ToUserDirectoryListResponse(result))
}

// UpdateUser godoc
// @Summary Update a user
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param user body dto.UpdateUserRequest true "User data"
// @Success 200 {object} dto.UserResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/users/{id} [put]
func (h *UserHandler) UpdateUser(c *gin.Context) {
	id, ok := handlerutil.ParseUUIDParam(c, "id")
	if !ok {
		return
	}

	var req dto.UpdateUserRequest
	if !handlerutil.BindJSON(c, &req) {
		return
	}

	ctx := c.Request.Context()

	usr, err := h.service.GetByID(ctx, id)
	if err != nil {
		handlerutil.RespondDomainError(
			c,
			err,
			handlerutil.LocalizedError(http.StatusInternalServerError, "fetch_failed", "user.fetch_failed"),
			handlerutil.MapError(domain.ErrNotFound, handlerutil.LocalizedError(http.StatusNotFound, "not_found", "user.user_not_found")),
		)
		return
	}

	ApplyUserUpdate(usr, req)

	if err := h.service.UpdateWithPassword(ctx, usr, &req.Password); err != nil {
		handlerutil.RespondLocalizedError(c, http.StatusInternalServerError, "update_failed", "user.update_failed")
		return
	}

	c.JSON(http.StatusOK, ToUserResponse(usr))
}

// DeleteUser godoc
// @Summary Delete a user
// @Tags users
// @Produce json
// @Param id path string true "User ID"
// @Success 204
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/users/{id} [delete]
func (h *UserHandler) DeleteUser(c *gin.Context) {
	id, ok := handlerutil.ParseUUIDParam(c, "id")
	if !ok {
		return
	}

	if err := h.service.DeleteByID(c.Request.Context(), id); err != nil {
		handlerutil.RespondLocalizedError(c, http.StatusInternalServerError, "deletion_failed", "user.deletion_failed")
		return
	}

	c.Status(http.StatusNoContent)
}

// GetAllowedRoles godoc
// @Summary Get roles that the current user can assign
// @Tags users
// @Produce json
// @Success 200 {object} dto.AllowedRolesResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/users/allowed-roles [get]
func (h *UserHandler) GetAllowedRoles(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		handlerutil.RespondLocalizedError(c, http.StatusUnauthorized, "unauthorized", "errors.unauthorized")
		return
	}

	role, err := h.roleService.GetGlobalRole(c.Request.Context(), userID)
	if err != nil {
		handlerutil.RespondLocalizedError(c, http.StatusInternalServerError, "fetch_failed", "user.fetch_failed")
		return
	}

	allowedRoles, err := h.roleService.GetAllowedRoles(c.Request.Context(), role)
	if err != nil {
		handlerutil.RespondLocalizedError(c, http.StatusInternalServerError, "fetch_failed", "user.fetch_failed")
		return
	}
	roleObjects := make([]dto.AllowedRole, len(allowedRoles))
	for i, r := range allowedRoles {
		roleObjects[i] = dto.AllowedRole{
			Role:        string(r),
			DisplayName: domainUser.RoleDisplayName(r),
		}
	}

	c.JSON(http.StatusOK, dto.AllowedRolesResponse{
		Roles: roleObjects,
	})
}
