package wire

import (
	"fmt"
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
	usermutationpolicy "github.com/besart951/go_infra_link/backend/internal/service/usermutationpolicy"
	userregistrationservice "github.com/besart951/go_infra_link/backend/internal/service/userregistration"
	"gorm.io/gorm"
)

// Services holds all service instances.
type Services struct {
	Project          *projectservice.Services
	Dashboard        *dashboardservice.Service
	Phase            *phaseservice.Service
	PhasePermission  *phasepermissionservice.Service
	User             *userservice.Service
	UserRegistration *userregistrationservice.Service
	Auth             *authservice.Service
	JWT              domainAuth.TokenService
	RBAC             *rbacservice.Service
	Team             *teamservice.Service
	Admin            *adminservice.Service
	UserDirectory    *userdirectoryservice.Service
	Notification     *notificationservice.Service
	Password         domainUser.PasswordHasher
	Export           *exportservice.Service
	History          HistoryRepository

	Facility *facilityservice.Services
}

// ServiceConfig contains configuration for services.
type ServiceConfig struct {
	JWTSecret       string
	Issuer          string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
	Export          exportservice.Config
	ExportDirectory string
	Runtime         *RuntimeAdapters
	AppPublicURL    string
}

type securityServices struct {
	jwt                domainAuth.TokenService
	rbac               *rbacservice.Service
	userMutationPolicy *usermutationpolicy.Policy
}

type userServices struct {
	user             *userservice.Service
	userRegistration *userregistrationservice.Service
	userDirectory    *userdirectoryservice.Service
	admin            *adminservice.Service
}

// NewServices creates all service instances from repositories and configuration.
func NewServices(gormDB *gorm.DB, repos *Repositories, cfg ServiceConfig) (*Services, error) {
	passwordService := passwordsvc.New()
	security := newSecurityServices(repos, cfg)
	userSvc := newUserServices(repos, passwordService, security.userMutationPolicy, cfg)
	facilityServices := newFacilityServices(gormDB, repos)

	exportSvc, err := newExportService(repos, cfg)
	if err != nil {
		return nil, fmt.Errorf("new export service: %w", err)
	}
	notificationSvc, err := newNotificationService(repos, cfg)
	if err != nil {
		return nil, fmt.Errorf("new notification service: %w", err)
	}

	return &Services{
		Project:          newProjectServices(gormDB, repos, facilityServices),
		Dashboard:        dashboardservice.New(repos.Project, repos.Phase, repos.Team, repos.TeamMember, repos.User),
		Phase:            phaseservice.NewPhaseService(repos.Phase),
		PhasePermission:  phasepermissionservice.New(repos.PhasePermissions, repos.Phase, repos.Permissions),
		User:             userSvc.user,
		UserRegistration: userSvc.userRegistration,
		Password:         passwordService,
		JWT:              security.jwt,
		RBAC:             security.rbac,
		Team:             teamservice.New(repos.Team, repos.TeamMember),
		Admin:            userSvc.admin,
		UserDirectory:    userSvc.userDirectory,
		Notification:     notificationSvc,
		Auth: authservice.NewService(
			security.jwt,
			repos.User,
			repos.UserEmail,
			repos.RefreshToken,
			passwordService,
			cfg.AccessTokenTTL,
			cfg.RefreshTokenTTL,
			cfg.Issuer,
		),
		Export:   exportSvc,
		History:  repos.History,
		Facility: facilityServices,
	}, nil
}

func newSecurityServices(repos *Repositories, cfg ServiceConfig) securityServices {
	rbacSvc := rbacservice.New(repos.User, repos.TeamMember, repos.Permissions, repos.RolePermissions)

	return securityServices{
		jwt:                authservice.NewJWTService(cfg.JWTSecret, cfg.Issuer),
		rbac:               rbacSvc,
		userMutationPolicy: usermutationpolicy.New(rbacSvc, repos.UserRegistration),
	}
}

func newUserServices(repos *Repositories, password domainUser.PasswordHasher, policy *usermutationpolicy.Policy, cfg ServiceConfig) userServices {
	return userServices{
		user:             userservice.New(repos.UserLifecycle, password, policy),
		userRegistration: userregistrationservice.New(repos.UserRegistration, policy, password, cfg.AppPublicURL),
		userDirectory:    userdirectoryservice.New(repos.User, repos.Team, repos.TeamMember, repos.RolePermissions),
		admin:            adminservice.New(repos.User, policy),
	}
}
