package facility

import (
	"net/http"

	dto "github.com/besart951/go_infra_link/backend/internal/handler/dto/facility"
	"github.com/besart951/go_infra_link/backend/internal/handler/middleware"
	"github.com/gin-gonic/gin"
)

type facilityReferenceDataErrorResponse = dto.ErrorResponse

type FacilityReferenceDataStreamHandler struct {
	streamer FacilityReferenceDataStreamer
}

func NewFacilityReferenceDataStreamHandler(streamer FacilityReferenceDataStreamer) *FacilityReferenceDataStreamHandler {
	return &FacilityReferenceDataStreamHandler{streamer: streamer}
}

// StreamFacilityReferenceData godoc
// @Summary Stream facility reference-data changes
// @Description Upgrades the authenticated request to the shared facility WebSocket. `facility_reference_data.changed` tells authorized clients to refresh cached apparats and system parts. User-scoped `facility.copy_job.progress` events contain a copy job ID, status, stage and 0-100 progress; they are only delivered to the user that started the job.
// @Tags facility-reference-data
// @Success 101
// @Failure 401 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /api/v1/facility/reference-data/stream [get]
func (h *FacilityReferenceDataStreamHandler) StreamFacilityReferenceData(c *gin.Context) {
	if h.streamer == nil {
		respondLocalizedError(c, http.StatusNotFound, "not_found", "facility.fetch_failed")
		return
	}
	userID, ok := middleware.GetUserID(c)
	if !ok {
		respondLocalizedError(c, http.StatusUnauthorized, "unauthorized", "errors.unauthorized")
		return
	}
	h.streamer.Stream(c.Writer, c.Request, userID)
}
