package wire

import (
	"time"

	"github.com/besart951/go_infra_link/backend/internal/handler"
	authhandler "github.com/besart951/go_infra_link/backend/internal/handler/auth"
	dashboardhandler "github.com/besart951/go_infra_link/backend/internal/handler/dashboard"
	facilityhandler "github.com/besart951/go_infra_link/backend/internal/handler/facility"
	i18nhandler "github.com/besart951/go_infra_link/backend/internal/handler/i18n"
	notificationhandler "github.com/besart951/go_infra_link/backend/internal/handler/notification"
	projecthandler "github.com/besart951/go_infra_link/backend/internal/handler/project"
	teamhandler "github.com/besart951/go_infra_link/backend/internal/handler/team"
	userhandler "github.com/besart951/go_infra_link/backend/internal/handler/user"
	"github.com/besart951/go_infra_link/backend/pkg/i18n"
)

// NewHandlers creates all HTTP handler instances from services.
func NewHandlers(services *Services, cookieSettings authhandler.CookieSettings, i18nLoader *i18n.Loader, accessTokenTTL, refreshTokenTTL time.Duration) *handler.Handlers {
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
		AlarmType:               services.Facility.AlarmType,
		Unit:                    services.Facility.Unit,
		AlarmField:              services.Facility.AlarmField,
		AlarmTypeField:          services.Facility.AlarmTypeField,
		BacnetAlarm:             services.Facility.BacnetAlarmValue,
	})

	projectHandlers := projecthandler.NewHandlers(services.Project, services.Phase, services.Facility.FieldDevice)
	userHandlers := userhandler.NewHandlers(services.User, services.Admin, services.RBAC, services.RBAC, services.RBAC)

	return &handler.Handlers{
		Auth: authhandler.NewAuthHandler(
			services.Auth,
			services.User,
			services.RBAC,
			accessTokenTTL,
			refreshTokenTTL,
			cookieSettings,
		),
		Dashboard:    dashboardhandler.NewDashboardHandler(services.Dashboard),
		I18n:         i18nhandler.NewI18nHandler(i18nLoader),
		Notification: notificationhandler.NewNotificationSettingsHandler(services.Notification),
		Project:      projectHandlers,
		Team:         teamhandler.NewTeamHandler(services.Team),
		User:         userHandlers,
		Facility:     facilityHandlers,
	}
}
