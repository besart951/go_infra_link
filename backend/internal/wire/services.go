package wire

import (
	"time"

	authservice "github.com/besart951/go_infra_link/backend/internal/service/auth"
	facilityservice "github.com/besart951/go_infra_link/backend/internal/service/facility"
	passwordsvc "github.com/besart951/go_infra_link/backend/internal/service/password"
	projectservice "github.com/besart951/go_infra_link/backend/internal/service/project"
	userservice "github.com/besart951/go_infra_link/backend/internal/service/user"
)

// Services holds all service instances.
type Services struct {
	Project  *projectservice.Service
	User     *userservice.Service
	Auth     *authservice.Service
	JWT      authservice.JWTService
	Password passwordsvc.Service

	FacilityBuilding       *facilityservice.BuildingService
	FacilitySystemType     *facilityservice.SystemTypeService
	FacilitySystemPart     *facilityservice.SystemPartService
	FacilitySpecification  *facilityservice.SpecificationService
	FacilityApparat        *facilityservice.ApparatService
	FacilityControlCabinet *facilityservice.ControlCabinetService
	FacilityFieldDevice    *facilityservice.FieldDeviceService
	FacilitySPSController  *facilityservice.SPSControllerService
}

// ServiceConfig contains configuration for services.
type ServiceConfig struct {
	JWTSecret       string
	Issuer          string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
}

// NewServices creates all service instances from repositories and configuration.
func NewServices(repos *Repositories, cfg ServiceConfig) *Services {
	passwordService := passwordsvc.New()
	jwtService := authservice.NewJWTService(cfg.JWTSecret, cfg.Issuer)

	return &Services{
		Project:  projectservice.New(repos.Project),
		User:     userservice.New(repos.User, passwordService),
		Password: passwordService,
		JWT:      jwtService,
		Auth: authservice.NewService(
			jwtService,
			repos.User,
			repos.UserEmail,
			repos.RefreshToken,
			passwordService,
			cfg.AccessTokenTTL,
			cfg.RefreshTokenTTL,
			cfg.Issuer,
		),

		FacilityBuilding:       facilityservice.NewBuildingService(repos.FacilityBuildings),
		FacilitySystemType:     facilityservice.NewSystemTypeService(repos.FacilitySystemTypes),
		FacilitySystemPart:     facilityservice.NewSystemPartService(repos.FacilitySystemParts),
		FacilitySpecification:  facilityservice.NewSpecificationService(repos.FacilitySpecifications),
		FacilityApparat:        facilityservice.NewApparatService(repos.FacilityApparats),
		FacilityControlCabinet: facilityservice.NewControlCabinetService(repos.FacilityControlCabinet),
		FacilityFieldDevice:    facilityservice.NewFieldDeviceService(repos.FacilityFieldDevices),
		FacilitySPSController:  facilityservice.NewSPSControllerService(repos.FacilitySPSControllers),
	}
}
