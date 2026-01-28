package facility

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type NotificationClassHandler struct {
	service NotificationClassService
}

func NewNotificationClassHandler(service NotificationClassService) *NotificationClassHandler {
	return &NotificationClassHandler{service: service}
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
		if respondNotFoundIf(c, err, "Notification class not found") {
			return
		}
		respondError(c, http.StatusInternalServerError, "fetch_failed", err.Error())
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
		respondError(c, http.StatusInternalServerError, "fetch_failed", err.Error())
		return
	}

	c.JSON(http.StatusOK, toNotificationClassListResponse(result))
}
