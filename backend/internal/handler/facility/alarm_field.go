package facility

import (
	"net/http"

	"github.com/besart951/go_infra_link/backend/internal/handler/dto"
	"github.com/gin-gonic/gin"
)

type AlarmFieldHandler struct {
	service AlarmFieldService
}

func NewAlarmFieldHandler(service AlarmFieldService) *AlarmFieldHandler {
	return &AlarmFieldHandler{service: service}
}

func (h *AlarmFieldHandler) CreateAlarmField(c *gin.Context) {
	var req dto.CreateAlarmFieldRequest
	if !bindJSON(c, &req) {
		return
	}
	item := toAlarmFieldModel(req)
	if err := h.service.Create(item); respondLocalizedValidationOrError(c, err, "facility.creation_failed") {
		return
	}
	c.JSON(http.StatusCreated, toAlarmFieldResponse(*item))
}

func (h *AlarmFieldHandler) ListAlarmFields(c *gin.Context) {
	query, ok := parsePaginationQuery(c)
	if !ok {
		return
	}
	result, err := h.service.List(query.Page, query.Limit, query.Search)
	if err != nil {
		respondLocalizedError(c, http.StatusInternalServerError, "fetch_failed", "facility.fetch_failed")
		return
	}
	c.JSON(http.StatusOK, toAlarmFieldListResponse(result))
}

func (h *AlarmFieldHandler) GetAlarmField(c *gin.Context) {
	id, ok := parseUUIDParam(c, "id")
	if !ok {
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
	c.JSON(http.StatusOK, toAlarmFieldResponse(*item))
}

func (h *AlarmFieldHandler) UpdateAlarmField(c *gin.Context) {
	id, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}
	var req dto.UpdateAlarmFieldRequest
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
	applyAlarmFieldUpdate(item, req)
	if err := h.service.Update(item); respondLocalizedValidationOrError(c, err, "facility.update_failed") {
		return
	}
	c.JSON(http.StatusOK, toAlarmFieldResponse(*item))
}

func (h *AlarmFieldHandler) DeleteAlarmField(c *gin.Context) {
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
