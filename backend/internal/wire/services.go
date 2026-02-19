package wire

import (
	"path/filepath"
	"time"

	domainAuth "github.com/besart951/go_infra_link/backend/internal/domain/auth"
	domainUser "github.com/besart951/go_infra_link/backend/internal/domain/user"
	exportinfra "github.com/besart951/go_infra_link/backend/internal/infrastructure/exporting"
	adminservice "github.com/besart951/go_infra_link/backend/internal/service/admin"
	authservice "github.com/besart951/go_infra_link/backend/internal/service/auth"
	exportservice "github.com/besart951/go_infra_link/backend/internal/service/exporting"
	facilityservice "github.com/besart951/go_infra_link/backend/internal/service/facility"
	passwordsvc "github.com/besart951/go_infra_link/backend/internal/service/password"
	phaseservice "github.com/besart951/go_infra_link/backend/internal/service/phase"
	projectservice "github.com/besart951/go_infra_link/backend/internal/service/project"
	rbacservice "github.com/besart951/go_infra_link/backend/internal/service/rbac"
	teamservice "github.com/besart951/go_infra_link/backend/internal/service/team"
	userservice "github.com/besart951/go_infra_link/backend/internal/service/user"
)

// Services holds all service instances.
type Services struct {
	Project         *projectservice.Service
	Phase           *phaseservice.Service
	PhasePermission *phaseservice.PermissionService
	User            *userservice.Service
	Auth            *authservice.Service
	JWT             domainAuth.TokenService
	RBAC            *rbacservice.Service
	Team            *teamservice.Service
	Admin           *adminservice.Service
	Password        domainUser.PasswordHasher
	Export          *exportservice.Service

	Facility *facilityservice.Services
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
	rbacSvc := rbacservice.New(repos.User, repos.TeamMember, repos.Permissions, repos.RolePermissions)
	facilityServices := facilityservice.NewServices(facilityservice.Repositories{
		Buildings:                repos.FacilityBuildings,
		SystemTypes:              repos.FacilitySystemTypes,
		SystemParts:              repos.FacilitySystemParts,
		Specifications:           repos.FacilitySpecifications,
		Apparats:                 repos.FacilityApparats,
		ControlCabinets:          repos.FacilityControlCabinet,
		FieldDevices:             repos.FacilityFieldDevices,
		SPSControllers:           repos.FacilitySPSControllers,
		SPSControllerSystemTypes: repos.FacilitySPSControllerSystemTypes,
		BacnetObjects:            repos.FacilityBacnetObjects,
		ObjectData:               repos.FacilityObjectData,
		ObjectDataBacnetObjects:  repos.FacilityObjectDataBacnetObjects,
		StateTexts:               repos.FacilityStateTexts,
		NotificationClasses:      repos.FacilityNotificationClasses,
		AlarmDefinitions:         repos.FacilityAlarmDefinitions,
		Units:                    repos.FacilityUnits,
		AlarmFields:              repos.FacilityAlarmFields,
		AlarmTypes:               repos.FacilityAlarmTypes,
		AlarmTypeFields:          repos.FacilityAlarmTypeFields,
		BacnetObjectAlarmValues:  repos.FacilityBacnetObjectAlarmValues,
	})

	jobStore := exportinfra.NewMemoryJobStore()
	fileStore, err := exportinfra.NewLocalFileStore(filepath.Join("data", "exports"))
	if err != nil {
		panic(err)
	}
	dataProvider := exportinfra.NewDataProvider(repos.FacilityFieldDevices, repos.FacilitySPSControllers, repos.FacilityControlCabinet)
	excelGenerator := exportinfra.NewExcelizeGenerator()
	exportSvc := exportservice.NewService(
		dataProvider,
		excelGenerator,
		excelGenerator,
		jobStore,
		fileStore,
		exportservice.Config{
			QueueSize:             200,
			MaxConcurrent:         1,
			SingleFileDeviceLimit: 5000,
			PageSize:              1000,
		},
	)

	return &Services{
		Project: projectservice.New(
			repos.Project,
			repos.ProjectControlCabinets,
			repos.ProjectSPSControllers,
			repos.ProjectFieldDevices,
			repos.User,
			repos.FacilityObjectData,
			repos.FacilityBacnetObjects,
		),
		Phase:           phaseservice.NewPhaseService(repos.Phase),
		PhasePermission: phaseservice.NewPhasePermissionService(repos.PhasePermission),
		User:            userservice.New(repos.User, passwordService),
		Password:        passwordService,
		JWT:             jwtService,
		RBAC:            rbacSvc,
		Team:            teamservice.New(repos.Team, repos.TeamMember),
		Admin:           adminservice.New(repos.User),
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
		Export:   exportSvc,
		Facility: facilityServices,
	}
}
