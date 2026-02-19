package facility

import (
	"net/http"

	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/besart951/go_infra_link/backend/internal/handler/dto"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type BacnetAlarmHandler struct {
	service BacnetAlarmValueService
}

func NewBacnetAlarmHandler(service BacnetAlarmValueService) *BacnetAlarmHandler {
	return &BacnetAlarmHandler{service: service}
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

	schema, err := h.service.GetSchema(id)
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

	values, err := h.service.GetValues(id)
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

	values := toAlarmValueModels(id, req.Values)

	if err := h.service.PutValues(id, values); respondLocalizedValidationOrError(c, err, "facility.update_failed") {
		return
	}

	updated, err := h.service.GetValues(id)
	if err != nil {
		respondLocalizedError(c, http.StatusInternalServerError, "fetch_failed", "facility.fetch_failed")
		return
	}

	c.JSON(http.StatusOK, toAlarmValuesResponse(updated))
}

func toAlarmValueModels(bacnetObjectID uuid.UUID, inputs []dto.AlarmValueInput) []domainFacility.BacnetObjectAlarmValue {
	values := make([]domainFacility.BacnetObjectAlarmValue, len(inputs))
	for i, inp := range inputs {
		source := inp.Source
		if source == "" {
			source = domainFacility.AlarmValueSourceUser
		}
		values[i] = domainFacility.BacnetObjectAlarmValue{
			BacnetObjectID:   bacnetObjectID,
			AlarmTypeFieldID: inp.AlarmTypeFieldID,
			ValueNumber:      inp.ValueNumber,
			ValueInteger:     inp.ValueInteger,
			ValueBoolean:     inp.ValueBoolean,
			ValueString:      inp.ValueString,
			ValueJSON:        inp.ValueJSON,
			UnitID:           inp.UnitID,
			Source:           source,
		}
	}
	return values
}
