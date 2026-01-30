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

	obj := toObjectDataModel(req)

	// Load apparats if IDs are provided
	if len(req.ApparatIDs) > 0 {
		apparats, err := h.apparatService.GetByIDs(req.ApparatIDs)
		if err != nil {
			respondError(c, http.StatusBadRequest, "invalid_apparats", "Failed to load apparats")
			return
		}
		obj.Apparats = apparats
	}

	if err := h.service.Create(obj); respondValidationOrError(c, err, "creation_failed") {
		return
	}

	if req.BacnetObjects != nil && len(*req.BacnetObjects) > 0 {
		for _, input := range *req.BacnetObjects {
			createReq := dto.CreateBacnetObjectRequest{
				ObjectDataID:      &obj.ID,
				BacnetObjectInput: input,
			}
			bacnetObject := toBacnetObjectModel(createReq)
			if err := h.bacnetService.CreateWithParent(bacnetObject, nil, &obj.ID); err != nil {
				if ve, ok := domain.AsValidationError(err); ok {
					respondValidationError(c, ve.Fields)
					return
				}
				if errors.Is(err, domain.ErrInvalidArgument) {
					respondInvalidArgument(c, "invalid bacnet object data")
					return
				}
				if errors.Is(err, domain.ErrNotFound) {
					respondInvalidReference(c)
					return
				}
				if errors.Is(err, domain.ErrConflict) {
					respondConflict(c, "entity conflict")
					return
				}
				respondError(c, http.StatusInternalServerError, "creation_failed", err.Error())
				return
			}
		}
	}

	if created, err := h.service.GetByID(obj.ID); err == nil && created != nil {
		c.JSON(http.StatusCreated, toObjectDataResponse(*created))
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

	apparatIDStr := c.Query("apparat_id")
	systemPartIDStr := c.Query("system_part_id")

	var apparatID *uuid.UUID
	var systemPartID *uuid.UUID

	if apparatIDStr != "" {
		id, err := parseUUIDString(apparatIDStr)
		if err != nil {
			respondError(c, http.StatusBadRequest, "invalid_apparat_id", "Invalid apparat_id format")
			return
		}
		apparatID = &id
	}

	if systemPartIDStr != "" {
		id, err := parseUUIDString(systemPartIDStr)
		if err != nil {
			respondError(c, http.StatusBadRequest, "invalid_system_part_id", "Invalid system_part_id format")
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
		result, err = h.service.ListByApparatAndSystemPartID(query.Page, query.Limit, query.Search, *apparatID, *systemPartID)
	case apparatID != nil:
		result, err = h.service.ListByApparatID(query.Page, query.Limit, query.Search, *apparatID)
	case systemPartID != nil:
		result, err = h.service.ListBySystemPartID(query.Page, query.Limit, query.Search, *systemPartID)
	default:
		result, err = h.service.List(query.Page, query.Limit, query.Search)
	}
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

	// Load apparats if IDs are provided
	if req.ApparatIDs != nil {
		if len(*req.ApparatIDs) > 0 {
			apparats, err := h.apparatService.GetByIDs(*req.ApparatIDs)
			if err != nil {
				respondError(c, http.StatusBadRequest, "invalid_apparats", "Failed to load apparats")
				return
			}
			obj.Apparats = apparats
		} else {
			// Empty array means clear all apparats
			obj.Apparats = []*domainFacility.Apparat{}
		}
	}

	if err := h.service.Update(obj); respondValidationOrError(c, err, "update_failed") {
		return
	}

	if req.BacnetObjects != nil {
		bacnetObjects := toFieldDeviceBacnetObjects(*req.BacnetObjects)
		if err := h.bacnetService.ReplaceForObjectData(obj.ID, bacnetObjects); err != nil {
			if ve, ok := domain.AsValidationError(err); ok {
				respondValidationError(c, ve.Fields)
				return
			}
			if respondNotFoundIf(c, err, "Object data not found") {
				return
			}
			respondError(c, http.StatusInternalServerError, "update_failed", err.Error())
			return
		}
	}

	updated, err := h.service.GetByID(obj.ID)
	if err != nil {
		if respondNotFoundIf(c, err, "Object data not found") {
			return
		}
		respondError(c, http.StatusInternalServerError, "fetch_failed", err.Error())
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

	if err := h.service.DeleteByID(id); err != nil {
		respondError(c, http.StatusInternalServerError, "deletion_failed", err.Error())
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

	bacnetObjectIDs, err := h.service.GetBacnetObjectIDs(id)
	if err != nil {
		if respondNotFoundIf(c, err, "Object data not found") {
			return
		}
		respondError(c, http.StatusInternalServerError, "fetch_failed", err.Error())
		return
	}

	if len(bacnetObjectIDs) == 0 {
		c.JSON(http.StatusOK, []dto.BacnetObjectResponse{})
		return
	}

	bacnetObjects, err := h.bacnetService.GetByIDs(bacnetObjectIDs)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "fetch_failed", err.Error())
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
