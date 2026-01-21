package facility

import (
	"net/http"

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

	fieldDevice := &domainFacility.FieldDevice{
		BMK:                       req.BMK,
		Description:               req.Description,
		ApparatNr:                 req.ApparatNr,
		SPSControllerSystemTypeID: req.SPSControllerSystemTypeID,
		SystemPartID:              req.SystemPartID,
		SpecificationID:           req.SpecificationID,
		ProjectID:                 req.ProjectID,
		ApparatID:                 req.ApparatID,
	}

	if err := h.service.Create(fieldDevice); err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "creation_failed",
			Message: err.Error(),
		})
		return
	}

	response := dto.FieldDeviceResponse{
		ID:                        fieldDevice.ID,
		BMK:                       fieldDevice.BMK,
		Description:               fieldDevice.Description,
		ApparatNr:                 fieldDevice.ApparatNr,
		SPSControllerSystemTypeID: fieldDevice.SPSControllerSystemTypeID,
		SystemPartID:              fieldDevice.SystemPartID,
		SpecificationID:           fieldDevice.SpecificationID,
		ProjectID:                 fieldDevice.ProjectID,
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
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "fetch_failed",
			Message: err.Error(),
		})
		return
	}

	if fieldDevice == nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Error:   "not_found",
			Message: "Field Device not found",
		})
		return
	}

	response := dto.FieldDeviceResponse{
		ID:                        fieldDevice.ID,
		BMK:                       fieldDevice.BMK,
		Description:               fieldDevice.Description,
		ApparatNr:                 fieldDevice.ApparatNr,
		SPSControllerSystemTypeID: fieldDevice.SPSControllerSystemTypeID,
		SystemPartID:              fieldDevice.SystemPartID,
		SpecificationID:           fieldDevice.SpecificationID,
		ProjectID:                 fieldDevice.ProjectID,
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
		items[i] = dto.FieldDeviceResponse{
			ID:                        fieldDevice.ID,
			BMK:                       fieldDevice.BMK,
			Description:               fieldDevice.Description,
			ApparatNr:                 fieldDevice.ApparatNr,
			SPSControllerSystemTypeID: fieldDevice.SPSControllerSystemTypeID,
			SystemPartID:              fieldDevice.SystemPartID,
			SpecificationID:           fieldDevice.SpecificationID,
			ProjectID:                 fieldDevice.ProjectID,
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
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "fetch_failed",
			Message: err.Error(),
		})
		return
	}

	if fieldDevice == nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Error:   "not_found",
			Message: "Field Device not found",
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
		fieldDevice.ApparatNr = req.ApparatNr
	}
	if req.SPSControllerSystemTypeID != uuid.Nil {
		fieldDevice.SPSControllerSystemTypeID = req.SPSControllerSystemTypeID
	}
	if req.SystemPartID != nil {
		fieldDevice.SystemPartID = req.SystemPartID
	}
	if req.SpecificationID != nil {
		fieldDevice.SpecificationID = req.SpecificationID
	}
	if req.ProjectID != nil {
		fieldDevice.ProjectID = req.ProjectID
	}
	if req.ApparatID != uuid.Nil {
		fieldDevice.ApparatID = req.ApparatID
	}

	if err := h.service.Update(fieldDevice); err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "update_failed",
			Message: err.Error(),
		})
		return
	}

	response := dto.FieldDeviceResponse{
		ID:                        fieldDevice.ID,
		BMK:                       fieldDevice.BMK,
		Description:               fieldDevice.Description,
		ApparatNr:                 fieldDevice.ApparatNr,
		SPSControllerSystemTypeID: fieldDevice.SPSControllerSystemTypeID,
		SystemPartID:              fieldDevice.SystemPartID,
		SpecificationID:           fieldDevice.SpecificationID,
		ProjectID:                 fieldDevice.ProjectID,
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
