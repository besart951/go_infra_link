package facility

import (
	"net/http"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/besart951/go_infra_link/backend/internal/handler/dto"
	facilityService "github.com/besart951/go_infra_link/backend/internal/service/facility"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type BuildingHandler struct {
	service *facilityService.Service
}

func NewBuildingHandler(service *facilityService.Service) *BuildingHandler {
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

	if err := h.service.Buildings.Create(building); err != nil {
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

	buildings, err := h.service.Buildings.GetByIds([]uuid.UUID{id})
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "fetch_failed",
			Message: err.Error(),
		})
		return
	}

	if len(buildings) == 0 {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Error:   "not_found",
			Message: "Building not found",
		})
		return
	}

	building := buildings[0]
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

	if query.Page == 0 {
		query.Page = 1
	}
	if query.Limit == 0 {
		query.Limit = 10
	}

	params := domain.PaginationParams{
		Page:   query.Page,
		Limit:  query.Limit,
		Search: query.Search,
	}

	result, err := h.service.Buildings.GetPaginatedList(params)
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

	buildings, err := h.service.Buildings.GetByIds([]uuid.UUID{id})
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "fetch_failed",
			Message: err.Error(),
		})
		return
	}

	if len(buildings) == 0 {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Error:   "not_found",
			Message: "Building not found",
		})
		return
	}

	building := buildings[0]
	if req.IWSCode != "" {
		building.IWSCode = req.IWSCode
	}
	if req.BuildingGroup != 0 {
		building.BuildingGroup = req.BuildingGroup
	}

	if err := h.service.Buildings.Update(building); err != nil {
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

	if err := h.service.Buildings.DeleteByIds([]uuid.UUID{id}); err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "deletion_failed",
			Message: err.Error(),
		})
		return
	}

	c.Status(http.StatusNoContent)
}
