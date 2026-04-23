package facility

import (
	"context"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	"github.com/google/uuid"
)

type BuildingRepository interface {
	domain.Repository[Building]
	ExistsIWSCodeGroup(ctx context.Context, iwsCode string, buildingGroup int, excludeID *uuid.UUID) (bool, error)
}
type SystemTypeRepository interface {
	domain.Repository[SystemType]
	ExistsName(ctx context.Context, name string, excludeID *uuid.UUID) (bool, error)
	ExistsOverlappingRange(ctx context.Context, numberMin, numberMax int, excludeID *uuid.UUID) (bool, error)
}
type SystemPartRepository interface {
	domain.Repository[SystemPart]
	ExistsShortName(ctx context.Context, shortName string, excludeID *uuid.UUID) (bool, error)
	ExistsName(ctx context.Context, name string, excludeID *uuid.UUID) (bool, error)
}
type SpecificationRepository = domain.Repository[Specification]
type StateTextRepository = domain.Repository[StateText]
type NotificationClassRepository = domain.Repository[NotificationClass]
type AlarmDefinitionRepository interface {
	domain.Repository[AlarmDefinition]
	FindOrCreateTemplateByAlarmTypeID(ctx context.Context, alarmTypeID uuid.UUID) (*AlarmDefinition, error)
}
type ApparatRepository interface {
	domain.Repository[Apparat]
	ExistsShortName(ctx context.Context, shortName string, excludeID *uuid.UUID) (bool, error)
	ExistsName(ctx context.Context, name string, excludeID *uuid.UUID) (bool, error)
	GetPaginatedListWithFilters(ctx context.Context, params domain.PaginationParams, filters ApparatFilterParams) (*domain.PaginatedList[Apparat], error)
}
type ObjectDataRepository = domain.Repository[ObjectData]
type ControlCabinetRepository interface {
	domain.Repository[ControlCabinet]
	GetPaginatedListByBuildingID(ctx context.Context, buildingID uuid.UUID, params domain.PaginationParams) (*domain.PaginatedList[ControlCabinet], error)
	GetIDsByBuildingID(ctx context.Context, buildingID uuid.UUID) ([]uuid.UUID, error)
	ExistsControlCabinetNr(ctx context.Context, buildingID uuid.UUID, controlCabinetNr string, excludeID *uuid.UUID) (bool, error)
}

type SPSControllerRepository interface {
	domain.Repository[SPSController]
	GetPaginatedListByControlCabinetID(ctx context.Context, controlCabinetID uuid.UUID, params domain.PaginationParams) (*domain.PaginatedList[SPSController], error)
	GetIDsByControlCabinetID(ctx context.Context, controlCabinetID uuid.UUID) ([]uuid.UUID, error)
	GetIDsByControlCabinetIDs(ctx context.Context, controlCabinetIDs []uuid.UUID) ([]uuid.UUID, error)
	ListGADevicesByControlCabinetID(ctx context.Context, controlCabinetID uuid.UUID) ([]string, error)
	ExistsDeviceName(ctx context.Context, controlCabinetID uuid.UUID, deviceName string, excludeID *uuid.UUID) (bool, error)
	ExistsGADevice(ctx context.Context, controlCabinetID uuid.UUID, gaDevice string, excludeID *uuid.UUID) (bool, error)
	ExistsIPAddressVlan(ctx context.Context, ipAddress string, vlan string, excludeID *uuid.UUID) (bool, error)
	GetByIdsForExport(ctx context.Context, ids []uuid.UUID) ([]SPSController, error)
}
type SPSControllerSystemTypeRepository = domain.Repository[SPSControllerSystemType]
type FieldDeviceRepository = domain.Repository[FieldDevice]
type BacnetObjectRepository = domain.Repository[BacnetObject]

type UnitRepository = domain.Repository[Unit]
type AlarmFieldRepository = domain.Repository[AlarmField]
type AlarmTypeRepository interface {
	domain.Repository[AlarmType]
	GetWithFields(ctx context.Context, id uuid.UUID) (*AlarmType, error)
	ListWithFields(ctx context.Context, params domain.PaginationParams) (*domain.PaginatedList[AlarmType], error)
}
type AlarmTypeFieldRepository = domain.Repository[AlarmTypeField]
type AlarmDefinitionFieldOverrideRepository = domain.Repository[AlarmDefinitionFieldOverride]
type BacnetObjectAlarmValueRepository interface {
	domain.Repository[BacnetObjectAlarmValue]
	GetByBacnetObjectID(ctx context.Context, bacnetObjectID uuid.UUID) ([]BacnetObjectAlarmValue, error)
	BulkCreate(ctx context.Context, values []*BacnetObjectAlarmValue, batchSize int) error
	ReplaceForBacnetObject(ctx context.Context, bacnetObjectID uuid.UUID, values []BacnetObjectAlarmValue) error
}
