package wire

import (
	"time"

	"github.com/besart951/go_infra_link/backend/internal/handler"
	facilityhandler "github.com/besart951/go_infra_link/backend/internal/handler/facility"
)

// NewHandlers creates all HTTP handler instances from services.
func NewHandlers(services *Services, cookieSettings handler.CookieSettings, devAuthCfg DevAuthConfig) *handler.Handlers {
	return &handler.Handlers{
		ProjectHandler: handler.NewProjectHandler(services.Project),
		UserHandler:    handler.NewUserHandler(services.User),
		TeamHandler:    handler.NewTeamHandler(services.Team),
		AdminHandler:   handler.NewAdminHandler(services.Admin, services.Auth),
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

		FacilityBuildingHandler:       facilityhandler.NewBuildingHandler(services.FacilityBuilding),
		FacilitySystemTypeHandler:     facilityhandler.NewSystemTypeHandler(services.FacilitySystemType),
		FacilitySystemPartHandler:     facilityhandler.NewSystemPartHandler(services.FacilitySystemPart),
		FacilitySpecificationHandler:  facilityhandler.NewSpecificationHandler(services.FacilitySpecification),
		FacilityApparatHandler:        facilityhandler.NewApparatHandler(services.FacilityApparat),
		FacilityControlCabinetHandler: facilityhandler.NewControlCabinetHandler(services.FacilityControlCabinet),
		FacilityFieldDeviceHandler:    facilityhandler.NewFieldDeviceHandler(services.FacilityFieldDevice),
		FacilityBacnetObjectHandler:   facilityhandler.NewBacnetObjectHandler(services.FacilityBacnetObject),
		FacilitySPSControllerHandler:  facilityhandler.NewSPSControllerHandler(services.FacilitySPSController),
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
