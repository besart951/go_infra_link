package facility

import (
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
)

// Service is a thin layer over the facility repository.
// It mirrors the repository capabilities but gives you a stable place
// to add validation, business rules, auth checks, etc.
type Service struct {
	Buildings                   domainFacility.BuildingRepository
	SystemTypes                 domainFacility.SystemTypeRepository
	SystemParts                 domainFacility.SystemPartRepository
	Specifications              domainFacility.SpecificationRepository
	StateTexts                  domainFacility.StateTextRepository
	NotificationClasses         domainFacility.NotificationClassRepository
	AlarmDefinitions            domainFacility.AlarmDefinitionRepository
	Apparats                    domainFacility.ApparatRepository
	ObjectData                  domainFacility.ObjectDataRepository
	ControlCabinets             domainFacility.ControlCabinetRepository
	SPSControllers              domainFacility.SPSControllerRepository
	SPSControllerSystemTypes    domainFacility.SPSControllerSystemTypeRepository
	FieldDevices                domainFacility.FieldDeviceRepository
	BacnetObjects               domainFacility.BacnetObjectRepository
	ObjectDataHistory           domainFacility.ObjectDataHistoryRepository
	ProjectControlCabinets      domainFacility.ProjectControlCabinetStore
	ProjectSPSControllers       domainFacility.ProjectSPSControllerStore
	ProjectFieldDevices         domainFacility.ProjectFieldDeviceStore
}

func New(
	buildings domainFacility.BuildingRepository,
	systemTypes domainFacility.SystemTypeRepository,
	systemParts domainFacility.SystemPartRepository,
	specifications domainFacility.SpecificationRepository,
	stateTexts domainFacility.StateTextRepository,
	notificationClasses domainFacility.NotificationClassRepository,
	alarmDefinitions domainFacility.AlarmDefinitionRepository,
	apparats domainFacility.ApparatRepository,
	objectData domainFacility.ObjectDataRepository,
	controlCabinets domainFacility.ControlCabinetRepository,
	spsControllers domainFacility.SPSControllerRepository,
	spsControllerSystemTypes domainFacility.SPSControllerSystemTypeRepository,
	fieldDevices domainFacility.FieldDeviceRepository,
	bacnetObjects domainFacility.BacnetObjectRepository,
	objectDataHistory domainFacility.ObjectDataHistoryRepository,
	projectControlCabinets domainFacility.ProjectControlCabinetStore,
	projectSPSControllers domainFacility.ProjectSPSControllerStore,
	projectFieldDevices domainFacility.ProjectFieldDeviceStore,
) *Service {
	return &Service{
		Buildings:                buildings,
		SystemTypes:              systemTypes,
		SystemParts:              systemParts,
		Specifications:           specifications,
		StateTexts:               stateTexts,
		NotificationClasses:      notificationClasses,
		AlarmDefinitions:         alarmDefinitions,
		Apparats:                 apparats,
		ObjectData:               objectData,
		ControlCabinets:          controlCabinets,
		SPSControllers:           spsControllers,
		SPSControllerSystemTypes: spsControllerSystemTypes,
		FieldDevices:             fieldDevices,
		BacnetObjects:            bacnetObjects,
		ObjectDataHistory:        objectDataHistory,
		ProjectControlCabinets:   projectControlCabinets,
		ProjectSPSControllers:    projectSPSControllers,
		ProjectFieldDevices:      projectFieldDevices,
	}
}
