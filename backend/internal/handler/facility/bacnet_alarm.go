package facility

import (
	"errors"
	"net/http"

	appbacnetobject "github.com/besart951/go_infra_link/backend/internal/application/facility/bacnetobject"
	dto "github.com/besart951/go_infra_link/backend/internal/handler/dto/facility"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type BacnetAlarmHandler struct {
	service  BacnetAlarmValueService
	replacer BacnetAlarmValueReplacer
}

func NewBacnetAlarmHandler(
	service BacnetAlarmValueService,
	replacer BacnetAlarmValueReplacer,
) *BacnetAlarmHandler {
	return &BacnetAlarmHandler{service: service, replacer: replacer}
}

// GetAlarmSchema godoc
// @Summary Get alarm field schema for a BacnetObject
// @Tags facility-bacnet-alarm
// @Produce json
// @Param id path string true "BacnetObject ID"
// @Success 200 {object} dto.AlarmTypeResponse
// @Router /api/v1/facility/bacnet-objects/{id}/alarm-schema [get]
func (h *BacnetAlarmHandler) GetAlarmSchema(c *gin.Context) {
	id, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}

	schema, err := h.service.GetSchema(c.Request.Context(), id)
	if err != nil {
		respondLocalizedError(c, http.StatusInternalServerError, "fetch_failed", "facility.fetch_failed")
		return
	}
	if schema == nil {
		c.JSON(http.StatusOK, nil)
		return
	}

	c.JSON(http.StatusOK, toAlarmTypeResponse(*schema))
}

// GetAlarmValues godoc
// @Summary Get alarm values for a BacnetObject
// @Tags facility-bacnet-alarm
// @Produce json
// @Param id path string true "BacnetObject ID"
// @Success 200 {object} dto.AlarmValuesResponse
// @Router /api/v1/facility/bacnet-objects/{id}/alarm-values [get]
func (h *BacnetAlarmHandler) GetAlarmValues(c *gin.Context) {
	id, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}

	values, err := h.service.GetValues(c.Request.Context(), id)
	if err != nil {
		respondLocalizedError(c, http.StatusInternalServerError, "fetch_failed", "facility.fetch_failed")
		return
	}

	c.JSON(http.StatusOK, toAlarmValuesResponse(values))
}

// PutAlarmValues godoc
// @Summary Replace alarm values for a BacnetObject
// @Tags facility-bacnet-alarm
// @Accept json
// @Produce json
// @Param id path string true "BacnetObject ID"
// @Param values body dto.PutAlarmValuesRequest true "Alarm values"
// @Success 200 {object} dto.AlarmValuesResponse
// @Router /api/v1/facility/bacnet-objects/{id}/alarm-values [put]
func (h *BacnetAlarmHandler) PutAlarmValues(c *gin.Context) {
	id, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}

	var req dto.PutAlarmValuesRequest
	if !bindJSON(c, &req) {
		return
	}

	if h.replacer == nil {
		respondLocalizedError(c, http.StatusInternalServerError, "update_failed", "facility.update_failed")
		return
	}

	updated, err := h.replacer.ReplaceAlarmValues(
		c.Request.Context(),
		toReplaceAlarmValuesCommand(id, req.Values),
	)
	if err != nil {
		var reloadErr *appbacnetobject.AlarmValuesReloadError
		if errors.As(err, &reloadErr) {
			respondLocalizedError(c, http.StatusInternalServerError, "fetch_failed", "facility.fetch_failed")
			return
		}
		respondLocalizedValidationOrError(c, err, "facility.update_failed")
		return
	}

	c.JSON(http.StatusOK, toAlarmValuesResponse(updated))
}

func toReplaceAlarmValuesCommand(
	bacnetObjectID uuid.UUID,
	inputs []dto.AlarmValueInput,
) appbacnetobject.ReplaceAlarmValuesCommand {
	values := make([]appbacnetobject.AlarmValueInput, len(inputs))
	for i, inp := range inputs {
		values[i] = appbacnetobject.AlarmValueInput{
			AlarmTypeFieldID: inp.AlarmTypeFieldID,
			ValueNumber:      inp.ValueNumber,
			ValueInteger:     inp.ValueInteger,
			ValueBoolean:     inp.ValueBoolean,
			ValueString:      inp.ValueString,
			ValueJSON:        inp.ValueJSON,
			UnitID:           inp.UnitID,
			Source:           inp.Source,
		}
	}
	return appbacnetobject.ReplaceAlarmValuesCommand{
		BacnetObjectID: bacnetObjectID,
		Values:         values,
	}
}
