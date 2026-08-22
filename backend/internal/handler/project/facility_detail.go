package project

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	domainUser "github.com/besart951/go_infra_link/backend/internal/domain/user"
	facilitydto "github.com/besart951/go_infra_link/backend/internal/handler/dto/facility"
	"github.com/besart951/go_infra_link/backend/internal/handler/middleware"
	projectshared "github.com/besart951/go_infra_link/backend/internal/handler/project/shared"
	"github.com/besart951/go_infra_link/backend/internal/handlerutil"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// FacilityDetailHandler exposes hierarchy details inside a project without
// requiring global read permissions. Every result is additionally constrained
// to the project's linked facility objects; update actions require both the
// project capability and their matching global permission.
type FacilityDetailHandler struct {
	access   ProjectAccessPolicyService
	links    ProjectFacilityLinkService
	services FacilityDetailServices
	auth     middleware.AuthorizationChecker
	notify   projectshared.ProjectMutationNotifier
}

func NewFacilityDetailHandler(access ProjectAccessPolicyService, links ProjectFacilityLinkService, services FacilityDetailServices, auth middleware.AuthorizationChecker, notify projectshared.ProjectMutationNotifier) *FacilityDetailHandler {
	return &FacilityDetailHandler{access: access, links: links, services: services, auth: auth, notify: notify}
}

func (h *FacilityDetailHandler) projectID(c *gin.Context) (uuid.UUID, bool) {
	return handlerutil.ParseUUIDParam(c, "id")
}

func (h *FacilityDetailHandler) entityID(c *gin.Context, name string) (uuid.UUID, bool) {
	return handlerutil.ParseUUIDParam(c, name)
}

func (h *FacilityDetailHandler) canProject(c *gin.Context, projectID uuid.UUID, permission string) bool {
	if h.access == nil {
		return false
	}
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return false
	}
	var role *domainUser.Role
	if value, ok := middleware.GetUserRole(c); ok {
		role = &value
	}
	allowed, err := h.access.CanUseProjectPermissionForProject(c.Request.Context(), userID, projectID, role, permission)
	return err == nil && allowed
}

func (h *FacilityDetailHandler) canGlobal(c *gin.Context, permission string) bool {
	if h.auth == nil {
		return false
	}
	role, ok := middleware.GetUserRole(c)
	if !ok {
		return false
	}
	allowed, err := h.auth.HasPermission(c.Request.Context(), role, permission)
	return err == nil && allowed
}

func projectDetailPage(c *gin.Context) (int, int, bool) {
	page, limit := 1, 12
	if raw := strings.TrimSpace(c.Query("page")); raw != "" {
		value, err := strconv.Atoi(raw)
		if err != nil || value < 1 {
			handlerutil.RespondLocalizedError(c, http.StatusBadRequest, "validation_error", "errors.validation_error")
			return 0, 0, false
		}
		page = value
	}
	if raw := strings.TrimSpace(c.Query("limit")); raw != "" {
		value, err := strconv.Atoi(raw)
		if err != nil || value < 1 || value > 50 {
			handlerutil.RespondLocalizedError(c, http.StatusBadRequest, "validation_error", "errors.validation_error")
			return 0, 0, false
		}
		limit = value
	}
	return page, limit, true
}

func (h *FacilityDetailHandler) projectCabinets(ctx *gin.Context, projectID uuid.UUID) ([]domainFacility.ControlCabinet, error) {
	links, err := h.links.ListControlCabinets(ctx.Request.Context(), projectID, 1, 1000)
	if err != nil || len(links.Items) == 0 {
		return nil, err
	}
	ids := make([]uuid.UUID, 0, len(links.Items))
	for _, item := range links.Items {
		ids = append(ids, item.ControlCabinetID)
	}
	return h.services.ControlCabinet.GetByIDs(ctx.Request.Context(), ids)
}

func (h *FacilityDetailHandler) projectControllers(ctx *gin.Context, projectID uuid.UUID) ([]domainFacility.SPSController, error) {
	links, err := h.links.ListSPSControllers(ctx.Request.Context(), projectID, 1, 1000)
	if err != nil || len(links.Items) == 0 {
		return nil, err
	}
	ids := make([]uuid.UUID, 0, len(links.Items))
	for _, item := range links.Items {
		ids = append(ids, item.SPSControllerID)
	}
	return h.services.SPSController.GetByIDs(ctx.Request.Context(), ids)
}

func (h *FacilityDetailHandler) projectFieldDeviceIDs(ctx *gin.Context, projectID uuid.UUID) (map[uuid.UUID]struct{}, error) {
	links, err := h.links.ListFieldDevices(ctx.Request.Context(), projectID, 1, 1000)
	if err != nil {
		return nil, err
	}
	ids := make(map[uuid.UUID]struct{}, len(links.Items))
	for _, item := range links.Items {
		ids[item.FieldDeviceID] = struct{}{}
	}
	return ids, nil
}

func containsCabinet(items []domainFacility.ControlCabinet, id uuid.UUID) (domainFacility.ControlCabinet, bool) {
	for _, item := range items {
		if item.ID == id {
			return item, true
		}
	}
	return domainFacility.ControlCabinet{}, false
}

func containsController(items []domainFacility.SPSController, id uuid.UUID) (domainFacility.SPSController, bool) {
	for _, item := range items {
		if item.ID == id {
			return item, true
		}
	}
	return domainFacility.SPSController{}, false
}

func projectDetailForbidden(c *gin.Context) {
	handlerutil.RespondLocalizedError(c, http.StatusForbidden, "forbidden", "errors.forbidden")
}

func projectDetailFetchFailed(c *gin.Context) {
	handlerutil.RespondLocalizedError(c, http.StatusInternalServerError, "fetch_failed", "project.fetch_failed")
}

// GetBuilding godoc
// @Summary Get a project-scoped, read-only building detail
// @Tags project-facility-details
// @Produce json
// @Param id path string true "Project ID"
// @Param buildingId path string true "Building ID"
// @Param page query int false "Relationship page" default(1)
// @Param limit query int false "Relationship page size" default(12)
// @Success 200 {object} facilitydto.BuildingDetailResponse
// @Failure 403 {object} facilitydto.ErrorResponse
// @Failure 404 {object} facilitydto.ErrorResponse
// @Router /api/v1/projects/{id}/facility/buildings/{buildingId} [get]
func (h *FacilityDetailHandler) GetBuilding(c *gin.Context) {
	projectID, ok := h.projectID(c)
	if !ok || !projectshared.EnsureProjectAccessAndPermission(c, h.access, projectID, domainUser.PermissionProjectControlCabinetRead) {
		return
	}
	buildingID, ok := h.entityID(c, "buildingId")
	if !ok {
		return
	}
	page, limit, ok := projectDetailPage(c)
	if !ok {
		return
	}
	cabinets, err := h.projectCabinets(c, projectID)
	if err != nil {
		projectDetailFetchFailed(c)
		return
	}
	children := make([]facilitydto.DetailRelationItem, 0)
	for _, cabinet := range cabinets {
		if cabinet.BuildingID == buildingID {
			children = append(children, cabinetRelationItem(cabinet))
		}
	}
	if len(children) == 0 {
		projectDetailForbidden(c)
		return
	}
	building, err := h.services.Building.GetByID(c.Request.Context(), buildingID)
	if err != nil {
		handlerutil.RespondLocalizedError(c, http.StatusNotFound, "not_found", "facility.building_not_found")
		return
	}
	c.JSON(http.StatusOK, facilitydto.BuildingDetailResponse{
		Building:     projectBuildingResponse(*building),
		Relations:    []facilitydto.DetailRelation{projectRelation("control_cabinets", "Schaltschränke", "control-cabinets", children, page, limit)},
		Capabilities: facilitydto.DetailCapabilities{CanUpdate: false},
	})
}

// GetControlCabinet godoc
// @Summary Get a project-scoped control cabinet detail
// @Tags project-facility-details
// @Produce json
// @Param id path string true "Project ID"
// @Param controlCabinetId path string true "Control cabinet ID"
// @Param page query int false "Relationship page" default(1)
// @Param limit query int false "Relationship page size" default(12)
// @Success 200 {object} facilitydto.ControlCabinetDetailResponse
// @Failure 403 {object} facilitydto.ErrorResponse
// @Router /api/v1/projects/{id}/facility/control-cabinets/{controlCabinetId} [get]
func (h *FacilityDetailHandler) GetControlCabinet(c *gin.Context) {
	projectID, ok := h.projectID(c)
	if !ok || !projectshared.EnsureProjectAccessAndPermission(c, h.access, projectID, domainUser.PermissionProjectControlCabinetRead) {
		return
	}
	cabinetID, ok := h.entityID(c, "controlCabinetId")
	if !ok {
		return
	}
	page, limit, ok := projectDetailPage(c)
	if !ok {
		return
	}
	cabinets, err := h.projectCabinets(c, projectID)
	if err != nil {
		projectDetailFetchFailed(c)
		return
	}
	cabinet, exists := containsCabinet(cabinets, cabinetID)
	if !exists {
		projectDetailForbidden(c)
		return
	}
	relations := []facilitydto.DetailRelation{}
	if building, getErr := h.services.Building.GetByID(c.Request.Context(), cabinet.BuildingID); getErr == nil {
		relations = append(relations, projectSingleton("building", "Gebäude", "buildings", buildingRelationItem(*building)))
	}
	if h.canProject(c, projectID, domainUser.PermissionProjectSPSControllerRead) {
		controllers, controllerErr := h.projectControllers(c, projectID)
		if controllerErr != nil {
			projectDetailFetchFailed(c)
			return
		}
		items := make([]facilitydto.DetailRelationItem, 0)
		for _, controller := range controllers {
			if controller.ControlCabinetID == cabinetID {
				items = append(items, controllerRelationItem(controller))
			}
		}
		relations = append(relations, projectRelation("sps_controllers", "SPS-Regler", "sps-controllers", items, page, limit))
	}
	c.JSON(http.StatusOK, facilitydto.ControlCabinetDetailResponse{
		ControlCabinet: projectControlCabinetResponse(cabinet), Relations: relations,
		Capabilities: facilitydto.DetailCapabilities{CanUpdate: h.canProject(c, projectID, domainUser.PermissionProjectControlCabinetUpdate) && h.canGlobal(c, domainUser.PermissionControlCabinetUpdate)},
	})
}

// GetSPSController godoc
// @Summary Get a project-scoped SPS controller detail
// @Tags project-facility-details
// @Produce json
// @Param id path string true "Project ID"
// @Param spsControllerId path string true "SPS controller ID"
// @Param page query int false "Relationship page" default(1)
// @Param limit query int false "Relationship page size" default(12)
// @Success 200 {object} facilitydto.SPSControllerDetailResponse
// @Failure 403 {object} facilitydto.ErrorResponse
// @Router /api/v1/projects/{id}/facility/sps-controllers/{spsControllerId} [get]
func (h *FacilityDetailHandler) GetSPSController(c *gin.Context) {
	projectID, ok := h.projectID(c)
	if !ok || !projectshared.EnsureProjectAccessAndPermission(c, h.access, projectID, domainUser.PermissionProjectSPSControllerRead) {
		return
	}
	controllerID, ok := h.entityID(c, "spsControllerId")
	if !ok {
		return
	}
	page, limit, ok := projectDetailPage(c)
	if !ok {
		return
	}
	controllers, err := h.projectControllers(c, projectID)
	if err != nil {
		projectDetailFetchFailed(c)
		return
	}
	controller, exists := containsController(controllers, controllerID)
	if !exists {
		projectDetailForbidden(c)
		return
	}
	relations := h.projectControllerParents(c, projectID, controller.ControlCabinetID)
	if h.canProject(c, projectID, domainUser.PermissionProjectSPSControllerSystemTypeRead) {
		types, listErr := h.services.SPSControllerSystemType.ListBySPSControllerID(c.Request.Context(), controllerID, page, limit, "")
		if listErr != nil {
			projectDetailFetchFailed(c)
			return
		}
		items := make([]facilitydto.DetailRelationItem, 0, len(types.Items))
		for _, item := range types.Items {
			items = append(items, systemTypeRelationItem(item))
		}
		relations = append(relations, facilitydto.DetailRelation{Key: "sps_controller_system_types", Label: "SPS-Systemtypen", Resource: "sps-controller-system-types", Count: types.Total, Items: items, Page: types.Page, TotalPages: types.TotalPages})
	}
	c.JSON(http.StatusOK, facilitydto.SPSControllerDetailResponse{
		SPSController: projectSPSControllerResponse(controller), Relations: relations,
		Capabilities: facilitydto.DetailCapabilities{CanUpdate: h.canProject(c, projectID, domainUser.PermissionProjectSPSControllerUpdate) && h.canGlobal(c, domainUser.PermissionSPSControllerUpdate)},
	})
}

// GetSPSControllerSystemType godoc
// @Summary Get a project-scoped SPS controller system type detail
// @Tags project-facility-details
// @Produce json
// @Param id path string true "Project ID"
// @Param spsControllerSystemTypeId path string true "SPS controller system type ID"
// @Param page query int false "Relationship page" default(1)
// @Param limit query int false "Relationship page size" default(12)
// @Success 200 {object} facilitydto.SPSControllerSystemTypeDetailResponse
// @Failure 403 {object} facilitydto.ErrorResponse
// @Router /api/v1/projects/{id}/facility/sps-controller-system-types/{spsControllerSystemTypeId} [get]
func (h *FacilityDetailHandler) GetSPSControllerSystemType(c *gin.Context) {
	projectID, ok := h.projectID(c)
	if !ok || !projectshared.EnsureProjectAccessAndPermission(c, h.access, projectID, domainUser.PermissionProjectSPSControllerSystemTypeRead) {
		return
	}
	typeID, ok := h.entityID(c, "spsControllerSystemTypeId")
	if !ok {
		return
	}
	page, limit, ok := projectDetailPage(c)
	if !ok {
		return
	}
	item, err := h.services.SPSControllerSystemType.GetByID(c.Request.Context(), typeID)
	if err != nil {
		handlerutil.RespondLocalizedError(c, http.StatusNotFound, "not_found", "facility.sps_controller_system_type_not_found")
		return
	}
	controllers, err := h.projectControllers(c, projectID)
	if err != nil {
		projectDetailFetchFailed(c)
		return
	}
	controller, exists := containsController(controllers, item.SPSControllerID)
	if !exists {
		projectDetailForbidden(c)
		return
	}
	relations := []facilitydto.DetailRelation{projectSingleton("sps_controller", "SPS-Regler", "sps-controllers", controllerRelationItem(controller))}
	relations = append(relations, h.projectControllerParents(c, projectID, controller.ControlCabinetID)...)
	if h.canProject(c, projectID, domainUser.PermissionProjectFieldDeviceRead) {
		devices, listErr := h.services.FieldDevice.ListWithFilters(c.Request.Context(), domain.PaginationParams{Page: page, Limit: limit}, domainFacility.FieldDeviceFilterParams{ProjectID: &projectID, SPSControllerSystemTypeID: &typeID})
		if listErr != nil {
			projectDetailFetchFailed(c)
			return
		}
		items := make([]facilitydto.DetailRelationItem, 0, len(devices.Items))
		for _, device := range devices.Items {
			items = append(items, fieldDeviceRelationItem(device))
		}
		relations = append(relations, facilitydto.DetailRelation{Key: "field_devices", Label: "Feldgeräte", Resource: "field-devices", Count: devices.Total, Items: items, Page: devices.Page, TotalPages: devices.TotalPages})
	}
	c.JSON(http.StatusOK, facilitydto.SPSControllerSystemTypeDetailResponse{
		SPSControllerSystemType: projectSPSControllerSystemTypeResponse(*item), Relations: relations,
		Capabilities: facilitydto.DetailCapabilities{CanUpdate: h.canProject(c, projectID, domainUser.PermissionProjectSPSControllerSystemTypeUpdate) && h.canGlobal(c, domainUser.PermissionSPSControllerSystemTypeUpdate)},
	})
}

// GetFieldDevice godoc
// @Summary Get a project-scoped field device detail
// @Tags project-facility-details
// @Produce json
// @Param id path string true "Project ID"
// @Param fieldDeviceId path string true "Field device ID"
// @Success 200 {object} facilitydto.FieldDeviceDetailResponse
// @Failure 403 {object} facilitydto.ErrorResponse
// @Router /api/v1/projects/{id}/facility/field-devices/{fieldDeviceId} [get]
func (h *FacilityDetailHandler) GetFieldDevice(c *gin.Context) {
	projectID, ok := h.projectID(c)
	if !ok || !projectshared.EnsureProjectAccessAndPermission(c, h.access, projectID, domainUser.PermissionProjectFieldDeviceRead) {
		return
	}
	deviceID, ok := h.entityID(c, "fieldDeviceId")
	if !ok {
		return
	}
	deviceIDs, err := h.projectFieldDeviceIDs(c, projectID)
	if err != nil {
		projectDetailFetchFailed(c)
		return
	}
	if _, exists := deviceIDs[deviceID]; !exists {
		projectDetailForbidden(c)
		return
	}
	device, err := h.services.FieldDevice.GetByID(c.Request.Context(), deviceID)
	if err != nil {
		handlerutil.RespondLocalizedError(c, http.StatusNotFound, "not_found", "facility.field_device_not_found")
		return
	}
	relations := h.projectFieldDeviceRelations(c, projectID, *device)
	c.JSON(http.StatusOK, facilitydto.FieldDeviceDetailResponse{
		FieldDevice: projectFieldDeviceResponse(*device), Relations: relations,
		Capabilities: facilitydto.DetailCapabilities{CanUpdate: h.canProject(c, projectID, domainUser.PermissionProjectFieldDeviceUpdate) && h.canGlobal(c, domainUser.PermissionFieldDeviceUpdate)},
	})
}

func (h *FacilityDetailHandler) projectControllerParents(c *gin.Context, projectID, cabinetID uuid.UUID) []facilitydto.DetailRelation {
	if !h.canProject(c, projectID, domainUser.PermissionProjectControlCabinetRead) {
		return []facilitydto.DetailRelation{}
	}
	cabinets, err := h.projectCabinets(c, projectID)
	if err != nil {
		return []facilitydto.DetailRelation{}
	}
	cabinet, exists := containsCabinet(cabinets, cabinetID)
	if !exists {
		return []facilitydto.DetailRelation{}
	}
	relations := []facilitydto.DetailRelation{projectSingleton("control_cabinet", "Schaltschrank", "control-cabinets", cabinetRelationItem(cabinet))}
	if building, getErr := h.services.Building.GetByID(c.Request.Context(), cabinet.BuildingID); getErr == nil {
		relations = append(relations, projectSingleton("building", "Gebäude", "buildings", buildingRelationItem(*building)))
	}
	return relations
}

func (h *FacilityDetailHandler) projectFieldDeviceRelations(c *gin.Context, projectID uuid.UUID, device domainFacility.FieldDevice) []facilitydto.DetailRelation {
	relations := make([]facilitydto.DetailRelation, 0, 6)
	if h.canProject(c, projectID, domainUser.PermissionProjectSPSControllerSystemTypeRead) {
		if systemType, getErr := h.services.SPSControllerSystemType.GetByID(c.Request.Context(), device.SPSControllerSystemTypeID); getErr == nil {
			relations = append(relations, projectSingleton("sps_controller_system_type", "SPS-Systemtyp", "sps-controller-system-types", systemTypeRelationItem(*systemType)))
			if h.canProject(c, projectID, domainUser.PermissionProjectSPSControllerRead) {
				controllers, listErr := h.projectControllers(c, projectID)
				if listErr == nil {
					if controller, exists := containsController(controllers, systemType.SPSControllerID); exists {
						relations = append(relations, projectSingleton("sps_controller", "SPS-Regler", "sps-controllers", controllerRelationItem(controller)))
						relations = append(relations, h.projectControllerParents(c, projectID, controller.ControlCabinetID)...)
					}
				}
			}
		}
	}
	if h.canGlobal(c, domainUser.PermissionApparatRead) {
		if apparat, getErr := h.services.Apparat.GetByID(c.Request.Context(), device.ApparatID); getErr == nil {
			relations = append(relations, projectSingleton("apparat", "Apparat", "references", apparatRelationItem(*apparat)))
		}
	}
	if device.SystemPartID != uuid.Nil && h.canGlobal(c, domainUser.PermissionSystemPartRead) {
		if part, getErr := h.services.SystemPart.GetByID(c.Request.Context(), device.SystemPartID); getErr == nil {
			relations = append(relations, projectSingleton("system_part", "Systemteil", "references", systemPartRelationItem(*part)))
		}
	}
	if device.Specification != nil && h.canProject(c, projectID, domainUser.PermissionProjectFieldDeviceSpecificationRead) {
		relations = append(relations, projectSingleton("specification", "Spezifikation", "references", specificationRelationItem(*device.Specification)))
	}
	if h.canProject(c, projectID, domainUser.PermissionProjectFieldDeviceBacnetObjectsRead) {
		if objects, listErr := h.services.FieldDevice.ListBacnetObjects(c.Request.Context(), device.ID); listErr == nil && len(objects) > 0 {
			items := make([]facilitydto.DetailRelationItem, 0, len(objects))
			for _, object := range objects {
				items = append(items, facilitydto.DetailRelationItem{ID: object.ID.String(), Label: nonEmptyDetail(object.TextFix, nonEmptyDetail(stringPtrDetail(object.Description), "BACnet-Objekt")), Subtitle: string(object.SoftwareType)})
			}
			relations = append(relations, facilitydto.DetailRelation{Key: "bacnet_objects", Label: "BACnet-Objekte", Resource: "references", Count: int64(len(items)), Items: items, Page: 1, TotalPages: 1})
		}
	}
	return relations
}

// UpdateControlCabinet godoc
// @Summary Patch a project control cabinet when both permission layers allow it
// @Tags project-facility-details
// @Accept json
// @Produce json
// @Param id path string true "Project ID"
// @Param controlCabinetId path string true "Control cabinet ID"
// @Param control_cabinet body facilitydto.UpdateControlCabinetRequest true "Control cabinet patch"
// @Success 200 {object} facilitydto.ControlCabinetResponse
// @Failure 403 {object} facilitydto.ErrorResponse
// @Failure 409 {object} facilitydto.ErrorResponse
// @Router /api/v1/projects/{id}/facility/control-cabinets/{controlCabinetId} [patch]
func (h *FacilityDetailHandler) UpdateControlCabinet(c *gin.Context) {
	projectID, cabinetID, cabinet, ok := h.editableCabinet(c)
	if !ok {
		return
	}
	var req facilitydto.UpdateControlCabinetRequest
	if !handlerutil.BindJSON(c, &req) {
		return
	}
	baseVersion := req.BaseVersion
	cabinet.Version = req.BaseVersion
	if req.ControlCabinetNr != nil {
		cabinet.ControlCabinetNr = req.ControlCabinetNr
	}
	if req.BuildingID != uuid.Nil {
		cabinets, err := h.projectCabinets(c, projectID)
		if err != nil || !projectContainsBuilding(cabinets, req.BuildingID) {
			handlerutil.RespondLocalizedError(c, http.StatusBadRequest, "invalid_reference", "facility.invalid_reference")
			return
		}
		cabinet.BuildingID = req.BuildingID
	}
	if err := h.services.ControlCabinet.Update(c.Request.Context(), cabinet); err != nil {
		h.projectWriteError(c, err, "control_cabinet", cabinetID, baseVersion, req.ChangedFields(), func() (uint64, any, bool) {
			current, getErr := h.services.ControlCabinet.GetByID(c.Request.Context(), cabinetID)
			if getErr != nil {
				return 0, nil, false
			}
			return current.Version, projectControlCabinetResponse(*current), true
		})
		return
	}
	h.notifyUpdate(c, projectID, "project.control_cabinet.updated", req.ChangedFields(), cabinetID)
	c.JSON(http.StatusOK, projectControlCabinetResponse(*cabinet))
}

// UpdateSPSController godoc
// @Summary Patch a project SPS controller when both permission layers allow it
// @Tags project-facility-details
// @Accept json
// @Produce json
// @Param id path string true "Project ID"
// @Param spsControllerId path string true "SPS controller ID"
// @Param sps_controller body facilitydto.UpdateSPSControllerRequest true "SPS controller patch"
// @Success 200 {object} facilitydto.SPSControllerResponse
// @Failure 403 {object} facilitydto.ErrorResponse
// @Failure 409 {object} facilitydto.ErrorResponse
// @Router /api/v1/projects/{id}/facility/sps-controllers/{spsControllerId} [patch]
func (h *FacilityDetailHandler) UpdateSPSController(c *gin.Context) {
	projectID, controllerID, controller, ok := h.editableController(c)
	if !ok {
		return
	}
	var req facilitydto.UpdateSPSControllerRequest
	if !handlerutil.BindJSON(c, &req) {
		return
	}
	baseVersion := req.BaseVersion
	controller.Version = req.BaseVersion
	if req.ControlCabinetID != uuid.Nil {
		cabinets, err := h.projectCabinets(c, projectID)
		if err != nil || !projectContainsCabinet(cabinets, req.ControlCabinetID) {
			handlerutil.RespondLocalizedError(c, http.StatusBadRequest, "invalid_reference", "facility.invalid_reference")
			return
		}
		controller.ControlCabinetID = req.ControlCabinetID
	}
	if req.GADevice != nil {
		controller.GADevice = req.GADevice
	}
	if req.DeviceName != "" {
		controller.DeviceName = req.DeviceName
	}
	if req.DeviceDescription != nil {
		controller.DeviceDescription = req.DeviceDescription
	}
	if req.DeviceLocation != nil {
		controller.DeviceLocation = req.DeviceLocation
	}
	if req.IPAddress != nil {
		controller.IPAddress = req.IPAddress
	}
	if req.Subnet != nil {
		controller.Subnet = req.Subnet
	}
	if req.Gateway != nil {
		controller.Gateway = req.Gateway
	}
	if req.Vlan != nil {
		controller.Vlan = req.Vlan
	}
	if err := h.services.SPSController.Update(c.Request.Context(), controller); err != nil {
		h.projectWriteError(c, err, "sps_controller", controllerID, baseVersion, req.ChangedFields(), func() (uint64, any, bool) {
			current, getErr := h.services.SPSController.GetByID(c.Request.Context(), controllerID)
			if getErr != nil {
				return 0, nil, false
			}
			return current.Version, projectSPSControllerResponse(*current), true
		})
		return
	}
	h.notifyUpdate(c, projectID, "project.sps_controller.updated", req.ChangedFields(), controllerID)
	c.JSON(http.StatusOK, projectSPSControllerResponse(*controller))
}

// UpdateSPSControllerSystemType godoc
// @Summary Patch a project SPS controller system type when both permission layers allow it
// @Tags project-facility-details
// @Accept json
// @Produce json
// @Param id path string true "Project ID"
// @Param spsControllerSystemTypeId path string true "SPS controller system type ID"
// @Param system_type body facilitydto.UpdateSPSControllerSystemTypeRequest true "System type patch"
// @Success 200 {object} facilitydto.SPSControllerSystemTypeResponse
// @Failure 403 {object} facilitydto.ErrorResponse
// @Failure 409 {object} facilitydto.ErrorResponse
// @Router /api/v1/projects/{id}/facility/sps-controller-system-types/{spsControllerSystemTypeId} [patch]
func (h *FacilityDetailHandler) UpdateSPSControllerSystemType(c *gin.Context) {
	projectID, typeID, item, ok := h.editableSystemType(c)
	if !ok {
		return
	}
	var req facilitydto.UpdateSPSControllerSystemTypeRequest
	if !handlerutil.BindJSON(c, &req) {
		return
	}
	baseVersion := req.BaseVersion
	item.Version = req.BaseVersion
	if req.Number != nil {
		item.Number = req.Number
	}
	if req.DocumentName != nil {
		item.DocumentName = req.DocumentName
	}
	if err := h.services.SPSControllerSystemType.Update(c.Request.Context(), item); err != nil {
		h.projectWriteError(c, err, "sps_controller_system_type", typeID, baseVersion, req.ChangedFields(), func() (uint64, any, bool) {
			current, getErr := h.services.SPSControllerSystemType.GetByID(c.Request.Context(), typeID)
			if getErr != nil {
				return 0, nil, false
			}
			return current.Version, projectSPSControllerSystemTypeResponse(*current), true
		})
		return
	}
	h.notifyUpdate(c, projectID, "project.sps_controller_system_type.updated", req.ChangedFields(), typeID)
	c.JSON(http.StatusOK, projectSPSControllerSystemTypeResponse(*item))
}

// UpdateFieldDevice godoc
// @Summary Patch a project field device when both permission layers allow it
// @Tags project-facility-details
// @Accept json
// @Produce json
// @Param id path string true "Project ID"
// @Param fieldDeviceId path string true "Field device ID"
// @Param field_device body facilitydto.UpdateFieldDeviceRequest true "Field device patch"
// @Success 200 {object} facilitydto.FieldDeviceResponse
// @Failure 403 {object} facilitydto.ErrorResponse
// @Failure 409 {object} facilitydto.ErrorResponse
// @Router /api/v1/projects/{id}/facility/field-devices/{fieldDeviceId} [patch]
func (h *FacilityDetailHandler) UpdateFieldDevice(c *gin.Context) {
	projectID, deviceID, device, ok := h.editableFieldDevice(c)
	if !ok {
		return
	}
	var req facilitydto.UpdateFieldDeviceRequest
	if !handlerutil.BindJSON(c, &req) {
		return
	}
	if req.BacnetObjects != nil {
		handlerutil.RespondLocalizedError(c, http.StatusBadRequest, "validation_error", "errors.validation_error")
		return
	}
	baseVersion := req.BaseVersion
	device.Version = req.BaseVersion
	if req.BMK != nil {
		device.BMK = req.BMK
	}
	if req.Description != nil {
		device.Description = req.Description
	}
	if req.TextIndividuell != nil {
		device.TextIndividuell = req.TextIndividuell
	}
	if req.ApparatNr != nil {
		device.ApparatNr = *req.ApparatNr
	}
	if req.ApparatID != uuid.Nil {
		device.ApparatID = req.ApparatID
	}
	if req.SystemPartID != uuid.Nil {
		device.SystemPartID = req.SystemPartID
	}
	if req.SPSControllerSystemTypeID != uuid.Nil {
		typeItem, err := h.services.SPSControllerSystemType.GetByID(c.Request.Context(), req.SPSControllerSystemTypeID)
		controllers, controllerErr := h.projectControllers(c, projectID)
		if err != nil || controllerErr != nil {
			projectDetailFetchFailed(c)
			return
		}
		if _, exists := containsController(controllers, typeItem.SPSControllerID); !exists {
			handlerutil.RespondLocalizedError(c, http.StatusBadRequest, "invalid_reference", "facility.invalid_reference")
			return
		}
		device.SPSControllerSystemTypeID = req.SPSControllerSystemTypeID
	}
	if err := h.services.FieldDevice.Update(c.Request.Context(), device); err != nil {
		h.projectWriteError(c, err, "field_device", deviceID, baseVersion, req.ChangedFields(), func() (uint64, any, bool) {
			current, getErr := h.services.FieldDevice.GetByID(c.Request.Context(), deviceID)
			if getErr != nil {
				return 0, nil, false
			}
			return current.Version, projectFieldDeviceResponse(*current), true
		})
		return
	}
	h.notifyUpdate(c, projectID, "project.field_device.updated", req.ChangedFields(), deviceID)
	c.JSON(http.StatusOK, projectFieldDeviceResponse(*device))
}

func (h *FacilityDetailHandler) editableCabinet(c *gin.Context) (uuid.UUID, uuid.UUID, *domainFacility.ControlCabinet, bool) {
	projectID, ok := h.projectID(c)
	if !ok {
		return uuid.Nil, uuid.Nil, nil, false
	}
	if !projectshared.EnsureProjectAccessAndPermission(c, h.access, projectID, domainUser.PermissionProjectControlCabinetUpdate) {
		return uuid.Nil, uuid.Nil, nil, false
	}
	if !h.canGlobal(c, domainUser.PermissionControlCabinetUpdate) {
		projectDetailForbidden(c)
		return uuid.Nil, uuid.Nil, nil, false
	}
	id, ok := h.entityID(c, "controlCabinetId")
	if !ok {
		return uuid.Nil, uuid.Nil, nil, false
	}
	cabinets, err := h.projectCabinets(c, projectID)
	if err != nil {
		projectDetailFetchFailed(c)
		return uuid.Nil, uuid.Nil, nil, false
	}
	item, exists := containsCabinet(cabinets, id)
	if !exists {
		projectDetailForbidden(c)
		return uuid.Nil, uuid.Nil, nil, false
	}
	return projectID, id, &item, true
}

func (h *FacilityDetailHandler) editableController(c *gin.Context) (uuid.UUID, uuid.UUID, *domainFacility.SPSController, bool) {
	projectID, ok := h.projectID(c)
	if !ok {
		return uuid.Nil, uuid.Nil, nil, false
	}
	if !projectshared.EnsureProjectAccessAndPermission(c, h.access, projectID, domainUser.PermissionProjectSPSControllerUpdate) {
		return uuid.Nil, uuid.Nil, nil, false
	}
	if !h.canGlobal(c, domainUser.PermissionSPSControllerUpdate) {
		projectDetailForbidden(c)
		return uuid.Nil, uuid.Nil, nil, false
	}
	id, ok := h.entityID(c, "spsControllerId")
	if !ok {
		return uuid.Nil, uuid.Nil, nil, false
	}
	controllers, err := h.projectControllers(c, projectID)
	if err != nil {
		projectDetailFetchFailed(c)
		return uuid.Nil, uuid.Nil, nil, false
	}
	item, exists := containsController(controllers, id)
	if !exists {
		projectDetailForbidden(c)
		return uuid.Nil, uuid.Nil, nil, false
	}
	return projectID, id, &item, true
}

func (h *FacilityDetailHandler) editableSystemType(c *gin.Context) (uuid.UUID, uuid.UUID, *domainFacility.SPSControllerSystemType, bool) {
	projectID, ok := h.projectID(c)
	if !ok {
		return uuid.Nil, uuid.Nil, nil, false
	}
	if !projectshared.EnsureProjectAccessAndPermission(c, h.access, projectID, domainUser.PermissionProjectSPSControllerSystemTypeUpdate) {
		return uuid.Nil, uuid.Nil, nil, false
	}
	if !h.canGlobal(c, domainUser.PermissionSPSControllerSystemTypeUpdate) {
		projectDetailForbidden(c)
		return uuid.Nil, uuid.Nil, nil, false
	}
	id, ok := h.entityID(c, "spsControllerSystemTypeId")
	if !ok {
		return uuid.Nil, uuid.Nil, nil, false
	}
	item, err := h.services.SPSControllerSystemType.GetByID(c.Request.Context(), id)
	if err != nil {
		handlerutil.RespondLocalizedError(c, http.StatusNotFound, "not_found", "facility.sps_controller_system_type_not_found")
		return uuid.Nil, uuid.Nil, nil, false
	}
	controllers, err := h.projectControllers(c, projectID)
	if err != nil {
		projectDetailFetchFailed(c)
		return uuid.Nil, uuid.Nil, nil, false
	}
	if _, exists := containsController(controllers, item.SPSControllerID); !exists {
		projectDetailForbidden(c)
		return uuid.Nil, uuid.Nil, nil, false
	}
	return projectID, id, item, true
}

func (h *FacilityDetailHandler) editableFieldDevice(c *gin.Context) (uuid.UUID, uuid.UUID, *domainFacility.FieldDevice, bool) {
	projectID, ok := h.projectID(c)
	if !ok {
		return uuid.Nil, uuid.Nil, nil, false
	}
	if !projectshared.EnsureProjectAccessAndPermission(c, h.access, projectID, domainUser.PermissionProjectFieldDeviceUpdate) {
		return uuid.Nil, uuid.Nil, nil, false
	}
	if !h.canGlobal(c, domainUser.PermissionFieldDeviceUpdate) {
		projectDetailForbidden(c)
		return uuid.Nil, uuid.Nil, nil, false
	}
	id, ok := h.entityID(c, "fieldDeviceId")
	if !ok {
		return uuid.Nil, uuid.Nil, nil, false
	}
	ids, err := h.projectFieldDeviceIDs(c, projectID)
	if err != nil {
		projectDetailFetchFailed(c)
		return uuid.Nil, uuid.Nil, nil, false
	}
	if _, exists := ids[id]; !exists {
		projectDetailForbidden(c)
		return uuid.Nil, uuid.Nil, nil, false
	}
	item, err := h.services.FieldDevice.GetByID(c.Request.Context(), id)
	if err != nil {
		handlerutil.RespondLocalizedError(c, http.StatusNotFound, "not_found", "facility.field_device_not_found")
		return uuid.Nil, uuid.Nil, nil, false
	}
	return projectID, id, item, true
}

func (h *FacilityDetailHandler) notifyUpdate(c *gin.Context, projectID uuid.UUID, event string, changedFields []string, entityID uuid.UUID) {
	if h.notify != nil {
		h.notify(c, projectID, event, changedFields, entityID.String())
	}
}

func (h *FacilityDetailHandler) projectWriteError(c *gin.Context, err error, aggregate string, id uuid.UUID, baseVersion uint64, fields []string, current func() (uint64, any, bool)) {
	if errors.Is(err, domain.ErrConflict) {
		if version, entity, ok := current(); ok {
			handlerutil.RespondWriteConflict(c, aggregate, id.String(), baseVersion, version, fields, entity)
			return
		}
	}
	handlerutil.RespondDomainError(c, err, handlerutil.LocalizedError(http.StatusInternalServerError, "update_failed", "facility.update_failed"), handlerutil.MapError(domain.ErrNotFound, handlerutil.LocalizedError(http.StatusNotFound, "not_found", "errors.not_found")))
}

func projectContainsCabinet(items []domainFacility.ControlCabinet, id uuid.UUID) bool {
	_, ok := containsCabinet(items, id)
	return ok
}
func projectContainsBuilding(items []domainFacility.ControlCabinet, id uuid.UUID) bool {
	for _, item := range items {
		if item.BuildingID == id {
			return true
		}
	}
	return false
}

func projectRelation(key, label, resource string, items []facilitydto.DetailRelationItem, page, limit int) facilitydto.DetailRelation {
	count := len(items)
	start := (page - 1) * limit
	if start > count {
		start = count
	}
	end := start + limit
	if end > count {
		end = count
	}
	totalPages := 1
	if count > 0 {
		totalPages = (count + limit - 1) / limit
	}
	return facilitydto.DetailRelation{Key: key, Label: label, Resource: resource, Count: int64(count), Items: items[start:end], Page: page, TotalPages: totalPages}
}

func projectSingleton(key, label, resource string, item facilitydto.DetailRelationItem) facilitydto.DetailRelation {
	return facilitydto.DetailRelation{Key: key, Label: label, Resource: resource, Count: 1, Items: []facilitydto.DetailRelationItem{item}, Page: 1, TotalPages: 1}
}

func projectBuildingResponse(item domainFacility.Building) facilitydto.BuildingResponse {
	return facilitydto.BuildingResponse{ID: item.ID, Version: item.Version, IWSCode: item.IWSCode, BuildingGroup: item.BuildingGroup, CreatedAt: item.CreatedAt, UpdatedAt: item.UpdatedAt}
}
func projectControlCabinetResponse(item domainFacility.ControlCabinet) facilitydto.ControlCabinetResponse {
	return facilitydto.ControlCabinetResponse{ID: item.ID, Version: item.Version, BuildingID: item.BuildingID, ControlCabinetNr: item.ControlCabinetNr, CreatedAt: item.CreatedAt, UpdatedAt: item.UpdatedAt}
}
func projectSPSControllerResponse(item domainFacility.SPSController) facilitydto.SPSControllerResponse {
	return facilitydto.SPSControllerResponse{ID: item.ID, Version: item.Version, ControlCabinetID: item.ControlCabinetID, GADevice: item.GADevice, DeviceName: item.DeviceName, DeviceDescription: item.DeviceDescription, DeviceLocation: item.DeviceLocation, IPAddress: item.IPAddress, Subnet: item.Subnet, Gateway: item.Gateway, Vlan: item.Vlan, CreatedAt: item.CreatedAt, UpdatedAt: item.UpdatedAt}
}
func projectSPSControllerSystemTypeResponse(item domainFacility.SPSControllerSystemType) facilitydto.SPSControllerSystemTypeResponse {
	return facilitydto.SPSControllerSystemTypeResponse{ID: item.ID, Version: item.Version, AggregateVersion: item.Version, SPSControllerID: item.SPSControllerID, SystemTypeID: item.SystemTypeID, SPSControllerName: item.SPSController.DeviceName, SystemTypeName: item.SystemType.Name, Number: item.Number, DocumentName: item.DocumentName, FieldDevicesCount: item.FieldDevicesCount, CreatedAt: item.CreatedAt, UpdatedAt: item.UpdatedAt}
}
func projectFieldDeviceResponse(item domainFacility.FieldDevice) facilitydto.FieldDeviceResponse {
	var partID *uuid.UUID
	if item.SystemPartID != uuid.Nil {
		id := item.SystemPartID
		partID = &id
	}
	return facilitydto.FieldDeviceResponse{ID: item.ID, Version: item.Version, BMK: item.BMK, Description: item.Description, TextIndividuell: item.TextIndividuell, ApparatNr: &item.ApparatNr, SPSControllerSystemTypeID: item.SPSControllerSystemTypeID, SystemPartID: partID, SpecificationID: item.SpecificationID, ApparatID: item.ApparatID, CreatedAt: item.CreatedAt, UpdatedAt: item.UpdatedAt}
}

func buildingRelationItem(item domainFacility.Building) facilitydto.DetailRelationItem {
	return facilitydto.DetailRelationItem{ID: item.ID.String(), Label: nonEmptyDetail(item.IWSCode, "Gebäude"), Subtitle: "Gruppe " + strconv.Itoa(item.BuildingGroup)}
}
func cabinetRelationItem(item domainFacility.ControlCabinet) facilitydto.DetailRelationItem {
	return facilitydto.DetailRelationItem{ID: item.ID.String(), Label: nonEmptyDetail(stringPtrDetail(item.ControlCabinetNr), "Schaltschrank")}
}
func controllerRelationItem(item domainFacility.SPSController) facilitydto.DetailRelationItem {
	return facilitydto.DetailRelationItem{ID: item.ID.String(), Label: nonEmptyDetail(item.DeviceName, "SPS-Regler"), Subtitle: stringPtrDetail(item.GADevice)}
}
func systemTypeRelationItem(item domainFacility.SPSControllerSystemType) facilitydto.DetailRelationItem {
	label := nonEmptyDetail(item.SystemType.Name, "SPS-Systemtyp")
	if item.Number != nil {
		label += " " + strconv.Itoa(*item.Number)
	}
	return facilitydto.DetailRelationItem{ID: item.ID.String(), Label: label, Subtitle: stringPtrDetail(item.DocumentName)}
}
func fieldDeviceRelationItem(item domainFacility.FieldDevice) facilitydto.DetailRelationItem {
	return facilitydto.DetailRelationItem{ID: item.ID.String(), Label: nonEmptyDetail(stringPtrDetail(item.BMK), nonEmptyDetail(stringPtrDetail(item.Description), "Feldgerät")), Subtitle: stringPtrDetail(item.Description)}
}
func apparatRelationItem(item domainFacility.Apparat) facilitydto.DetailRelationItem {
	return facilitydto.DetailRelationItem{ID: item.ID.String(), Label: nonEmptyDetail(item.ShortName, nonEmptyDetail(item.Name, "Apparat")), Subtitle: item.Name}
}
func systemPartRelationItem(item domainFacility.SystemPart) facilitydto.DetailRelationItem {
	return facilitydto.DetailRelationItem{ID: item.ID.String(), Label: nonEmptyDetail(item.ShortName, nonEmptyDetail(item.Name, "Systemteil")), Subtitle: item.Name}
}
func specificationRelationItem(item domainFacility.Specification) facilitydto.DetailRelationItem {
	parts := []string{
		stringPtrDetail(item.SpecificationSupplier),
		stringPtrDetail(item.SpecificationBrand),
		stringPtrDetail(item.SpecificationType),
	}
	values := make([]string, 0, len(parts))
	for _, part := range parts {
		if part != "" {
			values = append(values, part)
		}
	}
	return facilitydto.DetailRelationItem{ID: item.ID.String(), Label: nonEmptyDetail(strings.Join(values, " · "), "Spezifikation")}
}
func nonEmptyDetail(value, fallback string) string {
	if value = strings.TrimSpace(value); value != "" {
		return value
	}
	return fallback
}
func stringPtrDetail(value *string) string {
	if value == nil {
		return ""
	}
	return strings.TrimSpace(*value)
}
