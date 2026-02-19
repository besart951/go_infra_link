package facility

import domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"

// Repositories groups facility repositories for service construction.
type Repositories struct {
	Buildings                domainFacility.BuildingRepository
	SystemTypes              domainFacility.SystemTypeRepository
	SystemParts              domainFacility.SystemPartRepository
	Specifications           domainFacility.SpecificationStore
	Apparats                 domainFacility.ApparatRepository
	ControlCabinets          domainFacility.ControlCabinetRepository
	FieldDevices             domainFacility.FieldDeviceStore
	SPSControllers           domainFacility.SPSControllerRepository
	SPSControllerSystemTypes domainFacility.SPSControllerSystemTypeStore
	BacnetObjects            domainFacility.BacnetObjectStore
	ObjectData               domainFacility.ObjectDataStore
	ObjectDataBacnetObjects  domainFacility.ObjectDataBacnetObjectStore
	StateTexts               domainFacility.StateTextRepository
	NotificationClasses      domainFacility.NotificationClassRepository
	AlarmDefinitions         domainFacility.AlarmDefinitionRepository
	AlarmTypes               domainFacility.AlarmTypeRepository
	BacnetObjectAlarmValues  domainFacility.BacnetObjectAlarmValueRepository
}

// Services bundles all facility services.
type Services struct {
	Building                *BuildingService
	SystemType              *SystemTypeService
	SystemPart              *SystemPartService
	Apparat                 *ApparatService
	ControlCabinet          *ControlCabinetService
	FieldDevice             *FieldDeviceService
	BacnetObject            *BacnetObjectService
	SPSController           *SPSControllerService
	StateText               *StateTextService
	NotificationClass       *NotificationClassService
	AlarmDefinition         *AlarmDefinitionService
	ObjectData              *ObjectDataService
	SPSControllerSystemType *SPSControllerSystemTypeService
	AlarmType               *AlarmTypeService
	BacnetAlarmValue        *BacnetAlarmValueService
}

// NewServices creates facility services using a factory-style constructor.
func NewServices(repos Repositories) *Services {
	return &Services{
		Building: NewBuildingService(
			repos.Buildings,
			repos.ControlCabinets,
			repos.SPSControllers,
			repos.SPSControllerSystemTypes,
			repos.FieldDevices,
			repos.Specifications,
			repos.BacnetObjects,
		),
		SystemType: NewSystemTypeService(repos.SystemTypes),
		SystemPart: NewSystemPartService(repos.SystemParts),
		Apparat:    NewApparatService(repos.Apparats),
		ControlCabinet: NewControlCabinetService(
			repos.ControlCabinets,
			repos.Buildings,
			repos.SPSControllers,
			repos.SPSControllerSystemTypes,
			repos.FieldDevices,
			repos.BacnetObjects,
			repos.Specifications,
		),
		FieldDevice: NewFieldDeviceService(
			repos.FieldDevices,
			repos.SPSControllerSystemTypes,
			repos.SPSControllers,
			repos.ControlCabinets,
			repos.SystemTypes,
			repos.Buildings,
			repos.Apparats,
			repos.SystemParts,
			repos.Specifications,
			repos.BacnetObjects,
			repos.ObjectData,
		),
		BacnetObject: NewBacnetObjectService(
			repos.BacnetObjects,
			repos.FieldDevices,
			repos.ObjectData,
			repos.ObjectDataBacnetObjects,
		),
		SPSController: NewSPSControllerService(
			repos.SPSControllers,
			repos.ControlCabinets,
			repos.Buildings,
			repos.SystemTypes,
			repos.SPSControllerSystemTypes,
			repos.FieldDevices,
			repos.Specifications,
			repos.BacnetObjects,
		),
		StateText:         NewStateTextService(repos.StateTexts),
		NotificationClass: NewNotificationClassService(repos.NotificationClasses),
		AlarmDefinition:   NewAlarmDefinitionService(repos.AlarmDefinitions),
		ObjectData:        NewObjectDataService(repos.ObjectData),
		SPSControllerSystemType: NewSPSControllerSystemTypeService(
			repos.SPSControllerSystemTypes,
			repos.SystemTypes,
			repos.FieldDevices,
			repos.Specifications,
			repos.BacnetObjects,
		),
		AlarmType: NewAlarmTypeService(repos.AlarmTypes),
		BacnetAlarmValue: NewBacnetAlarmValueService(
			repos.BacnetObjectAlarmValues,
			repos.AlarmTypes,
			repos.BacnetObjects,
			repos.AlarmDefinitions,
		),
	}
}
