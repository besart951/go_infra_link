package facility

import (
	"errors"
	"net/http"

	appcontrolcabinet "github.com/besart951/go_infra_link/backend/internal/application/facility/controlcabinet"
	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	dto "github.com/besart951/go_infra_link/backend/internal/handler/dto/facility"
	"github.com/gin-gonic/gin"
)

type ControlCabinetHandler struct {
	service ControlCabinetService
	creator ControlCabinetCreator
	cloner  ControlCabinetCloner
	updater ControlCabinetUpdater
	deleter ControlCabinetDeleter
}

func NewControlCabinetHandler(
	service ControlCabinetService,
	creator ControlCabinetCreator,
	cloner ControlCabinetCloner,
	updater ControlCabinetUpdater,
	deleter ControlCabinetDeleter,
) *ControlCabinetHandler {
	return &ControlCabinetHandler{
		service: service,
		creator: creator,
		cloner:  cloner,
		updater: updater,
		deleter: deleter,
	}
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

	if h.creator == nil {
		respondLocalizedError(c, http.StatusInternalServerError, "creation_failed", "facility.creation_failed")
		return
	}
	controlCabinet, err := h.creator.Create(
		c.Request.Context(),
		toControlCabinetCreateCommand(req),
	)
	if err != nil {
		respondLocalizedDomainError(c, err, "creation_failed", "facility.creation_failed",
			localizedInvalidReference(),
		)
		return
	}

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
// @Success 201 {object} dto.ControlCabinetResponse
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

	if h.cloner == nil {
		respondLocalizedError(c, http.StatusInternalServerError, "creation_failed", "facility.creation_failed")
		return
	}
	copyEntity, err := h.cloner.Clone(c.Request.Context(), appcontrolcabinet.CloneCommand{
		SourceControlCabinetID: id,
	})
	if err != nil {
		respondLocalizedDomainError(c, err, "creation_failed", "facility.creation_failed",
			localizedNotFound("facility.control_cabinet_not_found"),
			localizedConflict("facility.update_failed"),
		)
		return
	}

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

	if h.updater == nil {
		respondLocalizedError(c, http.StatusInternalServerError, "update_failed", "facility.update_failed")
		return
	}

	controlCabinet, err := h.updater.Update(
		c.Request.Context(),
		toControlCabinetUpdateCommand(id, req),
	)
	if err != nil {
		var loadErr *appcontrolcabinet.LoadError
		if errors.As(err, &loadErr) {
			if respondLocalizedNotFoundIf(c, loadErr.Err, "facility.control_cabinet_not_found") {
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

	if h.deleter == nil {
		respondLocalizedError(c, http.StatusInternalServerError, "deletion_failed", "facility.deletion_failed")
		return
	}
	if err := h.deleter.Delete(c.Request.Context(), appcontrolcabinet.DeleteCommand{
		ControlCabinetID: id,
	}); err != nil {
		respondLocalizedDomainError(c, err, "deletion_failed", "facility.deletion_failed",
			localizedNotFound("facility.control_cabinet_not_found"),
		)
		return
	}

	c.Status(http.StatusNoContent)
}
