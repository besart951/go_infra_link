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

type SpecificationHandler struct {
	service SpecificationService
}

func NewSpecificationHandler(service SpecificationService) *SpecificationHandler {
	return &SpecificationHandler{service: service}
}

// CreateSpecification godoc
// @Summary Create a new specification
// @Tags facility-specifications
// @Accept json
// @Produce json
// @Param specification body dto.CreateSpecificationRequest true "Specification data"
// @Success 201 {object} dto.SpecificationResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/facility/specifications [post]
func (h *SpecificationHandler) CreateSpecification(c *gin.Context) {
	var req dto.CreateSpecificationRequest
	if !bindJSON(c, &req) {
		return
	}

	specification := &domainFacility.Specification{
		FieldDeviceID:                             &req.FieldDeviceID,
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

	if err := h.service.Create(specification); err != nil {
		if errors.Is(err, domain.ErrInvalidArgument) {
			respondError(c, http.StatusBadRequest, "validation_error", "field_device_id is required")
			return
		}
		if errors.Is(err, domain.ErrConflict) {
			respondError(c, http.StatusConflict, "conflict", "Specification already exists for this field device")
			return
		}
		respondError(c, http.StatusInternalServerError, "creation_failed", err.Error())
		return
	}

	response := dto.SpecificationResponse{
		ID:                       specification.ID,
		FieldDeviceID:            specification.FieldDeviceID,
		SpecificationSupplier:    specification.SpecificationSupplier,
		SpecificationBrand:       specification.SpecificationBrand,
		SpecificationType:        specification.SpecificationType,
		AdditionalInfoMotorValve: specification.AdditionalInfoMotorValve,
		AdditionalInfoSize:       specification.AdditionalInfoSize,
		AdditionalInformationInstallationLocation: specification.AdditionalInformationInstallationLocation,
		ElectricalConnectionPH:                    specification.ElectricalConnectionPH,
		ElectricalConnectionACDC:                  specification.ElectricalConnectionACDC,
		ElectricalConnectionAmperage:              specification.ElectricalConnectionAmperage,
		ElectricalConnectionPower:                 specification.ElectricalConnectionPower,
		ElectricalConnectionRotation:              specification.ElectricalConnectionRotation,
		CreatedAt:                                 specification.CreatedAt,
		UpdatedAt:                                 specification.UpdatedAt,
	}

	c.JSON(http.StatusCreated, response)
}

// GetSpecification godoc
// @Summary Get a specification by ID
// @Tags facility-specifications
// @Produce json
// @Param id path string true "Specification ID"
// @Success 200 {object} dto.SpecificationResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/facility/specifications/{id} [get]
func (h *SpecificationHandler) GetSpecification(c *gin.Context) {
	id, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}

	specification, err := h.service.GetByID(id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			respondNotFound(c, "Specification not found")
			return
		}
		respondError(c, http.StatusInternalServerError, "fetch_failed", err.Error())
		return
	}

	response := dto.SpecificationResponse{
		ID:                       specification.ID,
		FieldDeviceID:            specification.FieldDeviceID,
		SpecificationSupplier:    specification.SpecificationSupplier,
		SpecificationBrand:       specification.SpecificationBrand,
		SpecificationType:        specification.SpecificationType,
		AdditionalInfoMotorValve: specification.AdditionalInfoMotorValve,
		AdditionalInfoSize:       specification.AdditionalInfoSize,
		AdditionalInformationInstallationLocation: specification.AdditionalInformationInstallationLocation,
		ElectricalConnectionPH:                    specification.ElectricalConnectionPH,
		ElectricalConnectionACDC:                  specification.ElectricalConnectionACDC,
		ElectricalConnectionAmperage:              specification.ElectricalConnectionAmperage,
		ElectricalConnectionPower:                 specification.ElectricalConnectionPower,
		ElectricalConnectionRotation:              specification.ElectricalConnectionRotation,
		CreatedAt:                                 specification.CreatedAt,
		UpdatedAt:                                 specification.UpdatedAt,
	}

	c.JSON(http.StatusOK, response)
}

// ListSpecifications godoc
// @Summary List specifications with pagination
// @Tags facility-specifications
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Param search query string false "Search query"
// @Success 200 {object} dto.SpecificationListResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/facility/specifications [get]
func (h *SpecificationHandler) ListSpecifications(c *gin.Context) {
	var query dto.PaginationQuery
	if !bindQuery(c, &query) {
		return
	}

	result, err := h.service.List(query.Page, query.Limit, query.Search)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "fetch_failed", err.Error())
		return
	}

	items := make([]dto.SpecificationResponse, len(result.Items))
	for i, specification := range result.Items {
		items[i] = dto.SpecificationResponse{
			ID:                       specification.ID,
			FieldDeviceID:            specification.FieldDeviceID,
			SpecificationSupplier:    specification.SpecificationSupplier,
			SpecificationBrand:       specification.SpecificationBrand,
			SpecificationType:        specification.SpecificationType,
			AdditionalInfoMotorValve: specification.AdditionalInfoMotorValve,
			AdditionalInfoSize:       specification.AdditionalInfoSize,
			AdditionalInformationInstallationLocation: specification.AdditionalInformationInstallationLocation,
			ElectricalConnectionPH:                    specification.ElectricalConnectionPH,
			ElectricalConnectionACDC:                  specification.ElectricalConnectionACDC,
			ElectricalConnectionAmperage:              specification.ElectricalConnectionAmperage,
			ElectricalConnectionPower:                 specification.ElectricalConnectionPower,
			ElectricalConnectionRotation:              specification.ElectricalConnectionRotation,
			CreatedAt:                                 specification.CreatedAt,
			UpdatedAt:                                 specification.UpdatedAt,
		}
	}

	response := dto.SpecificationListResponse{
		Items:      items,
		Total:      result.Total,
		Page:       result.Page,
		TotalPages: result.TotalPages,
	}

	c.JSON(http.StatusOK, response)
}

// UpdateSpecification godoc
// @Summary Update a specification
// @Tags facility-specifications
// @Accept json
// @Produce json
// @Param id path string true "Specification ID"
// @Param specification body dto.UpdateSpecificationRequest true "Specification data"
// @Success 200 {object} dto.SpecificationResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/facility/specifications/{id} [put]
func (h *SpecificationHandler) UpdateSpecification(c *gin.Context) {
	id, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}

	var req dto.UpdateSpecificationRequest
	if !bindJSON(c, &req) {
		return
	}

	specification, err := h.service.GetByID(id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			respondNotFound(c, "Specification not found")
			return
		}
		respondError(c, http.StatusInternalServerError, "fetch_failed", err.Error())
		return
	}

	if req.SpecificationSupplier != nil {
		specification.SpecificationSupplier = req.SpecificationSupplier
	}
	if req.SpecificationBrand != nil {
		specification.SpecificationBrand = req.SpecificationBrand
	}
	if req.SpecificationType != nil {
		specification.SpecificationType = req.SpecificationType
	}
	if req.AdditionalInfoMotorValve != nil {
		specification.AdditionalInfoMotorValve = req.AdditionalInfoMotorValve
	}
	if req.AdditionalInfoSize != nil {
		specification.AdditionalInfoSize = req.AdditionalInfoSize
	}
	if req.AdditionalInformationInstallationLocation != nil {
		specification.AdditionalInformationInstallationLocation = req.AdditionalInformationInstallationLocation
	}
	if req.ElectricalConnectionPH != nil {
		specification.ElectricalConnectionPH = req.ElectricalConnectionPH
	}
	if req.ElectricalConnectionACDC != nil {
		specification.ElectricalConnectionACDC = req.ElectricalConnectionACDC
	}
	if req.ElectricalConnectionAmperage != nil {
		specification.ElectricalConnectionAmperage = req.ElectricalConnectionAmperage
	}
	if req.ElectricalConnectionPower != nil {
		specification.ElectricalConnectionPower = req.ElectricalConnectionPower
	}
	if req.ElectricalConnectionRotation != nil {
		specification.ElectricalConnectionRotation = req.ElectricalConnectionRotation
	}

	if err := h.service.Update(specification); err != nil {
		respondError(c, http.StatusInternalServerError, "update_failed", err.Error())
		return
	}

	response := dto.SpecificationResponse{
		ID:                       specification.ID,
		FieldDeviceID:            specification.FieldDeviceID,
		SpecificationSupplier:    specification.SpecificationSupplier,
		SpecificationBrand:       specification.SpecificationBrand,
		SpecificationType:        specification.SpecificationType,
		AdditionalInfoMotorValve: specification.AdditionalInfoMotorValve,
		AdditionalInfoSize:       specification.AdditionalInfoSize,
		AdditionalInformationInstallationLocation: specification.AdditionalInformationInstallationLocation,
		ElectricalConnectionPH:                    specification.ElectricalConnectionPH,
		ElectricalConnectionACDC:                  specification.ElectricalConnectionACDC,
		ElectricalConnectionAmperage:              specification.ElectricalConnectionAmperage,
		ElectricalConnectionPower:                 specification.ElectricalConnectionPower,
		ElectricalConnectionRotation:              specification.ElectricalConnectionRotation,
		CreatedAt:                                 specification.CreatedAt,
		UpdatedAt:                                 specification.UpdatedAt,
	}

	c.JSON(http.StatusOK, response)
}

// DeleteSpecification godoc
// @Summary Delete a specification
// @Tags facility-specifications
// @Produce json
// @Param id path string true "Specification ID"
// @Success 204
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/facility/specifications/{id} [delete]
func (h *SpecificationHandler) DeleteSpecification(c *gin.Context) {
	id, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}

	if err := h.service.DeleteByIds([]uuid.UUID{id}); err != nil {
		respondError(c, http.StatusInternalServerError, "deletion_failed", err.Error())
		return
	}

	c.Status(http.StatusNoContent)
}
