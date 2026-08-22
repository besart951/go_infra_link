package facility

import (
	"context"
	"errors"
	"net/http"

	fielddeviceimport "github.com/besart951/go_infra_link/backend/internal/application/fielddeviceimport"
	"github.com/besart951/go_infra_link/backend/internal/handler/middleware"
	"github.com/gin-gonic/gin"
)

const maxFieldDeviceImportBytes = int64(2 << 30)

type FieldDeviceImportService interface {
	Import(ctx context.Context, command fielddeviceimport.Command) (fielddeviceimport.Result, error)
}

type ImportHandler struct{ service FieldDeviceImportService }

func NewImportHandler(service FieldDeviceImportService) *ImportHandler {
	return &ImportHandler{service: service}
}

// ImportFieldDevices godoc
// @Summary Import a versioned field-device workbook
// @Tags Facility - Field Devices
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "Versioned XLSX export or ZIP workbook package"
// @Success 200 {object} fielddeviceimport.Result
// @Failure 422 {object} fielddeviceimport.Result
// @Router /api/v1/facility/imports/field-devices [post]
func (h *ImportHandler) ImportFieldDevices(c *gin.Context) {
	if h.service == nil {
		respondError(c, http.StatusServiceUnavailable, "import_unavailable", "Field-device import is unavailable")
		return
	}
	ownerID, ok := middleware.GetUserID(c)
	if !ok {
		respondLocalizedError(c, http.StatusUnauthorized, "unauthorized", "errors.unauthorized")
		return
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxFieldDeviceImportBytes)
	fileHeader, err := c.FormFile("file")
	if err != nil {
		respondError(c, http.StatusBadRequest, "invalid_import_file", err.Error())
		return
	}
	file, err := fileHeader.Open()
	if err != nil {
		respondError(c, http.StatusBadRequest, "invalid_import_file", err.Error())
		return
	}
	defer file.Close()
	result, err := h.service.Import(c.Request.Context(), fielddeviceimport.Command{OwnerID: ownerID, Source: file})
	if errors.Is(err, fielddeviceimport.ErrInvalidWorkbook) {
		c.JSON(http.StatusUnprocessableEntity, result)
		return
	}
	if err != nil {
		respondError(c, http.StatusInternalServerError, "field_device_import_failed", err.Error())
		return
	}
	c.JSON(http.StatusOK, result)
}
