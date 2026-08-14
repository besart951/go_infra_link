package facility

import (
	"net/http"

	dto "github.com/besart951/go_infra_link/backend/internal/handler/dto/facility"
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
// @Description Upgrades the authenticated request to a WebSocket. Each event follows the `facility_reference_data.changed` contract and causes clients to refresh cached apparats and system parts through their authorized HTTP endpoints.
// @Tags facility-reference-data
// @Success 101
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /api/v1/facility/reference-data/stream [get]
func (h *FacilityReferenceDataStreamHandler) StreamFacilityReferenceData(c *gin.Context) {
	if h.streamer == nil {
		respondLocalizedError(c, http.StatusNotFound, "not_found", "facility.fetch_failed")
		return
	}
	h.streamer.Stream(c.Writer, c.Request)
}
