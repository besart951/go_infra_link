package handler

import (
	"errors"
	"net/http"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	"github.com/besart951/go_infra_link/backend/internal/handler/dto"
	"github.com/besart951/go_infra_link/backend/internal/handler/mapper"
	"github.com/besart951/go_infra_link/backend/internal/handler/middleware"
	"github.com/besart951/go_infra_link/backend/internal/handlerutil"
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
	if !handlerutil.BindJSON(c, &req) {
		return
	}

	proj := mapper.ToProjectModel(req)

	creatorID, ok := middleware.GetUserID(c)
	if !ok {
		handlerutil.RespondError(c, http.StatusUnauthorized, "unauthorized", "User not authenticated")
		return
	}
	proj.CreatorID = creatorID

	if err := h.service.Create(proj); err != nil {
		handlerutil.RespondError(c, http.StatusInternalServerError, "creation_failed", err.Error())
		return
	}

	c.JSON(http.StatusCreated, mapper.ToProjectResponse(proj))
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

	proj, err := h.service.GetByID(id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			handlerutil.RespondNotFound(c, "Project not found")
			return
		}
		handlerutil.RespondError(c, http.StatusInternalServerError, "fetch_failed", err.Error())
		return
	}

	c.JSON(http.StatusOK, mapper.ToProjectResponse(proj))
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
	if !handlerutil.BindQuery(c, &query) {
		return
	}

	result, err := h.service.List(query.Page, query.Limit, query.Search)
	if err != nil {
		handlerutil.RespondError(c, http.StatusInternalServerError, "fetch_failed", err.Error())
		return
	}

	response := dto.ProjectListResponse{
		Items:      mapper.ToProjectListResponse(result.Items),
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
	id, ok := handlerutil.ParseUUIDParam(c, "id")
	if !ok {
		return
	}

	var req dto.UpdateProjectRequest
	if !handlerutil.BindJSON(c, &req) {
		return
	}

	proj, err := h.service.GetByID(id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			handlerutil.RespondNotFound(c, "Project not found")
			return
		}
		handlerutil.RespondError(c, http.StatusInternalServerError, "fetch_failed", err.Error())
		return
	}

	mapper.ApplyProjectUpdate(proj, req)

	if err := h.service.Update(proj); err != nil {
		handlerutil.RespondError(c, http.StatusInternalServerError, "update_failed", err.Error())
		return
	}

	c.JSON(http.StatusOK, mapper.ToProjectResponse(proj))
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

	if err := h.service.DeleteByID(id); err != nil {
		handlerutil.RespondError(c, http.StatusInternalServerError, "deletion_failed", err.Error())
		return
	}

	c.Status(http.StatusNoContent)
}

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

	var req dto.CreateProjectUserRequest
	if !handlerutil.BindJSON(c, &req) {
		return
	}

	if err := h.service.InviteUser(projectID, req.UserID); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			handlerutil.RespondNotFound(c, "Project or user not found")
			return
		}
		handlerutil.RespondError(c, http.StatusInternalServerError, "invite_failed", err.Error())
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

	users, err := h.service.ListUsers(projectID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			handlerutil.RespondNotFound(c, "Project not found")
			return
		}
		handlerutil.RespondError(c, http.StatusInternalServerError, "fetch_failed", err.Error())
		return
	}

	c.JSON(http.StatusOK, dto.ProjectUserListResponse{Items: mapper.ToUserListResponse(users)})
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
	userID, ok := handlerutil.ParseUUIDParam(c, "userId")
	if !ok {
		return
	}

	if err := h.service.RemoveUser(projectID, userID); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			handlerutil.RespondNotFound(c, "Project or user not found")
			return
		}
		handlerutil.RespondError(c, http.StatusInternalServerError, "remove_failed", err.Error())
		return
	}

	c.Status(http.StatusNoContent)
}

// CreateProjectControlCabinet godoc
// @Summary Create project control cabinet link
// @Tags projects
// @Accept json
// @Produce json
// @Param id path string true "Project ID"
// @Param link body dto.CreateProjectControlCabinetRequest true "Link data"
// @Success 201 {object} dto.ProjectControlCabinetResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/projects/{id}/control-cabinets [post]
func (h *ProjectHandler) CreateProjectControlCabinet(c *gin.Context) {
	projectID, ok := handlerutil.ParseUUIDParam(c, "id")
	if !ok {
		return
	}

	if _, err := h.service.GetByID(projectID); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			handlerutil.RespondNotFound(c, "Project not found")
			return
		}
		handlerutil.RespondError(c, http.StatusInternalServerError, "fetch_failed", err.Error())
		return
	}

	var req dto.CreateProjectControlCabinetRequest
	if !handlerutil.BindJSON(c, &req) {
		return
	}

	created, err := h.service.CreateControlCabinet(projectID, req.ControlCabinetID)
	if err != nil {
		handlerutil.RespondError(c, http.StatusInternalServerError, "creation_failed", err.Error())
		return
	}

	c.JSON(http.StatusCreated, mapper.ToProjectControlCabinetResponse(*created))
}

// UpdateProjectControlCabinet godoc
// @Summary Update project control cabinet link
// @Tags projects
// @Accept json
// @Produce json
// @Param id path string true "Project ID"
// @Param linkId path string true "Link ID"
// @Param link body dto.UpdateProjectControlCabinetRequest true "Link data"
// @Success 200 {object} dto.ProjectControlCabinetResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/projects/{id}/control-cabinets/{linkId} [put]
func (h *ProjectHandler) UpdateProjectControlCabinet(c *gin.Context) {
	projectID, ok := handlerutil.ParseUUIDParam(c, "id")
	if !ok {
		return
	}
	linkID, ok := handlerutil.ParseUUIDParam(c, "linkId")
	if !ok {
		return
	}

	if _, err := h.service.GetByID(projectID); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			handlerutil.RespondNotFound(c, "Project not found")
			return
		}
		handlerutil.RespondError(c, http.StatusInternalServerError, "fetch_failed", err.Error())
		return
	}

	var req dto.UpdateProjectControlCabinetRequest
	if !handlerutil.BindJSON(c, &req) {
		return
	}

	updated, err := h.service.UpdateControlCabinet(linkID, projectID, req.ControlCabinetID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			handlerutil.RespondNotFound(c, "Link not found")
			return
		}
		handlerutil.RespondError(c, http.StatusInternalServerError, "update_failed", err.Error())
		return
	}

	c.JSON(http.StatusOK, mapper.ToProjectControlCabinetResponse(*updated))
}

// DeleteProjectControlCabinet godoc
// @Summary Delete project control cabinet link
// @Tags projects
// @Produce json
// @Param id path string true "Project ID"
// @Param linkId path string true "Link ID"
// @Success 204
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/projects/{id}/control-cabinets/{linkId} [delete]
func (h *ProjectHandler) DeleteProjectControlCabinet(c *gin.Context) {
	projectID, ok := handlerutil.ParseUUIDParam(c, "id")
	if !ok {
		return
	}
	linkID, ok := handlerutil.ParseUUIDParam(c, "linkId")
	if !ok {
		return
	}

	if _, err := h.service.GetByID(projectID); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			handlerutil.RespondNotFound(c, "Project not found")
			return
		}
		handlerutil.RespondError(c, http.StatusInternalServerError, "fetch_failed", err.Error())
		return
	}

	if err := h.service.DeleteControlCabinet(linkID, projectID); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			handlerutil.RespondNotFound(c, "Link not found")
			return
		}
		handlerutil.RespondError(c, http.StatusInternalServerError, "deletion_failed", err.Error())
		return
	}

	c.Status(http.StatusNoContent)
}

// CreateProjectSPSController godoc
// @Summary Create project SPS controller link
// @Tags projects
// @Accept json
// @Produce json
// @Param id path string true "Project ID"
// @Param link body dto.CreateProjectSPSControllerRequest true "Link data"
// @Success 201 {object} dto.ProjectSPSControllerResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/projects/{id}/sps-controllers [post]
func (h *ProjectHandler) CreateProjectSPSController(c *gin.Context) {
	projectID, ok := handlerutil.ParseUUIDParam(c, "id")
	if !ok {
		return
	}

	if _, err := h.service.GetByID(projectID); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			handlerutil.RespondNotFound(c, "Project not found")
			return
		}
		handlerutil.RespondError(c, http.StatusInternalServerError, "fetch_failed", err.Error())
		return
	}

	var req dto.CreateProjectSPSControllerRequest
	if !handlerutil.BindJSON(c, &req) {
		return
	}

	created, err := h.service.CreateSPSController(projectID, req.SPSControllerID)
	if err != nil {
		handlerutil.RespondError(c, http.StatusInternalServerError, "creation_failed", err.Error())
		return
	}

	c.JSON(http.StatusCreated, mapper.ToProjectSPSControllerResponse(*created))
}

// UpdateProjectSPSController godoc
// @Summary Update project SPS controller link
// @Tags projects
// @Accept json
// @Produce json
// @Param id path string true "Project ID"
// @Param linkId path string true "Link ID"
// @Param link body dto.UpdateProjectSPSControllerRequest true "Link data"
// @Success 200 {object} dto.ProjectSPSControllerResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/projects/{id}/sps-controllers/{linkId} [put]
func (h *ProjectHandler) UpdateProjectSPSController(c *gin.Context) {
	projectID, ok := handlerutil.ParseUUIDParam(c, "id")
	if !ok {
		return
	}
	linkID, ok := handlerutil.ParseUUIDParam(c, "linkId")
	if !ok {
		return
	}

	if _, err := h.service.GetByID(projectID); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			handlerutil.RespondNotFound(c, "Project not found")
			return
		}
		handlerutil.RespondError(c, http.StatusInternalServerError, "fetch_failed", err.Error())
		return
	}

	var req dto.UpdateProjectSPSControllerRequest
	if !handlerutil.BindJSON(c, &req) {
		return
	}

	updated, err := h.service.UpdateSPSController(linkID, projectID, req.SPSControllerID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			handlerutil.RespondNotFound(c, "Link not found")
			return
		}
		handlerutil.RespondError(c, http.StatusInternalServerError, "update_failed", err.Error())
		return
	}

	c.JSON(http.StatusOK, mapper.ToProjectSPSControllerResponse(*updated))
}

// DeleteProjectSPSController godoc
// @Summary Delete project SPS controller link
// @Tags projects
// @Produce json
// @Param id path string true "Project ID"
// @Param linkId path string true "Link ID"
// @Success 204
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/projects/{id}/sps-controllers/{linkId} [delete]
func (h *ProjectHandler) DeleteProjectSPSController(c *gin.Context) {
	projectID, ok := handlerutil.ParseUUIDParam(c, "id")
	if !ok {
		return
	}
	linkID, ok := handlerutil.ParseUUIDParam(c, "linkId")
	if !ok {
		return
	}

	if _, err := h.service.GetByID(projectID); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			handlerutil.RespondNotFound(c, "Project not found")
			return
		}
		handlerutil.RespondError(c, http.StatusInternalServerError, "fetch_failed", err.Error())
		return
	}

	if err := h.service.DeleteSPSController(linkID, projectID); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			handlerutil.RespondNotFound(c, "Link not found")
			return
		}
		handlerutil.RespondError(c, http.StatusInternalServerError, "deletion_failed", err.Error())
		return
	}

	c.Status(http.StatusNoContent)
}

// CreateProjectFieldDevice godoc
// @Summary Create project field device link
// @Tags projects
// @Accept json
// @Produce json
// @Param id path string true "Project ID"
// @Param link body dto.CreateProjectFieldDeviceRequest true "Link data"
// @Success 201 {object} dto.ProjectFieldDeviceResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/projects/{id}/field-devices [post]
func (h *ProjectHandler) CreateProjectFieldDevice(c *gin.Context) {
	projectID, ok := handlerutil.ParseUUIDParam(c, "id")
	if !ok {
		return
	}

	if _, err := h.service.GetByID(projectID); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			handlerutil.RespondNotFound(c, "Project not found")
			return
		}
		handlerutil.RespondError(c, http.StatusInternalServerError, "fetch_failed", err.Error())
		return
	}

	var req dto.CreateProjectFieldDeviceRequest
	if !handlerutil.BindJSON(c, &req) {
		return
	}

	created, err := h.service.CreateFieldDevice(projectID, req.FieldDeviceID)
	if err != nil {
		handlerutil.RespondError(c, http.StatusInternalServerError, "creation_failed", err.Error())
		return
	}

	c.JSON(http.StatusCreated, mapper.ToProjectFieldDeviceResponse(*created))
}

// UpdateProjectFieldDevice godoc
// @Summary Update project field device link
// @Tags projects
// @Accept json
// @Produce json
// @Param id path string true "Project ID"
// @Param linkId path string true "Link ID"
// @Param link body dto.UpdateProjectFieldDeviceRequest true "Link data"
// @Success 200 {object} dto.ProjectFieldDeviceResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/projects/{id}/field-devices/{linkId} [put]
func (h *ProjectHandler) UpdateProjectFieldDevice(c *gin.Context) {
	projectID, ok := handlerutil.ParseUUIDParam(c, "id")
	if !ok {
		return
	}
	linkID, ok := handlerutil.ParseUUIDParam(c, "linkId")
	if !ok {
		return
	}

	if _, err := h.service.GetByID(projectID); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			handlerutil.RespondNotFound(c, "Project not found")
			return
		}
		handlerutil.RespondError(c, http.StatusInternalServerError, "fetch_failed", err.Error())
		return
	}

	var req dto.UpdateProjectFieldDeviceRequest
	if !handlerutil.BindJSON(c, &req) {
		return
	}

	updated, err := h.service.UpdateFieldDevice(linkID, projectID, req.FieldDeviceID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			handlerutil.RespondNotFound(c, "Link not found")
			return
		}
		handlerutil.RespondError(c, http.StatusInternalServerError, "update_failed", err.Error())
		return
	}

	c.JSON(http.StatusOK, mapper.ToProjectFieldDeviceResponse(*updated))
}

// DeleteProjectFieldDevice godoc
// @Summary Delete project field device link
// @Tags projects
// @Produce json
// @Param id path string true "Project ID"
// @Param linkId path string true "Link ID"
// @Success 204
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/projects/{id}/field-devices/{linkId} [delete]
func (h *ProjectHandler) DeleteProjectFieldDevice(c *gin.Context) {
	projectID, ok := handlerutil.ParseUUIDParam(c, "id")
	if !ok {
		return
	}
	linkID, ok := handlerutil.ParseUUIDParam(c, "linkId")
	if !ok {
		return
	}

	if _, err := h.service.GetByID(projectID); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			handlerutil.RespondNotFound(c, "Project not found")
			return
		}
		handlerutil.RespondError(c, http.StatusInternalServerError, "fetch_failed", err.Error())
		return
	}

	if err := h.service.DeleteFieldDevice(linkID, projectID); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			handlerutil.RespondNotFound(c, "Link not found")
			return
		}
		handlerutil.RespondError(c, http.StatusInternalServerError, "deletion_failed", err.Error())
		return
	}

	c.Status(http.StatusNoContent)
}

// ListProjectControlCabinets godoc
// @Summary List project control cabinets with pagination
// @Tags projects
// @Produce json
// @Param id path string true "Project ID"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {object} dto.ProjectControlCabinetListResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/projects/{id}/control-cabinets [get]
func (h *ProjectHandler) ListProjectControlCabinets(c *gin.Context) {
	projectID, ok := handlerutil.ParseUUIDParam(c, "id")
	if !ok {
		return
	}

	if _, err := h.service.GetByID(projectID); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			handlerutil.RespondNotFound(c, "Project not found")
			return
		}
		handlerutil.RespondError(c, http.StatusInternalServerError, "fetch_failed", err.Error())
		return
	}

	var query dto.PaginationQuery
	if !handlerutil.BindQuery(c, &query) {
		return
	}

	result, err := h.service.ListControlCabinets(projectID, query.Page, query.Limit)
	if err != nil {
		handlerutil.RespondError(c, http.StatusInternalServerError, "fetch_failed", err.Error())
		return
	}

	response := dto.ProjectControlCabinetListResponse{
		Items:      mapper.ToProjectControlCabinetList(result.Items),
		Total:      result.Total,
		Page:       result.Page,
		TotalPages: result.TotalPages,
	}

	c.JSON(http.StatusOK, response)
}

// ListProjectSPSControllers godoc
// @Summary List project SPS controllers with pagination
// @Tags projects
// @Produce json
// @Param id path string true "Project ID"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {object} dto.ProjectSPSControllerListResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/projects/{id}/sps-controllers [get]
func (h *ProjectHandler) ListProjectSPSControllers(c *gin.Context) {
	projectID, ok := handlerutil.ParseUUIDParam(c, "id")
	if !ok {
		return
	}

	if _, err := h.service.GetByID(projectID); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			handlerutil.RespondNotFound(c, "Project not found")
			return
		}
		handlerutil.RespondError(c, http.StatusInternalServerError, "fetch_failed", err.Error())
		return
	}

	var query dto.PaginationQuery
	if !handlerutil.BindQuery(c, &query) {
		return
	}

	result, err := h.service.ListSPSControllers(projectID, query.Page, query.Limit)
	if err != nil {
		handlerutil.RespondError(c, http.StatusInternalServerError, "fetch_failed", err.Error())
		return
	}

	response := dto.ProjectSPSControllerListResponse{
		Items:      mapper.ToProjectSPSControllerList(result.Items),
		Total:      result.Total,
		Page:       result.Page,
		TotalPages: result.TotalPages,
	}

	c.JSON(http.StatusOK, response)
}

// ListProjectFieldDevices godoc
// @Summary List project field devices with pagination
// @Tags projects
// @Produce json
// @Param id path string true "Project ID"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {object} dto.ProjectFieldDeviceListResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/projects/{id}/field-devices [get]
func (h *ProjectHandler) ListProjectFieldDevices(c *gin.Context) {
	projectID, ok := handlerutil.ParseUUIDParam(c, "id")
	if !ok {
		return
	}

	if _, err := h.service.GetByID(projectID); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			handlerutil.RespondNotFound(c, "Project not found")
			return
		}
		handlerutil.RespondError(c, http.StatusInternalServerError, "fetch_failed", err.Error())
		return
	}

	var query dto.PaginationQuery
	if !handlerutil.BindQuery(c, &query) {
		return
	}

	result, err := h.service.ListFieldDevices(projectID, query.Page, query.Limit)
	if err != nil {
		handlerutil.RespondError(c, http.StatusInternalServerError, "fetch_failed", err.Error())
		return
	}

	response := dto.ProjectFieldDeviceListResponse{
		Items:      mapper.ToProjectFieldDeviceList(result.Items),
		Total:      result.Total,
		Page:       result.Page,
		TotalPages: result.TotalPages,
	}

	c.JSON(http.StatusOK, response)
}

// ListProjectObjectData godoc
// @Summary List project object data with pagination
// @Tags projects
// @Produce json
// @Param id path string true "Project ID"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {object} dto.ObjectDataListResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/projects/{id}/object-data [get]
func (h *ProjectHandler) ListProjectObjectData(c *gin.Context) {
	projectID, ok := handlerutil.ParseUUIDParam(c, "id")
	if !ok {
		return
	}

	if _, err := h.service.GetByID(projectID); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			handlerutil.RespondNotFound(c, "Project not found")
			return
		}
		handlerutil.RespondError(c, http.StatusInternalServerError, "fetch_failed", err.Error())
		return
	}

	var query dto.PaginationQuery
	if !handlerutil.BindQuery(c, &query) {
		return
	}

	result, err := h.service.ListObjectData(projectID, query.Page, query.Limit)
	if err != nil {
		handlerutil.RespondError(c, http.StatusInternalServerError, "fetch_failed", err.Error())
		return
	}

	response := dto.ObjectDataListResponse{
		Items:      mapper.ToObjectDataList(result.Items),
		Total:      result.Total,
		Page:       result.Page,
		TotalPages: result.TotalPages,
	}

	c.JSON(http.StatusOK, response)
}

// AddProjectObjectData godoc
// @Summary Attach object data to project
// @Tags projects
// @Accept json
// @Produce json
// @Param id path string true "Project ID"
// @Param payload body dto.CreateProjectObjectDataRequest true "Object data link"
// @Success 201 {object} dto.ObjectDataResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 409 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/projects/{id}/object-data [post]
func (h *ProjectHandler) AddProjectObjectData(c *gin.Context) {
	projectID, ok := handlerutil.ParseUUIDParam(c, "id")
	if !ok {
		return
	}

	var req dto.CreateProjectObjectDataRequest
	if !handlerutil.BindJSON(c, &req) {
		return
	}

	obj, err := h.service.AddObjectData(projectID, req.ObjectDataID)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrNotFound):
			handlerutil.RespondNotFound(c, "Project or object data not found")
			return
		case errors.Is(err, domain.ErrConflict):
			handlerutil.RespondError(c, http.StatusConflict, "conflict", "Object data already linked to another project")
			return
		default:
			handlerutil.RespondError(c, http.StatusInternalServerError, "update_failed", err.Error())
			return
		}
	}

	c.JSON(http.StatusCreated, mapper.ToObjectDataResponse(*obj))
}

// RemoveProjectObjectData godoc
// @Summary Detach object data from project
// @Tags projects
// @Produce json
// @Param id path string true "Project ID"
// @Param objectDataId path string true "Object Data ID"
// @Success 200 {object} dto.ObjectDataResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/projects/{id}/object-data/{objectDataId} [delete]
func (h *ProjectHandler) RemoveProjectObjectData(c *gin.Context) {
	projectID, ok := handlerutil.ParseUUIDParam(c, "id")
	if !ok {
		return
	}
	objectDataID, ok := handlerutil.ParseUUIDParam(c, "objectDataId")
	if !ok {
		return
	}

	obj, err := h.service.RemoveObjectData(projectID, objectDataID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			handlerutil.RespondNotFound(c, "Project or object data not found")
			return
		}
		handlerutil.RespondError(c, http.StatusInternalServerError, "update_failed", err.Error())
		return
	}

	c.JSON(http.StatusOK, mapper.ToObjectDataResponse(*obj))
}
