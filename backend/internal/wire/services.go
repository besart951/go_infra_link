package wire

import (
	"time"

	adminservice "github.com/besart951/go_infra_link/backend/internal/service/admin"
	authservice "github.com/besart951/go_infra_link/backend/internal/service/auth"
	facilityservice "github.com/besart951/go_infra_link/backend/internal/service/facility"
	passwordsvc "github.com/besart951/go_infra_link/backend/internal/service/password"
	projectservice "github.com/besart951/go_infra_link/backend/internal/service/project"
	rbacservice "github.com/besart951/go_infra_link/backend/internal/service/rbac"
	teamservice "github.com/besart951/go_infra_link/backend/internal/service/team"
	userservice "github.com/besart951/go_infra_link/backend/internal/service/user"
)

// Services holds all service instances.
type Services struct {
	Project  *projectservice.Service
	User     *userservice.Service
	Auth     *authservice.Service
	JWT      authservice.JWTService
	RBAC     *rbacservice.Service
	Team     *teamservice.Service
	Admin    *adminservice.Service
	Password passwordsvc.Service

	FacilityBuilding       *facilityservice.BuildingService
	FacilitySystemType     *facilityservice.SystemTypeService
	FacilitySystemPart     *facilityservice.SystemPartService
	FacilitySpecification  *facilityservice.SpecificationService
	FacilityApparat        *facilityservice.ApparatService
	FacilityControlCabinet *facilityservice.ControlCabinetService
	FacilityFieldDevice    *facilityservice.FieldDeviceService
	FacilityBacnetObject   *facilityservice.BacnetObjectService
	FacilitySPSController  *facilityservice.SPSControllerService

	FacilityStateText               *facilityservice.StateTextService
	FacilityNotificationClass       *facilityservice.NotificationClassService
	FacilityAlarmDefinition         *facilityservice.AlarmDefinitionService
	FacilityObjectData              *facilityservice.ObjectDataService
	FacilitySPSControllerSystemType *facilityservice.SPSControllerSystemTypeService
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
	rbacSvc := rbacservice.New(repos.User, repos.TeamMember)

	return &Services{
		Project:  projectservice.New(repos.Project, repos.FacilityObjectData, repos.FacilityBacnetObjects),
		User:     userservice.New(repos.User, passwordService),
		Password: passwordService,
		JWT:      jwtService,
		RBAC:     rbacSvc,
		Team:     teamservice.New(repos.Team, repos.TeamMember),
		Admin:    adminservice.New(repos.User),
		Auth: authservice.NewService(
			jwtService,
			repos.User,
			repos.UserEmail,
			repos.RefreshToken,
			repos.LoginAttempt,
			repos.PasswordReset,
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
		FacilityControlCabinet: facilityservice.NewControlCabinetService(repos.FacilityControlCabinet, repos.FacilityBuildings),
		FacilityFieldDevice: facilityservice.NewFieldDeviceService(
			repos.FacilityFieldDevices,
			repos.FacilitySPSControllerSystemTypes,
			repos.FacilitySPSControllers,
			repos.FacilityControlCabinet,
			repos.FacilitySystemTypes,
			repos.FacilityBuildings,
			repos.FacilityApparats,
			repos.FacilitySystemParts,
			repos.FacilitySpecifications,
			repos.FacilityBacnetObjects,
			repos.FacilityObjectData,
		),
		FacilityBacnetObject: facilityservice.NewBacnetObjectService(
			repos.FacilityBacnetObjects,
			repos.FacilityFieldDevices,
			repos.FacilityObjectData,
			repos.FacilityObjectDataBacnetObjects,
		),
		FacilitySPSController: facilityservice.NewSPSControllerService(
			repos.FacilitySPSControllers,
			repos.FacilityControlCabinet,
			repos.FacilitySystemTypes,
			repos.FacilitySPSControllerSystemTypes,
		),

		FacilityStateText:               facilityservice.NewStateTextService(repos.FacilityStateTexts),
		FacilityNotificationClass:       facilityservice.NewNotificationClassService(repos.FacilityNotificationClasses),
		FacilityAlarmDefinition:         facilityservice.NewAlarmDefinitionService(repos.FacilityAlarmDefinitions),
		FacilityObjectData:              facilityservice.NewObjectDataService(repos.FacilityObjectData),
		FacilitySPSControllerSystemType: facilityservice.NewSPSControllerSystemTypeService(repos.FacilitySPSControllerSystemTypes),
	}
}
