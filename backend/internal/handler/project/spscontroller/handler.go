package spscontroller

import (
	"context"
	"net/http"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	domainProject "github.com/besart951/go_infra_link/backend/internal/domain/project"
	domainUser "github.com/besart951/go_infra_link/backend/internal/domain/user"
	facilitydto "github.com/besart951/go_infra_link/backend/internal/handler/dto/facility"
	dto "github.com/besart951/go_infra_link/backend/internal/handler/dto/project"
	projectshared "github.com/besart951/go_infra_link/backend/internal/handler/project/shared"
	"github.com/besart951/go_infra_link/backend/internal/handlerutil"
	facilityservice "github.com/besart951/go_infra_link/backend/internal/service/facility"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type FacilityLinkService interface {
	CreateSPSController(ctx context.Context, projectID, spsControllerID uuid.UUID) (*domainProject.ProjectSPSController, error)
	CopySPSController(ctx context.Context, projectID, spsControllerID uuid.UUID) (*domainFacility.SPSController, error)
	CopySPSControllerSystemType(ctx context.Context, projectID, systemTypeID uuid.UUID) (*domainFacility.SPSControllerSystemType, error)
	UpdateSPSController(ctx context.Context, linkID, projectID, spsControllerID uuid.UUID) (*domainProject.ProjectSPSController, error)
	DeleteSPSController(ctx context.Context, linkID, projectID uuid.UUID) error
	ListSPSControllers(ctx context.Context, projectID uuid.UUID, page, limit int) (*domain.PaginatedList[domainProject.ProjectSPSController], error)
}

type Handler struct {
	base         *projectshared.FacilityLinkHandler
	facilityLink FacilityLinkService
	notifyDelta  ProjectSPSControllerDeltaNotifier
}

type ProjectSPSControllerDeltaNotifier func(*gin.Context, uuid.UUID, domainFacility.SPSController)

// copyJobResponseContract keeps Swag's imported DTO reference available while
// the shared project-link module writes the asynchronous response.
type copyJobResponseContract = facilitydto.CopyJobResponse

func NewHandler(access projectshared.AccessPolicyService, facilityLink FacilityLinkService, notify projectshared.ProjectChangeNotifier, notifyDelta ...ProjectSPSControllerDeltaNotifier) *Handler {
	var delta ProjectSPSControllerDeltaNotifier
	if len(notifyDelta) > 0 {
		delta = notifyDelta[0]
	}
	return &Handler{base: projectshared.NewFacilityLinkHandler(access, notify), facilityLink: facilityLink, notifyDelta: delta}
}

func (h *Handler) ConfigureCopyJobs(copyJobs *facilityservice.CopyJobManager) {
	h.base.ConfigureCopyJobs(copyJobs)
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
func (h *Handler) CreateProjectSPSController(c *gin.Context) {
	projectID, ok := h.base.ProjectIDWithPermission(c, domainUser.PermissionProjectSPSControllerCreate)
	if !ok {
		return
	}

	var req dto.CreateProjectSPSControllerRequest
	if !handlerutil.BindJSON(c, &req) {
		return
	}

	created, err := h.facilityLink.CreateSPSController(c.Request.Context(), projectID, req.SPSControllerID)
	if err != nil {
		handlerutil.RespondDomainError(c, err,
			handlerutil.LocalizedError(http.StatusInternalServerError, "creation_failed", "project.creation_failed"),
			handlerutil.MapError(domain.ErrNotFound, handlerutil.LocalizedError(http.StatusNotFound, "not_found", "facility.sps_controller_not_found")),
			handlerutil.MapError(domain.ErrConflict, handlerutil.LocalizedError(http.StatusConflict, "conflict", "project.creation_failed")),
		)
		return
	}

	h.base.Notify(c, projectID, "project.sps_controller.created", created.SPSControllerID.String())

	c.JSON(http.StatusCreated, toProjectSPSControllerResponse(*created))
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
func (h *Handler) ListProjectSPSControllers(c *gin.Context) {
	projectID, ok := h.base.ProjectIDWithPermission(c, domainUser.PermissionProjectSPSControllerRead)
	if !ok {
		return
	}

	var query dto.PaginationQuery
	if !handlerutil.BindQuery(c, &query) {
		return
	}

	result, err := h.facilityLink.ListSPSControllers(c.Request.Context(), projectID, query.Page, query.Limit)
	if err != nil {
		handlerutil.RespondLocalizedError(c, http.StatusInternalServerError, "fetch_failed", "project.fetch_failed")
		return
	}

	response := dto.ProjectSPSControllerListResponse{
		Items:      toProjectSPSControllerList(result.Items),
		Total:      result.Total,
		Page:       result.Page,
		TotalPages: result.TotalPages,
	}

	c.JSON(http.StatusOK, response)
}

// CopyProjectSPSController godoc
// @Summary Copy a project SPS controller asynchronously
// @Tags projects
// @Produce json
// @Param id path string true "Project ID"
// @Param spsControllerId path string true "SPS Controller ID"
// @Param X-Copy-Operation-ID header string false "Client-generated copy operation UUID"
// @Success 202 {object} facilitydto.CopyJobResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 503 {object} dto.ErrorResponse
// @Router /api/v1/projects/{id}/sps-controllers/{spsControllerId}/copy [post]
func (h *Handler) CopyProjectSPSController(c *gin.Context) {
	projectID, ok := h.base.ProjectIDWithPermission(c, domainUser.PermissionProjectSPSControllerCreate)
	if !ok {
		return
	}

	spsControllerID, ok := handlerutil.ParseUUIDParam(c, "spsControllerId")
	if !ok {
		return
	}
	h.base.StartCopy(c, projectshared.ProjectCopyCommand{
		ProjectID: projectID,
		SourceID:  spsControllerID,
		Kind:      facilityservice.CopyJobKindSPSController,
		Task:      domainProject.TaskCopyProjectSPSController,
	})
}

// CopyProjectSPSControllerSystemType godoc
// @Summary Copy a project SPS controller system type asynchronously
// @Tags projects
// @Produce json
// @Param id path string true "Project ID"
// @Param systemTypeId path string true "SPS Controller System Type ID"
// @Param X-Copy-Operation-ID header string false "Client-generated copy operation UUID"
// @Success 202 {object} facilitydto.CopyJobResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 503 {object} dto.ErrorResponse
// @Router /api/v1/projects/{id}/sps-controller-system-types/{systemTypeId}/copy [post]
func (h *Handler) CopyProjectSPSControllerSystemType(c *gin.Context) {
	projectID, ok := h.base.ProjectIDWithPermission(c, domainUser.PermissionProjectSPSControllerSystemTypeCreate)
	if !ok {
		return
	}

	systemTypeID, ok := handlerutil.ParseUUIDParam(c, "systemTypeId")
	if !ok {
		return
	}
	h.base.StartCopy(c, projectshared.ProjectCopyCommand{
		ProjectID: projectID,
		SourceID:  systemTypeID,
		Kind:      facilityservice.CopyJobKindSPSControllerSystemType,
		Task:      domainProject.TaskCopyProjectSPSControllerSystemType,
	})
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
func (h *Handler) UpdateProjectSPSController(c *gin.Context) {
	projectID, ok := h.base.ProjectIDWithPermission(c, domainUser.PermissionProjectSPSControllerUpdate)
	if !ok {
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

	var updated *domainProject.ProjectSPSController
	var err error
	if req.BaseVersion != nil {
		if versioned, ok := h.facilityLink.(interface {
			UpdateSPSControllerAtVersion(context.Context, uuid.UUID, uuid.UUID, uuid.UUID, uint64) (*domainProject.ProjectSPSController, error)
		}); ok {
			updated, err = versioned.UpdateSPSControllerAtVersion(c.Request.Context(), linkID, projectID, req.SPSControllerID, *req.BaseVersion)
		} else {
			updated, err = h.facilityLink.UpdateSPSController(c.Request.Context(), linkID, projectID, req.SPSControllerID)
		}
	} else {
		updated, err = h.facilityLink.UpdateSPSController(c.Request.Context(), linkID, projectID, req.SPSControllerID)
	}
	if err != nil {
		handlerutil.RespondDomainError(c, err,
			handlerutil.LocalizedError(http.StatusInternalServerError, "update_failed", "project.update_failed"),
			handlerutil.MapError(domain.ErrNotFound, handlerutil.LocalizedError(http.StatusNotFound, "not_found", "project.link_not_found")),
			handlerutil.MapError(domain.ErrConflict, handlerutil.LocalizedError(http.StatusConflict, "write_conflict", "errors.write_conflict")),
		)
		return
	}

	h.base.Notify(c, projectID, "project.sps_controller.updated", updated.SPSControllerID.String())

	c.JSON(http.StatusOK, toProjectSPSControllerResponse(*updated))
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
func (h *Handler) DeleteProjectSPSController(c *gin.Context) {
	projectID, ok := h.base.ProjectIDWithPermission(c, domainUser.PermissionProjectSPSControllerDelete)
	if !ok {
		return
	}

	linkID, ok := handlerutil.ParseUUIDParam(c, "linkId")
	if !ok {
		return
	}

	if err := h.facilityLink.DeleteSPSController(c.Request.Context(), linkID, projectID); err != nil {
		handlerutil.RespondDomainError(c, err,
			handlerutil.LocalizedError(http.StatusInternalServerError, "deletion_failed", "project.deletion_failed"),
			handlerutil.MapError(domain.ErrNotFound, handlerutil.LocalizedError(http.StatusNotFound, "not_found", "project.link_not_found")),
		)
		return
	}

	h.base.Notify(c, projectID, "project.sps_controller.deleted")

	c.Status(http.StatusNoContent)
}

func toProjectSPSControllerResponse(item domainProject.ProjectSPSController) dto.ProjectSPSControllerResponse {
	return dto.ProjectSPSControllerResponse{
		ID:              item.ID,
		ProjectID:       item.ProjectID,
		SPSControllerID: item.SPSControllerID,
		CreatedAt:       item.CreatedAt,
		UpdatedAt:       item.UpdatedAt,
		Version:         item.Version,
	}
}

func toProjectSPSControllerList(items []domainProject.ProjectSPSController) []dto.ProjectSPSControllerResponse {
	out := make([]dto.ProjectSPSControllerResponse, len(items))
	for i, item := range items {
		out[i] = toProjectSPSControllerResponse(item)
	}
	return out
}
