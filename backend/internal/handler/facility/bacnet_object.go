package facility

import (
	"net/http"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	dto "github.com/besart951/go_infra_link/backend/internal/handler/dto/facility"
	"github.com/besart951/go_infra_link/backend/internal/handlerutil"
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

	if err := h.service.CreateWithParent(c.Request.Context(), obj, req.FieldDeviceID, req.ObjectDataID); err != nil {
		respondLocalizedDomainError(c, err, "creation_failed", "facility.creation_failed",
			localizedInvalidArgument("facility.exactly_one_required"),
			localizedInvalidReference(),
			localizedConflict("facility.entity_conflict"),
		)
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

	ctx := c.Request.Context()

	existing, err := h.service.GetByID(ctx, id)
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

	if err := h.service.Update(ctx, existing, req.ObjectDataID); err != nil {
		respondLocalizedDomainError(c, err, "update_failed", "facility.update_failed",
			handlerutil.MapError(domain.ErrInvalidArgument, handlerutil.PlainError(http.StatusBadRequest, "validation_error", err.Error())),
			localizedInvalidReference(),
			localizedConflict("facility.entity_conflict"),
		)
		return
	}

	c.JSON(http.StatusOK, toBacnetObjectResponse(*existing))
}
