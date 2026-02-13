package facility

import (
	"net/http"

	"github.com/besart951/go_infra_link/backend/internal/handler/dto"
	"github.com/gin-gonic/gin"
)

type NotificationClassHandler struct {
	service NotificationClassService
}

func NewNotificationClassHandler(service NotificationClassService) *NotificationClassHandler {
	return &NotificationClassHandler{service: service}
}

// CreateNotificationClass godoc
// @Summary Create a new notification class
// @Tags facility-notification-classes
// @Accept json
// @Produce json
// @Param notification_class body dto.CreateNotificationClassRequest true "Notification Class data"
// @Success 201 {object} dto.NotificationClassResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/facility/notification-classes [post]
func (h *NotificationClassHandler) CreateNotificationClass(c *gin.Context) {
	var req dto.CreateNotificationClassRequest
	if !bindJSON(c, &req) {
		return
	}

	nc := toNotificationClassModel(req)

	if err := h.service.Create(nc); respondLocalizedValidationOrError(c, err, "facility.creation_failed") {
		return
	}

	c.JSON(http.StatusCreated, toNotificationClassResponse(*nc))
}

// GetNotificationClass godoc
// @Summary Get a notification class by ID
// @Tags facility-notification-classes
// @Produce json
// @Param id path string true "Notification Class ID"
// @Success 200 {object} NotificationClassResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/facility/notification-classes/{id} [get]
func (h *NotificationClassHandler) GetNotificationClass(c *gin.Context) {
	id, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}

	nc, err := h.service.GetByID(id)
	if err != nil {
		if respondLocalizedNotFoundIf(c, err, "facility.notification_class_not_found") {
			return
		}
		respondLocalizedError(c, http.StatusInternalServerError, "fetch_failed", "facility.fetch_failed")
		return
	}

	c.JSON(http.StatusOK, toNotificationClassResponse(*nc))
}

// ListNotificationClasses godoc
// @Summary List notification classes with pagination
// @Tags facility-notification-classes
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Param search query string false "Search query"
// @Success 200 {object} NotificationClassListResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/facility/notification-classes [get]
func (h *NotificationClassHandler) ListNotificationClasses(c *gin.Context) {
	query, ok := parsePaginationQuery(c)
	if !ok {
		return
	}

	result, err := h.service.List(query.Page, query.Limit, query.Search)
	if err != nil {
		respondLocalizedError(c, http.StatusInternalServerError, "fetch_failed", "facility.fetch_failed")
		return
	}

	c.JSON(http.StatusOK, toNotificationClassListResponse(result))
}

// UpdateNotificationClass godoc
// @Summary Update a notification class
// @Tags facility-notification-classes
// @Accept json
// @Produce json
// @Param id path string true "Notification Class ID"
// @Param notification_class body dto.UpdateNotificationClassRequest true "Notification Class data"
// @Success 200 {object} dto.NotificationClassResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/facility/notification-classes/{id} [put]
func (h *NotificationClassHandler) UpdateNotificationClass(c *gin.Context) {
	id, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}

	var req dto.UpdateNotificationClassRequest
	if !bindJSON(c, &req) {
		return
	}

	nc, err := h.service.GetByID(id)
	if err != nil {
		if respondLocalizedNotFoundIf(c, err, "facility.notification_class_not_found") {
			return
		}
		respondLocalizedError(c, http.StatusInternalServerError, "fetch_failed", "facility.fetch_failed")
		return
	}

	applyNotificationClassUpdate(nc, req)

	if err := h.service.Update(nc); respondLocalizedValidationOrError(c, err, "facility.update_failed") {
		return
	}

	c.JSON(http.StatusOK, toNotificationClassResponse(*nc))
}

// DeleteNotificationClass godoc
// @Summary Delete a notification class
// @Tags facility-notification-classes
// @Produce json
// @Param id path string true "Notification Class ID"
// @Success 204
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/facility/notification-classes/{id} [delete]
func (h *NotificationClassHandler) DeleteNotificationClass(c *gin.Context) {
	id, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}

	if err := h.service.DeleteByID(id); err != nil {
		respondLocalizedError(c, http.StatusInternalServerError, "deletion_failed", "facility.deletion_failed")
		return
	}

	c.Status(http.StatusNoContent)
}
