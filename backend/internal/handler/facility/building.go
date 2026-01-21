package facility

import (
	"net/http"

	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/besart951/go_infra_link/backend/internal/handler/dto"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type BuildingHandler struct {
	service BuildingService
}

func NewBuildingHandler(service BuildingService) *BuildingHandler {
	return &BuildingHandler{service: service}
}

// CreateBuilding godoc
// @Summary Create a new building
// @Tags facility-buildings
// @Accept json
// @Produce json
// @Param building body dto.CreateBuildingRequest true "Building data"
// @Success 201 {object} dto.BuildingResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/facility/buildings [post]
func (h *BuildingHandler) CreateBuilding(c *gin.Context) {
	var req dto.CreateBuildingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "validation_error",
			Message: err.Error(),
		})
		return
	}

	building := &domainFacility.Building{
		IWSCode:       req.IWSCode,
		BuildingGroup: req.BuildingGroup,
	}

	if err := h.service.Create(building); err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "creation_failed",
			Message: err.Error(),
		})
		return
	}

	response := dto.BuildingResponse{
		ID:            building.ID,
		IWSCode:       building.IWSCode,
		BuildingGroup: building.BuildingGroup,
		CreatedAt:     building.CreatedAt,
		UpdatedAt:     building.UpdatedAt,
	}

	c.JSON(http.StatusCreated, response)
}

// GetBuilding godoc
// @Summary Get a building by ID
// @Tags facility-buildings
// @Produce json
// @Param id path string true "Building ID"
// @Success 200 {object} dto.BuildingResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/facility/buildings/{id} [get]
func (h *BuildingHandler) GetBuilding(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid UUID format",
		})
		return
	}

	building, err := h.service.GetByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "fetch_failed",
			Message: err.Error(),
		})
		return
	}

	if building == nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Error:   "not_found",
			Message: "Building not found",
		})
		return
	}

	response := dto.BuildingResponse{
		ID:            building.ID,
		IWSCode:       building.IWSCode,
		BuildingGroup: building.BuildingGroup,
		CreatedAt:     building.CreatedAt,
		UpdatedAt:     building.UpdatedAt,
	}

	c.JSON(http.StatusOK, response)
}

// ListBuildings godoc
// @Summary List buildings with pagination
// @Tags facility-buildings
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Param search query string false "Search query"
// @Success 200 {object} dto.BuildingListResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/facility/buildings [get]
func (h *BuildingHandler) ListBuildings(c *gin.Context) {
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

	items := make([]dto.BuildingResponse, len(result.Items))
	for i, building := range result.Items {
		items[i] = dto.BuildingResponse{
			ID:            building.ID,
			IWSCode:       building.IWSCode,
			BuildingGroup: building.BuildingGroup,
			CreatedAt:     building.CreatedAt,
			UpdatedAt:     building.UpdatedAt,
		}
	}

	response := dto.BuildingListResponse{
		Items:      items,
		Total:      result.Total,
		Page:       result.Page,
		TotalPages: result.TotalPages,
	}

	c.JSON(http.StatusOK, response)
}

// UpdateBuilding godoc
// @Summary Update a building
// @Tags facility-buildings
// @Accept json
// @Produce json
// @Param id path string true "Building ID"
// @Param building body dto.UpdateBuildingRequest true "Building data"
// @Success 200 {object} dto.BuildingResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/facility/buildings/{id} [put]
func (h *BuildingHandler) UpdateBuilding(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid UUID format",
		})
		return
	}

	var req dto.UpdateBuildingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "validation_error",
			Message: err.Error(),
		})
		return
	}

	building, err := h.service.GetByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "fetch_failed",
			Message: err.Error(),
		})
		return
	}

	if building == nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Error:   "not_found",
			Message: "Building not found",
		})
		return
	}

	if req.IWSCode != "" {
		building.IWSCode = req.IWSCode
	}
	if req.BuildingGroup != 0 {
		building.BuildingGroup = req.BuildingGroup
	}

	if err := h.service.Update(building); err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "update_failed",
			Message: err.Error(),
		})
		return
	}

	response := dto.BuildingResponse{
		ID:            building.ID,
		IWSCode:       building.IWSCode,
		BuildingGroup: building.BuildingGroup,
		CreatedAt:     building.CreatedAt,
		UpdatedAt:     building.UpdatedAt,
	}

	c.JSON(http.StatusOK, response)
}

// DeleteBuilding godoc
// @Summary Delete a building
// @Tags facility-buildings
// @Produce json
// @Param id path string true "Building ID"
// @Success 204
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/facility/buildings/{id} [delete]
func (h *BuildingHandler) DeleteBuilding(c *gin.Context) {
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
