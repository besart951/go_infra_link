package facility

import (
	"errors"
	"net/http"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/besart951/go_infra_link/backend/internal/handler/dto"
	"github.com/gin-gonic/gin"
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
	if !bindJSON(c, &req) {
		return
	}

	spsController := toSPSControllerModel(req)
	systemTypes := toSPSControllerSystemTypes(req.SystemTypes)

	if err := h.service.CreateWithSystemTypes(spsController, systemTypes); err != nil {
		if ve, ok := domain.AsValidationError(err); ok {
			respondValidationError(c, ve.Fields)
			return
		}
		if errors.Is(err, domain.ErrNotFound) {
			respondInvalidReference(c)
			return
		}
		respondError(c, http.StatusInternalServerError, "creation_failed", err.Error())
		return
	}

	c.JSON(http.StatusCreated, toSPSControllerResponse(*spsController))
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
	id, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}

	spsController, err := h.service.GetByID(id)
	if err != nil {
		if respondNotFoundIf(c, err, "SPS Controller not found") {
			return
		}
		respondError(c, http.StatusInternalServerError, "fetch_failed", err.Error())
		return
	}

	c.JSON(http.StatusOK, toSPSControllerResponse(*spsController))
}

// GetSPSControllersByIDs godoc
// @Summary Get multiple SPS controllers by IDs
// @Tags facility-sps-controllers
// @Accept json
// @Produce json
// @Param request body dto.SPSControllerBulkRequest true "SPS Controller IDs"
// @Success 200 {object} dto.SPSControllerBulkResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/facility/sps-controllers/bulk [post]
func (h *SPSControllerHandler) GetSPSControllersByIDs(c *gin.Context) {
	var req dto.SPSControllerBulkRequest
	if !bindJSON(c, &req) {
		return
	}
	if len(req.Ids) == 0 {
		respondInvalidArgument(c, "ids is required")
		return
	}

	items, err := h.service.GetByIDs(req.Ids)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "fetch_failed", err.Error())
		return
	}

	c.JSON(http.StatusOK, dto.SPSControllerBulkResponse{Items: toSPSControllerResponses(items)})
}

// ListSPSControllers godoc
// @Summary List SPS controllers with pagination
// @Tags facility-sps-controllers
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Param search query string false "Search query"
// @Param control_cabinet_id query string false "Control Cabinet ID"
// @Success 200 {object} dto.SPSControllerListResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/facility/sps-controllers [get]
func (h *SPSControllerHandler) ListSPSControllers(c *gin.Context) {
	query, ok := parsePaginationQuery(c)
	if !ok {
		return
	}

	controlCabinetID, ok := parseUUIDQueryParam(c, "control_cabinet_id")
	if !ok {
		return
	}

	var result *domain.PaginatedList[domainFacility.SPSController]
	var err error
	if controlCabinetID != nil {
		result, err = h.service.ListByControlCabinetID(*controlCabinetID, query.Page, query.Limit, query.Search)
	} else {
		result, err = h.service.List(query.Page, query.Limit, query.Search)
	}
	if err != nil {
		respondError(c, http.StatusInternalServerError, "fetch_failed", err.Error())
		return
	}

	c.JSON(http.StatusOK, toSPSControllerListResponse(result))
}

// GetNextAvailableGADevice godoc
// @Summary Suggest next available GA device for a control cabinet
// @Tags facility-sps-controllers
// @Produce json
// @Param control_cabinet_id query string true "Control Cabinet ID"
// @Param exclude_id query string false "SPS Controller ID to exclude"
// @Success 200 {object} dto.NextAvailableGADeviceResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 409 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/facility/sps-controllers/next-ga-device [get]
func (h *SPSControllerHandler) GetNextAvailableGADevice(c *gin.Context) {
	controlCabinetID, ok := parseUUIDQueryParam(c, "control_cabinet_id")
	if !ok {
		return
	}
	if controlCabinetID == nil {
		respondInvalidArgument(c, "control_cabinet_id is required")
		return
	}

	excludeID, ok := parseUUIDQueryParam(c, "exclude_id")
	if !ok {
		return
	}

	gaDevice, err := h.service.NextAvailableGADevice(*controlCabinetID, excludeID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			respondNotFound(c, "Control cabinet not found")
			return
		}
		if errors.Is(err, domain.ErrConflict) {
			respondConflict(c, "No available GA device")
			return
		}
		respondError(c, http.StatusInternalServerError, "fetch_failed", err.Error())
		return
	}

	c.JSON(http.StatusOK, dto.NextAvailableGADeviceResponse{GADevice: gaDevice})
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
	id, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}

	var req dto.UpdateSPSControllerRequest
	if !bindJSON(c, &req) {
		return
	}

	spsController, err := h.service.GetByID(id)
	if err != nil {
		if respondNotFoundIf(c, err, "SPS Controller not found") {
			return
		}
		respondError(c, http.StatusInternalServerError, "fetch_failed", err.Error())
		return
	}

	applySPSControllerUpdate(spsController, req)

	var updateErr error
	if req.SystemTypes != nil {
		systemTypes := toSPSControllerSystemTypes(*req.SystemTypes)
		updateErr = h.service.UpdateWithSystemTypes(spsController, systemTypes)
	} else {
		updateErr = h.service.Update(spsController)
	}
	if updateErr != nil {
		if ve, ok := domain.AsValidationError(updateErr); ok {
			respondValidationError(c, ve.Fields)
			return
		}
		if errors.Is(updateErr, domain.ErrNotFound) {
			respondInvalidReference(c)
			return
		}
		respondError(c, http.StatusInternalServerError, "update_failed", updateErr.Error())
		return
	}

	c.JSON(http.StatusOK, toSPSControllerResponse(*spsController))
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
