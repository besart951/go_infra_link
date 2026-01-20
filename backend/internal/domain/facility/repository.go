package facility

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
)

type BuildingRepository = domain.Repository[Building]
type SystemTypeRepository = domain.Repository[SystemType]
type SystemPartRepository = domain.Repository[SystemPart]
type SpecificationRepository = domain.Repository[Specification]
type StateTextRepository = domain.Repository[StateText]
type NotificationClassRepository = domain.Repository[NotificationClass]
type AlarmDefinitionRepository = domain.Repository[AlarmDefinition]
type ApparatRepository = domain.Repository[Apparat]
type ObjectDataRepository = domain.Repository[ObjectData]
type ControlCabinetRepository = domain.Repository[ControlCabinet]
type SPSControllerRepository = domain.Repository[SPSController]
type SPSControllerSystemTypeRepository = domain.Repository[SPSControllerSystemType]
type FieldDeviceRepository = domain.Repository[FieldDevice]
type BacnetObjectRepository = domain.Repository[BacnetObject]

type ObjectDataHistoryRepository = domain.AppendOnlyRepository[ObjectDataHistory]
