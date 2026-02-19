package facility

import (
	"net/http"

	"github.com/besart951/go_infra_link/backend/internal/handler/dto"
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
	if err := h.service.Create(item); respondLocalizedValidationOrError(c, err, "facility.creation_failed") {
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
	item, err := h.service.GetByID(id)
	if err != nil {
		if respondLocalizedNotFoundIf(c, err, "facility.not_found") {
			return
		}
		respondLocalizedError(c, http.StatusInternalServerError, "fetch_failed", "facility.fetch_failed")
		return
	}
	applyAlarmTypeFieldUpdate(item, req)
	if err := h.service.Update(item); respondLocalizedValidationOrError(c, err, "facility.update_failed") {
		return
	}
	c.JSON(http.StatusOK, toAlarmTypeFieldResponse(*item))
}

func (h *AlarmTypeFieldHandler) DeleteAlarmTypeField(c *gin.Context) {
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
