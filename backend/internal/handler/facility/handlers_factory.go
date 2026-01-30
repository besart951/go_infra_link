package facility

import facilityservice "github.com/besart951/go_infra_link/backend/internal/service/facility"

// Handlers groups all facility HTTP handlers.
type Handlers struct {
	Building                *BuildingHandler
	SystemType              *SystemTypeHandler
	SystemPart              *SystemPartHandler
	Specification           *SpecificationHandler
	Apparat                 *ApparatHandler
	ControlCabinet          *ControlCabinetHandler
	FieldDevice             *FieldDeviceHandler
	BacnetObject            *BacnetObjectHandler
	SPSController           *SPSControllerHandler
	StateText               *StateTextHandler
	NotificationClass       *NotificationClassHandler
	AlarmDefinition         *AlarmDefinitionHandler
	ObjectData              *ObjectDataHandler
	SPSControllerSystemType *SPSControllerSystemTypeHandler
}

// NewHandlers creates facility handlers using facility services.
func NewHandlers(services *facilityservice.Services) *Handlers {
	return &Handlers{
		Building:                NewBuildingHandler(services.Building),
		SystemType:              NewSystemTypeHandler(services.SystemType),
		SystemPart:              NewSystemPartHandler(services.SystemPart, services.Apparat, services.ObjectData),
		Specification:           NewSpecificationHandler(services.Specification),
		Apparat:                 NewApparatHandler(services.Apparat, services.SystemPart, services.ObjectData),
		ControlCabinet:          NewControlCabinetHandler(services.ControlCabinet),
		FieldDevice:             NewFieldDeviceHandler(services.FieldDevice),
		BacnetObject:            NewBacnetObjectHandler(services.BacnetObject),
		SPSController:           NewSPSControllerHandler(services.SPSController),
		StateText:               NewStateTextHandler(services.StateText),
		NotificationClass:       NewNotificationClassHandler(services.NotificationClass),
		AlarmDefinition:         NewAlarmDefinitionHandler(services.AlarmDefinition),
		ObjectData:              NewObjectDataHandler(services.ObjectData, services.BacnetObject, services.Apparat),
		SPSControllerSystemType: NewSPSControllerSystemTypeHandler(services.SPSControllerSystemType),
	}
}
