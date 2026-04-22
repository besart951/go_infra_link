package project

import (
	"net/http"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainProject "github.com/besart951/go_infra_link/backend/internal/domain/project"
	dto "github.com/besart951/go_infra_link/backend/internal/handler/dto/project"
	"github.com/besart951/go_infra_link/backend/internal/handler/middleware"
	"github.com/besart951/go_infra_link/backend/internal/handlerutil"
	"github.com/gin-gonic/gin"
)

// CreateProject godoc
// @Summary Create a new project
// @Tags projects
// @Accept json
// @Produce json
// @Param project body dto.CreateProjectRequest true "Project data"
// @Success 201 {object} dto.ProjectResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/projects [post]
func (h *ProjectHandler) CreateProject(c *gin.Context) {
	var req dto.CreateProjectRequest
	if !handlerutil.BindJSON(c, &req) {
		return
	}

	proj := ToProjectModel(req)

	creatorID, ok := middleware.GetUserID(c)
	if !ok {
		handlerutil.RespondLocalizedError(c, http.StatusUnauthorized, "unauthorized", "errors.unauthorized")
		return
	}
	proj.CreatorID = creatorID

	if err := h.service.Create(c.Request.Context(), proj); err != nil {
		handlerutil.RespondDomainError(c, err,
			handlerutil.LocalizedError(http.StatusInternalServerError, "creation_failed", "project.creation_failed"),
			handlerutil.MapError(domain.ErrNotFound, handlerutil.LocalizedError(http.StatusNotFound, "not_found", "project.project_or_object_data_not_found")),
			handlerutil.MapError(domain.ErrConflict, handlerutil.LocalizedError(http.StatusConflict, "conflict", "project.creation_failed")),
			handlerutil.MapError(domain.ErrInvalidArgument, handlerutil.LocalizedError(http.StatusBadRequest, "validation_error", "project.creation_failed")),
		)
		return
	}

	c.JSON(http.StatusCreated, ToProjectResponse(proj))
}

// GetProject godoc
// @Summary Get a project by ID
// @Tags projects
// @Produce json
// @Param id path string true "Project ID"
// @Success 200 {object} dto.ProjectResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/projects/{id} [get]
func (h *ProjectHandler) GetProject(c *gin.Context) {
	id, ok := handlerutil.ParseUUIDParam(c, "id")
	if !ok {
		return
	}

	if !h.ensureProjectAccess(c, id) {
		return
	}

	proj, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		handlerutil.RespondDomainError(c, err,
			handlerutil.LocalizedError(http.StatusInternalServerError, "fetch_failed", "project.fetch_failed"),
			handlerutil.MapError(domain.ErrNotFound, handlerutil.LocalizedError(http.StatusNotFound, "not_found", "project.project_not_found")),
		)
		return
	}

	c.JSON(http.StatusOK, ToProjectResponse(proj))
}

// ListProjects godoc
// @Summary List projects with pagination
// @Tags projects
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Param search query string false "Search query"
// @Success 200 {object} dto.ProjectListResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/projects [get]
func (h *ProjectHandler) ListProjects(c *gin.Context) {
	var query dto.ListProjectsQuery
	if !handlerutil.BindQuery(c, &query) {
		return
	}

	userID, ok := middleware.GetUserID(c)
	if !ok {
		handlerutil.RespondLocalizedError(c, http.StatusUnauthorized, "unauthorized", "errors.unauthorized")
		return
	}

	var status *domainProject.ProjectStatus
	if query.Status != "" {
		status = &query.Status
	}

	result, err := h.service.List(c.Request.Context(), userID, query.Page, query.Limit, query.Search, status)
	if err != nil {
		handlerutil.RespondLocalizedError(c, http.StatusInternalServerError, "fetch_failed", "project.fetch_failed")
		return
	}

	response := dto.ProjectListResponse{
		Items:      ToProjectListResponse(result.Items),
		Total:      result.Total,
		Page:       result.Page,
		TotalPages: result.TotalPages,
	}

	c.JSON(http.StatusOK, response)
}

// UpdateProject godoc
// @Summary Update a project
// @Description PATCH-like update: omitted fields remain unchanged, present string fields are applied even when empty, and start_date:null clears the date. PUT is kept for compatibility; PATCH is the preferred method.
// @Tags projects
// @Accept json
// @Produce json
// @Param id path string true "Project ID"
// @Param project body dto.UpdateProjectRequest true "Partial project data"
// @Success 200 {object} dto.ProjectResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/projects/{id} [patch]
// @Router /api/v1/projects/{id} [put]
func (h *ProjectHandler) UpdateProject(c *gin.Context) {
	id, ok := handlerutil.ParseUUIDParam(c, "id")
	if !ok {
		return
	}

	if !h.ensureProjectAccess(c, id) {
		return
	}

	var req dto.UpdateProjectRequest
	if !handlerutil.BindJSON(c, &req) {
		return
	}

	ctx := c.Request.Context()

	proj, err := h.service.GetByID(ctx, id)
	if err != nil {
		handlerutil.RespondDomainError(c, err,
			handlerutil.LocalizedError(http.StatusInternalServerError, "fetch_failed", "project.fetch_failed"),
			handlerutil.MapError(domain.ErrNotFound, handlerutil.LocalizedError(http.StatusNotFound, "not_found", "project.project_not_found")),
		)
		return
	}

	ApplyProjectUpdate(proj, req)

	if err := h.service.Update(ctx, proj); err != nil {
		handlerutil.RespondDomainError(c, err,
			handlerutil.LocalizedError(http.StatusInternalServerError, "update_failed", "project.update_failed"),
			handlerutil.MapError(domain.ErrNotFound, handlerutil.LocalizedError(http.StatusNotFound, "not_found", "project.project_not_found")),
			handlerutil.MapError(domain.ErrConflict, handlerutil.LocalizedError(http.StatusConflict, "conflict", "project.update_failed")),
			handlerutil.MapError(domain.ErrInvalidArgument, handlerutil.LocalizedError(http.StatusBadRequest, "validation_error", "project.update_failed")),
		)
		return
	}

	h.notifyProjectChange(c, id, "project.updated")

	c.JSON(http.StatusOK, ToProjectResponse(proj))
}

// DeleteProject godoc
// @Summary Delete a project
// @Tags projects
// @Produce json
// @Param id path string true "Project ID"
// @Success 204
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/projects/{id} [delete]
func (h *ProjectHandler) DeleteProject(c *gin.Context) {
	id, ok := handlerutil.ParseUUIDParam(c, "id")
	if !ok {
		return
	}

	if !h.ensureProjectAccess(c, id) {
		return
	}

	if err := h.service.DeleteByID(c.Request.Context(), id); err != nil {
		handlerutil.RespondDomainError(c, err,
			handlerutil.LocalizedError(http.StatusInternalServerError, "deletion_failed", "project.deletion_failed"),
			handlerutil.MapError(domain.ErrNotFound, handlerutil.LocalizedError(http.StatusNotFound, "not_found", "project.project_not_found")),
		)
		return
	}

	h.notifyProjectChange(c, id, "project.deleted")

	c.Status(http.StatusNoContent)
}
