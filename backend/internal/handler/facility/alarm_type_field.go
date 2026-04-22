package facility

import (
	"net/http"

	dto "github.com/besart951/go_infra_link/backend/internal/handler/dto/facility"
	"github.com/gin-gonic/gin"
)

type AlarmTypeFieldHandler struct {
	service AlarmTypeFieldService
}

func NewAlarmTypeFieldHandler(service AlarmTypeFieldService) *AlarmTypeFieldHandler {
	return &AlarmTypeFieldHandler{service: service}
}

func (h *AlarmTypeFieldHandler) CreateAlarmTypeField(c *gin.Context) {
	alarmTypeID, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}
	var req dto.CreateAlarmTypeFieldRequest
	if !bindJSON(c, &req) {
		return
	}
	item := toAlarmTypeFieldModel(alarmTypeID, req)
	if err := h.service.Create(c.Request.Context(), item); respondLocalizedValidationOrError(c, err, "facility.creation_failed") {
		return
	}
	c.JSON(http.StatusCreated, toAlarmTypeFieldResponse(*item))
}

func (h *AlarmTypeFieldHandler) UpdateAlarmTypeField(c *gin.Context) {
	id, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}
	var req dto.UpdateAlarmTypeFieldRequest
	if !bindJSON(c, &req) {
		return
	}
	ctx := c.Request.Context()
	item, err := h.service.GetByID(ctx, id)
	if err != nil {
		if respondLocalizedNotFoundIf(c, err, "facility.not_found") {
			return
		}
		respondLocalizedError(c, http.StatusInternalServerError, "fetch_failed", "facility.fetch_failed")
		return
	}
	applyAlarmTypeFieldUpdate(item, req)
	if err := h.service.Update(ctx, item); respondLocalizedValidationOrError(c, err, "facility.update_failed") {
		return
	}
	c.JSON(http.StatusOK, toAlarmTypeFieldResponse(*item))
}

func (h *AlarmTypeFieldHandler) DeleteAlarmTypeField(c *gin.Context) {
	id, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}
	if err := h.service.DeleteByID(c.Request.Context(), id); err != nil {
		respondLocalizedDomainError(c, err, "deletion_failed", "facility.deletion_failed",
			localizedNotFound("facility.not_found"),
		)
		return
	}
	c.Status(http.StatusNoContent)
}
