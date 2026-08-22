package facility

import (
	"context"
	"net/http"

	domainExport "github.com/besart951/go_infra_link/backend/internal/domain/exporting"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	facilityservice "github.com/besart951/go_infra_link/backend/internal/service/facility"
	"github.com/google/uuid"
)

type ProjectRefreshBroadcaster interface {
	BroadcastRefreshForControlCabinet(ctx context.Context, actorID *uuid.UUID, controlCabinetID uuid.UUID, scope string)
	BroadcastRefreshForSPSController(ctx context.Context, actorID *uuid.UUID, spsControllerID uuid.UUID, scope string)
	BroadcastControlCabinetDelta(ctx context.Context, actorID *uuid.UUID, controlCabinet domainFacility.ControlCabinet, changedFields ...string)
	BroadcastSPSControllerDelta(ctx context.Context, actorID *uuid.UUID, spsController domainFacility.SPSController, changedFields ...string)
	BroadcastSPSControllerSystemTypeChange(ctx context.Context, actorID *uuid.UUID, spsControllerID, systemTypeID uuid.UUID, action string, changedFields ...string)
}

type ProjectFieldDeviceChangeBroadcaster interface {
	BroadcastFieldDeviceChange(ctx context.Context, actorID *uuid.UUID, fieldDeviceID uuid.UUID, action string, changedFields ...string)
}

type FacilityReferenceDataBroadcaster interface {
	BroadcastFacilityReferenceDataChange(ctx context.Context, resources ...string)
}

type FacilityChangeBroadcaster interface {
	BroadcastFacilityChange(ctx context.Context, resource, action string, ids []uuid.UUID, actorID *uuid.UUID)
}

type FacilityReferenceDataStreamer interface {
	Stream(w http.ResponseWriter, r *http.Request, userID uuid.UUID, readableResources map[string]struct{})
}

type FacilityReferenceDataRealtime interface {
	FacilityReferenceDataBroadcaster
	FacilityChangeBroadcaster
	FacilityReferenceDataStreamer
}

// FacilityMutationBroadcaster is the narrow seam required by route wrappers.
// Streaming remains independent of mutation delivery.
type FacilityMutationBroadcaster interface {
	FacilityReferenceDataBroadcaster
	FacilityChangeBroadcaster
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
	Import                  FieldDeviceImportService
	ExportDownload          domainExport.DownloadAuthorizer
	AlarmType               AlarmTypeService
	Unit                    UnitService
	AlarmField              AlarmFieldService
	AlarmTypeField          AlarmTypeFieldService
	BacnetAlarm             BacnetAlarmValueService
	BacnetReferenceUsage    BacnetReferenceUsageService
	DeleteImpact            DeleteImpactService
	FacilityJobs            *facilityservice.FacilityJobManager
	Collaboration           ProjectRefreshBroadcaster
	ReferenceData           FacilityReferenceDataRealtime
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
	Import                  *ImportHandler
	Validation              *ValidationHandler
	AlarmType               *AlarmTypeHandler
	Unit                    *UnitHandler
	AlarmField              *AlarmFieldHandler
	AlarmTypeField          *AlarmTypeFieldHandler
	BacnetAlarm             *BacnetAlarmHandler
	BacnetReferenceUsage    *BacnetReferenceUsageHandler
	DeleteImpact            *DeleteImpactHandler
	FacilityJob             *FacilityJobHandler
	ReferenceData           *FacilityReferenceDataStreamHandler
	Details                 *FacilityDetailHandler
	Realtime                FacilityMutationBroadcaster
}

// NewHandlers creates facility handlers using service dependencies.
func NewHandlers(deps ServiceDeps) *Handlers {
	handlers := &Handlers{}
	handlers.Realtime = deps.ReferenceData
	registerFacilityHierarchyHandlers(handlers, deps)
	registerFacilityLookupHandlers(handlers, deps)
	registerFacilityAlarmHandlers(handlers, deps)
	handlers.Export = NewExportHandler(deps.Export, deps.ExportDownload)
	handlers.Import = NewImportHandler(deps.Import)
	return handlers
}

func registerFacilityHierarchyHandlers(handlers *Handlers, deps ServiceDeps) {
	handlers.Building = NewBuildingHandler(deps.Building)
	handlers.ControlCabinet = NewControlCabinetHandler(deps.ControlCabinet, deps.Collaboration, deps.FacilityJobs)
	handlers.SPSController = NewSPSControllerHandler(deps.SPSController, deps.Collaboration, deps.FacilityJobs)
	handlers.SPSControllerSystemType = NewSPSControllerSystemTypeHandlerWithFacilityJobs(deps.SPSControllerSystemType, deps.Collaboration, deps.FacilityJobs)
	handlers.FieldDevice = NewFieldDeviceHandlerWithFacilityJobs(deps.FieldDevice, deps.Collaboration, deps.FacilityJobs)
	handlers.BacnetObject = NewBacnetObjectHandler(deps.BacnetObject, deps.Collaboration)
	handlers.ObjectData = NewObjectDataHandlerWithFacilityJobs(deps.ObjectData, deps.BacnetObject, deps.Apparat, deps.FacilityJobs)
	handlers.Validation = NewValidationHandler(deps.Building, deps.ControlCabinet, deps.SPSController)
	handlers.Details = NewFacilityDetailHandler(
		deps.Building,
		deps.ControlCabinet,
		deps.SPSController,
		deps.SPSControllerSystemType,
		deps.FieldDevice,
		deps.Apparat,
		deps.SystemPart,
	)
}

func registerFacilityLookupHandlers(handlers *Handlers, deps ServiceDeps) {
	handlers.SystemType = NewSystemTypeHandler(deps.SystemType)
	handlers.SystemPart = NewSystemPartHandler(deps.SystemPart, deps.Apparat, deps.ObjectData)
	handlers.Apparat = NewApparatHandler(deps.Apparat)
	handlers.ReferenceData = NewFacilityReferenceDataStreamHandler(deps.ReferenceData)
	handlers.StateText = NewStateTextHandler(deps.StateText)
	handlers.NotificationClass = NewNotificationClassHandler(deps.NotificationClass)
	handlers.BacnetReferenceUsage = NewBacnetReferenceUsageHandler(deps.BacnetReferenceUsage)
	handlers.DeleteImpact = NewDeleteImpactHandler(deps.DeleteImpact)
	handlers.FacilityJob = NewFacilityJobHandler(deps.FacilityJobs)
}

func registerFacilityAlarmHandlers(handlers *Handlers, deps ServiceDeps) {
	handlers.AlarmDefinition = NewAlarmDefinitionHandler(deps.AlarmDefinition)
	handlers.AlarmType = NewAlarmTypeHandler(deps.AlarmType)
	handlers.Unit = NewUnitHandler(deps.Unit)
	handlers.AlarmField = NewAlarmFieldHandler(deps.AlarmField)
	handlers.AlarmTypeField = NewAlarmTypeFieldHandler(deps.AlarmTypeField)
	handlers.BacnetAlarm = NewBacnetAlarmHandler(deps.BacnetAlarm)
}
