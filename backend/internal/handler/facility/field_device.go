package facility

import (
	"errors"
	"net/http"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/besart951/go_infra_link/backend/internal/handler/dto"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "validation_error",
			Message: err.Error(),
		})
		return
	}

	var apparatNr int
	if req.ApparatNr != nil {
		apparatNr = *req.ApparatNr
	}
	var systemPartID uuid.UUID
	if req.SystemPartID != nil {
		systemPartID = *req.SystemPartID
	}

	fieldDevice := &domainFacility.FieldDevice{
		BMK:                       req.BMK,
		Description:               req.Description,
		ApparatNr:                 apparatNr,
		SPSControllerSystemTypeID: req.SPSControllerSystemTypeID,
		SystemPartID:              systemPartID,
		ApparatID:                 req.ApparatID,
	}

	bacnetObjects := make([]domainFacility.BacnetObject, 0, len(req.BacnetObjects))
	for _, bo := range req.BacnetObjects {
		bacnetObjects = append(bacnetObjects, domainFacility.BacnetObject{
			TextFix:             bo.TextFix,
			Description:         bo.Description,
			GMSVisible:          bo.GMSVisible,
			Optional:            bo.Optional,
			TextIndividual:      bo.TextIndividual,
			SoftwareType:        domainFacility.BacnetSoftwareType(bo.SoftwareType),
			SoftwareNumber:      uint16(bo.SoftwareNumber),
			HardwareType:        domainFacility.BacnetHardwareType(bo.HardwareType),
			HardwareQuantity:    uint8(bo.HardwareQuantity),
			SoftwareReferenceID: bo.SoftwareReferenceID,
			StateTextID:         bo.StateTextID,
			NotificationClassID: bo.NotificationClassID,
			AlarmDefinitionID:   bo.AlarmDefinitionID,
		})
	}

	if err := h.service.CreateWithBacnetObjects(fieldDevice, req.ObjectDataID, bacnetObjects); err != nil {
		if errors.Is(err, domain.ErrInvalidArgument) {
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{
				Error:   "validation_error",
				Message: "object_data_id and bacnet_objects are mutually exclusive",
			})
			return
		}
		if errors.Is(err, domain.ErrNotFound) {
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{
				Error:   "invalid_reference",
				Message: "Referenced entity not found or deleted",
			})
			return
		}
		if errors.Is(err, domain.ErrConflict) {
			c.JSON(http.StatusConflict, dto.ErrorResponse{
				Error:   "conflict",
				Message: "apparat_nr is already used in this scope",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "creation_failed",
			Message: err.Error(),
		})
		return
	}

	var sysPartID *uuid.UUID
	if fieldDevice.SystemPartID != uuid.Nil {
		sysPartID = &fieldDevice.SystemPartID
	}

	response := dto.FieldDeviceResponse{
		ID:                        fieldDevice.ID,
		BMK:                       fieldDevice.BMK,
		Description:               fieldDevice.Description,
		ApparatNr:                 &fieldDevice.ApparatNr,
		SPSControllerSystemTypeID: fieldDevice.SPSControllerSystemTypeID,
		SystemPartID:              sysPartID,
		SpecificationID:           fieldDevice.SpecificationID,
		ApparatID:                 fieldDevice.ApparatID,
		CreatedAt:                 fieldDevice.CreatedAt,
		UpdatedAt:                 fieldDevice.UpdatedAt,
	}

	c.JSON(http.StatusCreated, response)
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
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid UUID format",
		})
		return
	}

	fieldDevice, err := h.service.GetByID(id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			c.JSON(http.StatusNotFound, dto.ErrorResponse{
				Error:   "not_found",
				Message: "Field Device not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "fetch_failed",
			Message: err.Error(),
		})
		return
	}

	var sysPartID *uuid.UUID
	if fieldDevice.SystemPartID != uuid.Nil {
		sysPartID = &fieldDevice.SystemPartID
	}

	response := dto.FieldDeviceResponse{
		ID:                        fieldDevice.ID,
		BMK:                       fieldDevice.BMK,
		Description:               fieldDevice.Description,
		ApparatNr:                 &fieldDevice.ApparatNr,
		SPSControllerSystemTypeID: fieldDevice.SPSControllerSystemTypeID,
		SystemPartID:              sysPartID,
		SpecificationID:           fieldDevice.SpecificationID,
		ApparatID:                 fieldDevice.ApparatID,
		CreatedAt:                 fieldDevice.CreatedAt,
		UpdatedAt:                 fieldDevice.UpdatedAt,
	}

	c.JSON(http.StatusOK, response)
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
	var query dto.PaginationQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "validation_error",
			Message: err.Error(),
		})
		return
	}

	result, err := h.service.List(query.Page, query.Limit, query.Search)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "fetch_failed",
			Message: err.Error(),
		})
		return
	}

	items := make([]dto.FieldDeviceResponse, len(result.Items))
	for i, fieldDevice := range result.Items {
		var sysPartID *uuid.UUID
		if fieldDevice.SystemPartID != uuid.Nil {
			sysPartID = &fieldDevice.SystemPartID
		}
		items[i] = dto.FieldDeviceResponse{
			ID:                        fieldDevice.ID,
			BMK:                       fieldDevice.BMK,
			Description:               fieldDevice.Description,
			ApparatNr:                 &fieldDevice.ApparatNr,
			SPSControllerSystemTypeID: fieldDevice.SPSControllerSystemTypeID,
			SystemPartID:              sysPartID,
			SpecificationID:           fieldDevice.SpecificationID,
			ApparatID:                 fieldDevice.ApparatID,
			CreatedAt:                 fieldDevice.CreatedAt,
			UpdatedAt:                 fieldDevice.UpdatedAt,
		}
	}

	response := dto.FieldDeviceListResponse{
		Items:      items,
		Total:      result.Total,
		Page:       result.Page,
		TotalPages: result.TotalPages,
	}

	c.JSON(http.StatusOK, response)
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
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid UUID format",
		})
		return
	}

	var req dto.UpdateFieldDeviceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "validation_error",
			Message: err.Error(),
		})
		return
	}

	fieldDevice, err := h.service.GetByID(id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			c.JSON(http.StatusNotFound, dto.ErrorResponse{
				Error:   "not_found",
				Message: "Field Device not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "fetch_failed",
			Message: err.Error(),
		})
		return
	}

	if req.BMK != nil {
		fieldDevice.BMK = req.BMK
	}
	if req.Description != nil {
		fieldDevice.Description = req.Description
	}
	if req.ApparatNr != nil {
		fieldDevice.ApparatNr = *req.ApparatNr
	}
	if req.SPSControllerSystemTypeID != uuid.Nil {
		fieldDevice.SPSControllerSystemTypeID = req.SPSControllerSystemTypeID
	}
	if req.SystemPartID != nil {
		fieldDevice.SystemPartID = *req.SystemPartID
	}
	if req.ApparatID != uuid.Nil {
		fieldDevice.ApparatID = req.ApparatID
	}

	var bacnetObjects *[]domainFacility.BacnetObject
	if req.BacnetObjects != nil {
		mapped := make([]domainFacility.BacnetObject, 0, len(*req.BacnetObjects))
		for _, bo := range *req.BacnetObjects {
			mapped = append(mapped, domainFacility.BacnetObject{
				TextFix:             bo.TextFix,
				Description:         bo.Description,
				GMSVisible:          bo.GMSVisible,
				Optional:            bo.Optional,
				TextIndividual:      bo.TextIndividual,
				SoftwareType:        domainFacility.BacnetSoftwareType(bo.SoftwareType),
				SoftwareNumber:      uint16(bo.SoftwareNumber),
				HardwareType:        domainFacility.BacnetHardwareType(bo.HardwareType),
				HardwareQuantity:    uint8(bo.HardwareQuantity),
				SoftwareReferenceID: bo.SoftwareReferenceID,
				StateTextID:         bo.StateTextID,
				NotificationClassID: bo.NotificationClassID,
				AlarmDefinitionID:   bo.AlarmDefinitionID,
			})
		}
		bacnetObjects = &mapped
	}

	if err := h.service.UpdateWithBacnetObjects(fieldDevice, req.ObjectDataID, bacnetObjects); err != nil {
		if errors.Is(err, domain.ErrInvalidArgument) {
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{
				Error:   "validation_error",
				Message: "object_data_id and bacnet_objects are mutually exclusive",
			})
			return
		}
		if errors.Is(err, domain.ErrNotFound) {
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{
				Error:   "invalid_reference",
				Message: "Referenced entity not found or deleted",
			})
			return
		}
		if errors.Is(err, domain.ErrConflict) {
			c.JSON(http.StatusConflict, dto.ErrorResponse{
				Error:   "conflict",
				Message: "apparat_nr is already used in this scope",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "update_failed",
			Message: err.Error(),
		})
		return
	}

	var sysPartID *uuid.UUID
	if fieldDevice.SystemPartID != uuid.Nil {
		sysPartID = &fieldDevice.SystemPartID
	}

	response := dto.FieldDeviceResponse{
		ID:                        fieldDevice.ID,
		BMK:                       fieldDevice.BMK,
		Description:               fieldDevice.Description,
		ApparatNr:                 &fieldDevice.ApparatNr,
		SPSControllerSystemTypeID: fieldDevice.SPSControllerSystemTypeID,
		SystemPartID:              sysPartID,
		SpecificationID:           fieldDevice.SpecificationID,
		ApparatID:                 fieldDevice.ApparatID,
		CreatedAt:                 fieldDevice.CreatedAt,
		UpdatedAt:                 fieldDevice.UpdatedAt,
	}

	c.JSON(http.StatusOK, response)
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
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid UUID format",
		})
		return
	}

	if err := h.service.DeleteByIds([]uuid.UUID{id}); err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "deletion_failed",
			Message: err.Error(),
		})
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
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid UUID format",
		})
		return
	}

	objs, err := h.service.ListBacnetObjects(id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			c.JSON(http.StatusNotFound, dto.ErrorResponse{
				Error:   "not_found",
				Message: "Field Device not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "fetch_failed",
			Message: err.Error(),
		})
		return
	}

	out := make([]dto.BacnetObjectResponse, 0, len(objs))
	for _, o := range objs {
		out = append(out, dto.BacnetObjectResponse{
			ID:                  o.ID.String(),
			TextFix:             o.TextFix,
			Description:         o.Description,
			GMSVisible:          o.GMSVisible,
			Optional:            o.Optional,
			TextIndividual:      o.TextIndividual,
			SoftwareType:        string(o.SoftwareType),
			SoftwareNumber:      int(o.SoftwareNumber),
			HardwareType:        string(o.HardwareType),
			HardwareQuantity:    int(o.HardwareQuantity),
			FieldDeviceID:       o.FieldDeviceID,
			SoftwareReferenceID: o.SoftwareReferenceID,
			StateTextID:         o.StateTextID,
			NotificationClassID: o.NotificationClassID,
			AlarmDefinitionID:   o.AlarmDefinitionID,
			CreatedAt:           o.CreatedAt,
			UpdatedAt:           o.UpdatedAt,
		})
	}

	c.JSON(http.StatusOK, out)
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
	fieldDeviceID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid_id", Message: "Invalid UUID format"})
		return
	}

	var req dto.CreateFieldDeviceSpecificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "validation_error", Message: err.Error()})
		return
	}

	spec := &domainFacility.Specification{
		SpecificationSupplier:                     req.SpecificationSupplier,
		SpecificationBrand:                        req.SpecificationBrand,
		SpecificationType:                         req.SpecificationType,
		AdditionalInfoMotorValve:                  req.AdditionalInfoMotorValve,
		AdditionalInfoSize:                        req.AdditionalInfoSize,
		AdditionalInformationInstallationLocation: req.AdditionalInformationInstallationLocation,
		ElectricalConnectionPH:                    req.ElectricalConnectionPH,
		ElectricalConnectionACDC:                  req.ElectricalConnectionACDC,
		ElectricalConnectionAmperage:              req.ElectricalConnectionAmperage,
		ElectricalConnectionPower:                 req.ElectricalConnectionPower,
		ElectricalConnectionRotation:              req.ElectricalConnectionRotation,
	}

	if err := h.service.CreateSpecification(fieldDeviceID, spec); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: "not_found", Message: "Field Device not found"})
			return
		}
		if errors.Is(err, domain.ErrConflict) {
			c.JSON(http.StatusConflict, dto.ErrorResponse{Error: "conflict", Message: "Specification already exists for this field device"})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "creation_failed", Message: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, dto.SpecificationResponse{
		ID:                       spec.ID,
		FieldDeviceID:            spec.FieldDeviceID,
		SpecificationSupplier:    spec.SpecificationSupplier,
		SpecificationBrand:       spec.SpecificationBrand,
		SpecificationType:        spec.SpecificationType,
		AdditionalInfoMotorValve: spec.AdditionalInfoMotorValve,
		AdditionalInfoSize:       spec.AdditionalInfoSize,
		AdditionalInformationInstallationLocation: spec.AdditionalInformationInstallationLocation,
		ElectricalConnectionPH:                    spec.ElectricalConnectionPH,
		ElectricalConnectionACDC:                  spec.ElectricalConnectionACDC,
		ElectricalConnectionAmperage:              spec.ElectricalConnectionAmperage,
		ElectricalConnectionPower:                 spec.ElectricalConnectionPower,
		ElectricalConnectionRotation:              spec.ElectricalConnectionRotation,
		CreatedAt:                                 spec.CreatedAt,
		UpdatedAt:                                 spec.UpdatedAt,
	})
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
	fieldDeviceID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid_id", Message: "Invalid UUID format"})
		return
	}

	var req dto.UpdateFieldDeviceSpecificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "validation_error", Message: err.Error()})
		return
	}

	patch := &domainFacility.Specification{
		SpecificationSupplier:                     req.SpecificationSupplier,
		SpecificationBrand:                        req.SpecificationBrand,
		SpecificationType:                         req.SpecificationType,
		AdditionalInfoMotorValve:                  req.AdditionalInfoMotorValve,
		AdditionalInfoSize:                        req.AdditionalInfoSize,
		AdditionalInformationInstallationLocation: req.AdditionalInformationInstallationLocation,
		ElectricalConnectionPH:                    req.ElectricalConnectionPH,
		ElectricalConnectionACDC:                  req.ElectricalConnectionACDC,
		ElectricalConnectionAmperage:              req.ElectricalConnectionAmperage,
		ElectricalConnectionPower:                 req.ElectricalConnectionPower,
		ElectricalConnectionRotation:              req.ElectricalConnectionRotation,
	}

	spec, err := h.service.UpdateSpecification(fieldDeviceID, patch)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: "not_found", Message: "Field Device or specification not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "update_failed", Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.SpecificationResponse{
		ID:                       spec.ID,
		FieldDeviceID:            spec.FieldDeviceID,
		SpecificationSupplier:    spec.SpecificationSupplier,
		SpecificationBrand:       spec.SpecificationBrand,
		SpecificationType:        spec.SpecificationType,
		AdditionalInfoMotorValve: spec.AdditionalInfoMotorValve,
		AdditionalInfoSize:       spec.AdditionalInfoSize,
		AdditionalInformationInstallationLocation: spec.AdditionalInformationInstallationLocation,
		ElectricalConnectionPH:                    spec.ElectricalConnectionPH,
		ElectricalConnectionACDC:                  spec.ElectricalConnectionACDC,
		ElectricalConnectionAmperage:              spec.ElectricalConnectionAmperage,
		ElectricalConnectionPower:                 spec.ElectricalConnectionPower,
		ElectricalConnectionRotation:              spec.ElectricalConnectionRotation,
		CreatedAt:                                 spec.CreatedAt,
		UpdatedAt:                                 spec.UpdatedAt,
	})
}
