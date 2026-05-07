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
		tx, err := infratransaction.GormDB(unit)
		if err != nil {
			return facilityservice.Repositories{}, fmt.Errorf("facility transaction unit: %w", err)
		}
		txRepos, err := NewRepositories(tx)
		if err != nil {
			return facilityservice.Repositories{}, fmt.Errorf("transaction repositories: %w", err)
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
	}
}

func newProjectServices(gormDB *gorm.DB, repos *Repositories, facilityServices *facilityservice.Services) *projectservice.Services {
	txDependencies := func(unit apptransaction.UnitOfWork) (projectservice.Dependencies, error) {
		tx, err := infratransaction.GormDB(unit)
		if err != nil {
			return projectservice.Dependencies{}, fmt.Errorf("project transaction unit: %w", err)
		}
		txRepos, err := NewRepositories(tx)
		if err != nil {
			return projectservice.Dependencies{}, fmt.Errorf("transaction repositories: %w", err)
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
