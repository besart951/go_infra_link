// Package wire provides dependency injection wiring for the application.
// It separates the construction of dependencies from business logic.
package wire

import (
	"context"

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
	UserEmail                domainUser.UserEmailRepository
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

type HistoryRepository interface {
	ListTimeline(ctx context.Context, filter domainHistory.TimelineFilter) (*domain.PaginatedList[domainHistory.ChangeEvent], error)
	GetEvent(ctx context.Context, id uuid.UUID) (*domainHistory.ChangeEvent, error)
	RestoreEntityToEvent(ctx context.Context, eventID uuid.UUID, mode domainHistory.RestoreMode) (*domainHistory.RestoreResult, error)
	RestoreControlCabinet(ctx context.Context, controlCabinetID uuid.UUID, req domainHistory.RestoreControlCabinetRequest) (*domainHistory.RestoreResult, error)
}

// NewRepositories creates all repository instances from the database connection.
func NewRepositories(gormDB *gorm.DB) (*Repositories, error) {
	userRepo := userrepo.NewUserRepository(gormDB)
	permissionRepo := userrepo.NewPermissionRepository(gormDB)
	rolePermissionRepo := userrepo.NewRolePermissionRepository(gormDB)
	historyStore := historyrepo.NewStore(gormDB)
	userEmailRepo, ok := userRepo.(domainUser.UserEmailRepository)
	if !ok {
		return nil, ErrUserRepoMissingEmailLookup
	}
	facilitySystemParts := historycapture.WrapSystemPart(facilityrepo.NewSystemPartRepository(gormDB), historyStore)
	facilityApparats := historycapture.WrapApparat(facilityrepo.NewApparatRepository(gormDB), historyStore)
	facilityApparats, facilitySystemParts = facilitycache.WrapReferenceData(facilityApparats, facilitySystemParts)

	return &Repositories{
		Project:                  historycapture.WrapProject(projectrepo.NewProjectRepository(gormDB), historyStore),
		Phase:                    projectrepo.NewPhaseRepository(gormDB),
		PhasePermissions:         projectrepo.NewPhasePermissionRepository(gormDB),
		ProjectControlCabinets:   historycapture.WrapProjectControlCabinet(projectsqlrepo.NewProjectControlCabinetRepository(gormDB), historyStore),
		ProjectSPSControllers:    historycapture.WrapProjectSPSController(projectsqlrepo.NewProjectSPSControllerRepository(gormDB), historyStore),
		ProjectFieldDevices:      historycapture.WrapProjectFieldDevice(projectsqlrepo.NewProjectFieldDeviceRepository(gormDB), historyStore),
		History:                  historyStore,
		User:                     userRepo,
		UserEmail:                userEmailRepo,
		Permissions:              permissionRepo,
		RolePermissions:          rolePermissionRepo,
		RefreshToken:             authrepo.NewRefreshTokenRepository(gormDB),
		NotificationSMTPSettings: notificationrepo.NewSMTPSettingsRepository(gormDB),
		NotificationPreferences:  notificationrepo.NewUserPreferenceRepository(gormDB),
		SystemNotifications:      notificationrepo.NewSystemNotificationRepository(gormDB),
		NotificationEmailOutbox:  notificationrepo.NewEmailOutboxRepository(gormDB),
		NotificationRules:        notificationrepo.NewNotificationRuleRepository(gormDB),
		Team:                     teamrepo.NewTeamRepository(gormDB),
		TeamMember:               teamrepo.NewTeamMemberRepository(gormDB),

		FacilityBuildings:                historycapture.WrapBuilding(facilityrepo.NewBuildingRepository(gormDB), historyStore),
		FacilitySystemTypes:              historycapture.WrapSystemType(facilityrepo.NewSystemTypeRepository(gormDB), historyStore),
		FacilitySystemParts:              facilitySystemParts,
		FacilitySpecifications:           historycapture.WrapSpecification(facilityrepo.NewSpecificationRepository(gormDB), historyStore),
		FacilityApparats:                 facilityApparats,
		FacilityControlCabinet:           historycapture.WrapControlCabinet(facilityrepo.NewControlCabinetRepository(gormDB), historyStore),
		FacilityFieldDevices:             historycapture.WrapFieldDevice(facilityrepo.NewFieldDeviceRepository(gormDB), historyStore),
		FacilitySPSControllers:           historycapture.WrapSPSController(facilityrepo.NewSPSControllerRepository(gormDB), historyStore),
		FacilitySPSControllerSystemTypes: historycapture.WrapSPSControllerSystemType(facilityrepo.NewSPSControllerSystemTypeRepository(gormDB), historyStore),
		FacilityBacnetObjects:            historycapture.WrapBacnetObject(facilityrepo.NewBacnetObjectRepository(gormDB), historyStore),
		FacilityObjectData:               historycapture.WrapObjectData(facilityrepo.NewObjectDataRepository(gormDB), historyStore),
		FacilityObjectDataBacnetObjects:  facilityrepo.NewObjectDataBacnetObjectRepository(gormDB),
		FacilityStateTexts:               historycapture.WrapRepository("state_texts", facilityrepo.NewStateTextRepository(gormDB), historyStore),
		FacilityNotificationClasses:      historycapture.WrapRepository("notification_classes", facilityrepo.NewNotificationClassRepository(gormDB), historyStore),
		FacilityAlarmDefinitions:         historycapture.WrapAlarmDefinition(facilityrepo.NewAlarmDefinitionRepository(gormDB), historyStore),

		FacilityUnits:                   historycapture.WrapRepository("units", facilityrepo.NewUnitRepository(gormDB), historyStore),
		FacilityAlarmFields:             historycapture.WrapRepository("alarm_fields", facilityrepo.NewAlarmFieldRepository(gormDB), historyStore),
		FacilityAlarmTypes:              historycapture.WrapAlarmType(facilityrepo.NewAlarmTypeRepository(gormDB), historyStore),
		FacilityAlarmTypeFields:         historycapture.WrapRepository("alarm_type_fields", facilityrepo.NewAlarmTypeFieldRepository(gormDB), historyStore),
		FacilityBacnetObjectAlarmValues: historycapture.WrapBacnetObjectAlarmValue(facilityrepo.NewBacnetObjectAlarmValueRepository(gormDB), historyStore),
	}, nil
}
