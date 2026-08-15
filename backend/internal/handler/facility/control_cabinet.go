package facility

import (
	"context"
	"net/http"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	dto "github.com/besart951/go_infra_link/backend/internal/handler/dto/facility"
	"github.com/besart951/go_infra_link/backend/internal/handler/middleware"
	facilityservice "github.com/besart951/go_infra_link/backend/internal/service/facility"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ControlCabinetHandler struct {
	service       ControlCabinetService
	collaboration ProjectRefreshBroadcaster
	copyJobs      *facilityservice.CopyJobManager
}

func NewControlCabinetHandler(service ControlCabinetService, collaboration ProjectRefreshBroadcaster, copyJobs ...*facilityservice.CopyJobManager) *ControlCabinetHandler {
	handler := &ControlCabinetHandler{service: service, collaboration: collaboration}
	if len(copyJobs) > 0 {
		handler.copyJobs = copyJobs[0]
	}
	return handler
}

func (h *ControlCabinetHandler) broadcastProjectRefresh(ctx context.Context, actorID *uuid.UUID, controlCabinetID uuid.UUID) {
	if h.collaboration == nil {
		return
	}
	h.collaboration.BroadcastRefreshForControlCabinet(ctx, actorID, controlCabinetID, "control_cabinet")
}

func (h *ControlCabinetHandler) broadcastProjectDelta(ctx context.Context, actorID *uuid.UUID, controlCabinet *domainFacility.ControlCabinet) {
	if h.collaboration == nil || controlCabinet == nil {
		return
	}

	h.collaboration.BroadcastControlCabinetDelta(ctx, actorID, *controlCabinet)
}

// CreateControlCabinet godoc
// @Summary Create a new control cabinet
// @Tags facility-control-cabinets
// @Accept json
// @Produce json
// @Param control_cabinet body dto.CreateControlCabinetRequest true "Control Cabinet data"
// @Success 201 {object} dto.ControlCabinetResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/facility/control-cabinets [post]
func (h *ControlCabinetHandler) CreateControlCabinet(c *gin.Context) {
	var req dto.CreateControlCabinetRequest
	if !bindJSON(c, &req) {
		return
	}

	controlCabinet := toControlCabinetModel(req)

	if err := h.service.Create(c.Request.Context(), controlCabinet); err != nil {
		respondLocalizedDomainError(c, err, "creation_failed", "facility.creation_failed",
			localizedInvalidReference(),
		)
		return
	}

	h.broadcastProjectDelta(c.Request.Context(), currentActorID(c), controlCabinet)
	c.JSON(http.StatusCreated, toControlCabinetResponse(*controlCabinet))
}

// GetControlCabinet godoc
// @Summary Get a control cabinet by ID
// @Tags facility-control-cabinets
// @Produce json
// @Param id path string true "Control Cabinet ID"
// @Success 200 {object} dto.ControlCabinetResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/facility/control-cabinets/{id} [get]
func (h *ControlCabinetHandler) GetControlCabinet(c *gin.Context) {
	id, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}

	controlCabinet, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		if respondLocalizedNotFoundIf(c, err, "facility.control_cabinet_not_found") {
			return
		}
		respondLocalizedError(c, http.StatusInternalServerError, "fetch_failed", "facility.fetch_failed")
		return
	}

	c.JSON(http.StatusOK, toControlCabinetResponse(*controlCabinet))
}

// GetControlCabinetsByIDs godoc
// @Summary Get multiple control cabinets by IDs
// @Tags facility-control-cabinets
// @Accept json
// @Produce json
// @Param request body dto.ControlCabinetBulkRequest true "Control Cabinet IDs"
// @Success 200 {object} dto.ControlCabinetBulkResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/facility/control-cabinets/bulk [post]
func (h *ControlCabinetHandler) GetControlCabinetsByIDs(c *gin.Context) {
	var req dto.ControlCabinetBulkRequest
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

	c.JSON(http.StatusOK, dto.ControlCabinetBulkResponse{Items: toControlCabinetResponses(items)})
}

// CopyControlCabinet godoc
// @Summary Copy a control cabinet
// @Tags facility-control-cabinets
// @Produce json
// @Param id path string true "Control Cabinet ID"
// @Param X-Copy-Operation-ID header string false "Client-generated copy operation UUID"
// @Success 202 {object} dto.CopyJobResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 409 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/facility/control-cabinets/{id}/copy [post]
func (h *ControlCabinetHandler) CopyControlCabinet(c *gin.Context) {
	id, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}
	if h.copyJobs != nil {
		operationID, ok := copyOperationID(c)
		if !ok {
			return
		}
		actorID, ok := middleware.GetUserID(c)
		if !ok {
			respondLocalizedError(c, http.StatusUnauthorized, "unauthorized", "errors.unauthorized")
			return
		}
		job := h.copyJobs.Start(actorID, operationID, facilityservice.CopyJobKindControlCabinet, func(ctx context.Context) error {
			copyEntity, err := h.service.CopyByID(ctx, id)
			if err == nil {
				h.broadcastProjectDelta(ctx, &actorID, copyEntity)
			}
			return err
		})
		c.JSON(http.StatusAccepted, toCopyJobResponse(job))
		return
	}

	copyEntity, err := h.service.CopyByID(c.Request.Context(), id)
	if err != nil {
		respondLocalizedDomainError(c, err, "creation_failed", "facility.creation_failed",
			localizedNotFound("facility.control_cabinet_not_found"),
			localizedConflict("facility.update_failed"),
		)
		return
	}

	h.broadcastProjectDelta(c.Request.Context(), currentActorID(c), copyEntity)
	c.JSON(http.StatusCreated, toControlCabinetResponse(*copyEntity))
}

// GetControlCabinetDeleteImpact godoc
// @Summary Preview delete impact for a control cabinet
// @Tags facility-control-cabinets
// @Produce json
// @Param id path string true "Control Cabinet ID"
// @Success 200 {object} dto.ControlCabinetDeleteImpactResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/facility/control-cabinets/{id}/delete-impact [get]
func (h *ControlCabinetHandler) GetControlCabinetDeleteImpact(c *gin.Context) {
	id, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}

	impact, err := h.service.GetDeleteImpact(c.Request.Context(), id)
	if err != nil {
		if respondLocalizedNotFoundIf(c, err, "facility.control_cabinet_not_found") {
			return
		}
		respondLocalizedError(c, http.StatusInternalServerError, "fetch_failed", "facility.fetch_failed")
		return
	}

	c.JSON(http.StatusOK, dto.ControlCabinetDeleteImpactResponse{
		ControlCabinetID:              impact.ControlCabinetID,
		SPSControllersCount:           impact.SPSControllersCount,
		SPSControllerSystemTypesCount: impact.SPSControllerSystemTypesCount,
		FieldDevicesCount:             impact.FieldDevicesCount,
		BacnetObjectsCount:            impact.BacnetObjectsCount,
		SpecificationsCount:           impact.SpecificationsCount,
	})
}

// ListControlCabinets godoc
// @Summary List control cabinets with pagination
// @Tags facility-control-cabinets
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Param search query string false "Search query"
// @Param building_id query string false "Building ID"
// @Success 200 {object} dto.ControlCabinetListResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/facility/control-cabinets [get]
func (h *ControlCabinetHandler) ListControlCabinets(c *gin.Context) {
	query, ok := parsePaginationQuery(c)
	if !ok {
		return
	}

	buildingID, ok := parseUUIDQueryParam(c, "building_id")
	if !ok {
		return
	}

	ctx := c.Request.Context()

	var result *domain.PaginatedList[domainFacility.ControlCabinet]
	var err error
	if buildingID != nil {
		result, err = h.service.ListByBuildingID(ctx, *buildingID, query.Page, query.Limit, query.Search)
	} else {
		result, err = h.service.List(ctx, query.Page, query.Limit, query.Search)
	}
	if err != nil {
		respondLocalizedError(c, http.StatusInternalServerError, "fetch_failed", "facility.fetch_failed")
		return
	}

	c.JSON(http.StatusOK, toControlCabinetListResponse(result))
}

// UpdateControlCabinet godoc
// @Summary Update a control cabinet
// @Tags facility-control-cabinets
// @Accept json
// @Produce json
// @Param id path string true "Control Cabinet ID"
// @Param control_cabinet body dto.UpdateControlCabinetRequest true "Control Cabinet data"
// @Success 200 {object} dto.ControlCabinetResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/facility/control-cabinets/{id} [put]
func (h *ControlCabinetHandler) UpdateControlCabinet(c *gin.Context) {
	id, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}

	var req dto.UpdateControlCabinetRequest
	if !bindJSON(c, &req) {
		return
	}

	ctx := c.Request.Context()

	controlCabinet, err := h.service.GetByID(ctx, id)
	if err != nil {
		if respondLocalizedNotFoundIf(c, err, "facility.control_cabinet_not_found") {
			return
		}
		respondLocalizedError(c, http.StatusInternalServerError, "fetch_failed", "facility.fetch_failed")
		return
	}
	baseVersion := controlCabinet.Version
	if req.BaseVersion != nil {
		baseVersion = *req.BaseVersion
	}

	applyControlCabinetUpdate(controlCabinet, req)

	if err := h.service.Update(ctx, controlCabinet); err != nil {
		if current, getErr := h.service.GetByID(ctx, id); getErr == nil && respondWriteConflict(c, err, "control_cabinet", id, baseVersion, controlCabinetUpdatePaths(req), current.Version, toControlCabinetResponse(*current)) {
			return
		}
		respondLocalizedDomainError(c, err, "update_failed", "facility.update_failed",
			localizedInvalidReference(),
		)
		return
	}

	h.broadcastProjectDelta(ctx, currentActorID(c), controlCabinet)
	c.JSON(http.StatusOK, toControlCabinetResponse(*controlCabinet))
}

// DeleteControlCabinet godoc
// @Summary Delete a control cabinet
// @Tags facility-control-cabinets
// @Produce json
// @Param id path string true "Control Cabinet ID"
// @Success 204
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/facility/control-cabinets/{id} [delete]
func (h *ControlCabinetHandler) DeleteControlCabinet(c *gin.Context) {
	id, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}

	ctx := c.Request.Context()
	projectIDs := captureDeleteAudience(ctx, h.collaboration, "control_cabinet", id)
	if err := h.service.DeleteByID(ctx, id); err != nil {
		respondLocalizedDomainError(c, err, "deletion_failed", "facility.deletion_failed",
			localizedNotFound("facility.control_cabinet_not_found"),
		)
		return
	}

	if !broadcastCapturedDelete(ctx, h.collaboration, currentActorID(c), projectIDs, "control_cabinet", id) {
		h.broadcastProjectRefresh(ctx, currentActorID(c), id)
	}
	c.Status(http.StatusNoContent)
}
