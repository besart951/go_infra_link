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
	authz    middleware.AuthorizationChecker
}

func NewFacilityReferenceDataStreamHandler(streamer FacilityReferenceDataStreamer) *FacilityReferenceDataStreamHandler {
	return &FacilityReferenceDataStreamHandler{streamer: streamer}
}

func (h *FacilityReferenceDataStreamHandler) SetAuthorizationChecker(authz middleware.AuthorizationChecker) {
	h.authz = authz
}

// StreamFacilityReferenceData godoc
// @Summary Stream facility reference-data changes
// @Description Upgrades the authenticated request to the shared facility WebSocket. `facility_reference_data.changed` tells authorized clients to refresh cached apparats and system parts. `facility.changed` carries authorized facility resource changes with an action, IDs, actor and timestamp. User-scoped `facility.job.progress` events contain job type, status, stage and 0-100 progress; copy jobs temporarily also emit the legacy `facility.copy_job.progress` alias. Events are delivered only to the user that started the job.
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
	h.streamer.Stream(c.Writer, c.Request, userID, h.readableResources(c))
}

func (h *FacilityReferenceDataStreamHandler) readableResources(c *gin.Context) map[string]struct{} {
	if h.authz == nil {
		return nil
	}
	role, ok := middleware.GetUserRole(c)
	if !ok {
		return map[string]struct{}{}
	}
	readable := make(map[string]struct{}, len(facilityResourceCatalog))
	for _, definition := range facilityResourceCatalog {
		allowed, err := h.authz.HasPermission(c.Request.Context(), role, definition.readPermission)
		if err == nil && allowed {
			readable[definition.name] = struct{}{}
		}
	}
	return readable
}
