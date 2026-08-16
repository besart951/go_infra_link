package facility

import (
	"net/http"

	domainUser "github.com/besart951/go_infra_link/backend/internal/domain/user"
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
// @Description Upgrades the authenticated request to the shared facility WebSocket. `facility_reference_data.changed` tells authorized clients to refresh cached apparats and system parts. `facility.changed` carries authorized facility resource changes with an action, IDs, actor and timestamp. User-scoped `facility.copy_job.progress` events contain a copy job ID, status, stage and 0-100 progress; they are only delivered to the user that started the job.
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
	permissions := map[string]string{
		"buildings":                   domainUser.PermissionBuildingRead,
		"system_types":                domainUser.PermissionSystemTypeRead,
		"system_parts":                domainUser.PermissionSystemPartRead,
		"apparats":                    domainUser.PermissionApparatRead,
		"control_cabinets":            domainUser.PermissionControlCabinetRead,
		"sps_controllers":             domainUser.PermissionSPSControllerRead,
		"sps_controller_system_types": domainUser.PermissionSPSControllerSystemTypeRead,
		"field_devices":               domainUser.PermissionFieldDeviceRead,
		"bacnet_objects":              domainUser.PermissionBacnetObjectRead,
		"object_data":                 domainUser.PermissionObjectDataRead,
		"state_texts":                 domainUser.PermissionStateTextRead,
		"notification_classes":        domainUser.PermissionNotificationClassRead,
		"alarm_definitions":           domainUser.PermissionAlarmDefinitionRead,
		"alarm_types":                 domainUser.PermissionAlarmTypeRead,
		"alarm_type_fields":           domainUser.PermissionAlarmFieldRead,
		"alarm_fields":                domainUser.PermissionAlarmFieldRead,
		"units":                       domainUser.PermissionUnitRead,
	}
	readable := make(map[string]struct{}, len(permissions))
	for resource, permission := range permissions {
		allowed, err := h.authz.HasPermission(c.Request.Context(), role, permission)
		if err == nil && allowed {
			readable[resource] = struct{}{}
		}
	}
	return readable
}
