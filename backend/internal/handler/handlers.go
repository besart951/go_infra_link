package handler

import (
	authhandler "github.com/besart951/go_infra_link/backend/internal/handler/auth"
	dashboardhandler "github.com/besart951/go_infra_link/backend/internal/handler/dashboard"
	facilityhandler "github.com/besart951/go_infra_link/backend/internal/handler/facility"
	i18nhandler "github.com/besart951/go_infra_link/backend/internal/handler/i18n"
	notificationhandler "github.com/besart951/go_infra_link/backend/internal/handler/notification"
	projecthandler "github.com/besart951/go_infra_link/backend/internal/handler/project"
	teamhandler "github.com/besart951/go_infra_link/backend/internal/handler/team"
	userhandler "github.com/besart951/go_infra_link/backend/internal/handler/user"
)

type Handlers struct {
	Auth         *authhandler.AuthHandler
	Dashboard    *dashboardhandler.DashboardHandler
	I18n         *i18nhandler.I18nHandler
	Notification *notificationhandler.NotificationSettingsHandler
	Project      *projecthandler.Handlers
	Team         *teamhandler.TeamHandler
	User         *userhandler.Handlers
	Facility     *facilityhandler.Handlers
}
