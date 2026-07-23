package fielddevice

import (
	"context"
	"net/http"

	appfielddevice "github.com/besart951/go_infra_link/backend/internal/application/facility/fielddevice"
	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	domainProject "github.com/besart951/go_infra_link/backend/internal/domain/project"
	domainUser "github.com/besart951/go_infra_link/backend/internal/domain/user"
	facilitydto "github.com/besart951/go_infra_link/backend/internal/handler/dto/facility"
	dto "github.com/besart951/go_infra_link/backend/internal/handler/dto/project"
	sharedpresenter "github.com/besart951/go_infra_link/backend/internal/handler/presenter/shared"
	projectshared "github.com/besart951/go_infra_link/backend/internal/handler/project/shared"
	"github.com/besart951/go_infra_link/backend/internal/handlerutil"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type FacilityLinkService interface {
	DeleteFieldDevice(ctx context.Context, linkID, projectID uuid.UUID) error
	ListFieldDevices(ctx context.Context, projectID uuid.UUID, page, limit int) (*domain.PaginatedList[domainProject.ProjectFieldDevice], error)
}

type ProjectFieldDeviceMultiCreator interface {
	MultiCreateForProject(
		context.Context,
		appfielddevice.MultiCreateForProjectCommand,
	) (*domainFacility.FieldDeviceMultiCreateResult, error)
}

type ProjectFieldDeviceAssigner interface {
	AssignToProject(
		context.Context,
		appfielddevice.AssignToProjectCommand,
	) (*domainProject.ProjectFieldDevice, error)
}

type ProjectFieldDeviceBulkAssigner interface {
	BulkAssignToProject(
		context.Context,
		appfielddevice.BulkAssignToProjectCommand,
	) appfielddevice.BulkAssignToProjectResult
}

type ProjectFieldDeviceReassigner interface {
	ReassignProjectLink(
		context.Context,
		appfielddevice.ReassignProjectLinkCommand,
	) (*domainProject.ProjectFieldDevice, error)
}

type OptionsService interface {
	GetFieldDeviceOptionsForProject(ctx context.Context, projectID uuid.UUID) (*domainFacility.FieldDeviceOptions, error)
}

type Handler struct {
	access       projectshared.AccessPolicyService
	facilityLink FacilityLinkService
	multiCreator ProjectFieldDeviceMultiCreator
	assigner     ProjectFieldDeviceAssigner
	bulkAssigner ProjectFieldDeviceBulkAssigner
	reassigner   ProjectFieldDeviceReassigner
	notify       projectshared.ProjectChangeNotifier
	notifyEvent  projectshared.ProjectChangeNotifier
}

type OptionsHandler struct {
	access  projectshared.AccessPolicyService
	service OptionsService
}

func NewHandler(
	access projectshared.AccessPolicyService,
	facilityLink FacilityLinkService,
	multiCreator ProjectFieldDeviceMultiCreator,
	assigner ProjectFieldDeviceAssigner,
	bulkAssigner ProjectFieldDeviceBulkAssigner,
	reassigner ProjectFieldDeviceReassigner,
	notify projectshared.ProjectChangeNotifier,
	notifyEvent projectshared.ProjectChangeNotifier,
) *Handler {
	return &Handler{
		access:       access,
		facilityLink: facilityLink,
		multiCreator: multiCreator,
		assigner:     assigner,
		bulkAssigner: bulkAssigner,
		reassigner:   reassigner,
		notify:       notify,
		notifyEvent:  notifyEvent,
	}
}

func NewOptionsHandler(access projectshared.AccessPolicyService, service OptionsService) *OptionsHandler {
	return &OptionsHandler{access: access, service: service}
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
func (h *Handler) CreateProjectFieldDevice(c *gin.Context) {
	projectID, ok := handlerutil.ParseUUIDParam(c, "id")
	if !ok {
		return
	}

	if !projectshared.EnsureProjectAccessAndAnyPermission(
		c,
		h.access,
		projectID,
		domainUser.PermissionProjectFieldDeviceCreate,
	) {
		return
	}

	var req dto.CreateProjectFieldDeviceRequest
	if !handlerutil.BindJSON(c, &req) {
		return
	}

	if h.assigner == nil {
		handlerutil.RespondLocalizedError(c, http.StatusInternalServerError, "creation_failed", "project.creation_failed")
		return
	}
	created, err := h.assigner.AssignToProject(
		c.Request.Context(),
		appfielddevice.AssignToProjectCommand{
			ProjectID:     projectID,
			FieldDeviceID: req.FieldDeviceID,
		},
	)
	if err != nil {
		handlerutil.RespondDomainError(c, err,
			handlerutil.LocalizedError(http.StatusInternalServerError, "creation_failed", "project.creation_failed"),
			handlerutil.MapError(domain.ErrNotFound, handlerutil.LocalizedError(http.StatusNotFound, "not_found", "facility.field_device_not_found")),
			handlerutil.MapError(domain.ErrConflict, handlerutil.LocalizedError(http.StatusConflict, "conflict", "project.creation_failed")),
		)
		return
	}

	if h.notifyEvent != nil {
		h.notifyEvent(c, projectID, "project.field_device.created", created.FieldDeviceID.String())
	}

	c.JSON(http.StatusCreated, toProjectFieldDeviceResponse(*created))
}

// MultiCreateProjectFieldDevices godoc
// @Summary Create multiple project field device links
// @Tags projects
// @Accept json
// @Produce json
// @Param id path string true "Project ID"
// @Param link body dto.MultiCreateProjectFieldDeviceRequest true "Link data"
// @Success 200 {object} dto.MultiCreateProjectFieldDeviceResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/projects/{id}/field-devices/multi-create [post]
func (h *Handler) MultiCreateProjectFieldDevices(c *gin.Context) {
	projectID, ok := handlerutil.ParseUUIDParam(c, "id")
	if !ok {
		return
	}

	if !projectshared.EnsureProjectAccessAndAnyPermission(
		c,
		h.access,
		projectID,
		domainUser.PermissionProjectFieldDeviceCreate,
	) {
		return
	}

	var req dto.MultiCreateProjectFieldDeviceRequest
	if !handlerutil.BindJSON(c, &req) {
		return
	}

	if len(req.FieldDevices) > 0 {
		h.multiCreateAndAssignProjectFieldDevices(c, projectID, req.FieldDevices)
		return
	}
	if len(req.FieldDeviceIDs) == 0 {
		handlerutil.RespondLocalizedError(c, http.StatusBadRequest, "validation_error", "project.field_device_required")
		return
	}

	if h.bulkAssigner == nil {
		handlerutil.RespondLocalizedError(c, http.StatusInternalServerError, "creation_failed", "project.creation_failed")
		return
	}
	result := h.bulkAssigner.BulkAssignToProject(
		c.Request.Context(),
		appfielddevice.BulkAssignToProjectCommand{
			ProjectID:      projectID,
			FieldDeviceIDs: req.FieldDeviceIDs,
		},
	)

	if h.notifyEvent != nil && len(result.SuccessFieldDeviceIDs) > 0 {
		entityIDs := make([]string, len(result.SuccessFieldDeviceIDs))
		for i, id := range result.SuccessFieldDeviceIDs {
			entityIDs[i] = id.String()
		}
		h.notifyEvent(c, projectID, "project.field_device.multi_created", entityIDs...)
	}

	c.JSON(http.StatusOK, dto.MultiCreateProjectFieldDeviceResponse{
		SuccessFieldDeviceIDs: result.SuccessFieldDeviceIDs,
		AssociationErrors:     result.AssociationErrors,
	})
}

func (h *Handler) multiCreateAndAssignProjectFieldDevices(c *gin.Context, projectID uuid.UUID, reqItems []facilitydto.CreateFieldDeviceRequest) {
	items := toFieldDeviceCreateItems(reqItems)

	if h.multiCreator == nil {
		handlerutil.RespondLocalizedError(c, http.StatusInternalServerError, "creation_failed", "project.creation_failed")
		return
	}
	result, err := h.multiCreator.MultiCreateForProject(
		c.Request.Context(),
		appfielddevice.MultiCreateForProjectCommand{
			ProjectID: projectID,
			Items:     items,
		},
	)
	if err != nil {
		handlerutil.RespondDomainError(c, err,
			handlerutil.LocalizedError(http.StatusInternalServerError, "creation_failed", "project.creation_failed"),
			handlerutil.MapError(domain.ErrNotFound, handlerutil.LocalizedError(http.StatusNotFound, "not_found", "project.project_not_found")),
			handlerutil.MapError(domain.ErrInvalidArgument, handlerutil.LocalizedError(http.StatusBadRequest, "validation_error", "project.creation_failed")),
			handlerutil.MapError(domain.ErrConflict, handlerutil.LocalizedError(http.StatusConflict, "conflict", "project.creation_failed")),
		)
		return
	}

	c.JSON(http.StatusOK, toMultiCreateFieldDeviceResponse(result))
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
func (h *Handler) ListProjectFieldDevices(c *gin.Context) {
	projectID, ok := handlerutil.ParseUUIDParam(c, "id")
	if !ok {
		return
	}

	if !projectshared.EnsureProjectAccessAndPermission(c, h.access, projectID, domainUser.PermissionProjectFieldDeviceRead) {
		return
	}

	var query dto.PaginationQuery
	if !handlerutil.BindQuery(c, &query) {
		return
	}

	result, err := h.facilityLink.ListFieldDevices(c.Request.Context(), projectID, query.Page, query.Limit)
	if err != nil {
		handlerutil.RespondLocalizedError(c, http.StatusInternalServerError, "fetch_failed", "project.fetch_failed")
		return
	}

	response := dto.ProjectFieldDeviceListResponse{
		Items:      toProjectFieldDeviceList(result.Items),
		Total:      result.Total,
		Page:       result.Page,
		TotalPages: result.TotalPages,
	}

	c.JSON(http.StatusOK, response)
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
func (h *Handler) UpdateProjectFieldDevice(c *gin.Context) {
	projectID, ok := handlerutil.ParseUUIDParam(c, "id")
	if !ok {
		return
	}

	if !projectshared.EnsureProjectAccessAndAnyPermission(
		c,
		h.access,
		projectID,
		domainUser.PermissionProjectFieldDeviceUpdate,
	) {
		return
	}

	linkID, ok := handlerutil.ParseUUIDParam(c, "linkId")
	if !ok {
		return
	}

	var req dto.UpdateProjectFieldDeviceRequest
	if !handlerutil.BindJSON(c, &req) {
		return
	}

	if h.reassigner == nil {
		handlerutil.RespondLocalizedError(c, http.StatusInternalServerError, "update_failed", "project.update_failed")
		return
	}
	updated, err := h.reassigner.ReassignProjectLink(
		c.Request.Context(),
		appfielddevice.ReassignProjectLinkCommand{
			ProjectID:     projectID,
			LinkID:        linkID,
			FieldDeviceID: req.FieldDeviceID,
		},
	)
	if err != nil {
		handlerutil.RespondDomainError(c, err,
			handlerutil.LocalizedError(http.StatusInternalServerError, "update_failed", "project.update_failed"),
			handlerutil.MapError(domain.ErrNotFound, handlerutil.LocalizedError(http.StatusNotFound, "not_found", "project.link_not_found")),
		)
		return
	}

	if h.notifyEvent != nil {
		h.notifyEvent(c, projectID, "project.field_device.updated", updated.FieldDeviceID.String())
	}

	c.JSON(http.StatusOK, toProjectFieldDeviceResponse(*updated))
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
func (h *Handler) DeleteProjectFieldDevice(c *gin.Context) {
	projectID, ok := handlerutil.ParseUUIDParam(c, "id")
	if !ok {
		return
	}

	if !projectshared.EnsureProjectAccessAndAnyPermission(
		c,
		h.access,
		projectID,
		domainUser.PermissionProjectFieldDeviceDelete,
	) {
		return
	}

	linkID, ok := handlerutil.ParseUUIDParam(c, "linkId")
	if !ok {
		return
	}

	if err := h.facilityLink.DeleteFieldDevice(c.Request.Context(), linkID, projectID); err != nil {
		handlerutil.RespondDomainError(c, err,
			handlerutil.LocalizedError(http.StatusInternalServerError, "deletion_failed", "project.deletion_failed"),
			handlerutil.MapError(domain.ErrNotFound, handlerutil.LocalizedError(http.StatusNotFound, "not_found", "project.link_not_found")),
		)
		return
	}

	if h.notify != nil {
		h.notify(c, projectID, "project.field_device.deleted")
	}

	c.Status(http.StatusNoContent)
}

// GetFieldDeviceOptionsForProject godoc
// @Summary Get all metadata needed for creating/editing field devices within a project
// @Description Returns all apparats, system parts, object datas and their relationships for a specific project. This returns project-specific object data (object data where project_id = :id and is_active = true).
// @Tags projects
// @Produce json
// @Param id path string true "Project ID"
// @Success 200 {object} FieldDeviceOptionsResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/projects/{id}/field-device-options [get]
func (h *OptionsHandler) GetFieldDeviceOptionsForProject(c *gin.Context) {
	projectID, ok := handlerutil.ParseUUIDParam(c, "id")
	if !ok {
		return
	}

	if !projectshared.EnsureProjectAccessAndPermission(c, h.access, projectID, domainUser.PermissionProjectFieldDeviceRead) {
		return
	}

	options, err := h.service.GetFieldDeviceOptionsForProject(c.Request.Context(), projectID)
	if err != nil {
		handlerutil.RespondLocalizedError(c, http.StatusInternalServerError, "fetch_failed", "facility.fetch_failed")
		return
	}

	c.JSON(http.StatusOK, sharedpresenter.ToFieldDeviceOptionsResponse(options))
}

func toProjectFieldDeviceResponse(item domainProject.ProjectFieldDevice) dto.ProjectFieldDeviceResponse {
	return dto.ProjectFieldDeviceResponse{
		ID:            item.ID,
		ProjectID:     item.ProjectID,
		FieldDeviceID: item.FieldDeviceID,
		CreatedAt:     item.CreatedAt,
		UpdatedAt:     item.UpdatedAt,
	}
}

func toProjectFieldDeviceList(items []domainProject.ProjectFieldDevice) []dto.ProjectFieldDeviceResponse {
	out := make([]dto.ProjectFieldDeviceResponse, len(items))
	for i, item := range items {
		out[i] = toProjectFieldDeviceResponse(item)
	}
	return out
}

func toFieldDeviceCreateItems(reqItems []facilitydto.CreateFieldDeviceRequest) []domainFacility.FieldDeviceCreateItem {
	items := make([]domainFacility.FieldDeviceCreateItem, len(reqItems))
	for i, req := range reqItems {
		apparatNr := 0
		if req.ApparatNr != nil {
			apparatNr = *req.ApparatNr
		}
		items[i] = domainFacility.FieldDeviceCreateItem{
			FieldDevice: &domainFacility.FieldDevice{
				BMK:                       req.BMK,
				Description:               req.Description,
				TextIndividuell:           req.TextIndividuell,
				ApparatNr:                 apparatNr,
				SPSControllerSystemTypeID: req.SPSControllerSystemTypeID,
				SystemPartID:              req.SystemPartID,
				ApparatID:                 req.ApparatID,
			},
			ObjectDataID:  req.ObjectDataID,
			BacnetObjects: toFieldDeviceBacnetObjects(req.BacnetObjects),
		}
	}
	return items
}

func toFieldDeviceBacnetObjects(inputs []facilitydto.BacnetObjectInput) []domainFacility.BacnetObject {
	items := make([]domainFacility.BacnetObject, 0, len(inputs))
	for _, bo := range inputs {
		items = append(items, domainFacility.BacnetObject{
			TextFix:             bo.TextFix,
			Description:         bo.Description,
			GMSVisible:          bo.GMSVisible,
			Optional:            bo.Optional,
			TextIndividual:      normalizeTextIndividual(bo.TextIndividual),
			SoftwareType:        domainFacility.BacnetSoftwareType(bo.SoftwareType),
			SoftwareNumber:      uint16(bo.SoftwareNumber),
			HardwareType:        domainFacility.BacnetHardwareType(bo.HardwareType),
			HardwareQuantity:    uint8(bo.HardwareQuantity),
			SoftwareReferenceID: bo.SoftwareReferenceID,
			StateTextID:         bo.StateTextID,
			NotificationClassID: bo.NotificationClassID,
			AlarmDefinitionID:   bo.AlarmDefinitionID,
			AlarmTypeID:         bo.AlarmTypeID,
		})
	}
	return items
}

func normalizeTextIndividual(value *string) *string {
	if value != nil && *value == "" {
		return nil
	}
	return value
}

func toMultiCreateFieldDeviceResponse(result *domainFacility.FieldDeviceMultiCreateResult) facilitydto.MultiCreateFieldDeviceResponse {
	if result == nil {
		return facilitydto.MultiCreateFieldDeviceResponse{}
	}

	results := make([]facilitydto.FieldDeviceCreateResultResponse, len(result.Results))
	for i, item := range result.Results {
		results[i] = facilitydto.FieldDeviceCreateResultResponse{
			Index:       item.Index,
			Success:     item.Success,
			FieldDevice: toFieldDeviceResponse(item.FieldDevice),
			Error:       item.Error,
			ErrorField:  item.ErrorField,
		}
	}

	return facilitydto.MultiCreateFieldDeviceResponse{
		Results:       results,
		TotalRequests: result.TotalRequests,
		SuccessCount:  result.SuccessCount,
		FailureCount:  result.FailureCount,
	}
}

func toFieldDeviceResponse(fieldDevice *domainFacility.FieldDevice) *facilitydto.FieldDeviceResponse {
	if fieldDevice == nil {
		return nil
	}

	apparatNr := fieldDevice.ApparatNr
	systemPartID := fieldDevice.SystemPartID
	return &facilitydto.FieldDeviceResponse{
		ID:                        fieldDevice.ID,
		BMK:                       fieldDevice.BMK,
		Description:               fieldDevice.Description,
		TextIndividuell:           fieldDevice.TextIndividuell,
		ApparatNr:                 &apparatNr,
		SPSControllerSystemTypeID: fieldDevice.SPSControllerSystemTypeID,
		SystemPartID:              &systemPartID,
		SpecificationID:           fieldDevice.SpecificationID,
		ApparatID:                 fieldDevice.ApparatID,
		CreatedAt:                 fieldDevice.CreatedAt,
		UpdatedAt:                 fieldDevice.UpdatedAt,
	}
}
