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

	controlCabinet := &domainFacility.ControlCabinet{
		BuildingID:       req.BuildingID,
		ControlCabinetNr: req.ControlCabinetNr,
	}

	if err := h.service.Create(controlCabinet); err != nil {
		if ve, ok := domain.AsValidationError(err); ok {
			respondValidationError(c, ve.Fields)
			return
		}
		if errors.Is(err, domain.ErrNotFound) {
			respondError(c, http.StatusBadRequest, "invalid_reference", "Building not found or deleted")
			return
		}
		respondError(c, http.StatusInternalServerError, "creation_failed", err.Error())
		return
	}

	response := dto.ControlCabinetResponse{
		ID:               controlCabinet.ID,
		BuildingID:       controlCabinet.BuildingID,
		ControlCabinetNr: controlCabinet.ControlCabinetNr,
		CreatedAt:        controlCabinet.CreatedAt,
		UpdatedAt:        controlCabinet.UpdatedAt,
	}

	c.JSON(http.StatusCreated, response)
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
		if errors.Is(err, domain.ErrNotFound) {
			respondNotFound(c, "Control Cabinet not found")
			return
		}
		respondError(c, http.StatusInternalServerError, "fetch_failed", err.Error())
		return
	}

	response := dto.ControlCabinetResponse{
		ID:               controlCabinet.ID,
		BuildingID:       controlCabinet.BuildingID,
		ControlCabinetNr: controlCabinet.ControlCabinetNr,
		CreatedAt:        controlCabinet.CreatedAt,
		UpdatedAt:        controlCabinet.UpdatedAt,
	}

	c.JSON(http.StatusOK, response)
}

// ListControlCabinets godoc
// @Summary List control cabinets with pagination
// @Tags facility-control-cabinets
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Param search query string false "Search query"
// @Success 200 {object} dto.ControlCabinetListResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/facility/control-cabinets [get]
func (h *ControlCabinetHandler) ListControlCabinets(c *gin.Context) {
	var query dto.PaginationQuery
	if !bindQuery(c, &query) {
		return
	}

	result, err := h.service.List(query.Page, query.Limit, query.Search)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "fetch_failed", err.Error())
		return
	}

	items := make([]dto.ControlCabinetResponse, len(result.Items))
	for i, controlCabinet := range result.Items {
		items[i] = dto.ControlCabinetResponse{
			ID:               controlCabinet.ID,
			BuildingID:       controlCabinet.BuildingID,
			ControlCabinetNr: controlCabinet.ControlCabinetNr,
			CreatedAt:        controlCabinet.CreatedAt,
			UpdatedAt:        controlCabinet.UpdatedAt,
		}
	}

	response := dto.ControlCabinetListResponse{
		Items:      items,
		Total:      result.Total,
		Page:       result.Page,
		TotalPages: result.TotalPages,
	}

	c.JSON(http.StatusOK, response)
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
		if errors.Is(err, domain.ErrNotFound) {
			respondNotFound(c, "Control Cabinet not found")
			return
		}
		respondError(c, http.StatusInternalServerError, "fetch_failed", err.Error())
		return
	}

	if req.BuildingID != uuid.Nil {
		controlCabinet.BuildingID = req.BuildingID
	}
	if req.ControlCabinetNr != nil {
		controlCabinet.ControlCabinetNr = req.ControlCabinetNr
	}

	if err := h.service.Update(controlCabinet); err != nil {
		if ve, ok := domain.AsValidationError(err); ok {
			respondValidationError(c, ve.Fields)
			return
		}
		if errors.Is(err, domain.ErrNotFound) {
			respondError(c, http.StatusBadRequest, "invalid_reference", "Building not found or deleted")
			return
		}
		respondError(c, http.StatusInternalServerError, "update_failed", err.Error())
		return
	}

	response := dto.ControlCabinetResponse{
		ID:               controlCabinet.ID,
		BuildingID:       controlCabinet.BuildingID,
		ControlCabinetNr: controlCabinet.ControlCabinetNr,
		CreatedAt:        controlCabinet.CreatedAt,
		UpdatedAt:        controlCabinet.UpdatedAt,
	}

	c.JSON(http.StatusOK, response)
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

	if err := h.service.DeleteByIds([]uuid.UUID{id}); err != nil {
		respondError(c, http.StatusInternalServerError, "deletion_failed", err.Error())
		return
	}

	c.Status(http.StatusNoContent)
}
