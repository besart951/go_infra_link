package facility

import (
	"net/http"

	"github.com/besart951/go_infra_link/backend/internal/handler/dto"
	"github.com/gin-gonic/gin"
)

type UnitHandler struct {
	service UnitService
}

func NewUnitHandler(service UnitService) *UnitHandler {
	return &UnitHandler{service: service}
}

func (h *UnitHandler) CreateUnit(c *gin.Context) {
	var req dto.CreateUnitRequest
	if !bindJSON(c, &req) {
		return
	}
	item := toUnitModel(req)
	if err := h.service.Create(item); respondLocalizedValidationOrError(c, err, "facility.creation_failed") {
		return
	}
	c.JSON(http.StatusCreated, toUnitResponse(*item))
}

func (h *UnitHandler) ListUnits(c *gin.Context) {
	query, ok := parsePaginationQuery(c)
	if !ok {
		return
	}
	result, err := h.service.List(query.Page, query.Limit, query.Search)
	if err != nil {
		respondLocalizedError(c, http.StatusInternalServerError, "fetch_failed", "facility.fetch_failed")
		return
	}
	c.JSON(http.StatusOK, toUnitListResponse(result))
}

func (h *UnitHandler) GetUnit(c *gin.Context) {
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
	c.JSON(http.StatusOK, toUnitResponse(*item))
}

func (h *UnitHandler) UpdateUnit(c *gin.Context) {
	id, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}
	var req dto.UpdateUnitRequest
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
	applyUnitUpdate(item, req)
	if err := h.service.Update(item); respondLocalizedValidationOrError(c, err, "facility.update_failed") {
		return
	}
	c.JSON(http.StatusOK, toUnitResponse(*item))
}

func (h *UnitHandler) DeleteUnit(c *gin.Context) {
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
