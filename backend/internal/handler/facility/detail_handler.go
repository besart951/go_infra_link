package facility

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	domainUser "github.com/besart951/go_infra_link/backend/internal/domain/user"
	dto "github.com/besart951/go_infra_link/backend/internal/handler/dto/facility"
	"github.com/besart951/go_infra_link/backend/internal/handler/middleware"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// FacilityDetailHandler assembles the small, label-first hierarchy views used
// by the UI. It keeps relationship visibility tied to the corresponding read
// permission instead of leaking IDs from adjacent resources.
type FacilityDetailHandler struct {
	buildings    BuildingService
	cabinets     ControlCabinetService
	controllers  SPSControllerService
	systemTypes  SPSControllerSystemTypeService
	fieldDevices FieldDeviceService
	apparats     ApparatService
	systemParts  SystemPartService
	auth         middleware.AuthorizationChecker
}

func NewFacilityDetailHandler(
	buildings BuildingService,
	cabinets ControlCabinetService,
	controllers SPSControllerService,
	systemTypes SPSControllerSystemTypeService,
	fieldDevices FieldDeviceService,
	apparats ApparatService,
	systemParts SystemPartService,
) *FacilityDetailHandler {
	return &FacilityDetailHandler{
		buildings: buildings, cabinets: cabinets, controllers: controllers, systemTypes: systemTypes,
		fieldDevices: fieldDevices, apparats: apparats, systemParts: systemParts,
	}
}

func (h *FacilityDetailHandler) SetAuthorizationChecker(auth middleware.AuthorizationChecker) {
	h.auth = auth
}

func (h *FacilityDetailHandler) can(c *gin.Context, permission string) bool {
	if h.auth == nil {
		return false
	}
	role, ok := middleware.GetUserRole(c)
	if !ok {
		return false
	}
	hasPermission, err := h.auth.HasPermission(c.Request.Context(), role, permission)
	return err == nil && hasPermission
}

func detailPagination(c *gin.Context) (int, int, bool) {
	query, ok := parsePaginationQuery(c)
	if !ok {
		return 0, 0, false
	}
	page, limit := domain.NormalizePagination(query.Page, query.Limit, 12)
	if limit > 50 {
		limit = 50
	}
	return page, limit, true
}

// GetBuildingDetail godoc
// @Summary Get a building detail with permitted hierarchy relations
// @Tags facility-details
// @Produce json
// @Param id path string true "Building ID"
// @Param page query int false "Relationship page" default(1)
// @Param limit query int false "Relationship page size" default(12)
// @Success 200 {object} dto.BuildingDetailResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/facility/buildings/{id}/detail [get]
func (h *FacilityDetailHandler) GetBuildingDetail(c *gin.Context) {
	id, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}
	page, limit, ok := detailPagination(c)
	if !ok {
		return
	}

	building, err := h.buildings.GetByID(c.Request.Context(), id)
	if err != nil {
		if respondLocalizedNotFoundIf(c, err, "facility.building_not_found") {
			return
		}
		respondLocalizedError(c, http.StatusInternalServerError, "fetch_failed", "facility.fetch_failed")
		return
	}

	relations := make([]dto.DetailRelation, 0, 1)
	if h.can(c, domainUser.PermissionControlCabinetRead) {
		items, listErr := h.cabinets.ListByBuildingID(c.Request.Context(), id, page, limit, "")
		if listErr != nil {
			respondLocalizedError(c, http.StatusInternalServerError, "fetch_failed", "facility.fetch_failed")
			return
		}
		relations = append(relations, cabinetRelation(items))
	}

	c.JSON(http.StatusOK, dto.BuildingDetailResponse{
		Building:  toBuildingResponse(*building),
		Relations: relations,
		Capabilities: dto.DetailCapabilities{
			CanUpdate: h.can(c, domainUser.PermissionBuildingUpdate),
		},
	})
}

// GetControlCabinetDetail godoc
// @Summary Get a control cabinet detail with permitted hierarchy relations
// @Tags facility-details
// @Produce json
// @Param id path string true "Control cabinet ID"
// @Param page query int false "Relationship page" default(1)
// @Param limit query int false "Relationship page size" default(12)
// @Success 200 {object} dto.ControlCabinetDetailResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/facility/control-cabinets/{id}/detail [get]
func (h *FacilityDetailHandler) GetControlCabinetDetail(c *gin.Context) {
	id, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}
	page, limit, ok := detailPagination(c)
	if !ok {
		return
	}

	cabinet, err := h.cabinets.GetByID(c.Request.Context(), id)
	if err != nil {
		if respondLocalizedNotFoundIf(c, err, "facility.control_cabinet_not_found") {
			return
		}
		respondLocalizedError(c, http.StatusInternalServerError, "fetch_failed", "facility.fetch_failed")
		return
	}

	relations := make([]dto.DetailRelation, 0, 2)
	if h.can(c, domainUser.PermissionBuildingRead) {
		if building, getErr := h.buildings.GetByID(c.Request.Context(), cabinet.BuildingID); getErr == nil {
			relations = append(relations, singletonRelation("building", "Gebäude", "buildings", buildingItem(*building)))
		}
	}
	if h.can(c, domainUser.PermissionSPSControllerRead) {
		items, listErr := h.controllers.ListByControlCabinetID(c.Request.Context(), id, page, limit, "")
		if listErr != nil {
			respondLocalizedError(c, http.StatusInternalServerError, "fetch_failed", "facility.fetch_failed")
			return
		}
		relations = append(relations, controllerRelation(items))
	}

	c.JSON(http.StatusOK, dto.ControlCabinetDetailResponse{
		ControlCabinet: toControlCabinetResponse(*cabinet), Relations: relations,
		Capabilities: dto.DetailCapabilities{CanUpdate: h.can(c, domainUser.PermissionControlCabinetUpdate)},
	})
}

// GetSPSControllerDetail godoc
// @Summary Get an SPS controller detail with permitted hierarchy relations
// @Tags facility-details
// @Produce json
// @Param id path string true "SPS controller ID"
// @Param page query int false "Relationship page" default(1)
// @Param limit query int false "Relationship page size" default(12)
// @Success 200 {object} dto.SPSControllerDetailResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/facility/sps-controllers/{id}/detail [get]
func (h *FacilityDetailHandler) GetSPSControllerDetail(c *gin.Context) {
	id, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}
	page, limit, ok := detailPagination(c)
	if !ok {
		return
	}

	controller, err := h.controllers.GetByID(c.Request.Context(), id)
	if err != nil {
		if respondLocalizedNotFoundIf(c, err, "facility.sps_controller_not_found") {
			return
		}
		respondLocalizedError(c, http.StatusInternalServerError, "fetch_failed", "facility.fetch_failed")
		return
	}

	relations := h.controllerParents(c, controller.ControlCabinetID)
	if h.can(c, domainUser.PermissionSPSControllerSystemTypeRead) {
		items, listErr := h.systemTypes.ListBySPSControllerID(c.Request.Context(), id, page, limit, "")
		if listErr != nil {
			respondLocalizedError(c, http.StatusInternalServerError, "fetch_failed", "facility.fetch_failed")
			return
		}
		relations = append(relations, systemTypeRelation(items))
	}

	c.JSON(http.StatusOK, dto.SPSControllerDetailResponse{
		SPSController: toSPSControllerResponse(*controller), Relations: relations,
		Capabilities: dto.DetailCapabilities{CanUpdate: h.can(c, domainUser.PermissionSPSControllerUpdate)},
	})
}

// GetSPSControllerSystemTypeDetail godoc
// @Summary Get an SPS controller system type detail with permitted hierarchy relations
// @Tags facility-details
// @Produce json
// @Param id path string true "SPS controller system type ID"
// @Param page query int false "Relationship page" default(1)
// @Param limit query int false "Relationship page size" default(12)
// @Success 200 {object} dto.SPSControllerSystemTypeDetailResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/facility/sps-controller-system-types/{id}/detail [get]
func (h *FacilityDetailHandler) GetSPSControllerSystemTypeDetail(c *gin.Context) {
	id, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}
	page, limit, ok := detailPagination(c)
	if !ok {
		return
	}

	systemType, err := h.systemTypes.GetByID(c.Request.Context(), id)
	if err != nil {
		if respondLocalizedNotFoundIf(c, err, "facility.sps_controller_system_type_not_found") {
			return
		}
		respondLocalizedError(c, http.StatusInternalServerError, "fetch_failed", "facility.fetch_failed")
		return
	}

	relations := make([]dto.DetailRelation, 0, 3)
	if h.can(c, domainUser.PermissionSPSControllerRead) {
		if controller, getErr := h.controllers.GetByID(c.Request.Context(), systemType.SPSControllerID); getErr == nil {
			relations = append(relations, singletonRelation("sps_controller", "SPS-Regler", "sps-controllers", controllerItem(*controller)))
			relations = append(relations, h.controllerParents(c, controller.ControlCabinetID)...)
		}
	}
	if h.can(c, domainUser.PermissionFieldDeviceRead) {
		items, listErr := h.fieldDevices.ListWithFilters(c.Request.Context(), domain.PaginationParams{Page: page, Limit: limit}, domainFacility.FieldDeviceFilterParams{SPSControllerSystemTypeID: &id})
		if listErr != nil {
			respondLocalizedError(c, http.StatusInternalServerError, "fetch_failed", "facility.fetch_failed")
			return
		}
		relations = append(relations, fieldDeviceRelation(items))
	}

	c.JSON(http.StatusOK, dto.SPSControllerSystemTypeDetailResponse{
		SPSControllerSystemType: toSPSControllerSystemTypeResponse(*systemType), Relations: relations,
		Capabilities: dto.DetailCapabilities{CanUpdate: h.can(c, domainUser.PermissionSPSControllerSystemTypeUpdate)},
	})
}

// GetFieldDeviceDetail godoc
// @Summary Get a field device detail with permitted hierarchy and references
// @Tags facility-details
// @Produce json
// @Param id path string true "Field device ID"
// @Success 200 {object} dto.FieldDeviceDetailResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/facility/field-devices/{id}/detail [get]
func (h *FacilityDetailHandler) GetFieldDeviceDetail(c *gin.Context) {
	id, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}

	fieldDevice, err := h.fieldDevices.GetByID(c.Request.Context(), id)
	if err != nil {
		if respondLocalizedNotFoundIf(c, err, "facility.field_device_not_found") {
			return
		}
		respondLocalizedError(c, http.StatusInternalServerError, "fetch_failed", "facility.fetch_failed")
		return
	}

	relations := make([]dto.DetailRelation, 0, 6)
	if h.can(c, domainUser.PermissionSPSControllerSystemTypeRead) {
		if systemType, getErr := h.systemTypes.GetByID(c.Request.Context(), fieldDevice.SPSControllerSystemTypeID); getErr == nil {
			relations = append(relations, singletonRelation("sps_controller_system_type", "SPS-Systemtyp", "sps-controller-system-types", systemTypeItem(*systemType)))
			if h.can(c, domainUser.PermissionSPSControllerRead) {
				if controller, controllerErr := h.controllers.GetByID(c.Request.Context(), systemType.SPSControllerID); controllerErr == nil {
					relations = append(relations, singletonRelation("sps_controller", "SPS-Regler", "sps-controllers", controllerItem(*controller)))
					relations = append(relations, h.controllerParents(c, controller.ControlCabinetID)...)
				}
			}
		}
	}
	if h.can(c, domainUser.PermissionApparatRead) {
		if apparat, getErr := h.apparats.GetByID(c.Request.Context(), fieldDevice.ApparatID); getErr == nil {
			relations = append(relations, singletonRelation("apparat", "Apparat", "apparats", apparatItem(*apparat)))
		}
	}
	if fieldDevice.SystemPartID != uuid.Nil && h.can(c, domainUser.PermissionSystemPartRead) {
		if systemPart, getErr := h.systemParts.GetByID(c.Request.Context(), fieldDevice.SystemPartID); getErr == nil {
			relations = append(relations, singletonRelation("system_part", "Systemteil", "system-parts", systemPartItem(*systemPart)))
		}
	}
	if fieldDevice.Specification != nil && h.can(c, domainUser.PermissionSpecificationRead) {
		relations = append(relations, singletonRelation("specification", "Spezifikation", "references", specificationItem(*fieldDevice.Specification)))
	}
	if h.can(c, domainUser.PermissionBacnetObjectRead) {
		if objects, listErr := h.fieldDevices.ListBacnetObjects(c.Request.Context(), id); listErr == nil && len(objects) > 0 {
			items := make([]dto.DetailRelationItem, 0, len(objects))
			for _, object := range objects {
				items = append(items, dto.DetailRelationItem{ID: object.ID.String(), Label: bacnetLabel(object), Subtitle: string(object.SoftwareType)})
			}
			relations = append(relations, dto.DetailRelation{Key: "bacnet_objects", Label: "BACnet-Objekte", Resource: "bacnet-objects", Count: int64(len(items)), Items: items, Page: 1, TotalPages: 1})
		}
	}
	response := toFieldDeviceResponse(*fieldDevice)
	if !h.can(c, domainUser.PermissionSpecificationRead) {
		response.SpecificationID = nil
		response.Specification = nil
	}

	c.JSON(http.StatusOK, dto.FieldDeviceDetailResponse{
		FieldDevice: response, Relations: relations,
		Capabilities: dto.DetailCapabilities{CanUpdate: h.can(c, domainUser.PermissionFieldDeviceUpdate)},
	})
}

func (h *FacilityDetailHandler) controllerParents(c *gin.Context, cabinetID uuid.UUID) []dto.DetailRelation {
	if !h.can(c, domainUser.PermissionControlCabinetRead) {
		return []dto.DetailRelation{}
	}
	cabinet, err := h.cabinets.GetByID(c.Request.Context(), cabinetID)
	if err != nil {
		return []dto.DetailRelation{}
	}
	relations := []dto.DetailRelation{singletonRelation("control_cabinet", "Schaltschrank", "control-cabinets", cabinetItem(*cabinet))}
	if !h.can(c, domainUser.PermissionBuildingRead) {
		return relations
	}
	building, err := h.buildings.GetByID(c.Request.Context(), cabinet.BuildingID)
	if err != nil {
		return relations
	}
	return append(relations, singletonRelation("building", "Gebäude", "buildings", buildingItem(*building)))
}

func singletonRelation(key, label, resource string, item dto.DetailRelationItem) dto.DetailRelation {
	return dto.DetailRelation{Key: key, Label: label, Resource: resource, Count: 1, Items: []dto.DetailRelationItem{item}, Page: 1, TotalPages: 1}
}

func cabinetRelation(list *domain.PaginatedList[domainFacility.ControlCabinet]) dto.DetailRelation {
	items := make([]dto.DetailRelationItem, 0, len(list.Items))
	for _, item := range list.Items {
		items = append(items, cabinetItem(item))
	}
	return dto.DetailRelation{Key: "control_cabinets", Label: "Schaltschränke", Resource: "control-cabinets", Count: list.Total, Items: items, Page: list.Page, TotalPages: list.TotalPages}
}

func controllerRelation(list *domain.PaginatedList[domainFacility.SPSController]) dto.DetailRelation {
	items := make([]dto.DetailRelationItem, 0, len(list.Items))
	for _, item := range list.Items {
		items = append(items, controllerItem(item))
	}
	return dto.DetailRelation{Key: "sps_controllers", Label: "SPS-Regler", Resource: "sps-controllers", Count: list.Total, Items: items, Page: list.Page, TotalPages: list.TotalPages}
}

func systemTypeRelation(list *domain.PaginatedList[domainFacility.SPSControllerSystemType]) dto.DetailRelation {
	items := make([]dto.DetailRelationItem, 0, len(list.Items))
	for _, item := range list.Items {
		items = append(items, systemTypeItem(item))
	}
	return dto.DetailRelation{Key: "sps_controller_system_types", Label: "SPS-Systemtypen", Resource: "sps-controller-system-types", Count: list.Total, Items: items, Page: list.Page, TotalPages: list.TotalPages}
}

func fieldDeviceRelation(list *domain.PaginatedList[domainFacility.FieldDevice]) dto.DetailRelation {
	items := make([]dto.DetailRelationItem, 0, len(list.Items))
	for _, item := range list.Items {
		items = append(items, fieldDeviceItem(item))
	}
	return dto.DetailRelation{Key: "field_devices", Label: "Feldgeräte", Resource: "field-devices", Count: list.Total, Items: items, Page: list.Page, TotalPages: list.TotalPages}
}

func buildingItem(item domainFacility.Building) dto.DetailRelationItem {
	return dto.DetailRelationItem{ID: item.ID.String(), Label: nonEmpty(item.IWSCode, "Gebäude"), Subtitle: fmt.Sprintf("Gruppe %d", item.BuildingGroup)}
}

func cabinetItem(item domainFacility.ControlCabinet) dto.DetailRelationItem {
	return dto.DetailRelationItem{ID: item.ID.String(), Label: nonEmpty(stringValue(item.ControlCabinetNr), "Schaltschrank")}
}

func controllerItem(item domainFacility.SPSController) dto.DetailRelationItem {
	return dto.DetailRelationItem{ID: item.ID.String(), Label: nonEmpty(item.DeviceName, "SPS-Regler"), Subtitle: stringValue(item.GADevice)}
}

func systemTypeItem(item domainFacility.SPSControllerSystemType) dto.DetailRelationItem {
	label := nonEmpty(item.SystemType.Name, "SPS-Systemtyp")
	if item.Number != nil {
		label = fmt.Sprintf("%s %d", label, *item.Number)
	}
	return dto.DetailRelationItem{ID: item.ID.String(), Label: label, Subtitle: stringValue(item.DocumentName)}
}

func fieldDeviceItem(item domainFacility.FieldDevice) dto.DetailRelationItem {
	label := nonEmpty(stringValue(item.BMK), nonEmpty(stringValue(item.Description), "Feldgerät"))
	return dto.DetailRelationItem{ID: item.ID.String(), Label: label, Subtitle: stringValue(item.Description)}
}

func apparatItem(item domainFacility.Apparat) dto.DetailRelationItem {
	return dto.DetailRelationItem{ID: item.ID.String(), Label: nonEmpty(item.ShortName, nonEmpty(item.Name, "Apparat")), Subtitle: item.Name}
}

func systemPartItem(item domainFacility.SystemPart) dto.DetailRelationItem {
	return dto.DetailRelationItem{ID: item.ID.String(), Label: nonEmpty(item.ShortName, nonEmpty(item.Name, "Systemteil")), Subtitle: item.Name}
}

func specificationItem(item domainFacility.Specification) dto.DetailRelationItem {
	parts := []string{
		stringValue(item.SpecificationSupplier),
		stringValue(item.SpecificationBrand),
		stringValue(item.SpecificationType),
	}
	values := make([]string, 0, len(parts))
	for _, part := range parts {
		if part != "" {
			values = append(values, part)
		}
	}
	return dto.DetailRelationItem{ID: item.ID.String(), Label: nonEmpty(strings.Join(values, " · "), "Spezifikation")}
}

func bacnetLabel(item domainFacility.BacnetObject) string {
	return nonEmpty(item.TextFix, nonEmpty(stringValue(item.Description), "BACnet-Objekt"))
}

func nonEmpty(value, fallback string) string {
	if value = strings.TrimSpace(value); value != "" {
		return value
	}
	return fallback
}

func stringValue(value *string) string {
	if value == nil {
		return ""
	}
	return strings.TrimSpace(*value)
}
