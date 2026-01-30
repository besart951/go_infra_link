package facility

import (
	"errors"
	"net/http"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/besart951/go_infra_link/backend/internal/handler/dto"
	"github.com/gin-gonic/gin"
)

type FieldDeviceHandler struct {
	service FieldDeviceService
}

func NewFieldDeviceHandler(service FieldDeviceService) *FieldDeviceHandler {
	return &FieldDeviceHandler{service: service}
}

// CreateFieldDevice godoc
// @Summary Create a new field device
// @Tags facility-field-devices
// @Accept json
// @Produce json
// @Param field_device body dto.CreateFieldDeviceRequest true "Field Device data"
// @Success 201 {object} dto.FieldDeviceResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/facility/field-devices [post]
func (h *FieldDeviceHandler) CreateFieldDevice(c *gin.Context) {
	var req dto.CreateFieldDeviceRequest
	if !bindJSON(c, &req) {
		return
	}

	fieldDevice := toFieldDeviceModel(req)
	bacnetObjects := toFieldDeviceBacnetObjects(req.BacnetObjects)

	if err := h.service.CreateWithBacnetObjects(fieldDevice, req.ObjectDataID, bacnetObjects); err != nil {
		if ve, ok := domain.AsValidationError(err); ok {
			respondValidationError(c, ve.Fields)
			return
		}
		if errors.Is(err, domain.ErrInvalidArgument) {
			respondInvalidArgument(c, "object_data_id and bacnet_objects are mutually exclusive")
			return
		}
		if errors.Is(err, domain.ErrNotFound) {
			respondInvalidReference(c)
			return
		}
		if errors.Is(err, domain.ErrConflict) {
			respondConflict(c, "apparat_nr is already used in this scope")
			return
		}
		respondError(c, http.StatusInternalServerError, "creation_failed", err.Error())
		return
	}

	c.JSON(http.StatusCreated, toFieldDeviceResponse(*fieldDevice))
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

	fieldDevice, err := h.service.GetByID(id)
	if err != nil {
		if respondNotFoundIf(c, err, "Field Device not found") {
			return
		}
		respondError(c, http.StatusInternalServerError, "fetch_failed", err.Error())
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
// @Param building_id query string false "Filter by building ID"
// @Param control_cabinet_id query string false "Filter by control cabinet ID"
// @Param sps_controller_id query string false "Filter by SPS controller ID"
// @Param sps_controller_system_type_id query string false "Filter by SPS controller system type ID"
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

	if buildingID, ok := parseUUIDQueryParam(c, "building_id"); ok && buildingID != nil {
		filters.BuildingID = buildingID
	}

	if controlCabinetID, ok := parseUUIDQueryParam(c, "control_cabinet_id"); ok && controlCabinetID != nil {
		filters.ControlCabinetID = controlCabinetID
	}

	if spsControllerID, ok := parseUUIDQueryParam(c, "sps_controller_id"); ok && spsControllerID != nil {
		filters.SPSControllerID = spsControllerID
	}

	if spsControllerSystemTypeID, ok := parseUUIDQueryParam(c, "sps_controller_system_type_id"); ok && spsControllerSystemTypeID != nil {
		filters.SPSControllerSystemTypeID = spsControllerSystemTypeID
	}

	result, err := h.service.ListWithFilters(query.Page, query.Limit, query.Search, filters)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "fetch_failed", err.Error())
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
// @Param system_part_id query string false "System Part ID"
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
		respondInvalidArgument(c, "sps_controller_system_type_id is required")
		return
	}

	apparatID, ok := parseUUIDQueryParam(c, "apparat_id")
	if !ok {
		return
	}
	if apparatID == nil {
		respondInvalidArgument(c, "apparat_id is required")
		return
	}

	systemPartID, ok := parseUUIDQueryParam(c, "system_part_id")
	if !ok {
		return
	}

	available, err := h.service.ListAvailableApparatNumbers(*spsControllerSystemTypeID, systemPartID, *apparatID)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "fetch_failed", err.Error())
		return
	}

	c.JSON(http.StatusOK, dto.AvailableApparatNumbersResponse{Available: available})
}

// GetFieldDeviceOptions godoc
// @Summary Get all metadata needed for creating/editing field devices
// @Description Returns all apparats, system parts, object datas and their relationships in a single call
// @Tags facility-field-devices
// @Produce json
// @Success 200 {object} dto.FieldDeviceOptionsResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/facility/field-devices/options [get]
func (h *FieldDeviceHandler) GetFieldDeviceOptions(c *gin.Context) {
	options, err := h.service.GetFieldDeviceOptions()
	if err != nil {
		respondError(c, http.StatusInternalServerError, "fetch_failed", err.Error())
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

	fieldDevice, err := h.service.GetByID(id)
	if err != nil {
		if respondNotFoundIf(c, err, "Field Device not found") {
			return
		}
		respondError(c, http.StatusInternalServerError, "fetch_failed", err.Error())
		return
	}

	applyFieldDeviceUpdate(fieldDevice, req)

	var bacnetObjects *[]domainFacility.BacnetObject
	if req.BacnetObjects != nil {
		mapped := toFieldDeviceBacnetObjects(*req.BacnetObjects)
		bacnetObjects = &mapped
	}

	if err := h.service.UpdateWithBacnetObjects(fieldDevice, req.ObjectDataID, bacnetObjects); err != nil {
		if ve, ok := domain.AsValidationError(err); ok {
			respondValidationError(c, ve.Fields)
			return
		}
		if errors.Is(err, domain.ErrInvalidArgument) {
			respondInvalidArgument(c, "object_data_id and bacnet_objects are mutually exclusive")
			return
		}
		if errors.Is(err, domain.ErrNotFound) {
			respondInvalidReference(c)
			return
		}
		if errors.Is(err, domain.ErrConflict) {
			respondConflict(c, "apparat_nr is already used in this scope")
			return
		}
		respondError(c, http.StatusInternalServerError, "update_failed", err.Error())
		return
	}

	c.JSON(http.StatusOK, toFieldDeviceResponse(*fieldDevice))
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

	if err := h.service.DeleteByID(id); err != nil {
		respondError(c, http.StatusInternalServerError, "deletion_failed", err.Error())
		return
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

	objs, err := h.service.ListBacnetObjects(id)
	if err != nil {
		if respondNotFoundIf(c, err, "Field Device not found") {
			return
		}
		respondError(c, http.StatusInternalServerError, "fetch_failed", err.Error())
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

	if err := h.service.CreateSpecification(fieldDeviceID, spec); err != nil {
		if respondNotFoundIf(c, err, "Field Device not found") {
			return
		}
		if errors.Is(err, domain.ErrConflict) {
			respondConflict(c, "Specification already exists for this field device")
			return
		}
		respondError(c, http.StatusInternalServerError, "creation_failed", err.Error())
		return
	}

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

	spec, err := h.service.UpdateSpecification(fieldDeviceID, patch)
	if err != nil {
		if respondNotFoundIf(c, err, "Field Device or specification not found") {
			return
		}
		respondError(c, http.StatusInternalServerError, "update_failed", err.Error())
		return
	}

	c.JSON(http.StatusOK, toSpecificationResponse(*spec))
}
