package facility

import (
	"net/http"

	"github.com/besart951/go_infra_link/backend/internal/handler/dto"
	"github.com/gin-gonic/gin"
)

type ObjectDataHandler struct {
	service ObjectDataService
}

func NewObjectDataHandler(service ObjectDataService) *ObjectDataHandler {
	return &ObjectDataHandler{service: service}
}

// CreateObjectData godoc
// @Summary Create object data
// @Tags facility-object-data
// @Accept json
// @Produce json
// @Param object_data body dto.CreateObjectDataRequest true "Object Data"
// @Success 201 {object} dto.ObjectDataResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/facility/object-data [post]
func (h *ObjectDataHandler) CreateObjectData(c *gin.Context) {
	var req dto.CreateObjectDataRequest
	if !bindJSON(c, &req) {
		return
	}

	obj := toObjectDataModel(req)

	if err := h.service.Create(obj); respondValidationOrError(c, err, "creation_failed") {
		return
	}

	c.JSON(http.StatusCreated, toObjectDataResponse(*obj))
}

// GetObjectData godoc
// @Summary Get object data by ID
// @Tags facility-object-data
// @Produce json
// @Param id path string true "Object Data ID"
// @Success 200 {object} ObjectDataResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/facility/object-data/{id} [get]
func (h *ObjectDataHandler) GetObjectData(c *gin.Context) {
	id, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}

	obj, err := h.service.GetByID(id)
	if err != nil {
		if respondNotFoundIf(c, err, "Object data not found") {
			return
		}
		respondError(c, http.StatusInternalServerError, "fetch_failed", err.Error())
		return
	}

	c.JSON(http.StatusOK, toObjectDataResponse(*obj))
}

// ListObjectData godoc
// @Summary List object data with pagination
// @Tags facility-object-data
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Param search query string false "Search query"
// @Success 200 {object} ObjectDataListResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/facility/object-data [get]
func (h *ObjectDataHandler) ListObjectData(c *gin.Context) {
	query, ok := parsePaginationQuery(c)
	if !ok {
		return
	}

	result, err := h.service.List(query.Page, query.Limit, query.Search)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "fetch_failed", err.Error())
		return
	}

	c.JSON(http.StatusOK, toObjectDataListResponse(result))
}

// UpdateObjectData godoc
// @Summary Update object data
// @Tags facility-object-data
// @Accept json
// @Produce json
// @Param id path string true "Object Data ID"
// @Param object_data body dto.UpdateObjectDataRequest true "Object Data"
// @Success 200 {object} dto.ObjectDataResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/facility/object-data/{id} [put]
func (h *ObjectDataHandler) UpdateObjectData(c *gin.Context) {
	id, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}

	var req dto.UpdateObjectDataRequest
	if !bindJSON(c, &req) {
		return
	}

	obj, err := h.service.GetByID(id)
	if err != nil {
		if respondNotFoundIf(c, err, "Object data not found") {
			return
		}
		respondError(c, http.StatusInternalServerError, "fetch_failed", err.Error())
		return
	}

	applyObjectDataUpdate(obj, req)

	if err := h.service.Update(obj); respondValidationOrError(c, err, "update_failed") {
		return
	}

	c.JSON(http.StatusOK, toObjectDataResponse(*obj))
}

// DeleteObjectData godoc
// @Summary Delete object data
// @Tags facility-object-data
// @Produce json
// @Param id path string true "Object Data ID"
// @Success 204
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/facility/object-data/{id} [delete]
func (h *ObjectDataHandler) DeleteObjectData(c *gin.Context) {
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
