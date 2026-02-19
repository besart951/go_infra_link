// Package wire provides dependency injection wiring for the application.
// It separates the construction of dependencies from business logic.
package wire

import (
	domainAuth "github.com/besart951/go_infra_link/backend/internal/domain/auth"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	domainProject "github.com/besart951/go_infra_link/backend/internal/domain/project"
	domainTeam "github.com/besart951/go_infra_link/backend/internal/domain/team"
	domainUser "github.com/besart951/go_infra_link/backend/internal/domain/user"
	authrepo "github.com/besart951/go_infra_link/backend/internal/repository/auth"
	facilityrepo "github.com/besart951/go_infra_link/backend/internal/repository/facilitysql"
	projectrepo "github.com/besart951/go_infra_link/backend/internal/repository/project"
	projectsqlrepo "github.com/besart951/go_infra_link/backend/internal/repository/projectsql"
	teamrepo "github.com/besart951/go_infra_link/backend/internal/repository/team"
	userrepo "github.com/besart951/go_infra_link/backend/internal/repository/user"
	"gorm.io/gorm"
)

// Repositories holds all repository instances.
type Repositories struct {
	Project                domainProject.ProjectRepository
	Phase                  domainProject.PhaseRepository
	PhasePermission        domainProject.PhasePermissionRepository
	ProjectControlCabinets domainProject.ProjectControlCabinetRepository
	ProjectSPSControllers  domainProject.ProjectSPSControllerRepository
	ProjectFieldDevices    domainProject.ProjectFieldDeviceRepository
	User                   domainUser.UserRepository
	UserEmail              domainUser.UserEmailRepository
	Permissions            domainUser.PermissionRepository
	RolePermissions        domainUser.RolePermissionRepository
	RefreshToken           domainAuth.RefreshTokenRepository
	LoginAttempt           domainAuth.LoginAttemptRepository
	PasswordReset          domainAuth.PasswordResetTokenRepository
	Team                   domainTeam.TeamRepository
	TeamMember             domainTeam.TeamMemberRepository

	FacilityBuildings                domainFacility.BuildingRepository
	FacilitySystemTypes              domainFacility.SystemTypeRepository
	FacilitySystemParts              domainFacility.SystemPartRepository
	FacilitySpecifications           domainFacility.SpecificationStore
	FacilityApparats                 domainFacility.ApparatRepository
	FacilityControlCabinet           domainFacility.ControlCabinetRepository
	FacilityFieldDevices             domainFacility.FieldDeviceStore
	FacilitySPSControllers           domainFacility.SPSControllerRepository
	FacilitySPSControllerSystemTypes domainFacility.SPSControllerSystemTypeStore
	FacilityBacnetObjects            domainFacility.BacnetObjectStore
	FacilityObjectData               domainFacility.ObjectDataStore
	FacilityObjectDataBacnetObjects  domainFacility.ObjectDataBacnetObjectStore

	FacilityStateTexts          domainFacility.StateTextRepository
	FacilityNotificationClasses domainFacility.NotificationClassRepository
	FacilityAlarmDefinitions    domainFacility.AlarmDefinitionRepository

	FacilityUnits                   domainFacility.UnitRepository
	FacilityAlarmFields             domainFacility.AlarmFieldRepository
	FacilityAlarmTypes              domainFacility.AlarmTypeRepository
	FacilityAlarmTypeFields         domainFacility.AlarmTypeFieldRepository
	FacilityBacnetObjectAlarmValues domainFacility.BacnetObjectAlarmValueRepository
}

// NewRepositories creates all repository instances from the database connection.
func NewRepositories(gormDB *gorm.DB) (*Repositories, error) {
	userRepo := userrepo.NewUserRepository(gormDB)
	permissionRepo := userrepo.NewPermissionRepository(gormDB)
	rolePermissionRepo := userrepo.NewRolePermissionRepository(gormDB)
	userEmailRepo, ok := userRepo.(domainUser.UserEmailRepository)
	if !ok {
		return nil, ErrUserRepoMissingEmailLookup
	}

	return &Repositories{
		Project:                projectrepo.NewProjectRepository(gormDB),
		Phase:                  projectrepo.NewPhaseRepository(gormDB),
		PhasePermission:        projectrepo.NewPhasePermissionRepository(gormDB),
		ProjectControlCabinets: projectsqlrepo.NewProjectControlCabinetRepository(gormDB),
		ProjectSPSControllers:  projectsqlrepo.NewProjectSPSControllerRepository(gormDB),
		ProjectFieldDevices:    projectsqlrepo.NewProjectFieldDeviceRepository(gormDB),
		User:                   userRepo,
		UserEmail:              userEmailRepo,
		Permissions:            permissionRepo,
		RolePermissions:        rolePermissionRepo,
		RefreshToken:           authrepo.NewRefreshTokenRepository(gormDB),
		LoginAttempt:           authrepo.NewLoginAttemptRepository(gormDB),
		PasswordReset:          authrepo.NewPasswordResetTokenRepository(gormDB),
		Team:                   teamrepo.NewTeamRepository(gormDB),
		TeamMember:             teamrepo.NewTeamMemberRepository(gormDB),

		FacilityBuildings:                facilityrepo.NewBuildingRepository(gormDB),
		FacilitySystemTypes:              facilityrepo.NewSystemTypeRepository(gormDB),
		FacilitySystemParts:              facilityrepo.NewSystemPartRepository(gormDB),
		FacilitySpecifications:           facilityrepo.NewSpecificationRepository(gormDB),
		FacilityApparats:                 facilityrepo.NewApparatRepository(gormDB),
		FacilityControlCabinet:           facilityrepo.NewControlCabinetRepository(gormDB),
		FacilityFieldDevices:             facilityrepo.NewFieldDeviceRepository(gormDB),
		FacilitySPSControllers:           facilityrepo.NewSPSControllerRepository(gormDB),
		FacilitySPSControllerSystemTypes: facilityrepo.NewSPSControllerSystemTypeRepository(gormDB),
		FacilityBacnetObjects:            facilityrepo.NewBacnetObjectRepository(gormDB),
		FacilityObjectData:               facilityrepo.NewObjectDataRepository(gormDB),
		FacilityObjectDataBacnetObjects:  facilityrepo.NewObjectDataBacnetObjectRepository(gormDB),
		FacilityStateTexts:               facilityrepo.NewStateTextRepository(gormDB),
		FacilityNotificationClasses:      facilityrepo.NewNotificationClassRepository(gormDB),
		FacilityAlarmDefinitions:         facilityrepo.NewAlarmDefinitionRepository(gormDB),

		FacilityUnits:                   facilityrepo.NewUnitRepository(gormDB),
		FacilityAlarmFields:             facilityrepo.NewAlarmFieldRepository(gormDB),
		FacilityAlarmTypes:              facilityrepo.NewAlarmTypeRepository(gormDB),
		FacilityAlarmTypeFields:         facilityrepo.NewAlarmTypeFieldRepository(gormDB),
		FacilityBacnetObjectAlarmValues: facilityrepo.NewBacnetObjectAlarmValueRepository(gormDB),
	}, nil
}
