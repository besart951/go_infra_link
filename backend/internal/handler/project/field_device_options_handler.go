package project

import (
	"net/http"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	facilityhandler "github.com/besart951/go_infra_link/backend/internal/handler/facility"
	"github.com/besart951/go_infra_link/backend/internal/handler/middleware"
	"github.com/besart951/go_infra_link/backend/internal/handlerutil"
	"github.com/gin-gonic/gin"
)

type FieldDeviceOptionsHandler struct {
	accessService  ProjectAccessPolicyService
	service        FieldDeviceOptionsService
}

func NewFieldDeviceOptionsHandler(accessService ProjectAccessPolicyService, service FieldDeviceOptionsService) *FieldDeviceOptionsHandler {
	return &FieldDeviceOptionsHandler{accessService: accessService, service: service}
}

// GetFieldDeviceOptionsForProject godoc
// @Summary Get all metadata needed for creating/editing field devices within a project
// @Description Returns all apparats, system parts, object datas and their relationships for a specific project. This returns project-specific object data (object data where project_id = :id and is_active = true).
// @Tags projects
// @Produce json
// @Param id path string true "Project ID"
// @Success 200 {object} FieldDeviceOptionsResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/projects/{id}/field-device-options [get]
func (h *FieldDeviceOptionsHandler) GetFieldDeviceOptionsForProject(c *gin.Context) {
	projectID, ok := handlerutil.ParseUUIDParam(c, "id")
	if !ok {
		return
	}

	userID, ok := middleware.GetUserID(c)
	if !ok {
		handlerutil.RespondLocalizedError(c, http.StatusUnauthorized, "unauthorized", "errors.unauthorized")
		return
	}

	hasAccess, err := h.accessService.CanAccessProject(c.Request.Context(), userID, projectID)
	if err != nil {
		handlerutil.RespondDomainError(c, err,
			handlerutil.LocalizedError(http.StatusInternalServerError, "fetch_failed", "project.fetch_failed"),
			handlerutil.MapError(domain.ErrNotFound, handlerutil.LocalizedError(http.StatusNotFound, "not_found", "project.project_not_found")),
		)
		return
	}
	if !hasAccess {
		handlerutil.RespondLocalizedError(c, http.StatusForbidden, "forbidden", "errors.forbidden")
		return
	}

	options, err := h.service.GetFieldDeviceOptionsForProject(c.Request.Context(), projectID)
	if err != nil {
		handlerutil.RespondLocalizedError(c, http.StatusInternalServerError, "fetch_failed", "facility.fetch_failed")
		return
	}

	c.JSON(http.StatusOK, facilityhandler.ToFieldDeviceOptionsResponse(options))
}
