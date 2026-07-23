package facility

import (
	"errors"
	"net/http"

	appbacnetobject "github.com/besart951/go_infra_link/backend/internal/application/facility/bacnetobject"
	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	dto "github.com/besart951/go_infra_link/backend/internal/handler/dto/facility"
	"github.com/besart951/go_infra_link/backend/internal/handlerutil"
	"github.com/gin-gonic/gin"
)

type BacnetObjectHandler struct {
	creator BacnetObjectCreator
	updater BacnetObjectUpdater
}

func NewBacnetObjectHandler(
	creator BacnetObjectCreator,
	updater BacnetObjectUpdater,
) *BacnetObjectHandler {
	return &BacnetObjectHandler{creator: creator, updater: updater}
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

	if h.creator == nil {
		respondLocalizedError(c, http.StatusInternalServerError, "creation_failed", "facility.creation_failed")
		return
	}

	var (
		obj *domainFacility.BacnetObject
		err error
	)
	switch {
	case req.FieldDeviceID != nil && req.ObjectDataID == nil:
		obj, err = h.creator.CreateForFieldDevice(
			c.Request.Context(),
			toBacnetObjectCreateForFieldDeviceCommand(*req.FieldDeviceID, req),
		)
	case req.FieldDeviceID == nil && req.ObjectDataID != nil:
		obj, err = h.creator.CreateForObjectData(
			c.Request.Context(),
			toBacnetObjectCreateForObjectDataCommand(*req.ObjectDataID, req),
		)
	default:
		err = domain.ErrInvalidArgument
	}

	if err != nil {
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
	if h.updater == nil {
		respondLocalizedError(c, http.StatusInternalServerError, "update_failed", "facility.update_failed")
		return
	}

	updated, err := h.updater.Update(
		c.Request.Context(),
		toBacnetObjectUpdateCommand(id, req),
	)
	if err != nil {
		var loadErr *appbacnetobject.LoadError
		if errors.As(err, &loadErr) {
			if respondLocalizedNotFoundIf(c, loadErr.Err, "facility.bacnet_object_not_found") {
				return
			}
			respondLocalizedError(c, http.StatusInternalServerError, "fetch_failed", "facility.fetch_failed")
			return
		}
		respondLocalizedDomainError(c, err, "update_failed", "facility.update_failed",
			handlerutil.MapError(domain.ErrInvalidArgument, handlerutil.PlainError(http.StatusBadRequest, "validation_error", err.Error())),
			localizedInvalidReference(),
			localizedConflict("facility.entity_conflict"),
		)
		return
	}

	c.JSON(http.StatusOK, toBacnetObjectResponse(*updated))
}
