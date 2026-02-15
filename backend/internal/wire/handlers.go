package wire

import (
	"time"

	"github.com/besart951/go_infra_link/backend/internal/handler"
	facilityhandler "github.com/besart951/go_infra_link/backend/internal/handler/facility"
	"github.com/besart951/go_infra_link/backend/pkg/i18n"
)

// NewHandlers creates all HTTP handler instances from services.
func NewHandlers(services *Services, cookieSettings handler.CookieSettings, i18nLoader *i18n.Loader, devAuthCfg DevAuthConfig) *handler.Handlers {
	facilityHandlers := facilityhandler.NewHandlers(facilityhandler.ServiceDeps{
		Building:                services.Facility.Building,
		SystemType:              services.Facility.SystemType,
		SystemPart:              services.Facility.SystemPart,
		Apparat:                 services.Facility.Apparat,
		ControlCabinet:          services.Facility.ControlCabinet,
		FieldDevice:             services.Facility.FieldDevice,
		BacnetObject:            services.Facility.BacnetObject,
		SPSController:           services.Facility.SPSController,
		StateText:               services.Facility.StateText,
		NotificationClass:       services.Facility.NotificationClass,
		AlarmDefinition:         services.Facility.AlarmDefinition,
		ObjectData:              services.Facility.ObjectData,
		SPSControllerSystemType: services.Facility.SPSControllerSystemType,
		Export:                  services.Export,
	})

	return &handler.Handlers{
		ProjectHandler:         handler.NewProjectHandler(services.Project),
		PhaseHandler:           handler.NewPhaseHandler(services.Phase),
		PhasePermissionHandler: handler.NewPhasePermissionHandler(services.PhasePermission),
		UserHandler:            handler.NewUserHandler(services.User, services.RBAC),
		TeamHandler:            handler.NewTeamHandler(services.Team),
		AdminHandler:           handler.NewAdminHandler(services.Admin, services.Auth),
		RoleHandler:            handler.NewRoleHandler(services.RBAC),
		PermissionHandler:      handler.NewPermissionHandler(services.RBAC),
		I18nHandler:            handler.NewI18nHandler(i18nLoader),
		AuthHandler: handler.NewAuthHandler(
			services.Auth,
			services.User,
			services.RBAC,
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
		FacilityApparatHandler:        facilityHandlers.Apparat,
		FacilityControlCabinetHandler: facilityHandlers.ControlCabinet,
		FacilityFieldDeviceHandler:    facilityHandlers.FieldDevice,
		FacilityBacnetObjectHandler:   facilityHandlers.BacnetObject,
		FacilitySPSControllerHandler:  facilityHandlers.SPSController,
		FacilityValidationHandler:     facilityHandlers.Validation,

		FacilityStateTextHandler:               facilityHandlers.StateText,
		FacilityNotificationClassHandler:       facilityHandlers.NotificationClass,
		FacilityAlarmDefinitionHandler:         facilityHandlers.AlarmDefinition,
		FacilityObjectDataHandler:              facilityHandlers.ObjectData,
		FacilitySPSControllerSystemTypeHandler: facilityHandlers.SPSControllerSystemType,
		FacilityExportHandler:                  facilityHandlers.Export,
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
