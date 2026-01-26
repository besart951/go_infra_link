package handler

import (
	"errors"
	"net/http"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	"github.com/besart951/go_infra_link/backend/internal/domain/project"
	"github.com/besart951/go_infra_link/backend/internal/handler/dto"
	"github.com/gin-gonic/gin"
)

type ProjectHandler struct {
	service ProjectService
}

func NewProjectHandler(service ProjectService) *ProjectHandler {
	return &ProjectHandler{service: service}
}

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
	if !BindJSON(c, &req) {
		return
	}

	proj := &project.Project{
		Name:        req.Name,
		Description: req.Description,
		Status:      project.ProjectStatus(req.Status),
		StartDate:   req.StartDate,
		CreatorID:   req.CreatorID,
	}

	if req.PhaseID != nil {
		proj.PhaseID = *req.PhaseID
	}

	if err := h.service.Create(proj); err != nil {
		RespondError(c, http.StatusInternalServerError, "creation_failed", err.Error())
		return
	}

	response := dto.ProjectResponse{
		ID:          proj.ID,
		Name:        proj.Name,
		Description: proj.Description,
		Status:      proj.Status,
		StartDate:   proj.StartDate,
		PhaseID:     proj.PhaseID,
		CreatorID:   proj.CreatorID,
		CreatedAt:   proj.CreatedAt,
		UpdatedAt:   proj.UpdatedAt,
	}

	c.JSON(http.StatusCreated, response)
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
	id, ok := ParseUUIDParam(c, "id")
	if !ok {
		return
	}

	proj, err := h.service.GetByID(id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			RespondNotFound(c, "Project not found")
			return
		}
		RespondError(c, http.StatusInternalServerError, "fetch_failed", err.Error())
		return
	}

	response := dto.ProjectResponse{
		ID:          proj.ID,
		Name:        proj.Name,
		Description: proj.Description,
		Status:      proj.Status,
		StartDate:   proj.StartDate,
		PhaseID:     proj.PhaseID,
		CreatorID:   proj.CreatorID,
		CreatedAt:   proj.CreatedAt,
		UpdatedAt:   proj.UpdatedAt,
	}

	c.JSON(http.StatusOK, response)
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
	var query dto.PaginationQuery
	if !BindQuery(c, &query) {
		return
	}

	result, err := h.service.List(query.Page, query.Limit, query.Search)
	if err != nil {
		RespondError(c, http.StatusInternalServerError, "fetch_failed", err.Error())
		return
	}

	items := make([]dto.ProjectResponse, len(result.Items))
	for i, proj := range result.Items {
		items[i] = dto.ProjectResponse{
			ID:          proj.ID,
			Name:        proj.Name,
			Description: proj.Description,
			Status:      proj.Status,
			StartDate:   proj.StartDate,
			PhaseID:     proj.PhaseID,
			CreatorID:   proj.CreatorID,
			CreatedAt:   proj.CreatedAt,
			UpdatedAt:   proj.UpdatedAt,
		}
	}

	response := dto.ProjectListResponse{
		Items:      items,
		Total:      result.Total,
		Page:       result.Page,
		TotalPages: result.TotalPages,
	}

	c.JSON(http.StatusOK, response)
}

// UpdateProject godoc
// @Summary Update a project
// @Tags projects
// @Accept json
// @Produce json
// @Param id path string true "Project ID"
// @Param project body dto.UpdateProjectRequest true "Project data"
// @Success 200 {object} dto.ProjectResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/projects/{id} [put]
func (h *ProjectHandler) UpdateProject(c *gin.Context) {
	id, ok := ParseUUIDParam(c, "id")
	if !ok {
		return
	}

	var req dto.UpdateProjectRequest
	if !BindJSON(c, &req) {
		return
	}

	proj, err := h.service.GetByID(id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			RespondNotFound(c, "Project not found")
			return
		}
		RespondError(c, http.StatusInternalServerError, "fetch_failed", err.Error())
		return
	}

	if req.Name != "" {
		proj.Name = req.Name
	}
	if req.Description != "" {
		proj.Description = req.Description
	}
	if req.Status != "" {
		proj.Status = req.Status
	}
	if req.StartDate != nil {
		proj.StartDate = req.StartDate
	}
	if req.PhaseID != nil {
		proj.PhaseID = *req.PhaseID
	}

	if err := h.service.Update(proj); err != nil {
		RespondError(c, http.StatusInternalServerError, "update_failed", err.Error())
		return
	}

	response := dto.ProjectResponse{
		ID:          proj.ID,
		Name:        proj.Name,
		Description: proj.Description,
		Status:      proj.Status,
		StartDate:   proj.StartDate,
		PhaseID:     proj.PhaseID,
		CreatorID:   proj.CreatorID,
		CreatedAt:   proj.CreatedAt,
		UpdatedAt:   proj.UpdatedAt,
	}

	c.JSON(http.StatusOK, response)
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
	id, ok := ParseUUIDParam(c, "id")
	if !ok {
		return
	}

	if err := h.service.DeleteByID(id); err != nil {
		RespondError(c, http.StatusInternalServerError, "deletion_failed", err.Error())
		return
	}

	c.Status(http.StatusNoContent)
}
