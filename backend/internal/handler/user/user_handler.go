package user

import (
	"net/http"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainAuth "github.com/besart951/go_infra_link/backend/internal/domain/auth"
	domainUser "github.com/besart951/go_infra_link/backend/internal/domain/user"
	dto "github.com/besart951/go_infra_link/backend/internal/handler/dto/user"
	"github.com/besart951/go_infra_link/backend/internal/handler/middleware"
	"github.com/besart951/go_infra_link/backend/internal/handlerutil"
	userdirectory "github.com/besart951/go_infra_link/backend/internal/service/userdirectory"
	userregistration "github.com/besart951/go_infra_link/backend/internal/service/userregistration"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type UserHandler struct {
	service        UserService
	roleService    RoleQueryService
	directory      UserDirectoryService
	registration   UserRegistrationService
	errorResponder *handlerutil.ErrorResponder
}

func NewUserHandler(service UserService, roleService RoleQueryService, directory UserDirectoryService, registration UserRegistrationService) *UserHandler {
	return &UserHandler{
		service:        service,
		roleService:    roleService,
		directory:      directory,
		registration:   registration,
		errorResponder: handlerutil.NewErrorResponder(),
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
	actorID, ok := middleware.GetUserID(c)
	if !ok {
		handlerutil.RespondLocalizedError(c, http.StatusUnauthorized, "unauthorized", "errors.unauthorized")
		return
	}
	var req dto.CreateUserRequest
	if !handlerutil.BindJSON(c, &req) {
		return
	}

	usr := ToUserModel(req)

	if err := h.service.CreateWithPasswordForActor(c.Request.Context(), actorID, usr, req.Password); err != nil {
		h.errorResponder.RespondUserError(c, err)
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
	if usr.IsIdentityHidden() {
		handlerutil.RespondLocalizedError(c, http.StatusNotFound, "not_found", "user.user_not_found")
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

	result, err := h.service.List(c.Request.Context(), query.Page, query.Limit, query.Search, query.OrderBy, query.Order, query.IncludeDeleted)
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
// @Param role query string false "Role filter"
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
		Role   string `form:"role"`
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
		query.Role,
		query.OrderBy,
		query.Order,
		query.IncludeDeleted,
	)
	if err != nil {
		if err == domainUser.ErrForbiddenUserDirectory {
			handlerutil.RespondLocalizedError(c, http.StatusForbidden, "forbidden", "errors.forbidden")
			return
		}
		handlerutil.RespondLocalizedError(c, http.StatusInternalServerError, "fetch_failed", "user.fetch_failed")
		return
	}

	processes, err := h.registrationProcesses(c, result)
	if err != nil {
		handlerutil.RespondLocalizedError(c, http.StatusInternalServerError, "fetch_failed", "user.fetch_failed")
		return
	}
	applyRegistrationCapabilities(result, processes)
	c.JSON(http.StatusOK, ToUserDirectoryListResponse(result, processes))
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
	actorID, ok := middleware.GetUserID(c)
	if !ok {
		handlerutil.RespondLocalizedError(c, http.StatusUnauthorized, "unauthorized", "errors.unauthorized")
		return
	}
	id, ok := handlerutil.ParseUUIDParam(c, "id")
	if !ok {
		return
	}

	var req dto.UpdateUserRequest
	if !handlerutil.BindJSON(c, &req) {
		return
	}
	if rejectRestrictedUserUpdate(c, req) {
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
	if usr.IsIdentityHidden() {
		handlerutil.RespondLocalizedError(c, http.StatusNotFound, "not_found", "user.user_not_found")
		return
	}

	ApplyUserUpdate(usr, req)

	if err := h.service.UpdateProfileForActor(ctx, actorID, usr); err != nil {
		h.errorResponder.RespondUserError(c, err)
		return
	}

	c.JSON(http.StatusOK, ToUserResponse(usr))
}

// UpdateOwnPassword godoc
// @Summary Update own password
// @Tags users
// @Accept json
// @Produce json
// @Param payload body dto.UpdateOwnPasswordRequest true "Password"
// @Success 200 {object} dto.UserResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/users/me/password [put]
func (h *UserHandler) UpdateOwnPassword(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		handlerutil.RespondLocalizedError(c, http.StatusUnauthorized, "unauthorized", "errors.unauthorized")
		return
	}
	var req dto.UpdateOwnPasswordRequest
	if !handlerutil.BindJSON(c, &req) {
		return
	}
	usr, err := h.service.UpdatePassword(c.Request.Context(), userID, req.CurrentPassword, req.NewPassword)
	if err != nil {
		handlerutil.RespondDomainError(
			c,
			err,
			handlerutil.LocalizedError(http.StatusInternalServerError, "update_failed", "user.update_failed"),
			handlerutil.MapError(domainAuth.ErrInvalidCredentials, handlerutil.LocalizedError(http.StatusUnauthorized, "invalid_credentials", "auth.invalid_credentials")),
		)
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
	actorID, ok := middleware.GetUserID(c)
	if !ok {
		handlerutil.RespondLocalizedError(c, http.StatusUnauthorized, "unauthorized", "errors.unauthorized")
		return
	}
	id, ok := handlerutil.ParseUUIDParam(c, "id")
	if !ok {
		return
	}

	if err := h.service.DeleteByIDForActor(c.Request.Context(), actorID, id); err != nil {
		h.errorResponder.RespondUserError(c, err)
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

func (h *UserHandler) registrationProcesses(c *gin.Context, result *userdirectory.ListResult) (map[uuid.UUID]*userregistration.Process, error) {
	processes := make(map[uuid.UUID]*userregistration.Process, len(result.Items))
	if h.registration == nil {
		return processes, nil
	}
	users := make([]domainUser.User, 0, len(result.Items))
	for _, item := range result.Items {
		users = append(users, item.User)
	}
	return h.registration.ListProcessesForUsers(c.Request.Context(), users)
}

func applyRegistrationCapabilities(result *userdirectory.ListResult, processes map[uuid.UUID]*userregistration.Process) {
	if result == nil {
		return
	}
	for i := range result.Items {
		process := processes[result.Items[i].User.ID]
		if process != nil && process.BlocksManualEnable {
			result.Items[i].Capabilities.CanEnable = false
		}
	}
}

func rejectRestrictedUserUpdate(c *gin.Context, req dto.UpdateUserRequest) bool {
	fields := map[string]string{}
	if req.Role != nil {
		fields["role"] = "use admin role endpoint"
	}
	if req.IsActive != nil {
		fields["is_active"] = "use admin enable or disable endpoint"
	}
	if req.Password != nil {
		fields["password"] = "use password endpoint"
	}
	if len(fields) == 0 {
		return false
	}
	handlerutil.RespondValidationError(c, fields)
	return true
}
