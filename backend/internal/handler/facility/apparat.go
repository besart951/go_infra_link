package facility

import (
	"net/http"

	"github.com/besart951/go_infra_link/backend/internal/handler/dto"
	"github.com/gin-gonic/gin"
)

type ApparatHandler struct {
	service ApparatService
}

func NewApparatHandler(service ApparatService) *ApparatHandler {
	return &ApparatHandler{service: service}
}

// CreateApparat godoc
// @Summary Create a new apparat
// @Tags facility-apparats
// @Accept json
// @Produce json
// @Param apparat body dto.CreateApparatRequest true "Apparat data"
// @Success 201 {object} dto.ApparatResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/facility/apparats [post]
func (h *ApparatHandler) CreateApparat(c *gin.Context) {
	var req dto.CreateApparatRequest
	if !bindJSON(c, &req) {
		return
	}

	apparat := toApparatModel(req)

	if err := h.service.Create(apparat); respondValidationOrError(c, err, "creation_failed") {
		return
	}

	c.JSON(http.StatusCreated, toApparatResponse(*apparat))
}

// GetApparat godoc
// @Summary Get an apparat by ID
// @Tags facility-apparats
// @Produce json
// @Param id path string true "Apparat ID"
// @Success 200 {object} dto.ApparatResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/facility/apparats/{id} [get]
func (h *ApparatHandler) GetApparat(c *gin.Context) {
	id, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}

	apparat, err := h.service.GetByID(id)
	if err != nil {
		if respondNotFoundIf(c, err, "Apparat not found") {
			return
		}
		respondError(c, http.StatusInternalServerError, "fetch_failed", err.Error())
		return
	}

	c.JSON(http.StatusOK, toApparatResponse(*apparat))
}

// ListApparats godoc
// @Summary List apparats with pagination
// @Tags facility-apparats
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Param search query string false "Search query"
// @Success 200 {object} dto.ApparatListResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/facility/apparats [get]
func (h *ApparatHandler) ListApparats(c *gin.Context) {
	query, ok := parsePaginationQuery(c)
	if !ok {
		return
	}

	result, err := h.service.List(query.Page, query.Limit, query.Search)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "fetch_failed", err.Error())
		return
	}

	c.JSON(http.StatusOK, toApparatListResponse(result))
}

// UpdateApparat godoc
// @Summary Update an apparat
// @Tags facility-apparats
// @Accept json
// @Produce json
// @Param id path string true "Apparat ID"
// @Param apparat body dto.UpdateApparatRequest true "Apparat data"
// @Success 200 {object} dto.ApparatResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/facility/apparats/{id} [put]
func (h *ApparatHandler) UpdateApparat(c *gin.Context) {
	id, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}

	var req dto.UpdateApparatRequest
	if !bindJSON(c, &req) {
		return
	}

	apparat, err := h.service.GetByID(id)
	if err != nil {
		if respondNotFoundIf(c, err, "Apparat not found") {
			return
		}
		respondError(c, http.StatusInternalServerError, "fetch_failed", err.Error())
		return
	}

	applyApparatUpdate(apparat, req)

	if err := h.service.Update(apparat); respondValidationOrError(c, err, "update_failed") {
		return
	}

	c.JSON(http.StatusOK, toApparatResponse(*apparat))
}

// DeleteApparat godoc
// @Summary Delete an apparat
// @Tags facility-apparats
// @Produce json
// @Param id path string true "Apparat ID"
// @Success 204
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/facility/apparats/{id} [delete]
func (h *ApparatHandler) DeleteApparat(c *gin.Context) {
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
