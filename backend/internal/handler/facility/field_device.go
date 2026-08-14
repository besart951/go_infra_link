package facility

import (
	"net/http"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	dto "github.com/besart951/go_infra_link/backend/internal/handler/dto/facility"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type FieldDeviceHandler struct {
	service       FieldDeviceService
	collaboration ProjectFieldDeviceChangeBroadcaster
}

func NewFieldDeviceHandler(service FieldDeviceService, broadcasters ...ProjectRefreshBroadcaster) *FieldDeviceHandler {
	h := &FieldDeviceHandler{service: service}
	if len(broadcasters) > 0 {
		h.collaboration, _ = broadcasters[0].(ProjectFieldDeviceChangeBroadcaster)
	}
	return h
}

func (h *FieldDeviceHandler) broadcastChange(c *gin.Context, id uuid.UUID, action string) {
	if h.collaboration != nil {
		h.collaboration.BroadcastFieldDeviceChange(c.Request.Context(), currentActorID(c), id, action)
	}
}

// MultiCreateFieldDevices godoc
// @Summary Create multiple field devices in a single operation
// @Description Creates multiple field devices with independent validation. Returns detailed results for each device. To link created devices to a project, use the CreateProjectFieldDevice endpoint with the returned field device IDs.
// @Tags facility-field-devices
// @Accept json
// @Produce json
// @Param request body dto.MultiCreateFieldDeviceRequest true "Multi-create request"
// @Success 200 {object} dto.MultiCreateFieldDeviceResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/facility/field-devices/multi-create [post]
func (h *FieldDeviceHandler) MultiCreateFieldDevices(c *gin.Context) {
	var req dto.MultiCreateFieldDeviceRequest
	if !bindJSON(c, &req) {
		return
	}

	// Convert DTOs to domain models
	items := make([]domainFacility.FieldDeviceCreateItem, len(req.FieldDevices))
	for i, fdReq := range req.FieldDevices {
		fieldDevice := toFieldDeviceModel(fdReq)
		bacnetObjects := toFieldDeviceBacnetObjects(fdReq.BacnetObjects)
		items[i] = domainFacility.FieldDeviceCreateItem{
			FieldDevice:   fieldDevice,
			ObjectDataID:  fdReq.ObjectDataID,
			BacnetObjects: bacnetObjects,
		}
	}

	// Create field devices
	result := h.service.MultiCreate(c.Request.Context(), items)

	c.JSON(http.StatusOK, toMultiCreateFieldDeviceResponse(result))
}

// GetFieldDevice godoc
// @Summary Get a field device by ID
// @Tags facility-field-devices
// @Produce json
// @Param id path string true "Field Device ID"
// @Success 200 {object} dto.FieldDeviceResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/facility/field-devices/{id} [get]
func (h *FieldDeviceHandler) GetFieldDevice(c *gin.Context) {
	id, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}

	fieldDevice, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		if suppressCanceledRequestError(c, err) {
			return
		}
		if respondLocalizedNotFoundIf(c, err, "facility.field_device_not_found") {
			return
		}
		respondLocalizedError(c, http.StatusInternalServerError, "fetch_failed", "facility.fetch_failed")
		return
	}

	c.JSON(http.StatusOK, toFieldDeviceResponse(*fieldDevice))
}

// ListFieldDevices godoc
// @Summary List field devices with pagination and filtering
// @Tags facility-field-devices
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(300)
// @Param search query string false "Search query"
// @Param order_by query string false "Order by (created_at,sps_system_type,bmk,description,apparat_nr,apparat,system_part,spec_supplier,spec_brand,spec_type,spec_motor_valve,spec_size,spec_install_loc,spec_ph,spec_acdc,spec_amperage,spec_power,spec_rotation)"
// @Param order query string false "Order direction (asc, desc)"
// @Param building_id query string false "Filter by building ID(s), accepts a single UUID or a | separated list"
// @Param control_cabinet_id query string false "Filter by control cabinet ID(s), accepts a single UUID or a | separated list"
// @Param sps_controller_id query string false "Filter by SPS controller ID(s), accepts a single UUID or a | separated list"
// @Param sps_controller_system_type_id query string false "Filter by SPS controller system type ID(s), accepts a single UUID or a | separated list"
// @Param project_id query string false "Filter by project ID"
// @Success 200 {object} dto.FieldDeviceListResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/facility/field-devices [get]
func (h *FieldDeviceHandler) ListFieldDevices(c *gin.Context) {
	query, ok := parsePaginationQuery(c)
	if !ok {
		return
	}

	// Parse optional filter parameters
	filters := domainFacility.FieldDeviceFilterParams{}

	buildingIDs, ok := parseUUIDListQueryParam(c, "building_id")
	if !ok {
		return
	}
	if len(buildingIDs) == 1 {
		filters.BuildingID = &buildingIDs[0]
	} else if len(buildingIDs) > 1 {
		filters.BuildingIDs = buildingIDs
	}

	controlCabinetIDs, ok := parseUUIDListQueryParam(c, "control_cabinet_id")
	if !ok {
		return
	}
	if len(controlCabinetIDs) == 1 {
		filters.ControlCabinetID = &controlCabinetIDs[0]
	} else if len(controlCabinetIDs) > 1 {
		filters.ControlCabinetIDs = controlCabinetIDs
	}

	spsControllerIDs, ok := parseUUIDListQueryParam(c, "sps_controller_id")
	if !ok {
		return
	}
	if len(spsControllerIDs) == 1 {
		filters.SPSControllerID = &spsControllerIDs[0]
	} else if len(spsControllerIDs) > 1 {
		filters.SPSControllerIDs = spsControllerIDs
	}

	spsControllerSystemTypeIDs, ok := parseUUIDListQueryParam(c, "sps_controller_system_type_id")
	if !ok {
		return
	}
	if len(spsControllerSystemTypeIDs) == 1 {
		filters.SPSControllerSystemTypeID = &spsControllerSystemTypeIDs[0]
	} else if len(spsControllerSystemTypeIDs) > 1 {
		filters.SPSControllerSystemTypeIDs = spsControllerSystemTypeIDs
	}

	projectID, ok := parseUUIDQueryParam(c, "project_id")
	if !ok {
		return
	}
	if projectID != nil {
		filters.ProjectID = projectID
	}

	params := domain.PaginationParams{
		Page:    query.Page,
		Limit:   query.Limit,
		Search:  query.Search,
		OrderBy: query.OrderBy,
		Order:   query.Order,
	}

	result, err := h.service.ListWithFilters(c.Request.Context(), params, filters)
	if err != nil {
		if suppressCanceledRequestError(c, err) {
			return
		}
		respondLocalizedError(c, http.StatusInternalServerError, "fetch_failed", "facility.fetch_failed")
		return
	}

	c.JSON(http.StatusOK, toFieldDeviceListResponse(result))
}

// ListAvailableApparatNumbers godoc
// @Summary List available apparat numbers for field devices
// @Tags facility-field-devices
// @Produce json
// @Param sps_controller_system_type_id query string true "SPS Controller System Type ID"
// @Param apparat_id query string true "Apparat ID"
// @Param system_part_id query string true "System Part ID"
// @Success 200 {object} dto.AvailableApparatNumbersResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/facility/field-devices/available-apparat-nr [get]
func (h *FieldDeviceHandler) ListAvailableApparatNumbers(c *gin.Context) {
	spsControllerSystemTypeID, ok := parseUUIDQueryParam(c, "sps_controller_system_type_id")
	if !ok {
		return
	}
	if spsControllerSystemTypeID == nil {
		respondLocalizedInvalidArgument(c, "facility.sps_controller_system_type_id_required")
		return
	}

	apparatID, ok := parseUUIDQueryParam(c, "apparat_id")
	if !ok {
		return
	}
	if apparatID == nil {
		respondLocalizedInvalidArgument(c, "facility.apparat_id_required")
		return
	}

	systemPartID, ok := parseUUIDQueryParam(c, "system_part_id")
	if !ok {
		return
	}
	if systemPartID == nil {
		respondLocalizedInvalidArgument(c, "facility.system_part_id_required")
		return
	}

	available, err := h.service.ListAvailableApparatNumbers(c.Request.Context(), *spsControllerSystemTypeID, *systemPartID, *apparatID)
	if err != nil {
		if suppressCanceledRequestError(c, err) {
			return
		}
		respondLocalizedError(c, http.StatusInternalServerError, "fetch_failed", "facility.fetch_failed")
		return
	}

	c.JSON(http.StatusOK, dto.AvailableApparatNumbersResponse{Available: available})
}

// GetFieldDeviceOptions godoc
// @Summary Get all metadata needed for creating/editing field devices
// @Description Returns all apparats, system parts, object datas and their relationships in a single call. This returns global templates (object data where project_id is null and is_active = true).
// @Tags facility-field-devices
// @Produce json
// @Success 200 {object} dto.FieldDeviceOptionsResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/facility/field-devices/options [get]
func (h *FieldDeviceHandler) GetFieldDeviceOptions(c *gin.Context) {
	options, err := h.service.GetFieldDeviceOptions(c.Request.Context())
	if err != nil {
		if suppressCanceledRequestError(c, err) {
			return
		}
		respondLocalizedError(c, http.StatusInternalServerError, "fetch_failed", "facility.fetch_failed")
		return
	}

	c.JSON(http.StatusOK, toFieldDeviceOptionsResponse(options))
}

// UpdateFieldDevice godoc
// @Summary Update a field device
// @Tags facility-field-devices
// @Accept json
// @Produce json
// @Param id path string true "Field Device ID"
// @Param field_device body dto.UpdateFieldDeviceRequest true "Field Device data"
// @Success 200 {object} dto.FieldDeviceResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/facility/field-devices/{id} [put]
func (h *FieldDeviceHandler) UpdateFieldDevice(c *gin.Context) {
	id, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}

	var req dto.UpdateFieldDeviceRequest
	if !bindJSON(c, &req) {
		return
	}

	ctx := c.Request.Context()

	fieldDevice, err := h.service.GetByID(ctx, id)
	if err != nil {
		if respondLocalizedNotFoundIf(c, err, "facility.field_device_not_found") {
			return
		}
		respondLocalizedError(c, http.StatusInternalServerError, "fetch_failed", "facility.fetch_failed")
		return
	}
	baseVersion := fieldDevice.Version
	if req.BaseVersion != nil {
		baseVersion = *req.BaseVersion
	}

	applyFieldDeviceUpdate(fieldDevice, req)

	var bacnetObjects *[]domainFacility.BacnetObject
	if req.BacnetObjects != nil {
		mapped := toFieldDeviceBacnetObjects(*req.BacnetObjects)
		bacnetObjects = &mapped
	}

	if err := h.service.UpdateWithBacnetObjects(ctx, fieldDevice, req.ObjectDataID, bacnetObjects); err != nil {
		if current, getErr := h.service.GetByID(ctx, id); getErr == nil && respondWriteConflict(c, err, "field_device", id, baseVersion, fieldDeviceUpdatePaths(req), current.Version, toFieldDeviceResponse(*current)) {
			return
		}
		respondLocalizedDomainError(c, err, "update_failed", "facility.update_failed",
			localizedInvalidArgument("facility.mutually_exclusive_error"),
			localizedInvalidReference(),
			localizedConflict("facility.apparat_nr_already_used"),
		)
		return
	}

	c.JSON(http.StatusOK, toFieldDeviceResponse(*fieldDevice))
	h.broadcastChange(c, fieldDevice.ID, "updated")
}

// DeleteFieldDevice godoc
// @Summary Delete a field device
// @Tags facility-field-devices
// @Produce json
// @Param id path string true "Field Device ID"
// @Success 204
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/facility/field-devices/{id} [delete]
func (h *FieldDeviceHandler) DeleteFieldDevice(c *gin.Context) {
	id, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}

	ctx := c.Request.Context()
	projectIDs := captureDeleteAudience(ctx, h.collaboration, "field_device", id)
	if err := h.service.DeleteByID(ctx, id); err != nil {
		respondLocalizedDomainError(c, err, "deletion_failed", "facility.deletion_failed",
			localizedNotFound("facility.field_device_not_found"),
		)
		return
	}
	if !broadcastCapturedDelete(ctx, h.collaboration, currentActorID(c), projectIDs, "field_device", id) {
		h.broadcastChange(c, id, "deleted")
	}

	c.Status(http.StatusNoContent)
}

// ListFieldDeviceBacnetObjects godoc
// @Summary List bacnet objects for a field device (hydration)
// @Tags facility-field-devices
// @Produce json
// @Param id path string true "Field Device ID"
// @Success 200 {array} dto.BacnetObjectResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/facility/field-devices/{id}/bacnet-objects [get]
func (h *FieldDeviceHandler) ListFieldDeviceBacnetObjects(c *gin.Context) {
	id, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}

	objs, err := h.service.ListBacnetObjects(c.Request.Context(), id)
	if err != nil {
		if respondLocalizedNotFoundIf(c, err, "facility.field_device_not_found") {
			return
		}
		respondLocalizedError(c, http.StatusInternalServerError, "fetch_failed", "facility.fetch_failed")
		return
	}

	c.JSON(http.StatusOK, toBacnetObjectResponses(objs))
}

// CreateFieldDeviceSpecification godoc
// @Summary Create specification for a field device
// @Tags facility-field-devices
// @Accept json
// @Produce json
// @Param id path string true "Field Device ID"
// @Param specification body dto.CreateFieldDeviceSpecificationRequest true "Specification data"
// @Success 201 {object} dto.SpecificationResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 409 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/facility/field-devices/{id}/specification [post]
func (h *FieldDeviceHandler) CreateFieldDeviceSpecification(c *gin.Context) {
	fieldDeviceID, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}

	var req dto.CreateFieldDeviceSpecificationRequest
	if !bindJSON(c, &req) {
		return
	}

	spec := toFieldDeviceSpecification(req)

	if err := h.service.CreateSpecification(c.Request.Context(), fieldDeviceID, spec); err != nil {
		respondLocalizedDomainError(c, err, "creation_failed", "facility.creation_failed",
			localizedNotFound("facility.field_device_not_found"),
			localizedConflict("facility.specification_already_exists"),
		)
		return
	}

	h.broadcastChange(c, fieldDeviceID, "updated")
	c.JSON(http.StatusCreated, toSpecificationResponse(*spec))
}

// UpdateFieldDeviceSpecification godoc
// @Summary Update specification for a field device
// @Tags facility-field-devices
// @Accept json
// @Produce json
// @Param id path string true "Field Device ID"
// @Param specification body dto.UpdateFieldDeviceSpecificationRequest true "Specification data"
// @Success 200 {object} dto.SpecificationResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/facility/field-devices/{id}/specification [put]
func (h *FieldDeviceHandler) UpdateFieldDeviceSpecification(c *gin.Context) {
	fieldDeviceID, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}

	var req dto.UpdateFieldDeviceSpecificationRequest
	if !bindJSON(c, &req) {
		return
	}

	patch := toFieldDeviceSpecificationPatch(req)

	spec, err := h.service.UpdateSpecificationPatch(c.Request.Context(), fieldDeviceID, patch)
	if err != nil {
		if respondLocalizedNotFoundIf(c, err, "facility.field_device_not_found") {
			return
		}
		respondLocalizedError(c, http.StatusInternalServerError, "update_failed", "facility.update_failed")
		return
	}

	h.broadcastChange(c, fieldDeviceID, "updated")
	c.JSON(http.StatusOK, toSpecificationResponse(*spec))
}

// BulkUpdateFieldDevices godoc
// @Summary Bulk update multiple field devices
// @Description Updates multiple field devices in a single operation. Supports nested specification and BACnet objects updates.
// @Tags facility-field-devices
// @Accept json
// @Produce json
// @Param request body dto.BulkUpdateFieldDeviceRequest true "Bulk update request"
// @Success 200 {object} dto.BulkUpdateFieldDeviceResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/facility/field-devices/bulk-update [patch]
func (h *FieldDeviceHandler) BulkUpdateFieldDevices(c *gin.Context) {
	var req dto.BulkUpdateFieldDeviceRequest
	if !bindJSON(c, &req) {
		return
	}

	// Convert DTOs to domain models
	updates := make([]domainFacility.BulkFieldDeviceUpdate, len(req.Updates))
	for i, item := range req.Updates {
		var spec *domainFacility.SpecificationPatch
		if item.Specification != nil {
			spec = toSpecificationPatchFromInput(item.Specification)
		}

		var bacnetObjs *[]domainFacility.BacnetObjectPatch
		if item.BacnetObjects != nil {
			mapped := toBacnetObjectPatches(*item.BacnetObjects)
			bacnetObjs = &mapped
		}

		updates[i] = domainFacility.BulkFieldDeviceUpdate{
			ID:                 item.ID,
			BaseVersion:        item.BaseVersion,
			BMK:                item.BMK.Value,
			HasBMK:             item.BMK.Set,
			Description:        item.Description.Value,
			HasDescription:     item.Description.Set,
			TextIndividuell:    item.TextIndividuell.Value,
			HasTextIndividuell: item.TextIndividuell.Set,
			ApparatNr:          item.ApparatNr,
			ApparatID:          item.ApparatID,
			SystemPartID:       item.SystemPartID,
			Specification:      spec,
			BacnetObjects:      bacnetObjs,
		}
	}

	result := h.service.BulkUpdate(c.Request.Context(), updates)
	for _, item := range result.Results {
		if item.Success {
			h.broadcastChange(c, item.ID, "updated")
		}
	}

	c.JSON(http.StatusOK, toBulkOperationResponse(result))
}

// BulkDeleteFieldDevices godoc
// @Summary Bulk delete multiple field devices
// @Tags facility-field-devices
// @Accept json
// @Produce json
// @Param request body dto.BulkDeleteFieldDeviceRequest true "Bulk delete request"
// @Success 200 {object} dto.BulkDeleteFieldDeviceResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/facility/field-devices/bulk-delete [delete]
func (h *FieldDeviceHandler) BulkDeleteFieldDevices(c *gin.Context) {
	var req dto.BulkDeleteFieldDeviceRequest
	if !bindJSON(c, &req) {
		return
	}

	result := h.service.BulkDelete(c.Request.Context(), req.IDs)
	for _, item := range result.Results {
		if item.Success {
			h.broadcastChange(c, item.ID, "deleted")
		}
	}

	c.JSON(http.StatusOK, toBulkOperationResponse(result))
}

func toBulkOperationResponse(result *domainFacility.BulkOperationResult) dto.BulkUpdateFieldDeviceResponse {
	results := make([]dto.BulkOperationResultItem, len(result.Results))
	for i, r := range result.Results {
		results[i] = dto.BulkOperationResultItem{
			ID:                r.ID,
			Success:           r.Success,
			Version:           r.Version,
			Merged:            r.Merged,
			FieldDevice:       fieldDeviceResponsePtr(r.FieldDevice),
			Error:             r.Error,
			Fields:            r.Fields,
			Suggestions:       r.Suggestions,
			SuggestionOptions: r.SuggestionOptions,
		}
	}
	return dto.BulkUpdateFieldDeviceResponse{
		Results:      results,
		TotalCount:   result.TotalCount,
		SuccessCount: result.SuccessCount,
		FailureCount: result.FailureCount,
	}
}

func fieldDeviceResponsePtr(item *domainFacility.FieldDevice) *dto.FieldDeviceResponse {
	if item == nil {
		return nil
	}
	response := toFieldDeviceResponse(*item)
	return &response
}
