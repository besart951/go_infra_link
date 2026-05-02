package wire

import (
	"fmt"

	facilityservice "github.com/besart951/go_infra_link/backend/internal/service/facility"
	projectservice "github.com/besart951/go_infra_link/backend/internal/service/project"
	"gorm.io/gorm"
)

func newFacilityServices(gormDB *gorm.DB, repos *Repositories) *facilityservice.Services {
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
	return facilityservice.NewServices(buildFacilityRepositories(repos), facilityservice.Config{
		TxRunner:       facilityTxRunner,
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
