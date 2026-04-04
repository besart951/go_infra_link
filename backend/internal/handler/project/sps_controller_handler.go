package project

import (
	"net/http"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	dto "github.com/besart951/go_infra_link/backend/internal/handler/dto/project"
	facilityhandler "github.com/besart951/go_infra_link/backend/internal/handler/facility"
	"github.com/besart951/go_infra_link/backend/internal/handlerutil"
	"github.com/gin-gonic/gin"
)

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

	if !h.ensureProjectAccess(c, projectID) {
		return
	}

	var req dto.CreateProjectSPSControllerRequest
	if !handlerutil.BindJSON(c, &req) {
		return
	}

	created, err := h.service.CreateSPSController(c.Request.Context(), projectID, req.SPSControllerID)
	if err != nil {
		handlerutil.RespondLocalizedError(c, http.StatusInternalServerError, "creation_failed", "project.creation_failed")
		return
	}

	h.notifyProjectChange(c, projectID, "project.sps_controller.created")

	c.JSON(http.StatusCreated, ToProjectSPSControllerResponse(*created))
}

func (h *ProjectHandler) CopyProjectSPSController(c *gin.Context) {
	projectID, ok := handlerutil.ParseUUIDParam(c, "id")
	if !ok {
		return
	}
	if !h.ensureProjectAccess(c, projectID) {
		return
	}

	spsControllerID, ok := handlerutil.ParseUUIDParam(c, "spsControllerId")
	if !ok {
		return
	}

	copyEntity, err := h.service.CopySPSController(c.Request.Context(), projectID, spsControllerID)
	if err != nil {
		handlerutil.RespondDomainError(c, err,
			handlerutil.LocalizedError(http.StatusInternalServerError, "creation_failed", "project.creation_failed"),
			handlerutil.MapError(domain.ErrNotFound, handlerutil.LocalizedError(http.StatusNotFound, "not_found", "facility.sps_controller_not_found")),
			handlerutil.MapError(domain.ErrConflict, handlerutil.LocalizedError(http.StatusConflict, "conflict", "project.creation_failed")),
		)
		return
	}

	h.notifyProjectChange(c, projectID, "project.sps_controller.copied")
	c.JSON(http.StatusCreated, facilityhandler.ToSPSControllerResponse(*copyEntity))
}

func (h *ProjectHandler) CopyProjectSPSControllerSystemType(c *gin.Context) {
	projectID, ok := handlerutil.ParseUUIDParam(c, "id")
	if !ok {
		return
	}
	if !h.ensureProjectAccess(c, projectID) {
		return
	}

	systemTypeID, ok := handlerutil.ParseUUIDParam(c, "systemTypeId")
	if !ok {
		return
	}

	copyEntity, err := h.service.CopySPSControllerSystemType(c.Request.Context(), projectID, systemTypeID)
	if err != nil {
		handlerutil.RespondDomainError(c, err,
			handlerutil.LocalizedError(http.StatusInternalServerError, "creation_failed", "project.creation_failed"),
			handlerutil.MapError(domain.ErrNotFound, handlerutil.LocalizedError(http.StatusNotFound, "not_found", "facility.sps_controller_system_type_not_found")),
		)
		return
	}

	h.notifyProjectChange(c, projectID, "project.sps_controller_system_type.copied")
	c.JSON(http.StatusCreated, facilityhandler.ToSPSControllerSystemTypeResponse(*copyEntity))
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

	if !h.ensureProjectAccess(c, projectID) {
		return
	}

	linkID, ok := handlerutil.ParseUUIDParam(c, "linkId")
	if !ok {
		return
	}

	var req dto.UpdateProjectSPSControllerRequest
	if !handlerutil.BindJSON(c, &req) {
		return
	}

	updated, err := h.service.UpdateSPSController(c.Request.Context(), linkID, projectID, req.SPSControllerID)
	if err != nil {
		handlerutil.RespondDomainError(c, err,
			handlerutil.LocalizedError(http.StatusInternalServerError, "update_failed", "project.update_failed"),
			handlerutil.MapError(domain.ErrNotFound, handlerutil.LocalizedError(http.StatusNotFound, "not_found", "project.link_not_found")),
		)
		return
	}

	h.notifyProjectChange(c, projectID, "project.sps_controller.updated")

	c.JSON(http.StatusOK, ToProjectSPSControllerResponse(*updated))
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

	if !h.ensureProjectAccess(c, projectID) {
		return
	}

	linkID, ok := handlerutil.ParseUUIDParam(c, "linkId")
	if !ok {
		return
	}

	if err := h.service.DeleteSPSController(c.Request.Context(), linkID, projectID); err != nil {
		handlerutil.RespondDomainError(c, err,
			handlerutil.LocalizedError(http.StatusInternalServerError, "deletion_failed", "project.deletion_failed"),
			handlerutil.MapError(domain.ErrNotFound, handlerutil.LocalizedError(http.StatusNotFound, "not_found", "project.link_not_found")),
		)
		return
	}

	h.notifyProjectChange(c, projectID, "project.sps_controller.deleted")

	c.Status(http.StatusNoContent)
}
