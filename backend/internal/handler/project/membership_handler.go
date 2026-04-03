package project

import (
	"errors"
	"net/http"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	dto "github.com/besart951/go_infra_link/backend/internal/handler/dto/project"
	userhandler "github.com/besart951/go_infra_link/backend/internal/handler/user"
	"github.com/besart951/go_infra_link/backend/internal/handlerutil"
	"github.com/gin-gonic/gin"
)

// InviteProjectUser godoc
// @Summary Invite user to project
// @Tags projects
// @Accept json
// @Produce json
// @Param id path string true "Project ID"
// @Param invite body dto.CreateProjectUserRequest true "Invite data"
// @Success 201 {object} dto.ProjectUserResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/projects/{id}/users [post]
func (h *ProjectHandler) InviteProjectUser(c *gin.Context) {
	projectID, ok := handlerutil.ParseUUIDParam(c, "id")
	if !ok {
		return
	}

	if !h.ensureProjectAccess(c, projectID) {
		return
	}

	var req dto.CreateProjectUserRequest
	if !handlerutil.BindJSON(c, &req) {
		return
	}

	if err := h.service.InviteUser(projectID, req.UserID); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			handlerutil.RespondLocalizedError(c, http.StatusNotFound, "not_found", "project.project_or_user_not_found")
			return
		}
		handlerutil.RespondLocalizedError(c, http.StatusInternalServerError, "invite_failed", "project.user_invited_failed")
		return
	}

	c.JSON(http.StatusCreated, dto.ProjectUserResponse{ProjectID: projectID, UserID: req.UserID})
}

// ListProjectUsers godoc
// @Summary List users in a project
// @Tags projects
// @Produce json
// @Param id path string true "Project ID"
// @Success 200 {object} dto.ProjectUserListResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/projects/{id}/users [get]
func (h *ProjectHandler) ListProjectUsers(c *gin.Context) {
	projectID, ok := handlerutil.ParseUUIDParam(c, "id")
	if !ok {
		return
	}

	if !h.ensureProjectAccess(c, projectID) {
		return
	}

	users, err := h.service.ListUsers(projectID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			handlerutil.RespondLocalizedError(c, http.StatusNotFound, "not_found", "project.project_not_found")
			return
		}
		handlerutil.RespondLocalizedError(c, http.StatusInternalServerError, "fetch_failed", "project.fetch_failed")
		return
	}

	c.JSON(http.StatusOK, dto.ProjectUserListResponse{Items: userhandler.ToUserListResponse(users)})
}

// RemoveProjectUser godoc
// @Summary Remove user from project
// @Tags projects
// @Produce json
// @Param id path string true "Project ID"
// @Param userId path string true "User ID"
// @Success 204
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/projects/{id}/users/{userId} [delete]
func (h *ProjectHandler) RemoveProjectUser(c *gin.Context) {
	projectID, ok := handlerutil.ParseUUIDParam(c, "id")
	if !ok {
		return
	}

	if !h.ensureProjectAccess(c, projectID) {
		return
	}

	userID, ok := handlerutil.ParseUUIDParam(c, "userId")
	if !ok {
		return
	}

	if err := h.service.RemoveUser(projectID, userID); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			handlerutil.RespondLocalizedError(c, http.StatusNotFound, "not_found", "project.project_or_user_not_found")
			return
		}
		handlerutil.RespondLocalizedError(c, http.StatusInternalServerError, "remove_failed", "project.user_remove_failed")
		return
	}

	c.Status(http.StatusNoContent)
}
