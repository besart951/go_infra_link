package facility

import (
	"errors"
	"net/http"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	"github.com/besart951/go_infra_link/backend/internal/handler/dto"
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
// @Success 200 {object} dto.NotificationClassResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/facility/notification-classes/{id} [get]
func (h *NotificationClassHandler) GetNotificationClass(c *gin.Context) {
	id, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}

	nc, err := h.service.GetByID(id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			respondNotFound(c, "Notification class not found")
			return
		}
		respondError(c, http.StatusInternalServerError, "fetch_failed", err.Error())
		return
	}

	response := dto.NotificationClassResponse{
		ID:                   nc.ID,
		EventCategory:        nc.EventCategory,
		Nc:                   nc.Nc,
		ObjectDescription:    nc.ObjectDescription,
		InternalDescription:  nc.InternalDescription,
		Meaning:              nc.Meaning,
		AckRequiredNotNormal: nc.AckRequiredNotNormal,
		AckRequiredError:     nc.AckRequiredError,
		AckRequiredNormal:    nc.AckRequiredNormal,
		NormNotNormal:        nc.NormNotNormal,
		NormError:            nc.NormError,
		NormNormal:           nc.NormNormal,
		CreatedAt:            nc.CreatedAt,
		UpdatedAt:            nc.UpdatedAt,
	}

	c.JSON(http.StatusOK, response)
}

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
func (h *NotificationClassHandler) ListNotificationClasses(c *gin.Context) {
	var query dto.PaginationQuery
	if !bindQuery(c, &query) {
		return
	}

	result, err := h.service.List(query.Page, query.Limit, query.Search)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "fetch_failed", err.Error())
		return
	}

	items := make([]dto.NotificationClassResponse, len(result.Items))
	for i, nc := range result.Items {
		items[i] = dto.NotificationClassResponse{
			ID:                   nc.ID,
			EventCategory:        nc.EventCategory,
			Nc:                   nc.Nc,
			ObjectDescription:    nc.ObjectDescription,
			InternalDescription:  nc.InternalDescription,
			Meaning:              nc.Meaning,
			AckRequiredNotNormal: nc.AckRequiredNotNormal,
			AckRequiredError:     nc.AckRequiredError,
			AckRequiredNormal:    nc.AckRequiredNormal,
			NormNotNormal:        nc.NormNotNormal,
			NormError:            nc.NormError,
			NormNormal:           nc.NormNormal,
			CreatedAt:            nc.CreatedAt,
			UpdatedAt:            nc.UpdatedAt,
		}
	}

	response := dto.NotificationClassListResponse{
		Items:      items,
		Total:      result.Total,
		Page:       result.Page,
		TotalPages: result.TotalPages,
	}

	c.JSON(http.StatusOK, response)
}
