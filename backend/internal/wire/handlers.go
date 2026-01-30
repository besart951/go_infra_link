package wire

import (
	"time"

	"github.com/besart951/go_infra_link/backend/internal/handler"
	facilityhandler "github.com/besart951/go_infra_link/backend/internal/handler/facility"
)

// NewHandlers creates all HTTP handler instances from services.
func NewHandlers(services *Services, cookieSettings handler.CookieSettings, devAuthCfg DevAuthConfig) *handler.Handlers {
	facilityHandlers := facilityhandler.NewHandlers(services.Facility)

	return &handler.Handlers{
		ProjectHandler:         handler.NewProjectHandler(services.Project),
		PhaseHandler:           handler.NewPhaseHandler(services.Phase),
		PhasePermissionHandler: handler.NewPhasePermissionHandler(services.PhasePermission),
		UserHandler:            handler.NewUserHandler(services.User, services.RBAC),
		TeamHandler:            handler.NewTeamHandler(services.Team),
		AdminHandler:           handler.NewAdminHandler(services.Admin, services.Auth),
		AuthHandler: handler.NewAuthHandler(
			services.Auth,
			services.User,
			devAuthCfg.AccessTokenTTL,
			devAuthCfg.RefreshTokenTTL,
			cookieSettings,
			devAuthCfg.Enabled,
			devAuthCfg.Email,
			devAuthCfg.Password,
		),

		FacilityBuildingHandler:       facilityHandlers.Building,
		FacilitySystemTypeHandler:     facilityHandlers.SystemType,
		FacilitySystemPartHandler:     facilityHandlers.SystemPart,
		FacilitySpecificationHandler:  facilityHandlers.Specification,
		FacilityApparatHandler:        facilityHandlers.Apparat,
		FacilityControlCabinetHandler: facilityHandlers.ControlCabinet,
		FacilityFieldDeviceHandler:    facilityHandlers.FieldDevice,
		FacilityBacnetObjectHandler:   facilityHandlers.BacnetObject,
		FacilitySPSControllerHandler:  facilityHandlers.SPSController,

		FacilityStateTextHandler:               facilityHandlers.StateText,
		FacilityNotificationClassHandler:       facilityHandlers.NotificationClass,
		FacilityAlarmDefinitionHandler:         facilityHandlers.AlarmDefinition,
		FacilityObjectDataHandler:              facilityHandlers.ObjectData,
		FacilitySPSControllerSystemTypeHandler: facilityHandlers.SPSControllerSystemType,
	}
}

// DevAuthConfig holds development authentication bypass configuration.
type DevAuthConfig struct {
	Enabled         bool
	Email           string
	Password        string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
}
