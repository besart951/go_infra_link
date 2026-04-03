package facility

import (
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	dto "github.com/besart951/go_infra_link/backend/internal/handler/dto/facility"
	"github.com/gin-gonic/gin"
)

type NotificationClassHandler struct {
	crud crudHandler[domainFacility.NotificationClass, dto.CreateNotificationClassRequest, dto.UpdateNotificationClassRequest]
}

func NewNotificationClassHandler(svc NotificationClassService) *NotificationClassHandler {
	return &NotificationClassHandler{crud: newCRUD(
		svc,
		toNotificationClassModel,
		applyNotificationClassUpdate,
		respFn(toNotificationClassResponse),
		listRespFn(toNotificationClassListResponse),
		"facility.notification_class_not_found",
	)}
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
func (h *NotificationClassHandler) CreateNotificationClass(c *gin.Context) { h.crud.handleCreate(c) }

// GetNotificationClass godoc
// @Summary Get a notification class by ID
// @Tags facility-notification-classes
// @Produce json
// @Param id path string true "Notification Class ID"
// @Success 200 {object} dto.NotificationClassResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/facility/notification-classes/{id} [get]
func (h *NotificationClassHandler) GetNotificationClass(c *gin.Context) { h.crud.handleGetByID(c) }

// ListNotificationClasses godoc
// @Summary List notification classes with pagination
// @Tags facility-notification-classes
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Param search query string false "Search query"
// @Success 200 {object} dto.NotificationClassListResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/facility/notification-classes [get]
func (h *NotificationClassHandler) ListNotificationClasses(c *gin.Context) { h.crud.handleList(c) }

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
func (h *NotificationClassHandler) UpdateNotificationClass(c *gin.Context) { h.crud.handleUpdate(c) }

// DeleteNotificationClass godoc
// @Summary Delete a notification class
// @Tags facility-notification-classes
// @Produce json
// @Param id path string true "Notification Class ID"
// @Success 204
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/facility/notification-classes/{id} [delete]
func (h *NotificationClassHandler) DeleteNotificationClass(c *gin.Context) { h.crud.handleDelete(c) }
