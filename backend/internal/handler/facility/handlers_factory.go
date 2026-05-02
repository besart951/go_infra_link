package facility

import (
	"context"

	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
)

type ProjectRefreshBroadcaster interface {
	BroadcastRefreshForControlCabinet(ctx context.Context, actorID *uuid.UUID, controlCabinetID uuid.UUID, scope string)
	BroadcastRefreshForSPSController(ctx context.Context, actorID *uuid.UUID, spsControllerID uuid.UUID, scope string)
	BroadcastControlCabinetDelta(ctx context.Context, actorID *uuid.UUID, controlCabinet domainFacility.ControlCabinet)
	BroadcastSPSControllerDelta(ctx context.Context, actorID *uuid.UUID, spsController domainFacility.SPSController)
}

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
	Collaboration           ProjectRefreshBroadcaster
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
	handlers := &Handlers{}
	registerFacilityHierarchyHandlers(handlers, deps)
	registerFacilityLookupHandlers(handlers, deps)
	registerFacilityAlarmHandlers(handlers, deps)
	handlers.Export = NewExportHandler(deps.Export)
	return handlers
}

func registerFacilityHierarchyHandlers(handlers *Handlers, deps ServiceDeps) {
	handlers.Building = NewBuildingHandler(deps.Building)
	handlers.ControlCabinet = NewControlCabinetHandler(deps.ControlCabinet, deps.Collaboration)
	handlers.SPSController = NewSPSControllerHandler(deps.SPSController, deps.Collaboration)
	handlers.SPSControllerSystemType = NewSPSControllerSystemTypeHandler(deps.SPSControllerSystemType)
	handlers.FieldDevice = NewFieldDeviceHandler(deps.FieldDevice)
	handlers.BacnetObject = NewBacnetObjectHandler(deps.BacnetObject)
	handlers.ObjectData = NewObjectDataHandler(deps.ObjectData, deps.BacnetObject, deps.Apparat)
	handlers.Validation = NewValidationHandler(deps.Building, deps.ControlCabinet, deps.SPSController)
}

func registerFacilityLookupHandlers(handlers *Handlers, deps ServiceDeps) {
	handlers.SystemType = NewSystemTypeHandler(deps.SystemType)
	handlers.SystemPart = NewSystemPartHandler(deps.SystemPart, deps.Apparat, deps.ObjectData)
	handlers.Apparat = NewApparatHandler(deps.Apparat)
	handlers.StateText = NewStateTextHandler(deps.StateText)
	handlers.NotificationClass = NewNotificationClassHandler(deps.NotificationClass)
}

func registerFacilityAlarmHandlers(handlers *Handlers, deps ServiceDeps) {
	handlers.AlarmDefinition = NewAlarmDefinitionHandler(deps.AlarmDefinition)
	handlers.AlarmType = NewAlarmTypeHandler(deps.AlarmType)
	handlers.Unit = NewUnitHandler(deps.Unit)
	handlers.AlarmField = NewAlarmFieldHandler(deps.AlarmField)
	handlers.AlarmTypeField = NewAlarmTypeFieldHandler(deps.AlarmTypeField)
	handlers.BacnetAlarm = NewBacnetAlarmHandler(deps.BacnetAlarm)
}
