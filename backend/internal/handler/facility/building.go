package facility

import (
	"net/http"

	"github.com/besart951/go_infra_link/backend/internal/handler/dto"
	"github.com/gin-gonic/gin"
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
	if !bindJSON(c, &req) {
		return
	}

	building := toBuildingModel(req)

	if err := h.service.Create(building); respondValidationOrError(c, err, "creation_failed") {
		return
	}

	c.JSON(http.StatusCreated, toBuildingResponse(*building))
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
	id, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}

	building, err := h.service.GetByID(id)
	if err != nil {
		if respondNotFoundIf(c, err, "Building not found") {
			return
		}
		respondError(c, http.StatusInternalServerError, "fetch_failed", err.Error())
		return
	}

	c.JSON(http.StatusOK, toBuildingResponse(*building))
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
	query, ok := parsePaginationQuery(c)
	if !ok {
		return
	}

	result, err := h.service.List(query.Page, query.Limit, query.Search)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "fetch_failed", err.Error())
		return
	}

	c.JSON(http.StatusOK, toBuildingListResponse(result))
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
	id, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}

	var req dto.UpdateBuildingRequest
	if !bindJSON(c, &req) {
		return
	}

	building, err := h.service.GetByID(id)
	if err != nil {
		if respondNotFoundIf(c, err, "Building not found") {
			return
		}
		respondError(c, http.StatusInternalServerError, "fetch_failed", err.Error())
		return
	}

	applyBuildingUpdate(building, req)

	if err := h.service.Update(building); respondValidationOrError(c, err, "update_failed") {
		return
	}

	c.JSON(http.StatusOK, toBuildingResponse(*building))
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
