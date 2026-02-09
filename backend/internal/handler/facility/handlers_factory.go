package facility

// ServiceDeps groups service dependencies for facility handler construction.
type ServiceDeps struct {
	Building                BuildingService
	SystemType              SystemTypeService
	SystemPart              SystemPartService
	Specification           SpecificationService
	Apparat                 ApparatService
	ControlCabinet          ControlCabinetService
	FieldDevice             FieldDeviceService
	BacnetObject            BacnetObjectService
	SPSController           SPSControllerService
	StateText               StateTextService
	NotificationClass       NotificationClassService
	AlarmDefinition         AlarmDefinitionService
	ObjectData              ObjectDataService
	SPSControllerSystemType SPSControllerSystemTypeService
}

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
	Validation              *ValidationHandler
}

// NewHandlers creates facility handlers using service dependencies.
func NewHandlers(deps ServiceDeps) *Handlers {
	return &Handlers{
		Building:                NewBuildingHandler(deps.Building),
		SystemType:              NewSystemTypeHandler(deps.SystemType),
		SystemPart:              NewSystemPartHandler(deps.SystemPart, deps.Apparat, deps.ObjectData),
		Specification:           NewSpecificationHandler(deps.Specification),
		Apparat:                 NewApparatHandler(deps.Apparat, deps.SystemPart, deps.ObjectData),
		ControlCabinet:          NewControlCabinetHandler(deps.ControlCabinet),
		FieldDevice:             NewFieldDeviceHandler(deps.FieldDevice),
		BacnetObject:            NewBacnetObjectHandler(deps.BacnetObject),
		SPSController:           NewSPSControllerHandler(deps.SPSController),
		StateText:               NewStateTextHandler(deps.StateText),
		NotificationClass:       NewNotificationClassHandler(deps.NotificationClass),
		AlarmDefinition:         NewAlarmDefinitionHandler(deps.AlarmDefinition),
		ObjectData:              NewObjectDataHandler(deps.ObjectData, deps.BacnetObject, deps.Apparat),
		SPSControllerSystemType: NewSPSControllerSystemTypeHandler(deps.SPSControllerSystemType),
		Validation:              NewValidationHandler(deps.Building, deps.ControlCabinet, deps.SPSController),
	}
}
