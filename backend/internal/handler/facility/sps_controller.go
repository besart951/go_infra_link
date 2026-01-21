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

type SPSControllerHandler struct {
	service SPSControllerService
}

func NewSPSControllerHandler(service SPSControllerService) *SPSControllerHandler {
	return &SPSControllerHandler{service: service}
}

// CreateSPSController godoc
// @Summary Create a new SPS controller
// @Tags facility-sps-controllers
// @Accept json
// @Produce json
// @Param sps_controller body dto.CreateSPSControllerRequest true "SPS Controller data"
// @Success 201 {object} dto.SPSControllerResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/facility/sps-controllers [post]
func (h *SPSControllerHandler) CreateSPSController(c *gin.Context) {
	var req dto.CreateSPSControllerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "validation_error",
			Message: err.Error(),
		})
		return
	}

	spsController := &domainFacility.SPSController{
		ControlCabinetID:  req.ControlCabinetID,
		ProjectID:         req.ProjectID,
		GADevice:          req.GADevice,
		DeviceName:        req.DeviceName,
		DeviceDescription: req.DeviceDescription,
		DeviceLocation:    req.DeviceLocation,
		IPAddress:         req.IPAddress,
		Subnet:            req.Subnet,
		Gateway:           req.Gateway,
		Vlan:              req.Vlan,
	}

	systemTypes := make([]domainFacility.SPSControllerSystemType, 0, len(req.SystemTypes))
	for _, st := range req.SystemTypes {
		systemTypes = append(systemTypes, domainFacility.SPSControllerSystemType{
			SystemTypeID: st.SystemTypeID,
			Number:       st.Number,
			DocumentName: st.DocumentName,
		})
	}

	if err := h.service.CreateWithSystemTypes(spsController, systemTypes); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{
				Error:   "invalid_reference",
				Message: "Referenced entity not found or deleted",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "creation_failed",
			Message: err.Error(),
		})
		return
	}

	response := dto.SPSControllerResponse{
		ID:                spsController.ID,
		ControlCabinetID:  spsController.ControlCabinetID,
		ProjectID:         spsController.ProjectID,
		GADevice:          spsController.GADevice,
		DeviceName:        spsController.DeviceName,
		DeviceDescription: spsController.DeviceDescription,
		DeviceLocation:    spsController.DeviceLocation,
		IPAddress:         spsController.IPAddress,
		Subnet:            spsController.Subnet,
		Gateway:           spsController.Gateway,
		Vlan:              spsController.Vlan,
		CreatedAt:         spsController.CreatedAt,
		UpdatedAt:         spsController.UpdatedAt,
	}

	c.JSON(http.StatusCreated, response)
}

// GetSPSController godoc
// @Summary Get an SPS controller by ID
// @Tags facility-sps-controllers
// @Produce json
// @Param id path string true "SPS Controller ID"
// @Success 200 {object} dto.SPSControllerResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/facility/sps-controllers/{id} [get]
func (h *SPSControllerHandler) GetSPSController(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid UUID format",
		})
		return
	}

	spsController, err := h.service.GetByID(id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			c.JSON(http.StatusNotFound, dto.ErrorResponse{
				Error:   "not_found",
				Message: "SPS Controller not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "fetch_failed",
			Message: err.Error(),
		})
		return
	}

	response := dto.SPSControllerResponse{
		ID:                spsController.ID,
		ControlCabinetID:  spsController.ControlCabinetID,
		ProjectID:         spsController.ProjectID,
		GADevice:          spsController.GADevice,
		DeviceName:        spsController.DeviceName,
		DeviceDescription: spsController.DeviceDescription,
		DeviceLocation:    spsController.DeviceLocation,
		IPAddress:         spsController.IPAddress,
		Subnet:            spsController.Subnet,
		Gateway:           spsController.Gateway,
		Vlan:              spsController.Vlan,
		CreatedAt:         spsController.CreatedAt,
		UpdatedAt:         spsController.UpdatedAt,
	}

	c.JSON(http.StatusOK, response)
}

// ListSPSControllers godoc
// @Summary List SPS controllers with pagination
// @Tags facility-sps-controllers
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Param search query string false "Search query"
// @Success 200 {object} dto.SPSControllerListResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/facility/sps-controllers [get]
func (h *SPSControllerHandler) ListSPSControllers(c *gin.Context) {
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

	items := make([]dto.SPSControllerResponse, len(result.Items))
	for i, spsController := range result.Items {
		items[i] = dto.SPSControllerResponse{
			ID:                spsController.ID,
			ControlCabinetID:  spsController.ControlCabinetID,
			ProjectID:         spsController.ProjectID,
			GADevice:          spsController.GADevice,
			DeviceName:        spsController.DeviceName,
			DeviceDescription: spsController.DeviceDescription,
			DeviceLocation:    spsController.DeviceLocation,
			IPAddress:         spsController.IPAddress,
			Subnet:            spsController.Subnet,
			Gateway:           spsController.Gateway,
			Vlan:              spsController.Vlan,
			CreatedAt:         spsController.CreatedAt,
			UpdatedAt:         spsController.UpdatedAt,
		}
	}

	response := dto.SPSControllerListResponse{
		Items:      items,
		Total:      result.Total,
		Page:       result.Page,
		TotalPages: result.TotalPages,
	}

	c.JSON(http.StatusOK, response)
}

// UpdateSPSController godoc
// @Summary Update an SPS controller
// @Tags facility-sps-controllers
// @Accept json
// @Produce json
// @Param id path string true "SPS Controller ID"
// @Param sps_controller body dto.UpdateSPSControllerRequest true "SPS Controller data"
// @Success 200 {object} dto.SPSControllerResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/facility/sps-controllers/{id} [put]
func (h *SPSControllerHandler) UpdateSPSController(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid UUID format",
		})
		return
	}

	var req dto.UpdateSPSControllerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "validation_error",
			Message: err.Error(),
		})
		return
	}

	spsController, err := h.service.GetByID(id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			c.JSON(http.StatusNotFound, dto.ErrorResponse{
				Error:   "not_found",
				Message: "SPS Controller not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "fetch_failed",
			Message: err.Error(),
		})
		return
	}

	if req.ControlCabinetID != uuid.Nil {
		spsController.ControlCabinetID = req.ControlCabinetID
	}
	if req.ProjectID != nil {
		spsController.ProjectID = req.ProjectID
	}
	if req.GADevice != nil {
		spsController.GADevice = req.GADevice
	}
	if req.DeviceName != "" {
		spsController.DeviceName = req.DeviceName
	}
	if req.DeviceDescription != nil {
		spsController.DeviceDescription = req.DeviceDescription
	}
	if req.DeviceLocation != nil {
		spsController.DeviceLocation = req.DeviceLocation
	}
	if req.IPAddress != nil {
		spsController.IPAddress = req.IPAddress
	}
	if req.Subnet != nil {
		spsController.Subnet = req.Subnet
	}
	if req.Gateway != nil {
		spsController.Gateway = req.Gateway
	}
	if req.Vlan != nil {
		spsController.Vlan = req.Vlan
	}

	var updateErr error
	if req.SystemTypes != nil {
		systemTypes := make([]domainFacility.SPSControllerSystemType, 0, len(*req.SystemTypes))
		for _, st := range *req.SystemTypes {
			systemTypes = append(systemTypes, domainFacility.SPSControllerSystemType{
				SystemTypeID: st.SystemTypeID,
				Number:       st.Number,
				DocumentName: st.DocumentName,
			})
		}
		updateErr = h.service.UpdateWithSystemTypes(spsController, systemTypes)
	} else {
		updateErr = h.service.Update(spsController)
	}
	if updateErr != nil {
		if errors.Is(updateErr, domain.ErrNotFound) {
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{
				Error:   "invalid_reference",
				Message: "Referenced entity not found or deleted",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "update_failed",
			Message: updateErr.Error(),
		})
		return
	}

	response := dto.SPSControllerResponse{
		ID:                spsController.ID,
		ControlCabinetID:  spsController.ControlCabinetID,
		ProjectID:         spsController.ProjectID,
		GADevice:          spsController.GADevice,
		DeviceName:        spsController.DeviceName,
		DeviceDescription: spsController.DeviceDescription,
		DeviceLocation:    spsController.DeviceLocation,
		IPAddress:         spsController.IPAddress,
		Subnet:            spsController.Subnet,
		Gateway:           spsController.Gateway,
		Vlan:              spsController.Vlan,
		CreatedAt:         spsController.CreatedAt,
		UpdatedAt:         spsController.UpdatedAt,
	}

	c.JSON(http.StatusOK, response)
}

// DeleteSPSController godoc
// @Summary Delete an SPS controller
// @Tags facility-sps-controllers
// @Produce json
// @Param id path string true "SPS Controller ID"
// @Success 204
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/facility/sps-controllers/{id} [delete]
func (h *SPSControllerHandler) DeleteSPSController(c *gin.Context) {
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
