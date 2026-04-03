package project

import (
	"errors"
	"net/http"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	dto "github.com/besart951/go_infra_link/backend/internal/handler/dto/project"
	"github.com/besart951/go_infra_link/backend/internal/handlerutil"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

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
func (h *ProjectHandler) ListProjectObjectData(c *gin.Context) {
	projectID, ok := handlerutil.ParseUUIDParam(c, "id")
	if !ok {
		return
	}

	if !h.ensureProjectAccess(c, projectID) {
		return
	}

	if _, err := h.service.GetByID(projectID); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			handlerutil.RespondLocalizedError(c, http.StatusNotFound, "not_found", "project.project_not_found")
			return
		}
		handlerutil.RespondLocalizedError(c, http.StatusInternalServerError, "fetch_failed", "project.fetch_failed")
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

	result, err := h.service.ListObjectData(projectID, query.Page, query.Limit, query.Search, apparatID, systemPartID)
	if err != nil {
		handlerutil.RespondLocalizedError(c, http.StatusInternalServerError, "fetch_failed", "project.fetch_failed")
		return
	}

	response := dto.ObjectDataListResponse{
		Items:      ToObjectDataList(result.Items),
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
func (h *ProjectHandler) AddProjectObjectData(c *gin.Context) {
	projectID, ok := handlerutil.ParseUUIDParam(c, "id")
	if !ok {
		return
	}

	if !h.ensureProjectAccess(c, projectID) {
		return
	}

	var req dto.CreateProjectObjectDataRequest
	if !handlerutil.BindJSON(c, &req) {
		return
	}

	obj, err := h.service.AddObjectData(projectID, req.ObjectDataID)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrNotFound):
			handlerutil.RespondLocalizedError(c, http.StatusNotFound, "not_found", "project.project_or_object_data_not_found")
			return
		case errors.Is(err, domain.ErrConflict):
			handlerutil.RespondLocalizedError(c, http.StatusConflict, "conflict", "project.object_data_already_linked")
			return
		default:
			handlerutil.RespondLocalizedError(c, http.StatusInternalServerError, "update_failed", "project.update_failed")
			return
		}
	}

	h.notifyProjectChange(c, projectID, "project.object_data.created")

	c.JSON(http.StatusCreated, ToObjectDataResponse(*obj))
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
func (h *ProjectHandler) RemoveProjectObjectData(c *gin.Context) {
	projectID, ok := handlerutil.ParseUUIDParam(c, "id")
	if !ok {
		return
	}

	if !h.ensureProjectAccess(c, projectID) {
		return
	}

	objectDataID, ok := handlerutil.ParseUUIDParam(c, "objectDataId")
	if !ok {
		return
	}

	obj, err := h.service.RemoveObjectData(projectID, objectDataID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			handlerutil.RespondLocalizedError(c, http.StatusNotFound, "not_found", "project.project_or_object_data_not_found")
			return
		}
		handlerutil.RespondLocalizedError(c, http.StatusInternalServerError, "update_failed", "project.update_failed")
		return
	}

	h.notifyProjectChange(c, projectID, "project.object_data.deleted")

	c.JSON(http.StatusOK, ToObjectDataResponse(*obj))
}
