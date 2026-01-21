package facility

import (
	"net/http"

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
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "validation_error",
			Message: err.Error(),
		})
		return
	}

	controlCabinet := &domainFacility.ControlCabinet{
		BuildingID:       req.BuildingID,
		ProjectID:        req.ProjectID,
		ControlCabinetNr: req.ControlCabinetNr,
	}

	if err := h.service.Create(controlCabinet); err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "creation_failed",
			Message: err.Error(),
		})
		return
	}

	response := dto.ControlCabinetResponse{
		ID:               controlCabinet.ID,
		BuildingID:       controlCabinet.BuildingID,
		ProjectID:        controlCabinet.ProjectID,
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
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid UUID format",
		})
		return
	}

	controlCabinet, err := h.service.GetByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "fetch_failed",
			Message: err.Error(),
		})
		return
	}

	if controlCabinet == nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Error:   "not_found",
			Message: "Control Cabinet not found",
		})
		return
	}

	response := dto.ControlCabinetResponse{
		ID:               controlCabinet.ID,
		BuildingID:       controlCabinet.BuildingID,
		ProjectID:        controlCabinet.ProjectID,
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

	items := make([]dto.ControlCabinetResponse, len(result.Items))
	for i, controlCabinet := range result.Items {
		items[i] = dto.ControlCabinetResponse{
			ID:               controlCabinet.ID,
			BuildingID:       controlCabinet.BuildingID,
			ProjectID:        controlCabinet.ProjectID,
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
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid UUID format",
		})
		return
	}

	var req dto.UpdateControlCabinetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "validation_error",
			Message: err.Error(),
		})
		return
	}

	controlCabinet, err := h.service.GetByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "fetch_failed",
			Message: err.Error(),
		})
		return
	}

	if controlCabinet == nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Error:   "not_found",
			Message: "Control Cabinet not found",
		})
		return
	}

	if req.BuildingID != uuid.Nil {
		controlCabinet.BuildingID = req.BuildingID
	}
	if req.ProjectID != nil {
		controlCabinet.ProjectID = req.ProjectID
	}
	if req.ControlCabinetNr != nil {
		controlCabinet.ControlCabinetNr = req.ControlCabinetNr
	}

	if err := h.service.Update(controlCabinet); err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "update_failed",
			Message: err.Error(),
		})
		return
	}

	response := dto.ControlCabinetResponse{
		ID:               controlCabinet.ID,
		BuildingID:       controlCabinet.BuildingID,
		ProjectID:        controlCabinet.ProjectID,
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
