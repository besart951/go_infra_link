package facility

import (
	"errors"
	"net/http"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/besart951/go_infra_link/backend/internal/handler/dto"
	"github.com/gin-gonic/gin"
)

type ControlCabinetHandler struct {
	service ControlCabinetService
}

func NewControlCabinetHandler(service ControlCabinetService) *ControlCabinetHandler {
	return &ControlCabinetHandler{service: service}
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

	if err := h.service.Create(controlCabinet); err != nil {
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

	controlCabinet, err := h.service.GetByID(id)
	if err != nil {
		if respondNotFoundIf(c, err, "Control Cabinet not found") {
			return
		}
		respondError(c, http.StatusInternalServerError, "fetch_failed", err.Error())
		return
	}

	c.JSON(http.StatusOK, toControlCabinetResponse(*controlCabinet))
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

	var result *domain.PaginatedList[domainFacility.ControlCabinet]
	var err error
	if buildingID != nil {
		result, err = h.service.ListByBuildingID(*buildingID, query.Page, query.Limit, query.Search)
	} else {
		result, err = h.service.List(query.Page, query.Limit, query.Search)
	}
	if err != nil {
		respondError(c, http.StatusInternalServerError, "fetch_failed", err.Error())
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

	controlCabinet, err := h.service.GetByID(id)
	if err != nil {
		if respondNotFoundIf(c, err, "Control Cabinet not found") {
			return
		}
		respondError(c, http.StatusInternalServerError, "fetch_failed", err.Error())
		return
	}

	applyControlCabinetUpdate(controlCabinet, req)

	if err := h.service.Update(controlCabinet); err != nil {
		if ve, ok := domain.AsValidationError(err); ok {
			respondValidationError(c, ve.Fields)
			return
		}
		if errors.Is(err, domain.ErrNotFound) {
			respondInvalidReference(c)
			return
		}
		respondError(c, http.StatusInternalServerError, "update_failed", err.Error())
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

	if err := h.service.DeleteByID(id); err != nil {
		respondError(c, http.StatusInternalServerError, "deletion_failed", err.Error())
		return
	}

	c.Status(http.StatusNoContent)
}
