package facility

import (
	"net/http"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	dto "github.com/besart951/go_infra_link/backend/internal/handler/dto/facility"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ObjectDataHandler struct {
	service        ObjectDataService
	bacnetService  BacnetObjectService
	apparatService ApparatService
}

func NewObjectDataHandler(service ObjectDataService, bacnetService BacnetObjectService, apparatService ApparatService) *ObjectDataHandler {
	return &ObjectDataHandler{service: service, bacnetService: bacnetService, apparatService: apparatService}
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

	ctx := c.Request.Context()
	created, err := h.service.CreateTemplate(ctx, toObjectDataTemplateCreate(req))
	if err != nil {
		respondLocalizedDomainError(c, err, "creation_failed", "facility.creation_failed",
			localizedInvalidArgument("facility.invalid_bacnet_object_data"),
			localizedInvalidReference(),
			localizedConflict("facility.entity_conflict"),
		)
		return
	}

	c.JSON(http.StatusCreated, toObjectDataResponse(*created))
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

	obj, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		if respondLocalizedNotFoundIf(c, err, "facility.object_data_not_found") {
			return
		}
		respondLocalizedError(c, http.StatusInternalServerError, "fetch_failed", "facility.fetch_failed")
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
// @Param apparat_id query string false "Filter by Apparat ID"
// @Param system_part_id query string false "Filter by System Part ID"
// @Success 200 {object} ObjectDataListResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/facility/object-data [get]
func (h *ObjectDataHandler) ListObjectData(c *gin.Context) {
	query, ok := parsePaginationQuery(c)
	if !ok {
		return
	}

	ctx := c.Request.Context()
	apparatIDStr := c.Query("apparat_id")
	systemPartIDStr := c.Query("system_part_id")

	var apparatID *uuid.UUID
	var systemPartID *uuid.UUID

	if apparatIDStr != "" {
		id, err := parseUUIDString(apparatIDStr)
		if err != nil {
			respondLocalizedError(c, http.StatusBadRequest, "invalid_apparat_id", "facility.invalid_apparat_id")
			return
		}
		apparatID = &id
	}

	if systemPartIDStr != "" {
		id, err := parseUUIDString(systemPartIDStr)
		if err != nil {
			respondLocalizedError(c, http.StatusBadRequest, "invalid_system_part_id", "facility.invalid_system_part_id")
			return
		}
		systemPartID = &id
	}

	var (
		result *domain.PaginatedList[domainFacility.ObjectData]
		err    error
	)

	switch {
	case apparatID != nil && systemPartID != nil:
		result, err = h.service.ListByApparatAndSystemPartID(ctx, query.Page, query.Limit, query.Search, *apparatID, *systemPartID)
	case apparatID != nil:
		result, err = h.service.ListByApparatID(ctx, query.Page, query.Limit, query.Search, *apparatID)
	case systemPartID != nil:
		result, err = h.service.ListBySystemPartID(ctx, query.Page, query.Limit, query.Search, *systemPartID)
	default:
		result, err = h.service.List(ctx, query.Page, query.Limit, query.Search)
	}
	if err != nil {
		respondLocalizedError(c, http.StatusInternalServerError, "fetch_failed", "facility.fetch_failed")
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

	ctx := c.Request.Context()

	updated, err := h.service.UpdateTemplate(ctx, id, toObjectDataTemplateUpdate(req))
	if err != nil {
		respondLocalizedDomainError(c, err, "update_failed", "facility.update_failed",
			localizedNotFound("facility.object_data_not_found"),
			localizedInvalidReference(),
			localizedConflict("facility.entity_conflict"),
		)
		return
	}

	c.JSON(http.StatusOK, toObjectDataResponse(*updated))
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

	if err := h.service.DeleteByID(c.Request.Context(), id); err != nil {
		respondLocalizedDomainError(c, err, "deletion_failed", "facility.deletion_failed",
			localizedNotFound("facility.object_data_not_found"),
		)
		return
	}

	c.Status(http.StatusNoContent)
}

// GetObjectDataBacnetObjects godoc
// @Summary Get bacnet objects for object data
// @Tags facility-object-data
// @Produce json
// @Param id path string true "Object Data ID"
// @Success 200 {array} dto.BacnetObjectResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/facility/object-data/{id}/bacnet-objects [get]
func (h *ObjectDataHandler) GetObjectDataBacnetObjects(c *gin.Context) {
	id, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}

	ctx := c.Request.Context()

	bacnetObjectIDs, err := h.service.GetBacnetObjectIDs(ctx, id)
	if err != nil {
		if respondLocalizedNotFoundIf(c, err, "facility.object_data_not_found") {
			return
		}
		respondLocalizedError(c, http.StatusInternalServerError, "fetch_failed", "facility.fetch_failed")
		return
	}

	if len(bacnetObjectIDs) == 0 {
		c.JSON(http.StatusOK, []dto.BacnetObjectResponse{})
		return
	}

	bacnetObjects, err := h.bacnetService.GetByIDs(ctx, bacnetObjectIDs)
	if err != nil {
		respondLocalizedError(c, http.StatusInternalServerError, "fetch_failed", "facility.fetch_failed")
		return
	}

	response := make([]dto.BacnetObjectResponse, 0, len(bacnetObjects))
	for _, obj := range bacnetObjects {
		if obj != nil {
			response = append(response, toBacnetObjectResponse(*obj))
		}
	}

	c.JSON(http.StatusOK, response)
}
