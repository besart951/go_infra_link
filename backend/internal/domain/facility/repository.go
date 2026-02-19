package facility

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
	"github.com/google/uuid"
)

type BuildingRepository interface {
	domain.Repository[Building]
	ExistsIWSCodeGroup(iwsCode string, buildingGroup int, excludeID *uuid.UUID) (bool, error)
}
type SystemTypeRepository interface {
	domain.Repository[SystemType]
	ExistsName(name string, excludeID *uuid.UUID) (bool, error)
	ExistsOverlappingRange(numberMin, numberMax int, excludeID *uuid.UUID) (bool, error)
}
type SystemPartRepository = domain.Repository[SystemPart]
type SpecificationRepository = domain.Repository[Specification]
type StateTextRepository = domain.Repository[StateText]
type NotificationClassRepository = domain.Repository[NotificationClass]
type AlarmDefinitionRepository = domain.Repository[AlarmDefinition]
type ApparatRepository = domain.Repository[Apparat]
type ObjectDataRepository = domain.Repository[ObjectData]
type ControlCabinetRepository interface {
	domain.Repository[ControlCabinet]
	GetPaginatedListByBuildingID(buildingID uuid.UUID, params domain.PaginationParams) (*domain.PaginatedList[ControlCabinet], error)
	GetIDsByBuildingID(buildingID uuid.UUID) ([]uuid.UUID, error)
	ExistsControlCabinetNr(buildingID uuid.UUID, controlCabinetNr string, excludeID *uuid.UUID) (bool, error)
}

type SPSControllerRepository interface {
	domain.Repository[SPSController]
	GetPaginatedListByControlCabinetID(controlCabinetID uuid.UUID, params domain.PaginationParams) (*domain.PaginatedList[SPSController], error)
	GetIDsByControlCabinetID(controlCabinetID uuid.UUID) ([]uuid.UUID, error)
	GetIDsByControlCabinetIDs(controlCabinetIDs []uuid.UUID) ([]uuid.UUID, error)
	ListGADevicesByControlCabinetID(controlCabinetID uuid.UUID) ([]string, error)
	ExistsGADevice(controlCabinetID uuid.UUID, gaDevice string, excludeID *uuid.UUID) (bool, error)
	ExistsIPAddressVlan(ipAddress string, vlan string, excludeID *uuid.UUID) (bool, error)
	GetByIdsForExport(ids []uuid.UUID) ([]SPSController, error)
}
type SPSControllerSystemTypeRepository = domain.Repository[SPSControllerSystemType]
type FieldDeviceRepository = domain.Repository[FieldDevice]
type BacnetObjectRepository = domain.Repository[BacnetObject]

type ObjectDataHistoryRepository = domain.AppendOnlyRepository[ObjectDataHistory]

type UnitRepository = domain.Repository[Unit]
type AlarmFieldRepository = domain.Repository[AlarmField]
type AlarmTypeRepository interface {
	domain.Repository[AlarmType]
	GetWithFields(id uuid.UUID) (*AlarmType, error)
	ListWithFields(params domain.PaginationParams) (*domain.PaginatedList[AlarmType], error)
}
type AlarmTypeFieldRepository = domain.Repository[AlarmTypeField]
type AlarmDefinitionFieldOverrideRepository = domain.Repository[AlarmDefinitionFieldOverride]
type BacnetObjectAlarmValueRepository interface {
	domain.Repository[BacnetObjectAlarmValue]
	GetByBacnetObjectID(bacnetObjectID uuid.UUID) ([]BacnetObjectAlarmValue, error)
	ReplaceForBacnetObject(bacnetObjectID uuid.UUID, values []BacnetObjectAlarmValue) error
}
