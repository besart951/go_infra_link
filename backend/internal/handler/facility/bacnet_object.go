package facility

import (
	"net/http"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	dto "github.com/besart951/go_infra_link/backend/internal/handler/dto/facility"
	"github.com/besart951/go_infra_link/backend/internal/handlerutil"
	"github.com/gin-gonic/gin"
)

type BacnetObjectHandler struct {
	service       BacnetObjectService
	collaboration ProjectFieldDeviceChangeBroadcaster
}

func NewBacnetObjectHandler(service BacnetObjectService, broadcasters ...ProjectRefreshBroadcaster) *BacnetObjectHandler {
	h := &BacnetObjectHandler{service: service}
	if len(broadcasters) > 0 {
		h.collaboration, _ = broadcasters[0].(ProjectFieldDeviceChangeBroadcaster)
	}
	return h
}

// GetBacnetObject godoc
// @Summary Get a BACnet instance or template by ID
// @Tags facility-bacnet-objects
// @Produce json
// @Param id path string true "BACnet Object ID"
// @Success 200 {object} dto.BacnetObjectResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /api/v1/facility/bacnet-objects/{id} [get]
func (h *BacnetObjectHandler) GetBacnetObject(c *gin.Context) {
	id, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}
	object, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		respondLocalizedDomainError(c, err, "fetch_failed", "facility.fetch_failed", localizedNotFound("facility.bacnet_object_not_found"))
		return
	}
	c.JSON(http.StatusOK, toBacnetObjectResponse(*object))
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
	if obj.FieldDeviceID != nil && h.collaboration != nil {
		h.collaboration.BroadcastFieldDeviceChange(c.Request.Context(), currentActorID(c), *obj.FieldDeviceID, "updated", dto.FieldDeviceBacnetObjectsChangedFields()...)
	}
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
	baseVersion := existing.Version
	previousFieldDeviceID := existing.FieldDeviceID
	if req.BaseVersion != nil {
		baseVersion = *req.BaseVersion
		existing.Version = *req.BaseVersion
	}

	applyBacnetObjectPatch(existing, req.BacnetObjectPatchInput)
	if req.FieldDeviceID != nil {
		existing.FieldDeviceID = req.FieldDeviceID
	}

	if err := h.service.Update(ctx, existing, req.ObjectDataID); err != nil {
		if current, getErr := h.service.GetByID(ctx, id); getErr == nil && respondWriteConflict(c, err, "field_device", uuidValue(current.FieldDeviceID), baseVersion, req.FieldDeviceConflictFields(id.String()), current.Version, toBacnetObjectResponse(*current)) {
			return
		}
		respondLocalizedDomainError(c, err, "update_failed", "facility.update_failed",
			handlerutil.MapError(domain.ErrInvalidArgument, handlerutil.PlainError(http.StatusBadRequest, "validation_error", err.Error())),
			localizedInvalidReference(),
			localizedConflict("facility.entity_conflict"),
		)
		return
	}

	c.JSON(http.StatusOK, toBacnetObjectResponse(*existing))
	if existing.FieldDeviceID != nil && h.collaboration != nil {
		h.collaboration.BroadcastFieldDeviceChange(ctx, currentActorID(c), *existing.FieldDeviceID, "updated", dto.FieldDeviceBacnetObjectsChangedFields()...)
		if previousFieldDeviceID != nil && *previousFieldDeviceID != *existing.FieldDeviceID {
			h.collaboration.BroadcastFieldDeviceChange(ctx, currentActorID(c), *previousFieldDeviceID, "updated", dto.FieldDeviceBacnetObjectsChangedFields()...)
		}
	}
}

// DeleteBacnetObject deletes one BACnet object.
func (h *BacnetObjectHandler) DeleteBacnetObject(c *gin.Context) {
	id, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}
	existing, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		respondLocalizedDomainError(c, err, "deletion_failed", "facility.deletion_failed", localizedNotFound("facility.bacnet_object_not_found"))
		return
	}
	if err := h.service.DeleteByID(c.Request.Context(), id); err != nil {
		respondLocalizedDomainError(c, err, "deletion_failed", "facility.deletion_failed",
			localizedNotFound("facility.bacnet_object_not_found"),
		)
		return
	}
	if existing.FieldDeviceID != nil && h.collaboration != nil {
		h.collaboration.BroadcastFieldDeviceChange(c.Request.Context(), currentActorID(c), *existing.FieldDeviceID, "updated", dto.FieldDeviceBacnetObjectsChangedFields()...)
	}
	c.Status(http.StatusNoContent)
}
