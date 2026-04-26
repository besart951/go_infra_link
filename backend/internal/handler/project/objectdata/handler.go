package objectdata

import (
	"context"
	"net/http"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	dto "github.com/besart951/go_infra_link/backend/internal/handler/dto/project"
	projectshared "github.com/besart951/go_infra_link/backend/internal/handler/project/shared"
	"github.com/besart951/go_infra_link/backend/internal/handlerutil"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type FacilityLinkService interface {
	ListObjectData(ctx context.Context, projectID uuid.UUID, page, limit int, search string, apparatID, systemPartID *uuid.UUID) (*domain.PaginatedList[domainFacility.ObjectData], error)
	AddObjectData(ctx context.Context, projectID, objectDataID uuid.UUID) (*domainFacility.ObjectData, error)
	RemoveObjectData(ctx context.Context, projectID, objectDataID uuid.UUID) (*domainFacility.ObjectData, error)
}

type Handler struct {
	access       projectshared.AccessPolicyService
	facilityLink FacilityLinkService
	notify       projectshared.ProjectChangeNotifier
}

func NewHandler(access projectshared.AccessPolicyService, facilityLink FacilityLinkService, notify projectshared.ProjectChangeNotifier) *Handler {
	return &Handler{access: access, facilityLink: facilityLink, notify: notify}
}

// ListProjectObjectData godoc
// @Summary List project object data with pagination
// @Tags projects
// @Produce json
// @Param id path string true "Project ID"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Param search query string false "Search query"
// @Param apparat_id query string false "Filter by Apparat ID"
// @Param system_part_id query string false "Filter by System Part ID"
// @Success 200 {object} dto.ObjectDataListResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/projects/{id}/object-data [get]
func (h *Handler) ListProjectObjectData(c *gin.Context) {
	projectID, ok := handlerutil.ParseUUIDParam(c, "id")
	if !ok {
		return
	}

	if !projectshared.EnsureProjectAccess(c, h.access, projectID) {
		return
	}

	var query dto.PaginationQuery
	if !handlerutil.BindQuery(c, &query) {
		return
	}

	apparatIDStr := c.Query("apparat_id")
	systemPartIDStr := c.Query("system_part_id")

	var apparatID *uuid.UUID
	var systemPartID *uuid.UUID

	if apparatIDStr != "" {
		id, err := uuid.Parse(apparatIDStr)
		if err != nil {
			handlerutil.RespondLocalizedError(c, http.StatusBadRequest, "invalid_apparat_id", "validation.invalid_uuid_format")
			return
		}
		apparatID = &id
	}

	if systemPartIDStr != "" {
		id, err := uuid.Parse(systemPartIDStr)
		if err != nil {
			handlerutil.RespondLocalizedError(c, http.StatusBadRequest, "invalid_system_part_id", "validation.invalid_uuid_format")
			return
		}
		systemPartID = &id
	}

	result, err := h.facilityLink.ListObjectData(c.Request.Context(), projectID, query.Page, query.Limit, query.Search, apparatID, systemPartID)
	if err != nil {
		handlerutil.RespondDomainError(c, err,
			handlerutil.LocalizedError(http.StatusInternalServerError, "fetch_failed", "project.fetch_failed"),
			handlerutil.MapError(domain.ErrNotFound, handlerutil.LocalizedError(http.StatusNotFound, "not_found", "project.project_or_object_data_not_found")),
		)
		return
	}

	response := dto.ObjectDataListResponse{
		Items:      toObjectDataList(result.Items),
		Total:      result.Total,
		Page:       result.Page,
		TotalPages: result.TotalPages,
	}

	c.JSON(http.StatusOK, response)
}

// AddProjectObjectData godoc
// @Summary Attach object data to project
// @Tags projects
// @Accept json
// @Produce json
// @Param id path string true "Project ID"
// @Param payload body dto.CreateProjectObjectDataRequest true "Object data link"
// @Success 201 {object} dto.ObjectDataResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 409 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/projects/{id}/object-data [post]
func (h *Handler) AddProjectObjectData(c *gin.Context) {
	projectID, ok := handlerutil.ParseUUIDParam(c, "id")
	if !ok {
		return
	}

	if !projectshared.EnsureProjectAccess(c, h.access, projectID) {
		return
	}

	var req dto.CreateProjectObjectDataRequest
	if !handlerutil.BindJSON(c, &req) {
		return
	}

	obj, err := h.facilityLink.AddObjectData(c.Request.Context(), projectID, req.ObjectDataID)
	if err != nil {
		handlerutil.RespondDomainError(c, err,
			handlerutil.LocalizedError(http.StatusInternalServerError, "update_failed", "project.update_failed"),
			handlerutil.MapError(domain.ErrNotFound, handlerutil.LocalizedError(http.StatusNotFound, "not_found", "project.project_or_object_data_not_found")),
			handlerutil.MapError(domain.ErrConflict, handlerutil.LocalizedError(http.StatusConflict, "conflict", "project.object_data_already_linked")),
		)
		return
	}

	if h.notify != nil {
		h.notify(c, projectID, "project.object_data.created")
	}

	c.JSON(http.StatusCreated, toObjectDataResponse(*obj))
}

// RemoveProjectObjectData godoc
// @Summary Detach object data from project
// @Tags projects
// @Produce json
// @Param id path string true "Project ID"
// @Param objectDataId path string true "Object Data ID"
// @Success 200 {object} dto.ObjectDataResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/projects/{id}/object-data/{objectDataId} [delete]
func (h *Handler) RemoveProjectObjectData(c *gin.Context) {
	projectID, ok := handlerutil.ParseUUIDParam(c, "id")
	if !ok {
		return
	}

	if !projectshared.EnsureProjectAccess(c, h.access, projectID) {
		return
	}

	objectDataID, ok := handlerutil.ParseUUIDParam(c, "objectDataId")
	if !ok {
		return
	}

	obj, err := h.facilityLink.RemoveObjectData(c.Request.Context(), projectID, objectDataID)
	if err != nil {
		handlerutil.RespondDomainError(c, err,
			handlerutil.LocalizedError(http.StatusInternalServerError, "update_failed", "project.update_failed"),
			handlerutil.MapError(domain.ErrNotFound, handlerutil.LocalizedError(http.StatusNotFound, "not_found", "project.project_or_object_data_not_found")),
		)
		return
	}

	if h.notify != nil {
		h.notify(c, projectID, "project.object_data.deleted")
	}

	c.JSON(http.StatusOK, toObjectDataResponse(*obj))
}

func toObjectDataResponse(item domainFacility.ObjectData) dto.ObjectDataResponse {
	bacnetObjects := []domainFacility.BacnetObject{}
	if len(item.BacnetObjects) > 0 {
		bacnetObjects = make([]domainFacility.BacnetObject, 0, len(item.BacnetObjects))
		for _, obj := range item.BacnetObjects {
			if obj == nil {
				continue
			}
			bacnetObjects = append(bacnetObjects, *obj)
		}
	}

	return dto.ObjectDataResponse{
		ID:            item.ID,
		Description:   item.Description,
		Version:       item.Version,
		IsActive:      item.IsActive,
		ProjectID:     item.ProjectID,
		BacnetObjects: mapBacnetObjectResponses(bacnetObjects),
		CreatedAt:     item.CreatedAt,
		UpdatedAt:     item.UpdatedAt,
	}
}

func toObjectDataList(items []domainFacility.ObjectData) []dto.ObjectDataResponse {
	out := make([]dto.ObjectDataResponse, len(items))
	for i, item := range items {
		out[i] = toObjectDataResponse(item)
	}
	return out
}

func mapBacnetObjectResponses(objs []domainFacility.BacnetObject) []dto.BacnetObjectResponse {
	items := make([]dto.BacnetObjectResponse, len(objs))
	for i, obj := range objs {
		items[i] = dto.BacnetObjectResponse{
			ID:                  obj.ID.String(),
			TextFix:             obj.TextFix,
			Description:         obj.Description,
			GMSVisible:          obj.GMSVisible,
			Optional:            obj.Optional,
			TextIndividual:      obj.TextIndividual,
			SoftwareType:        string(obj.SoftwareType),
			SoftwareNumber:      int(obj.SoftwareNumber),
			HardwareType:        string(obj.HardwareType),
			HardwareQuantity:    int(obj.HardwareQuantity),
			FieldDeviceID:       obj.FieldDeviceID,
			SoftwareReferenceID: obj.SoftwareReferenceID,
			StateTextID:         obj.StateTextID,
			NotificationClassID: obj.NotificationClassID,
			AlarmDefinitionID:   obj.AlarmDefinitionID,
			AlarmTypeID:         obj.AlarmTypeID,
			CreatedAt:           obj.CreatedAt,
			UpdatedAt:           obj.UpdatedAt,
		}
	}
	return items
}
