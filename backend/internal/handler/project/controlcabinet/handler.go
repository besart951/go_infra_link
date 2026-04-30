package controlcabinet

import (
	"context"
	"net/http"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	domainProject "github.com/besart951/go_infra_link/backend/internal/domain/project"
	domainUser "github.com/besart951/go_infra_link/backend/internal/domain/user"
	dto "github.com/besart951/go_infra_link/backend/internal/handler/dto/project"
	sharedpresenter "github.com/besart951/go_infra_link/backend/internal/handler/presenter/shared"
	projectshared "github.com/besart951/go_infra_link/backend/internal/handler/project/shared"
	"github.com/besart951/go_infra_link/backend/internal/handlerutil"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type FacilityLinkService interface {
	CreateControlCabinet(ctx context.Context, projectID, controlCabinetID uuid.UUID) (*domainProject.ProjectControlCabinet, error)
	CopyControlCabinet(ctx context.Context, projectID, controlCabinetID uuid.UUID) (*domainFacility.ControlCabinet, error)
	UpdateControlCabinet(ctx context.Context, linkID, projectID, controlCabinetID uuid.UUID) (*domainProject.ProjectControlCabinet, error)
	DeleteControlCabinet(ctx context.Context, linkID, projectID uuid.UUID) error
	ListControlCabinets(ctx context.Context, projectID uuid.UUID, page, limit int) (*domain.PaginatedList[domainProject.ProjectControlCabinet], error)
}

type Handler struct {
	access       projectshared.AccessPolicyService
	facilityLink FacilityLinkService
	notify       projectshared.ProjectChangeNotifier
	notifyDelta  ProjectControlCabinetDeltaNotifier
}

type ProjectControlCabinetDeltaNotifier func(*gin.Context, uuid.UUID, domainFacility.ControlCabinet)

func NewHandler(access projectshared.AccessPolicyService, facilityLink FacilityLinkService, notify projectshared.ProjectChangeNotifier, notifyDelta ...ProjectControlCabinetDeltaNotifier) *Handler {
	var delta ProjectControlCabinetDeltaNotifier
	if len(notifyDelta) > 0 {
		delta = notifyDelta[0]
	}
	return &Handler{access: access, facilityLink: facilityLink, notify: notify, notifyDelta: delta}
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
func (h *Handler) CreateProjectControlCabinet(c *gin.Context) {
	projectID, ok := handlerutil.ParseUUIDParam(c, "id")
	if !ok {
		return
	}

	if !projectshared.EnsureProjectAccessAndAnyPermission(
		c,
		h.access,
		projectID,
		domainUser.PermissionProjectControlCabinetCreate,
	) {
		return
	}

	var req dto.CreateProjectControlCabinetRequest
	if !handlerutil.BindJSON(c, &req) {
		return
	}

	created, err := h.facilityLink.CreateControlCabinet(c.Request.Context(), projectID, req.ControlCabinetID)
	if err != nil {
		handlerutil.RespondDomainError(c, err,
			handlerutil.LocalizedError(http.StatusInternalServerError, "creation_failed", "project.creation_failed"),
			handlerutil.MapError(domain.ErrNotFound, handlerutil.LocalizedError(http.StatusNotFound, "not_found", "facility.control_cabinet_not_found")),
			handlerutil.MapError(domain.ErrConflict, handlerutil.LocalizedError(http.StatusConflict, "conflict", "project.creation_failed")),
		)
		return
	}

	if h.notify != nil {
		h.notify(c, projectID, "project.control_cabinet.created", created.ControlCabinetID.String())
	}

	c.JSON(http.StatusCreated, toProjectControlCabinetResponse(*created))
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
func (h *Handler) ListProjectControlCabinets(c *gin.Context) {
	projectID, ok := handlerutil.ParseUUIDParam(c, "id")
	if !ok {
		return
	}

	if !projectshared.EnsureProjectAccess(c, h.access, projectID) {
		return
	}

	var query dto.PaginationQuery
	if !handlerutil.BindQuery(c, &query) {
		return
	}

	result, err := h.facilityLink.ListControlCabinets(c.Request.Context(), projectID, query.Page, query.Limit)
	if err != nil {
		handlerutil.RespondLocalizedError(c, http.StatusInternalServerError, "fetch_failed", "project.fetch_failed")
		return
	}

	response := dto.ProjectControlCabinetListResponse{
		Items:      toProjectControlCabinetList(result.Items),
		Total:      result.Total,
		Page:       result.Page,
		TotalPages: result.TotalPages,
	}

	c.JSON(http.StatusOK, response)
}

func (h *Handler) CopyProjectControlCabinet(c *gin.Context) {
	projectID, ok := handlerutil.ParseUUIDParam(c, "id")
	if !ok {
		return
	}

	if !projectshared.EnsureProjectAccessAndAnyPermission(
		c,
		h.access,
		projectID,
		domainUser.PermissionProjectControlCabinetCreate,
	) {
		return
	}

	controlCabinetID, ok := handlerutil.ParseUUIDParam(c, "controlCabinetId")
	if !ok {
		return
	}

	copyEntity, err := h.facilityLink.CopyControlCabinet(c.Request.Context(), projectID, controlCabinetID)
	if err != nil {
		handlerutil.RespondDomainError(c, err,
			handlerutil.LocalizedError(http.StatusInternalServerError, "creation_failed", "project.creation_failed"),
			handlerutil.MapError(domain.ErrNotFound, handlerutil.LocalizedError(http.StatusNotFound, "not_found", "facility.control_cabinet_not_found")),
			handlerutil.MapError(domain.ErrConflict, handlerutil.LocalizedError(http.StatusConflict, "conflict", "project.creation_failed")),
		)
		return
	}

	if h.notifyDelta != nil {
		h.notifyDelta(c, projectID, *copyEntity)
	} else if h.notify != nil {
		h.notify(c, projectID, "project.control_cabinet.copied", copyEntity.ID.String())
	}

	c.JSON(http.StatusCreated, sharedpresenter.ToControlCabinetResponse(*copyEntity))
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
func (h *Handler) UpdateProjectControlCabinet(c *gin.Context) {
	projectID, ok := handlerutil.ParseUUIDParam(c, "id")
	if !ok {
		return
	}

	if !projectshared.EnsureProjectAccessAndAnyPermission(
		c,
		h.access,
		projectID,
		domainUser.PermissionProjectControlCabinetUpdate,
	) {
		return
	}

	linkID, ok := handlerutil.ParseUUIDParam(c, "linkId")
	if !ok {
		return
	}

	var req dto.UpdateProjectControlCabinetRequest
	if !handlerutil.BindJSON(c, &req) {
		return
	}

	updated, err := h.facilityLink.UpdateControlCabinet(c.Request.Context(), linkID, projectID, req.ControlCabinetID)
	if err != nil {
		handlerutil.RespondDomainError(c, err,
			handlerutil.LocalizedError(http.StatusInternalServerError, "update_failed", "project.update_failed"),
			handlerutil.MapError(domain.ErrNotFound, handlerutil.LocalizedError(http.StatusNotFound, "not_found", "project.link_not_found")),
		)
		return
	}

	if h.notify != nil {
		h.notify(c, projectID, "project.control_cabinet.updated", updated.ControlCabinetID.String())
	}

	c.JSON(http.StatusOK, toProjectControlCabinetResponse(*updated))
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
func (h *Handler) DeleteProjectControlCabinet(c *gin.Context) {
	projectID, ok := handlerutil.ParseUUIDParam(c, "id")
	if !ok {
		return
	}

	if !projectshared.EnsureProjectAccessAndAnyPermission(
		c,
		h.access,
		projectID,
		domainUser.PermissionProjectControlCabinetDelete,
	) {
		return
	}

	linkID, ok := handlerutil.ParseUUIDParam(c, "linkId")
	if !ok {
		return
	}

	if err := h.facilityLink.DeleteControlCabinet(c.Request.Context(), linkID, projectID); err != nil {
		handlerutil.RespondDomainError(c, err,
			handlerutil.LocalizedError(http.StatusInternalServerError, "deletion_failed", "project.deletion_failed"),
			handlerutil.MapError(domain.ErrNotFound, handlerutil.LocalizedError(http.StatusNotFound, "not_found", "project.link_not_found")),
		)
		return
	}

	if h.notify != nil {
		h.notify(c, projectID, "project.control_cabinet.deleted")
	}

	c.Status(http.StatusNoContent)
}

func toProjectControlCabinetResponse(item domainProject.ProjectControlCabinet) dto.ProjectControlCabinetResponse {
	return dto.ProjectControlCabinetResponse{
		ID:               item.ID,
		ProjectID:        item.ProjectID,
		ControlCabinetID: item.ControlCabinetID,
		CreatedAt:        item.CreatedAt,
		UpdatedAt:        item.UpdatedAt,
	}
}

func toProjectControlCabinetList(items []domainProject.ProjectControlCabinet) []dto.ProjectControlCabinetResponse {
	out := make([]dto.ProjectControlCabinetResponse, len(items))
	for i, item := range items {
		out[i] = toProjectControlCabinetResponse(item)
	}
	return out
}
