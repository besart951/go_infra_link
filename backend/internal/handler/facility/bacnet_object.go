package facility

import (
	"errors"
	"net/http"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	"github.com/besart951/go_infra_link/backend/internal/handler/dto"
	"github.com/gin-gonic/gin"
)

type BacnetObjectHandler struct {
	service BacnetObjectService
}

func NewBacnetObjectHandler(service BacnetObjectService) *BacnetObjectHandler {
	return &BacnetObjectHandler{service: service}
}

// CreateBacnetObject godoc
// @Summary Create a bacnet object (for field device or object data)
// @Tags facility-bacnet-objects
// @Accept json
// @Produce json
// @Param bacnet_object body dto.CreateBacnetObjectRequest true "Bacnet Object data"
// @Success 201 {object} dto.BacnetObjectResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 409 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/facility/bacnet-objects [post]
func (h *BacnetObjectHandler) CreateBacnetObject(c *gin.Context) {
	var req dto.CreateBacnetObjectRequest
	if !bindJSON(c, &req) {
		return
	}

	obj := toBacnetObjectModel(req)

	if err := h.service.CreateWithParent(obj, req.FieldDeviceID, req.ObjectDataID); err != nil {
		if ve, ok := domain.AsValidationError(err); ok {
			respondValidationError(c, ve.Fields)
			return
		}
		if errors.Is(err, domain.ErrInvalidArgument) {
			respondLocalizedInvalidArgument(c, "facility.exactly_one_required")
			return
		}
		if errors.Is(err, domain.ErrNotFound) {
			respondInvalidReference(c)
			return
		}
		if errors.Is(err, domain.ErrConflict) {
			respondLocalizedConflict(c, "facility.entity_conflict")
			return
		}
		respondLocalizedError(c, http.StatusInternalServerError, "creation_failed", "facility.creation_failed")
		return
	}

	c.JSON(http.StatusCreated, toBacnetObjectResponse(*obj))
}

// UpdateBacnetObject godoc
// @Summary Update a bacnet object
// @Tags facility-bacnet-objects
// @Accept json
// @Produce json
// @Param id path string true "Bacnet Object ID"
// @Param bacnet_object body dto.UpdateBacnetObjectRequest true "Bacnet Object data"
// @Success 200 {object} dto.BacnetObjectResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 409 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/facility/bacnet-objects/{id} [put]
func (h *BacnetObjectHandler) UpdateBacnetObject(c *gin.Context) {
	id, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}

	var req dto.UpdateBacnetObjectRequest
	if !bindJSON(c, &req) {
		return
	}
	if req.FieldDeviceID != nil && req.ObjectDataID != nil {
		respondLocalizedError(c, http.StatusBadRequest, "validation_error", "facility.validation_error")
		return
	}

	existing, err := h.service.GetByID(id)
	if err != nil {
		if respondLocalizedNotFoundIf(c, err, "facility.bacnet_object_not_found") {
			return
		}
		respondLocalizedError(c, http.StatusInternalServerError, "fetch_failed", "facility.fetch_failed")
		return
	}

	applyBacnetObjectPatch(existing, req.BacnetObjectPatchInput)
	if req.FieldDeviceID != nil {
		existing.FieldDeviceID = req.FieldDeviceID
	}

	if err := h.service.Update(existing, req.ObjectDataID); err != nil {
		if ve, ok := domain.AsValidationError(err); ok {
			respondValidationError(c, ve.Fields)
			return
		}
		if errors.Is(err, domain.ErrInvalidArgument) {
			respondInvalidArgument(c, err.Error())
			return
		}
		if errors.Is(err, domain.ErrNotFound) {
			respondInvalidReference(c)
			return
		}
		if errors.Is(err, domain.ErrConflict) {
			respondLocalizedConflict(c, "facility.entity_conflict")
			return
		}
		respondLocalizedError(c, http.StatusInternalServerError, "update_failed", "facility.update_failed")
		return
	}

	c.JSON(http.StatusOK, toBacnetObjectResponse(*existing))
}
