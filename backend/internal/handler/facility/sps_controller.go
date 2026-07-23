package facility

import (
	"errors"
	"net/http"

	appspscontroller "github.com/besart951/go_infra_link/backend/internal/application/facility/spscontroller"
	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	dto "github.com/besart951/go_infra_link/backend/internal/handler/dto/facility"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type SPSControllerHandler struct {
	service SPSControllerService
	creator SPSControllerCreator
	cloner  SPSControllerCloner
	updater SPSControllerUpdater
	deleter SPSControllerDeleter
}

func NewSPSControllerHandler(
	service SPSControllerService,
	creator SPSControllerCreator,
	cloner SPSControllerCloner,
	updater SPSControllerUpdater,
	deleter SPSControllerDeleter,
) *SPSControllerHandler {
	return &SPSControllerHandler{
		service: service,
		creator: creator,
		cloner:  cloner,
		updater: updater,
		deleter: deleter,
	}
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

	if h.creator == nil {
		respondLocalizedError(c, http.StatusInternalServerError, "creation_failed", "facility.creation_failed")
		return
	}
	spsController, err := h.creator.Create(
		c.Request.Context(),
		toSPSControllerCreateCommand(req),
	)
	if err != nil {
		respondLocalizedDomainError(c, err, "creation_failed", "facility.creation_failed",
			localizedInvalidReference(),
		)
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

	spsController, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		if respondLocalizedNotFoundIf(c, err, "facility.sps_controller_not_found") {
			return
		}
		respondLocalizedError(c, http.StatusInternalServerError, "fetch_failed", "facility.fetch_failed")
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
		respondLocalizedInvalidArgument(c, "facility.ids_required")
		return
	}

	items, err := h.service.GetByIDs(c.Request.Context(), req.Ids)
	if err != nil {
		respondLocalizedError(c, http.StatusInternalServerError, "fetch_failed", "facility.fetch_failed")
		return
	}

	c.JSON(http.StatusOK, dto.SPSControllerBulkResponse{Items: toSPSControllerResponses(items)})
}

// CopySPSController godoc
// @Summary Copy an SPS controller
// @Tags facility-sps-controllers
// @Produce json
// @Param id path string true "SPS Controller ID"
// @Success 201 {object} dto.SPSControllerResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 409 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/facility/sps-controllers/{id}/copy [post]
func (h *SPSControllerHandler) CopySPSController(c *gin.Context) {
	id, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}

	if h.cloner == nil {
		respondLocalizedError(c, http.StatusInternalServerError, "creation_failed", "facility.creation_failed")
		return
	}
	copyEntity, err := h.cloner.Clone(
		c.Request.Context(),
		appspscontroller.CloneCommand{SourceSPSControllerID: id},
	)
	if err != nil {
		respondLocalizedDomainError(c, err, "creation_failed", "facility.creation_failed",
			localizedNotFound("facility.sps_controller_not_found"),
			localizedConflict("facility.update_failed"),
		)
		return
	}

	c.JSON(http.StatusCreated, toSPSControllerResponse(*copyEntity))
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

	ctx := c.Request.Context()

	var result *domain.PaginatedList[domainFacility.SPSController]
	var err error
	if controlCabinetID != nil {
		result, err = h.service.ListByControlCabinetID(ctx, *controlCabinetID, query.Page, query.Limit, query.Search)
	} else {
		result, err = h.service.List(ctx, query.Page, query.Limit, query.Search)
	}
	if err != nil {
		respondLocalizedError(c, http.StatusInternalServerError, "fetch_failed", "facility.fetch_failed")
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
		respondLocalizedInvalidArgument(c, "facility.control_cabinet_id_required")
		return
	}

	excludeID, ok := parseSPSControllerGAExcludeID(c)
	if !ok {
		return
	}

	gaDevice, err := h.service.NextAvailableGADevice(c.Request.Context(), *controlCabinetID, excludeID)
	if err != nil {
		respondLocalizedDomainError(c, err, "fetch_failed", "facility.fetch_failed",
			localizedNotFound("facility.control_cabinet_not_found"),
			localizedConflict("facility.no_available_ga_device"),
		)
		return
	}

	c.JSON(http.StatusOK, dto.NextAvailableGADeviceResponse{GADevice: gaDevice})
}

func parseSPSControllerGAExcludeID(c *gin.Context) (*uuid.UUID, bool) {
	excludeID, ok := parseUUIDQueryParam(c, "exclude_id")
	if !ok || excludeID != nil {
		return excludeID, ok
	}

	// `sps_controller_id` was emitted by the frontend before the backend query
	// contract was aligned. Keep it as a read-only compatibility alias while
	// new clients use the documented `exclude_id` parameter.
	return parseUUIDQueryParam(c, "sps_controller_id")
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

	if h.updater == nil {
		respondLocalizedError(c, http.StatusInternalServerError, "update_failed", "facility.update_failed")
		return
	}

	spsController, err := h.updater.Update(
		c.Request.Context(),
		toSPSControllerUpdateCommand(id, req),
	)
	if err != nil {
		var loadErr *appspscontroller.LoadError
		if errors.As(err, &loadErr) {
			if respondLocalizedNotFoundIf(c, loadErr.Err, "facility.sps_controller_not_found") {
				return
			}
			respondLocalizedError(c, http.StatusInternalServerError, "fetch_failed", "facility.fetch_failed")
			return
		}
		respondLocalizedDomainError(c, err, "update_failed", "facility.update_failed",
			localizedInvalidReference(),
		)
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

	if h.deleter == nil {
		respondLocalizedError(c, http.StatusInternalServerError, "deletion_failed", "facility.deletion_failed")
		return
	}
	if err := h.deleter.Delete(
		c.Request.Context(),
		appspscontroller.DeleteCommand{SPSControllerID: id},
	); err != nil {
		respondLocalizedDomainError(c, err, "deletion_failed", "facility.deletion_failed",
			localizedNotFound("facility.sps_controller_not_found"),
		)
		return
	}

	c.Status(http.StatusNoContent)
}
