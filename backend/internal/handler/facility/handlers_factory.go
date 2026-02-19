package facility

// ServiceDeps groups service dependencies for facility handler construction.
type ServiceDeps struct {
	Building                BuildingService
	SystemType              SystemTypeService
	SystemPart              SystemPartService
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
	Export                  ExportService
	AlarmType               AlarmTypeService
	Unit                    UnitService
	AlarmField              AlarmFieldService
	AlarmTypeField          AlarmTypeFieldService
	BacnetAlarm             BacnetAlarmValueService
}

// Handlers groups all facility HTTP handlers.
type Handlers struct {
	Building                *BuildingHandler
	SystemType              *SystemTypeHandler
	SystemPart              *SystemPartHandler
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
	Export                  *ExportHandler
	Validation              *ValidationHandler
	AlarmType               *AlarmTypeHandler
	Unit                    *UnitHandler
	AlarmField              *AlarmFieldHandler
	AlarmTypeField          *AlarmTypeFieldHandler
	BacnetAlarm             *BacnetAlarmHandler
}

// NewHandlers creates facility handlers using service dependencies.
func NewHandlers(deps ServiceDeps) *Handlers {
	return &Handlers{
		Building:                NewBuildingHandler(deps.Building),
		SystemType:              NewSystemTypeHandler(deps.SystemType),
		SystemPart:              NewSystemPartHandler(deps.SystemPart, deps.Apparat, deps.ObjectData),
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
		Export:                  NewExportHandler(deps.Export),
		Validation:              NewValidationHandler(deps.Building, deps.ControlCabinet, deps.SPSController),
		AlarmType:               NewAlarmTypeHandler(deps.AlarmType),
		Unit:                    NewUnitHandler(deps.Unit),
		AlarmField:              NewAlarmFieldHandler(deps.AlarmField),
		AlarmTypeField:          NewAlarmTypeFieldHandler(deps.AlarmTypeField),
		BacnetAlarm:             NewBacnetAlarmHandler(deps.BacnetAlarm),
	}
}
