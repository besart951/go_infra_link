package wire

import (
	"time"

	"github.com/besart951/go_infra_link/backend/internal/handler"
	facilityhandler "github.com/besart951/go_infra_link/backend/internal/handler/facility"
	"github.com/besart951/go_infra_link/backend/pkg/i18n"
)

// NewHandlers creates all HTTP handler instances from services.
func NewHandlers(services *Services, cookieSettings handler.CookieSettings, i18nLoader *i18n.Loader, accessTokenTTL, refreshTokenTTL time.Duration) *handler.Handlers {
	projectEvents := handler.NewProjectEventHub()

	facilityHandlers := facilityhandler.NewHandlers(facilityhandler.ServiceDeps{
		Building:                services.Facility.Building,
		SystemType:              services.Facility.SystemType,
		SystemPart:              services.Facility.SystemPart,
		Apparat:                 services.Facility.Apparat,
		ControlCabinet:          services.Facility.ControlCabinet,
		FieldDevice:             services.Facility.FieldDevice,
		ProjectAccess:           services.Project,
		BacnetObject:            services.Facility.BacnetObject,
		SPSController:           services.Facility.SPSController,
		StateText:               services.Facility.StateText,
		NotificationClass:       services.Facility.NotificationClass,
		AlarmDefinition:         services.Facility.AlarmDefinition,
		ObjectData:              services.Facility.ObjectData,
		SPSControllerSystemType: services.Facility.SPSControllerSystemType,
		Export:                  services.Export,
		AlarmType:               services.Facility.AlarmType,
		Unit:                    services.Facility.Unit,
		AlarmField:              services.Facility.AlarmField,
		AlarmTypeField:          services.Facility.AlarmTypeField,
		BacnetAlarm:             services.Facility.BacnetAlarmValue,
	})

	return &handler.Handlers{
		ProjectHandler:    handler.NewProjectHandler(services.Project, projectEvents),
		DashboardHandler:  handler.NewDashboardHandler(services.Dashboard),
		PhaseHandler:      handler.NewPhaseHandler(services.Phase),
		UserHandler:       handler.NewUserHandler(services.User, services.RBAC),
		TeamHandler:       handler.NewTeamHandler(services.Team),
		AdminHandler:      handler.NewAdminHandler(services.Admin),
		RoleHandler:       handler.NewRoleHandler(services.RBAC),
		PermissionHandler: handler.NewPermissionHandler(services.RBAC),
		I18nHandler:       handler.NewI18nHandler(i18nLoader),
		AuthHandler: handler.NewAuthHandler(
			services.Auth,
			services.User,
			services.RBAC,
			accessTokenTTL,
			refreshTokenTTL,
			cookieSettings,
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
		FacilityAlarmTypeHandler:               facilityHandlers.AlarmType,
		FacilityUnitHandler:                    facilityHandlers.Unit,
		FacilityAlarmFieldHandler:              facilityHandlers.AlarmField,
		FacilityAlarmTypeFieldHandler:          facilityHandlers.AlarmTypeField,
		FacilityBacnetAlarmHandler:             facilityHandlers.BacnetAlarm,
	}
}
