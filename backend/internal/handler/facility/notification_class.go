package facility

import (
	"errors"
	"net/http"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	"github.com/besart951/go_infra_link/backend/internal/handler/dto"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type NotificationClassHandler struct {
	service NotificationClassService
}

func NewNotificationClassHandler(service NotificationClassService) *NotificationClassHandler {
	return &NotificationClassHandler{service: service}
}

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
