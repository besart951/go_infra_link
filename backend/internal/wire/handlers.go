package wire

import (
	"time"

	"github.com/besart951/go_infra_link/backend/internal/handler"
	authhandler "github.com/besart951/go_infra_link/backend/internal/handler/auth"
	dashboardhandler "github.com/besart951/go_infra_link/backend/internal/handler/dashboard"
	historyhandler "github.com/besart951/go_infra_link/backend/internal/handler/history"
	i18nhandler "github.com/besart951/go_infra_link/backend/internal/handler/i18n"
	notificationhandler "github.com/besart951/go_infra_link/backend/internal/handler/notification"
	teamhandler "github.com/besart951/go_infra_link/backend/internal/handler/team"
	"github.com/besart951/go_infra_link/backend/pkg/i18n"
)

// NewHandlers creates all HTTP handler instances from services.
func NewHandlers(services *Services, runtime *RuntimeAdapters, cookieSettings authhandler.CookieSettings, i18nLoader *i18n.Loader, accessTokenTTL, refreshTokenTTL time.Duration) *handler.Handlers {
	runtime = runtimeOrDefault(runtime)

	projectHandlers := newProjectHandlers(services, runtime)

	facilityHandlers := newFacilityHandlers(services, projectHandlers.RefreshBroadcaster)
	userHandlers := newUserHandlers(services)

	authHandler := authhandler.NewAuthHandler(
		services.Auth,
		services.User,
		services.RBAC,
		services.JWT,
		accessTokenTTL,
		refreshTokenTTL,
		cookieSettings,
	)

	return &handler.Handlers{
		Auth:             authHandler,
		AuthRegistration: authhandler.NewRegistrationHandler(services.UserRegistration),
		Dashboard:        dashboardhandler.NewDashboardHandler(services.Dashboard),
		I18n:             i18nhandler.NewI18nHandler(i18nLoader),
		Notification:     notificationhandler.NewNotificationSettingsHandler(services.Notification, runtime.SystemNotificationStream),
		Project:          projectHandlers,
		Team:             teamhandler.NewTeamHandler(services.Team),
		User:             userHandlers,
		Facility:         facilityHandlers,
		History:          historyhandler.NewHandler(services.History),
	}
}
