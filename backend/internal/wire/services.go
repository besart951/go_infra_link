package wire

import (
	"time"

	domainAuth "github.com/besart951/go_infra_link/backend/internal/domain/auth"
	domainUser "github.com/besart951/go_infra_link/backend/internal/domain/user"
	adminservice "github.com/besart951/go_infra_link/backend/internal/service/admin"
	authservice "github.com/besart951/go_infra_link/backend/internal/service/auth"
	dashboardservice "github.com/besart951/go_infra_link/backend/internal/service/dashboard"
	exportservice "github.com/besart951/go_infra_link/backend/internal/service/exporting"
	facilityservice "github.com/besart951/go_infra_link/backend/internal/service/facility"
	notificationservice "github.com/besart951/go_infra_link/backend/internal/service/notification"
	passwordsvc "github.com/besart951/go_infra_link/backend/internal/service/password"
	phaseservice "github.com/besart951/go_infra_link/backend/internal/service/phase"
	phasepermissionservice "github.com/besart951/go_infra_link/backend/internal/service/phasepermission"
	projectservice "github.com/besart951/go_infra_link/backend/internal/service/project"
	rbacservice "github.com/besart951/go_infra_link/backend/internal/service/rbac"
	teamservice "github.com/besart951/go_infra_link/backend/internal/service/team"
	userservice "github.com/besart951/go_infra_link/backend/internal/service/user"
	userdirectoryservice "github.com/besart951/go_infra_link/backend/internal/service/userdirectory"
	"gorm.io/gorm"
)

// Services holds all service instances.
type Services struct {
	Project         *projectservice.Services
	Dashboard       *dashboardservice.Service
	Phase           *phaseservice.Service
	PhasePermission *phasepermissionservice.Service
	User            *userservice.Service
	Auth            *authservice.Service
	JWT             domainAuth.TokenService
	RBAC            *rbacservice.Service
	Team            *teamservice.Service
	Admin           *adminservice.Service
	UserDirectory   *userdirectoryservice.Service
	Notification    *notificationservice.Service
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
	Runtime         *RuntimeAdapters
}

// NewServices creates all service instances from repositories and configuration.
func NewServices(gormDB *gorm.DB, repos *Repositories, cfg ServiceConfig) (*Services, error) {
	passwordService := passwordsvc.New()
	jwtService := authservice.NewJWTService(cfg.JWTSecret, cfg.Issuer)
	rbacSvc := rbacservice.New(repos.User, repos.TeamMember, repos.Permissions, repos.RolePermissions)
	facilityServices := newFacilityServices(gormDB, repos)

	exportSvc, err := newExportService(repos)
	if err != nil {
		return nil, err
	}
	notificationSvc, err := newNotificationService(repos, cfg)
	if err != nil {
		return nil, err
	}

	return &Services{
		Project:         newProjectServices(gormDB, repos, facilityServices),
		Dashboard:       dashboardservice.New(repos.Project, repos.Phase, repos.Team, repos.TeamMember, repos.User),
		Phase:           phaseservice.NewPhaseService(repos.Phase),
		PhasePermission: phasepermissionservice.New(repos.PhasePermissions, repos.Phase, repos.Permissions),
		User:            userservice.New(repos.User, passwordService),
		Password:        passwordService,
		JWT:             jwtService,
		RBAC:            rbacSvc,
		Team:            teamservice.New(repos.Team, repos.TeamMember),
		Admin:           adminservice.New(repos.User),
		UserDirectory:   userdirectoryservice.New(repos.User, repos.Team, repos.TeamMember, repos.RolePermissions),
		Notification:    notificationSvc,
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
