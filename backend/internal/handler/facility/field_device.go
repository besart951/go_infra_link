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
// @Summary List field devices with pagination
// @Tags facility-field-devices
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Param search query string false "Search query"
// @Success 200 {object} dto.FieldDeviceListResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/facility/field-devices [get]
func (h *FieldDeviceHandler) ListFieldDevices(c *gin.Context) {
	query, ok := parsePaginationQuery(c)
	if !ok {
		return
	}

	result, err := h.service.List(query.Page, query.Limit, query.Search)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "fetch_failed", err.Error())
		return
	}

	c.JSON(http.StatusOK, toFieldDeviceListResponse(result))
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
