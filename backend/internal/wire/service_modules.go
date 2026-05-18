package wire

import (
	"fmt"

	apptransaction "github.com/besart951/go_infra_link/backend/internal/application/transaction"
	infratransaction "github.com/besart951/go_infra_link/backend/internal/infrastructure/transaction"
	facilityservice "github.com/besart951/go_infra_link/backend/internal/service/facility"
	projectservice "github.com/besart951/go_infra_link/backend/internal/service/project"
	"gorm.io/gorm"
)

func newFacilityServices(gormDB *gorm.DB, repos *Repositories) *facilityservice.Services {
	facilityTxRepositories := func(unit apptransaction.UnitOfWork) (facilityservice.Repositories, error) {
		txRepos, err := repositoriesFromUnit(unit)
		if err != nil {
			return facilityservice.Repositories{}, fmt.Errorf("facility transaction unit: %w", err)
		}
		return buildFacilityRepositories(txRepos), nil
	}
	return facilityservice.NewServices(buildFacilityRepositories(repos), facilityservice.Config{
		TxRunner:       infratransaction.NewGormRunner(gormDB),
		TxRepositories: facilityTxRepositories,
	})
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
		BacnetReferenceUsages:    repos.FacilityBacnetReferenceUsages,
	}
}

func newProjectServices(gormDB *gorm.DB, repos *Repositories, facilityServices *facilityservice.Services) *projectservice.Services {
	txDependencies := func(unit apptransaction.UnitOfWork) (projectservice.Dependencies, error) {
		txRepos, err := repositoriesFromUnit(unit)
		if err != nil {
			return projectservice.Dependencies{}, fmt.Errorf("project transaction unit: %w", err)
		}

		txFacilityServices := facilityservice.NewServices(buildFacilityRepositories(txRepos))
		return buildProjectDependencies(txRepos, txFacilityServices), nil
	}

	return projectservice.NewServices(buildProjectDependencies(repos, facilityServices), projectservice.Config{
		TxRunner:       infratransaction.NewGormRunner(gormDB),
		TxDependencies: txDependencies,
	})
}

func buildProjectDependencies(repos *Repositories, facilityServices *facilityservice.Services) projectservice.Dependencies {
	return projectservice.Dependencies{
		Projects:                 repos.Project,
		Phases:                   repos.Phase,
		PhasePermissions:         repos.PhasePermissions,
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

func repositoriesFromUnit(unit apptransaction.UnitOfWork) (*Repositories, error) {
	tx, err := infratransaction.GormDB(unit)
	if err != nil {
		return nil, fmt.Errorf("resolve transaction unit: %w", err)
	}
	return NewRepositories(tx)
}
