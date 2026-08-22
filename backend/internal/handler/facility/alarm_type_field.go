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

// CreateAlarmTypeField godoc
// @Summary Add a field to an alarm type
// @Tags facility-alarm-types
// @Accept json
// @Produce json
// @Param id path string true "Alarm Type ID"
// @Param field body dto.CreateAlarmTypeFieldRequest true "Alarm type field data"
// @Success 201 {object} dto.AlarmTypeFieldResponse
// @Router /api/v1/facility/alarm-types/{id}/fields [post]
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

// UpdateAlarmTypeField godoc
// @Summary Update an alarm type field mapping
// @Tags facility-alarm-types
// @Accept json
// @Produce json
// @Param id path string true "Alarm Type Field ID"
// @Param field body dto.UpdateAlarmTypeFieldRequest true "Alarm type field data"
// @Success 200 {object} dto.AlarmTypeFieldResponse
// @Router /api/v1/facility/alarm-type-fields/{id} [put]
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
	item.Version = req.BaseVersion
	if err := h.service.Update(ctx, item); respondLocalizedValidationOrError(c, err, "facility.update_failed") {
		return
	}
	c.JSON(http.StatusOK, toAlarmTypeFieldResponse(*item))
}

// DeleteAlarmTypeField godoc
// @Summary Delete an alarm type field mapping
// @Tags facility-alarm-types
// @Param id path string true "Alarm Type Field ID"
// @Param base_version query integer true "Expected aggregate version" minimum(1)
// @Success 204
// @Router /api/v1/facility/alarm-type-fields/{id} [delete]
func (h *AlarmTypeFieldHandler) DeleteAlarmTypeField(c *gin.Context) {
	id, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}
	version, ok := parseRequiredBaseVersion(c)
	if !ok {
		return
	}
	if err := deleteAtVersion(c.Request.Context(), h.service, id, version); err != nil {
		respondLocalizedDomainError(c, err, "deletion_failed", "facility.deletion_failed",
			localizedNotFound("facility.not_found"),
		)
		return
	}
	c.Status(http.StatusNoContent)
}
