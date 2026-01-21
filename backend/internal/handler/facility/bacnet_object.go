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

type BacnetObjectHandler struct {
	service BacnetObjectService
}

func NewBacnetObjectHandler(service BacnetObjectService) *BacnetObjectHandler {
	return &BacnetObjectHandler{service: service}
}

// CreateBacnetObject godoc
// @Summary Create a bacnet object (for field device or object data)
// @Tags facility-bacnet-objects
// @Accept json
// @Produce json
// @Param bacnet_object body dto.CreateBacnetObjectRequest true "Bacnet Object data"
// @Success 201 {object} dto.BacnetObjectResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 409 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/facility/bacnet-objects [post]
func (h *BacnetObjectHandler) CreateBacnetObject(c *gin.Context) {
	var req dto.CreateBacnetObjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "validation_error", Message: err.Error()})
		return
	}

	obj := &domainFacility.BacnetObject{
		TextFix:             req.TextFix,
		Description:         req.Description,
		GMSVisible:          req.GMSVisible,
		Optional:            req.Optional,
		TextIndividual:      req.TextIndividual,
		SoftwareType:        domainFacility.BacnetSoftwareType(req.SoftwareType),
		SoftwareNumber:      uint16(req.SoftwareNumber),
		HardwareType:        domainFacility.BacnetHardwareType(req.HardwareType),
		HardwareQuantity:    uint8(req.HardwareQuantity),
		SoftwareReferenceID: req.SoftwareReferenceID,
		StateTextID:         req.StateTextID,
		NotificationClassID: req.NotificationClassID,
		AlarmDefinitionID:   req.AlarmDefinitionID,
	}

	if err := h.service.CreateWithParent(obj, req.FieldDeviceID, req.ObjectDataID); err != nil {
		if errors.Is(err, domain.ErrInvalidArgument) {
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "validation_error", Message: "exactly one of field_device_id or object_data_id must be set"})
			return
		}
		if errors.Is(err, domain.ErrNotFound) {
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid_reference", Message: "Referenced entity not found or deleted"})
			return
		}
		if errors.Is(err, domain.ErrConflict) {
			c.JSON(http.StatusConflict, dto.ErrorResponse{Error: "conflict", Message: "entity conflict"})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "creation_failed", Message: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, dto.BacnetObjectResponse{
		ID:                  obj.ID.String(),
		TextFix:             obj.TextFix,
		Description:         obj.Description,
		GMSVisible:          obj.GMSVisible,
		Optional:            obj.Optional,
		TextIndividual:      obj.TextIndividual,
		SoftwareType:        string(obj.SoftwareType),
		SoftwareNumber:      int(obj.SoftwareNumber),
		HardwareType:        string(obj.HardwareType),
		HardwareQuantity:    int(obj.HardwareQuantity),
		FieldDeviceID:       obj.FieldDeviceID,
		SoftwareReferenceID: obj.SoftwareReferenceID,
		StateTextID:         obj.StateTextID,
		NotificationClassID: obj.NotificationClassID,
		AlarmDefinitionID:   obj.AlarmDefinitionID,
		CreatedAt:           obj.CreatedAt,
		UpdatedAt:           obj.UpdatedAt,
	})
}

// UpdateBacnetObject godoc
// @Summary Update a bacnet object
// @Tags facility-bacnet-objects
// @Accept json
// @Produce json
// @Param id path string true "Bacnet Object ID"
// @Param bacnet_object body dto.UpdateBacnetObjectRequest true "Bacnet Object data"
// @Success 200 {object} dto.BacnetObjectResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 409 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/facility/bacnet-objects/{id} [put]
func (h *BacnetObjectHandler) UpdateBacnetObject(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid_id", Message: "Invalid UUID format"})
		return
	}

	var req dto.UpdateBacnetObjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "validation_error", Message: err.Error()})
		return
	}
	if req.FieldDeviceID != nil && req.ObjectDataID != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "validation_error", Message: "field_device_id and object_data_id are mutually exclusive"})
		return
	}

	existing, err := h.service.GetByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "fetch_failed", Message: err.Error()})
		return
	}
	if existing == nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: "not_found", Message: "Bacnet Object not found"})
		return
	}

	existing.TextFix = req.TextFix
	existing.Description = req.Description
	existing.GMSVisible = req.GMSVisible
	existing.Optional = req.Optional
	existing.TextIndividual = req.TextIndividual
	existing.SoftwareType = domainFacility.BacnetSoftwareType(req.SoftwareType)
	existing.SoftwareNumber = uint16(req.SoftwareNumber)
	existing.HardwareType = domainFacility.BacnetHardwareType(req.HardwareType)
	existing.HardwareQuantity = uint8(req.HardwareQuantity)
	existing.SoftwareReferenceID = req.SoftwareReferenceID
	existing.StateTextID = req.StateTextID
	existing.NotificationClassID = req.NotificationClassID
	existing.AlarmDefinitionID = req.AlarmDefinitionID
	if req.FieldDeviceID != nil {
		existing.FieldDeviceID = req.FieldDeviceID
	}

	if err := h.service.Update(existing, req.ObjectDataID); err != nil {
		if errors.Is(err, domain.ErrInvalidArgument) {
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "validation_error", Message: err.Error()})
			return
		}
		if errors.Is(err, domain.ErrNotFound) {
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid_reference", Message: "Referenced entity not found or deleted"})
			return
		}
		if errors.Is(err, domain.ErrConflict) {
			c.JSON(http.StatusConflict, dto.ErrorResponse{Error: "conflict", Message: "entity conflict"})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "update_failed", Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.BacnetObjectResponse{
		ID:                  existing.ID.String(),
		TextFix:             existing.TextFix,
		Description:         existing.Description,
		GMSVisible:          existing.GMSVisible,
		Optional:            existing.Optional,
		TextIndividual:      existing.TextIndividual,
		SoftwareType:        string(existing.SoftwareType),
		SoftwareNumber:      int(existing.SoftwareNumber),
		HardwareType:        string(existing.HardwareType),
		HardwareQuantity:    int(existing.HardwareQuantity),
		FieldDeviceID:       existing.FieldDeviceID,
		SoftwareReferenceID: existing.SoftwareReferenceID,
		StateTextID:         existing.StateTextID,
		NotificationClassID: existing.NotificationClassID,
		AlarmDefinitionID:   existing.AlarmDefinitionID,
		CreatedAt:           existing.CreatedAt,
		UpdatedAt:           existing.UpdatedAt,
	})
}
