package wire

import (
	"fmt"
	"path/filepath"
	"time"

	domainAuth "github.com/besart951/go_infra_link/backend/internal/domain/auth"
	domainUser "github.com/besart951/go_infra_link/backend/internal/domain/user"
	exportinfra "github.com/besart951/go_infra_link/backend/internal/infrastructure/exporting"
	adminservice "github.com/besart951/go_infra_link/backend/internal/service/admin"
	authservice "github.com/besart951/go_infra_link/backend/internal/service/auth"
	dashboardservice "github.com/besart951/go_infra_link/backend/internal/service/dashboard"
	exportservice "github.com/besart951/go_infra_link/backend/internal/service/exporting"
	facilityservice "github.com/besart951/go_infra_link/backend/internal/service/facility"
	notificationservice "github.com/besart951/go_infra_link/backend/internal/service/notification"
	passwordsvc "github.com/besart951/go_infra_link/backend/internal/service/password"
	phaseservice "github.com/besart951/go_infra_link/backend/internal/service/phase"
	projectservice "github.com/besart951/go_infra_link/backend/internal/service/project"
	rbacservice "github.com/besart951/go_infra_link/backend/internal/service/rbac"
	teamservice "github.com/besart951/go_infra_link/backend/internal/service/team"
	userservice "github.com/besart951/go_infra_link/backend/internal/service/user"
	userdirectoryservice "github.com/besart951/go_infra_link/backend/internal/service/userdirectory"
	"gorm.io/gorm"
)

// Services holds all service instances.
type Services struct {
	Project       *projectservice.Services
	Dashboard     *dashboardservice.Service
	Phase         *phaseservice.Service
	User          *userservice.Service
	Auth          *authservice.Service
	JWT           domainAuth.TokenService
	RBAC          *rbacservice.Service
	Team          *teamservice.Service
	Admin         *adminservice.Service
	UserDirectory *userdirectoryservice.Service
	Notification  *notificationservice.Service
	Password      domainUser.PasswordHasher
	Export        *exportservice.Service

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
func NewServices(gormDB *gorm.DB, repos *Repositories, cfg ServiceConfig) (*Services, error) {
	passwordService := passwordsvc.New()
	jwtService := authservice.NewJWTService(cfg.JWTSecret, cfg.Issuer)
	rbacSvc := rbacservice.New(repos.User, repos.TeamMember, repos.Permissions, repos.RolePermissions)
	facilityTxRunner := facilityservice.TxRunner(func(run func(tx *gorm.DB) error) error {
		return gormDB.Transaction(run)
	})
	facilityTxRepositories := func(tx *gorm.DB) (facilityservice.Repositories, error) {
		txRepos, err := NewRepositories(tx)
		if err != nil {
			return facilityservice.Repositories{}, fmt.Errorf("transaction repositories: %w", err)
		}
		return buildFacilityRepositories(txRepos), nil
	}
	facilityServices := facilityservice.NewServices(buildFacilityRepositories(repos), facilityservice.Config{
		TxRunner:       facilityTxRunner,
		TxRepositories: facilityTxRepositories,
	})

	jobStore := exportinfra.NewMemoryJobStore()
	fileStore, err := exportinfra.NewLocalFileStore(filepath.Join("data", "exports"))
	if err != nil {
		return nil, fmt.Errorf("export file store: %w", err)
	}
	dataProvider := exportinfra.NewDataProvider(
		repos.FacilityFieldDevices,
		repos.FacilitySpecifications,
		repos.FacilityBacnetObjects,
		repos.FacilitySPSControllers,
		repos.FacilityControlCabinet,
	)
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
	secretCipher, err := notificationservice.NewAESCipher(cfg.JWTSecret)
	if err != nil {
		return nil, fmt.Errorf("notification secret cipher: %w", err)
	}
	notificationSvc := notificationservice.New(
		repos.NotificationSMTPSettings,
		secretCipher,
		notificationservice.NewSMTPStrategy(),
	)

	return &Services{
		Project:       buildProjectServices(gormDB, repos, facilityServices),
		Dashboard:     dashboardservice.New(repos.Project, repos.Phase, repos.Team, repos.TeamMember, repos.User),
		Phase:         phaseservice.NewPhaseService(repos.Phase),
		User:          userservice.New(repos.User, passwordService),
		Password:      passwordService,
		JWT:           jwtService,
		RBAC:          rbacSvc,
		Team:          teamservice.New(repos.Team, repos.TeamMember),
		Admin:         adminservice.New(repos.User),
		UserDirectory: userdirectoryservice.New(repos.User, repos.Team, repos.TeamMember, repos.RolePermissions),
		Notification:  notificationSvc,
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
		Export:   exportSvc,
		Facility: facilityServices,
	}, nil
}

func buildFacilityRepositories(repos *Repositories) facilityservice.Repositories {
	return facilityservice.Repositories{
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
	}
}

func buildProjectDependencies(repos *Repositories, facilityServices *facilityservice.Services) projectservice.Dependencies {
	return projectservice.Dependencies{
		Projects:                 repos.Project,
		ProjectControlCabinets:   repos.ProjectControlCabinets,
		ProjectSPSControllers:    repos.ProjectSPSControllers,
		ProjectFieldDevices:      repos.ProjectFieldDevices,
		Users:                    repos.User,
		RolePermissions:          repos.RolePermissions,
		ObjectData:               repos.FacilityObjectData,
		BacnetObjects:            repos.FacilityBacnetObjects,
		Specifications:           repos.FacilitySpecifications,
		ControlCabinets:          repos.FacilityControlCabinet,
		SPSControllers:           repos.FacilitySPSControllers,
		SPSControllerSystemTypes: repos.FacilitySPSControllerSystemTypes,
		FieldDevices:             repos.FacilityFieldDevices,
		HierarchyCopier:          facilityServices.HierarchyCopier,
		FieldDeviceCreator:       facilityServices.FieldDevice,
	}
}

func buildProjectServices(gormDB *gorm.DB, repos *Repositories, facilityServices *facilityservice.Services) *projectservice.Services {
	txRunner := projectservice.TxRunner(func(run func(tx *gorm.DB) error) error {
		return gormDB.Transaction(run)
	})

	txDependencies := func(tx *gorm.DB) (projectservice.Dependencies, error) {
		txRepos, err := NewRepositories(tx)
		if err != nil {
			return projectservice.Dependencies{}, fmt.Errorf("transaction repositories: %w", err)
		}

		txFacilityServices := facilityservice.NewServices(buildFacilityRepositories(txRepos))
		return buildProjectDependencies(txRepos, txFacilityServices), nil
	}

	return projectservice.NewServices(buildProjectDependencies(repos, facilityServices), projectservice.Config{
		TxRunner:       txRunner,
		TxDependencies: txDependencies,
	})
}
