package facility

import (
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"gorm.io/gorm"
)

// Repositories groups per-entity repositories for convenient wiring.
// Consumers should still depend only on the specific repo(s) they need.
type Repositories struct {
	Buildings                domainFacility.BuildingRepository
	SystemTypes              domainFacility.SystemTypeRepository
	SystemParts              domainFacility.SystemPartRepository
	Specifications           domainFacility.SpecificationRepository
	StateTexts               domainFacility.StateTextRepository
	NotificationClasses      domainFacility.NotificationClassRepository
	AlarmDefinitions         domainFacility.AlarmDefinitionRepository
	Apparats                 domainFacility.ApparatRepository
	ObjectData               domainFacility.ObjectDataRepository
	ControlCabinets          domainFacility.ControlCabinetRepository
	SPSControllers           domainFacility.SPSControllerRepository
	SPSControllerSystemTypes domainFacility.SPSControllerSystemTypeRepository
	FieldDevices             domainFacility.FieldDeviceRepository
	BacnetObjects            domainFacility.BacnetObjectRepository
	ObjectDataHistory        domainFacility.ObjectDataHistoryRepository
}

func NewRepositories(db *gorm.DB) Repositories {
	return Repositories{
		Buildings:                NewBuildingRepository(db),
		SystemTypes:              NewSystemTypeRepository(db),
		SystemParts:              NewSystemPartRepository(db),
		Specifications:           NewSpecificationRepository(db),
		StateTexts:               NewStateTextRepository(db),
		NotificationClasses:      NewNotificationClassRepository(db),
		AlarmDefinitions:         NewAlarmDefinitionRepository(db),
		Apparats:                 NewApparatRepository(db),
		ObjectData:               NewObjectDataRepository(db),
		ControlCabinets:          NewControlCabinetRepository(db),
		SPSControllers:           NewSPSControllerRepository(db),
		SPSControllerSystemTypes: NewSPSControllerSystemTypeRepository(db),
		FieldDevices:             NewFieldDeviceRepository(db),
		BacnetObjects:            NewBacnetObjectRepository(db),
		ObjectDataHistory:        NewObjectDataHistoryRepository(db),
	}
}
