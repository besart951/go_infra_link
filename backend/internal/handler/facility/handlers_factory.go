package facility

// ServiceDeps groups service dependencies for facility handler construction.
type ServiceDeps struct {
	Building                       BuildingService
	SystemType                     SystemTypeService
	SystemPart                     SystemPartService
	Apparat                        ApparatService
	ControlCabinet                 ControlCabinetService
	ControlCabinetCreator          ControlCabinetCreator
	ControlCabinetCloner           ControlCabinetCloner
	ControlCabinetUpdater          ControlCabinetUpdater
	ControlCabinetDeleter          ControlCabinetDeleter
	FieldDevice                    FieldDeviceService
	FieldDeviceMultiCreator        FieldDeviceMultiCreator
	FieldDeviceUpdater             FieldDeviceUpdater
	FieldDeviceDeleter             FieldDeviceDeleter
	FieldDeviceBulkUpdater         FieldDeviceBulkUpdater
	FieldDeviceBulkDeleter         FieldDeviceBulkDeleter
	BacnetObject                   BacnetObjectService
	BacnetObjectCreator            BacnetObjectCreator
	BacnetObjectUpdater            BacnetObjectUpdater
	SPSController                  SPSControllerService
	SPSControllerCreator           SPSControllerCreator
	SPSControllerCloner            SPSControllerCloner
	SPSControllerUpdater           SPSControllerUpdater
	SPSControllerDeleter           SPSControllerDeleter
	SPSControllerSystemTypeCloner  SPSControllerSystemTypeCloner
	SPSControllerSystemTypeDeleter SPSControllerSystemTypeDeleter
	StateText                      StateTextService
	NotificationClass              NotificationClassService
	AlarmDefinition                AlarmDefinitionService
	ObjectData                     ObjectDataService
	SPSControllerSystemType        SPSControllerSystemTypeService
	Export                         ExportService
	AlarmType                      AlarmTypeService
	Unit                           UnitService
	AlarmField                     AlarmFieldService
	AlarmTypeField                 AlarmTypeFieldService
	BacnetAlarm                    BacnetAlarmValueService
	BacnetAlarmReplacer            BacnetAlarmValueReplacer
	BacnetReferenceUsage           BacnetReferenceUsageService
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
	BacnetReferenceUsage    *BacnetReferenceUsageHandler
}

// NewHandlers creates facility handlers using service dependencies.
func NewHandlers(deps ServiceDeps) *Handlers {
	handlers := &Handlers{}
	registerFacilityHierarchyHandlers(handlers, deps)
	registerFacilityLookupHandlers(handlers, deps)
	registerFacilityAlarmHandlers(handlers, deps)
	handlers.Export = NewExportHandler(deps.Export)
	return handlers
}

func registerFacilityHierarchyHandlers(handlers *Handlers, deps ServiceDeps) {
	handlers.Building = NewBuildingHandler(deps.Building)
	handlers.ControlCabinet = NewControlCabinetHandler(
		deps.ControlCabinet,
		deps.ControlCabinetCreator,
		deps.ControlCabinetCloner,
		deps.ControlCabinetUpdater,
		deps.ControlCabinetDeleter,
	)
	handlers.SPSController = NewSPSControllerHandler(
		deps.SPSController,
		deps.SPSControllerCreator,
		deps.SPSControllerCloner,
		deps.SPSControllerUpdater,
		deps.SPSControllerDeleter,
	)
	handlers.SPSControllerSystemType = NewSPSControllerSystemTypeHandler(
		deps.SPSControllerSystemType,
		deps.SPSControllerSystemTypeCloner,
		deps.SPSControllerSystemTypeDeleter,
	)
	bulkUpdater := deps.FieldDeviceBulkUpdater
	if bulkUpdater == nil {
		bulkUpdater, _ = deps.FieldDevice.(FieldDeviceBulkUpdater)
	}
	bulkDeleter := deps.FieldDeviceBulkDeleter
	if bulkDeleter == nil {
		bulkDeleter, _ = deps.FieldDevice.(FieldDeviceBulkDeleter)
	}
	handlers.FieldDevice = NewFieldDeviceHandler(
		deps.FieldDevice,
		deps.FieldDeviceUpdater,
		deps.FieldDeviceDeleter,
		bulkUpdater,
		deps.FieldDeviceMultiCreator,
		bulkDeleter,
	)
	handlers.BacnetObject = NewBacnetObjectHandler(
		deps.BacnetObjectCreator,
		deps.BacnetObjectUpdater,
	)
	handlers.ObjectData = NewObjectDataHandler(deps.ObjectData, deps.BacnetObject, deps.Apparat)
	handlers.Validation = NewValidationHandler(deps.Building, deps.ControlCabinet, deps.SPSController)
}

func registerFacilityLookupHandlers(handlers *Handlers, deps ServiceDeps) {
	handlers.SystemType = NewSystemTypeHandler(deps.SystemType)
	handlers.SystemPart = NewSystemPartHandler(deps.SystemPart, deps.Apparat, deps.ObjectData)
	handlers.Apparat = NewApparatHandler(deps.Apparat)
	handlers.StateText = NewStateTextHandler(deps.StateText)
	handlers.NotificationClass = NewNotificationClassHandler(deps.NotificationClass)
	handlers.BacnetReferenceUsage = NewBacnetReferenceUsageHandler(deps.BacnetReferenceUsage)
}

func registerFacilityAlarmHandlers(handlers *Handlers, deps ServiceDeps) {
	handlers.AlarmDefinition = NewAlarmDefinitionHandler(deps.AlarmDefinition)
	handlers.AlarmType = NewAlarmTypeHandler(deps.AlarmType)
	handlers.Unit = NewUnitHandler(deps.Unit)
	handlers.AlarmField = NewAlarmFieldHandler(deps.AlarmField)
	handlers.AlarmTypeField = NewAlarmTypeFieldHandler(deps.AlarmTypeField)
	handlers.BacnetAlarm = NewBacnetAlarmHandler(deps.BacnetAlarm, deps.BacnetAlarmReplacer)
}
