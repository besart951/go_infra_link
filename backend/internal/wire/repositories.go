// Package wire provides dependency injection wiring for the application.
// It separates the construction of dependencies from business logic.
package wire

import (
	"context"
	"fmt"
	domainFieldDevice "github.com/besart951/go_infra_link/backend/internal/domain/facility/fielddevice"
	domainHierarchy "github.com/besart951/go_infra_link/backend/internal/domain/facility/hierarchy"
	domainObjectData "github.com/besart951/go_infra_link/backend/internal/domain/facility/objectdata"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainAuth "github.com/besart951/go_infra_link/backend/internal/domain/auth"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	domainHistory "github.com/besart951/go_infra_link/backend/internal/domain/history"
	domainNotification "github.com/besart951/go_infra_link/backend/internal/domain/notification"
	domainProject "github.com/besart951/go_infra_link/backend/internal/domain/project"
	domainTeam "github.com/besart951/go_infra_link/backend/internal/domain/team"
	domainUser "github.com/besart951/go_infra_link/backend/internal/domain/user"
	authrepo "github.com/besart951/go_infra_link/backend/internal/repository/auth"
	facilitycache "github.com/besart951/go_infra_link/backend/internal/repository/facilitycache"
	facilityrepo "github.com/besart951/go_infra_link/backend/internal/repository/facilitysql"
	historycapture "github.com/besart951/go_infra_link/backend/internal/repository/historycapture"
	historyrepo "github.com/besart951/go_infra_link/backend/internal/repository/historysql"
	notificationrepo "github.com/besart951/go_infra_link/backend/internal/repository/notification"
	projectrepo "github.com/besart951/go_infra_link/backend/internal/repository/project"
	projectsqlrepo "github.com/besart951/go_infra_link/backend/internal/repository/projectsql"
	teamrepo "github.com/besart951/go_infra_link/backend/internal/repository/team"
	userrepo "github.com/besart951/go_infra_link/backend/internal/repository/user"
	userregistrationrepo "github.com/besart951/go_infra_link/backend/internal/repository/userregistration"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Repositories holds all repository instances.
type Repositories struct {
	Project                  domainProject.ProjectRepository
	Phase                    domainProject.PhaseRepository
	PhasePermissions         domainProject.PhasePermissionRepository
	ProjectControlCabinets   domainProject.ProjectControlCabinetRepository
	ProjectSPSControllers    domainProject.ProjectSPSControllerRepository
	ProjectFieldDevices      domainProject.ProjectFieldDeviceRepository
	History                  HistoryRepository
	User                     domainUser.UserRepository
	UserLifecycle            domainUser.UserLifecycleRepository
	UserEmail                domainUser.UserEmailRepository
	UserRegistration         *userregistrationrepo.Store
	Permissions              domainUser.PermissionRepository
	RolePermissions          domainUser.RolePermissionRepository
	RefreshToken             domainAuth.RefreshTokenRepository
	NotificationSMTPSettings domainNotification.SMTPSettingsRepository
	NotificationPreferences  domainNotification.UserPreferenceRepository
	SystemNotifications      domainNotification.SystemNotificationRepository
	NotificationEmailOutbox  domainNotification.EmailOutboxRepository
	NotificationRules        domainNotification.NotificationRuleRepository
	Team                     domainTeam.TeamRepository
	TeamMember               domainTeam.TeamMemberRepository

	FacilityBuildings                domainFacility.BuildingRepository
	FacilitySystemTypes              domainFacility.SystemTypeRepository
	FacilitySystemParts              domainFacility.SystemPartRepository
	FacilitySpecifications           domainFieldDevice.SpecificationStore
	FacilityApparats                 domainFacility.ApparatRepository
	FacilityControlCabinet           domainFacility.ControlCabinetRepository
	FacilityFieldDevices             domainFieldDevice.FieldDeviceStore
	FacilitySPSControllers           domainFacility.SPSControllerRepository
	FacilitySPSControllerSystemTypes domainHierarchy.SPSControllerSystemTypeStore
	FacilityBacnetObjects            domainObjectData.BacnetObjectStore
	FacilityObjectData               domainObjectData.ObjectDataStore
	FacilityObjectDataBacnetObjects  domainObjectData.ObjectDataBacnetObjectStore

	FacilityStateTexts          domainFacility.StateTextRepository
	FacilityNotificationClasses domainFacility.NotificationClassRepository
	FacilityAlarmDefinitions    domainFacility.AlarmDefinitionRepository

	FacilityUnits                   domainFacility.UnitRepository
	FacilityAlarmFields             domainFacility.AlarmFieldRepository
	FacilityAlarmTypes              domainFacility.AlarmTypeRepository
	FacilityAlarmTypeFields         domainFacility.AlarmTypeFieldRepository
	FacilityBacnetObjectAlarmValues domainFacility.BacnetObjectAlarmValueRepository
	FacilityBacnetReferenceUsages   domainFacility.BacnetReferenceUsageRepository
}

type HistoryRepository interface {
	ListTimeline(ctx context.Context, filter domainHistory.TimelineFilter) (*domain.PaginatedList[domainHistory.ChangeEvent], error)
	GetEvent(ctx context.Context, id uuid.UUID) (*domainHistory.ChangeEvent, error)
	RestoreEntityToEvent(ctx context.Context, eventID uuid.UUID, mode domainHistory.RestoreMode) (*domainHistory.RestoreResult, error)
	RestoreControlCabinet(ctx context.Context, controlCabinetID uuid.UUID, req domainHistory.RestoreControlCabinetRequest) (*domainHistory.RestoreResult, error)
}

type (
	userRepositoryGroup struct {
		User             domainUser.UserRepository
		UserLifecycle    domainUser.UserLifecycleRepository
		UserEmail        domainUser.UserEmailRepository
		UserRegistration *userregistrationrepo.Store
		Permissions      domainUser.PermissionRepository
		RolePermissions  domainUser.RolePermissionRepository
		RefreshToken     domainAuth.RefreshTokenRepository
	}

	projectRepositoryGroup struct {
		Project                domainProject.ProjectRepository
		Phase                  domainProject.PhaseRepository
		PhasePermissions       domainProject.PhasePermissionRepository
		ProjectControlCabinets domainProject.ProjectControlCabinetRepository
		ProjectSPSControllers  domainProject.ProjectSPSControllerRepository
		ProjectFieldDevices    domainProject.ProjectFieldDeviceRepository
	}

	facilityRepositoryGroup struct {
		FacilityBuildings                domainFacility.BuildingRepository
		FacilitySystemTypes              domainFacility.SystemTypeRepository
		FacilitySystemParts              domainFacility.SystemPartRepository
		FacilitySpecifications           domainFieldDevice.SpecificationStore
		FacilityApparats                 domainFacility.ApparatRepository
		FacilityControlCabinet           domainFacility.ControlCabinetRepository
		FacilityFieldDevices             domainFieldDevice.FieldDeviceStore
		FacilitySPSControllers           domainFacility.SPSControllerRepository
		FacilitySPSControllerSystemTypes domainHierarchy.SPSControllerSystemTypeStore
		FacilityBacnetObjects            domainObjectData.BacnetObjectStore
		FacilityObjectData               domainObjectData.ObjectDataStore
		FacilityObjectDataBacnetObjects  domainObjectData.ObjectDataBacnetObjectStore
		FacilityStateTexts               domainFacility.StateTextRepository
		FacilityNotificationClasses      domainFacility.NotificationClassRepository
		FacilityAlarmDefinitions         domainFacility.AlarmDefinitionRepository
		FacilityUnits                    domainFacility.UnitRepository
		FacilityAlarmFields              domainFacility.AlarmFieldRepository
		FacilityAlarmTypes               domainFacility.AlarmTypeRepository
		FacilityAlarmTypeFields          domainFacility.AlarmTypeFieldRepository
		FacilityBacnetObjectAlarmValues  domainFacility.BacnetObjectAlarmValueRepository
		FacilityBacnetReferenceUsages    domainFacility.BacnetReferenceUsageRepository
	}

	notificationRepositoryGroup struct {
		NotificationSMTPSettings domainNotification.SMTPSettingsRepository
		NotificationPreferences  domainNotification.UserPreferenceRepository
		SystemNotifications      domainNotification.SystemNotificationRepository
		NotificationEmailOutbox  domainNotification.EmailOutboxRepository
		NotificationRules        domainNotification.NotificationRuleRepository
	}

	teamRepositoryGroup struct {
		Team       domainTeam.TeamRepository
		TeamMember domainTeam.TeamMemberRepository
	}
)

// NewRepositories creates all repository instances from the database connection.
func NewRepositories(gormDB *gorm.DB) (*Repositories, error) {
	historyStore := historyrepo.NewStore(gormDB)

	userRepos, err := newUserRepositories(gormDB)
	if err != nil {
		return nil, fmt.Errorf("user repositories: %w", err)
	}
	projectRepos := newProjectRepositories(gormDB, historyStore)
	facilityRepos := newFacilityRepositories(gormDB, historyStore)
	notificationRepos := newNotificationRepositories(gormDB)
	teamRepos := newTeamRepositories(gormDB)

	return composeRepositories(
		historyStore,
		userRepos,
		projectRepos,
		facilityRepos,
		notificationRepos,
		teamRepos,
	), nil
}

func newUserRepositories(gormDB *gorm.DB) (userRepositoryGroup, error) {
	userRepo := userrepo.NewUserRepository(gormDB)

	userEmailRepo, err := requireUserEmailRepository(userRepo)
	if err != nil {
		return userRepositoryGroup{}, fmt.Errorf("user email capability: %w", err)
	}
	userLifecycleStore, err := requireUserLifecycleStore(userRepo)
	if err != nil {
		return userRepositoryGroup{}, fmt.Errorf("user lifecycle capability: %w", err)
	}
	return userRepositoryGroup{
		User:             userRepo,
		UserLifecycle:    userLifecycleStore,
		UserEmail:        userEmailRepo,
		UserRegistration: userregistrationrepo.NewStore(gormDB),
		Permissions:      userrepo.NewPermissionRepository(gormDB),
		RolePermissions:  userrepo.NewRolePermissionRepository(gormDB),
		RefreshToken:     authrepo.NewRefreshTokenRepository(gormDB),
	}, nil
}

func newProjectRepositories(gormDB *gorm.DB, history *historyrepo.Store) projectRepositoryGroup {
	return projectRepositoryGroup{
		Project:                historycapture.WrapProject(projectrepo.NewProjectRepository(gormDB), history),
		Phase:                  projectrepo.NewPhaseRepository(gormDB),
		PhasePermissions:       projectrepo.NewPhasePermissionRepository(gormDB),
		ProjectControlCabinets: historycapture.WrapProjectControlCabinet(projectsqlrepo.NewProjectControlCabinetRepository(gormDB), history),
		ProjectSPSControllers:  historycapture.WrapProjectSPSController(projectsqlrepo.NewProjectSPSControllerRepository(gormDB), history),
		ProjectFieldDevices:    historycapture.WrapProjectFieldDevice(projectsqlrepo.NewProjectFieldDeviceRepository(gormDB), history),
	}
}

func newFacilityRepositories(gormDB *gorm.DB, history *historyrepo.Store) facilityRepositoryGroup {
	facilitySystemParts := historycapture.WrapSystemPart(facilityrepo.NewSystemPartRepository(gormDB), history)
	facilityApparats := historycapture.WrapApparat(facilityrepo.NewApparatRepository(gormDB), history)
	facilityApparats, facilitySystemParts = facilitycache.WrapReferenceData(facilityApparats, facilitySystemParts)

	return facilityRepositoryGroup{
		FacilityBuildings:                historycapture.WrapBuilding(facilityrepo.NewBuildingRepository(gormDB), history),
		FacilitySystemTypes:              historycapture.WrapSystemType(facilityrepo.NewSystemTypeRepository(gormDB), history),
		FacilitySystemParts:              facilitySystemParts,
		FacilitySpecifications:           historycapture.WrapSpecification(facilityrepo.NewSpecificationRepository(gormDB), history),
		FacilityApparats:                 facilityApparats,
		FacilityControlCabinet:           historycapture.WrapControlCabinet(facilityrepo.NewControlCabinetRepository(gormDB), history),
		FacilityFieldDevices:             historycapture.WrapFieldDevice(facilityrepo.NewFieldDeviceRepository(gormDB), history),
		FacilitySPSControllers:           historycapture.WrapSPSController(facilityrepo.NewSPSControllerRepository(gormDB), history),
		FacilitySPSControllerSystemTypes: historycapture.WrapSPSControllerSystemType(facilityrepo.NewSPSControllerSystemTypeRepository(gormDB), history),
		FacilityBacnetObjects:            historycapture.WrapBacnetObject(facilityrepo.NewBacnetObjectRepository(gormDB), history),
		FacilityObjectData:               historycapture.WrapObjectData(facilityrepo.NewObjectDataRepository(gormDB), history),
		FacilityObjectDataBacnetObjects:  facilityrepo.NewObjectDataBacnetObjectRepository(gormDB),
		FacilityStateTexts:               historycapture.WrapRepository("state_texts", facilityrepo.NewStateTextRepository(gormDB), history),
		FacilityNotificationClasses:      historycapture.WrapRepository("notification_classes", facilityrepo.NewNotificationClassRepository(gormDB), history),
		FacilityAlarmDefinitions:         historycapture.WrapAlarmDefinition(facilityrepo.NewAlarmDefinitionRepository(gormDB), history),
		FacilityUnits:                    historycapture.WrapRepository("units", facilityrepo.NewUnitRepository(gormDB), history),
		FacilityAlarmFields:              historycapture.WrapRepository("alarm_fields", facilityrepo.NewAlarmFieldRepository(gormDB), history),
		FacilityAlarmTypes:               historycapture.WrapAlarmType(facilityrepo.NewAlarmTypeRepository(gormDB), history),
		FacilityAlarmTypeFields:          historycapture.WrapRepository("alarm_type_fields", facilityrepo.NewAlarmTypeFieldRepository(gormDB), history),
		FacilityBacnetObjectAlarmValues:  historycapture.WrapBacnetObjectAlarmValue(facilityrepo.NewBacnetObjectAlarmValueRepository(gormDB), history),
		FacilityBacnetReferenceUsages:    facilityrepo.NewBacnetReferenceUsageRepository(gormDB),
	}
}

func newNotificationRepositories(gormDB *gorm.DB) notificationRepositoryGroup {
	return notificationRepositoryGroup{
		NotificationSMTPSettings: notificationrepo.NewSMTPSettingsRepository(gormDB),
		NotificationPreferences:  notificationrepo.NewUserPreferenceRepository(gormDB),
		SystemNotifications:      notificationrepo.NewSystemNotificationRepository(gormDB),
		NotificationEmailOutbox:  notificationrepo.NewEmailOutboxRepository(gormDB),
		NotificationRules:        notificationrepo.NewNotificationRuleRepository(gormDB),
	}
}

func newTeamRepositories(gormDB *gorm.DB) teamRepositoryGroup {
	return teamRepositoryGroup{
		Team:       teamrepo.NewTeamRepository(gormDB),
		TeamMember: teamrepo.NewTeamMemberRepository(gormDB),
	}
}

func composeRepositories(
	history HistoryRepository,
	users userRepositoryGroup,
	projects projectRepositoryGroup,
	facilities facilityRepositoryGroup,
	notifications notificationRepositoryGroup,
	teams teamRepositoryGroup,
) *Repositories {
	return &Repositories{
		History:                          history,
		Project:                          projects.Project,
		Phase:                            projects.Phase,
		PhasePermissions:                 projects.PhasePermissions,
		ProjectControlCabinets:           projects.ProjectControlCabinets,
		ProjectSPSControllers:            projects.ProjectSPSControllers,
		ProjectFieldDevices:              projects.ProjectFieldDevices,
		User:                             users.User,
		UserLifecycle:                    users.UserLifecycle,
		UserEmail:                        users.UserEmail,
		UserRegistration:                 users.UserRegistration,
		Permissions:                      users.Permissions,
		RolePermissions:                  users.RolePermissions,
		RefreshToken:                     users.RefreshToken,
		NotificationSMTPSettings:         notifications.NotificationSMTPSettings,
		NotificationPreferences:          notifications.NotificationPreferences,
		SystemNotifications:              notifications.SystemNotifications,
		NotificationEmailOutbox:          notifications.NotificationEmailOutbox,
		NotificationRules:                notifications.NotificationRules,
		Team:                             teams.Team,
		TeamMember:                       teams.TeamMember,
		FacilityBuildings:                facilities.FacilityBuildings,
		FacilitySystemTypes:              facilities.FacilitySystemTypes,
		FacilitySystemParts:              facilities.FacilitySystemParts,
		FacilitySpecifications:           facilities.FacilitySpecifications,
		FacilityApparats:                 facilities.FacilityApparats,
		FacilityControlCabinet:           facilities.FacilityControlCabinet,
		FacilityFieldDevices:             facilities.FacilityFieldDevices,
		FacilitySPSControllers:           facilities.FacilitySPSControllers,
		FacilitySPSControllerSystemTypes: facilities.FacilitySPSControllerSystemTypes,
		FacilityBacnetObjects:            facilities.FacilityBacnetObjects,
		FacilityObjectData:               facilities.FacilityObjectData,
		FacilityObjectDataBacnetObjects:  facilities.FacilityObjectDataBacnetObjects,
		FacilityStateTexts:               facilities.FacilityStateTexts,
		FacilityNotificationClasses:      facilities.FacilityNotificationClasses,
		FacilityAlarmDefinitions:         facilities.FacilityAlarmDefinitions,
		FacilityUnits:                    facilities.FacilityUnits,
		FacilityAlarmFields:              facilities.FacilityAlarmFields,
		FacilityAlarmTypes:               facilities.FacilityAlarmTypes,
		FacilityAlarmTypeFields:          facilities.FacilityAlarmTypeFields,
		FacilityBacnetObjectAlarmValues:  facilities.FacilityBacnetObjectAlarmValues,
		FacilityBacnetReferenceUsages:    facilities.FacilityBacnetReferenceUsages,
	}
}

func requireUserEmailRepository(repo domainUser.UserRepository) (domainUser.UserEmailRepository, error) {
	userEmailRepo, ok := repo.(domainUser.UserEmailRepository)
	if !ok {
		return nil, ErrUserRepoMissingEmailLookup
	}
	return userEmailRepo, nil
}

func requireUserLifecycleStore(repo domainUser.UserRepository) (domainUser.UserLifecycleRepository, error) {
	userLifecycleStore, ok := repo.(domainUser.UserLifecycleRepository)
	if !ok {
		return nil, ErrUserRepoMissingLifecycle
	}
	return userLifecycleStore, nil
}
