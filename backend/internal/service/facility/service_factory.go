package facility

import (
	"context"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	domainFieldDevice "github.com/besart951/go_infra_link/backend/internal/domain/facility/fielddevice"
	domainHierarchy "github.com/besart951/go_infra_link/backend/internal/domain/facility/hierarchy"
	domainObjectData "github.com/besart951/go_infra_link/backend/internal/domain/facility/objectdata"
	serviceAlarm "github.com/besart951/go_infra_link/backend/internal/service/facility/alarm"
	serviceFieldDevice "github.com/besart951/go_infra_link/backend/internal/service/facility/fielddevice"
	serviceHierarchy "github.com/besart951/go_infra_link/backend/internal/service/facility/hierarchy"
	serviceObjectData "github.com/besart951/go_infra_link/backend/internal/service/facility/objectdata"
	serviceReference "github.com/besart951/go_infra_link/backend/internal/service/facility/reference"
	"github.com/google/uuid"
)

// baseService provides GetByID, List, and DeleteByID for services whose
// repository satisfies domain.Repository[T].
// Embed this in a concrete service struct to avoid repeating these three methods.
type baseService[T any] struct {
	repo         domain.Repository[T]
	defaultLimit int
}

func newBase[T any](repo domain.Repository[T], defaultLimit int) baseService[T] {
	return baseService[T]{repo: repo, defaultLimit: defaultLimit}
}

func (s *baseService[T]) GetByID(ctx context.Context, id uuid.UUID) (*T, error) {
	return domain.GetByID(ctx, s.repo, id)
}

func (s *baseService[T]) List(ctx context.Context, page, limit int, search string) (*domain.PaginatedList[T], error) {
	page, limit = domain.NormalizePagination(page, limit, s.defaultLimit)
	return s.repo.GetPaginatedList(ctx, domain.PaginationParams{
		Page:   page,
		Limit:  limit,
		Search: search,
	})
}

func (s *baseService[T]) DeleteByID(ctx context.Context, id uuid.UUID) error {
	return s.repo.DeleteByIds(ctx, []uuid.UUID{id})
}

// derefSlice converts []*T to []T by dereferencing each element.
func derefSlice[T any](ptrs []*T) []T {
	items := make([]T, len(ptrs))
	for i, p := range ptrs {
		items[i] = *p
	}
	return items
}

// extractIDs extracts UUIDs from a nil-safe slice of entity pointers.
func extractIDs[T any](items []*T, id func(*T) uuid.UUID) []uuid.UUID {
	ids := make([]uuid.UUID, 0, len(items))
	for _, item := range items {
		if item != nil {
			ids = append(ids, id(item))
		}
	}
	return ids
}

// Repositories groups facility repositories for service construction.
type Repositories struct {
	Buildings                domainFacility.BuildingRepository
	SystemTypes              domainFacility.SystemTypeRepository
	SystemParts              domainFacility.SystemPartRepository
	Specifications           domainFieldDevice.SpecificationStore
	Apparats                 domainFacility.ApparatRepository
	ControlCabinets          domainFacility.ControlCabinetRepository
	FieldDevices             domainFieldDevice.FieldDeviceStore
	SPSControllers           domainFacility.SPSControllerRepository
	SPSControllerSystemTypes domainHierarchy.SPSControllerSystemTypeStore
	BacnetObjects            domainObjectData.BacnetObjectStore
	ObjectData               domainObjectData.ObjectDataStore
	ObjectDataBacnetObjects  domainObjectData.ObjectDataBacnetObjectStore
	StateTexts               domainFacility.StateTextRepository
	NotificationClasses      domainFacility.NotificationClassRepository
	AlarmDefinitions         domainFacility.AlarmDefinitionRepository
	Units                    domainFacility.UnitRepository
	AlarmFields              domainFacility.AlarmFieldRepository
	AlarmTypes               domainFacility.AlarmTypeRepository
	AlarmTypeFields          domainFacility.AlarmTypeFieldRepository
	BacnetObjectAlarmValues  domainFacility.BacnetObjectAlarmValueRepository
	BacnetReferenceUsages    domainFacility.BacnetReferenceUsageRepository
}

func (r Repositories) FieldDeviceModule() serviceFieldDevice.Repositories {
	return serviceFieldDevice.Repositories{
		FieldDevices:             r.FieldDevices,
		SPSControllerSystemTypes: r.SPSControllerSystemTypes,
		SystemTypes:              r.SystemTypes,
		Apparats:                 r.Apparats,
		SystemParts:              r.SystemParts,
		Specifications:           r.Specifications,
		BacnetObjects:            r.BacnetObjects,
		ObjectData:               r.ObjectData,
		AlarmTypes:               r.AlarmTypes,
		BacnetObjectAlarmValues:  r.BacnetObjectAlarmValues,
	}
}

func (r Repositories) ObjectDataModule() serviceObjectData.Repositories {
	return serviceObjectData.Repositories{
		ObjectData:              r.ObjectData,
		BacnetObjects:           r.BacnetObjects,
		ObjectDataBacnetObjects: r.ObjectDataBacnetObjects,
		Apparats:                r.Apparats,
		FieldDevices:            r.FieldDevices,
		AlarmDefinitions:        r.AlarmDefinitions,
		AlarmTypes:              r.AlarmTypes,
	}
}

func (r Repositories) HierarchyModule() serviceHierarchy.Repositories {
	return serviceHierarchy.Repositories{
		Buildings:                r.Buildings,
		ControlCabinets:          r.ControlCabinets,
		SPSControllers:           r.SPSControllers,
		SystemTypes:              r.SystemTypes,
		SPSControllerSystemTypes: r.SPSControllerSystemTypes,
		FieldDevices:             r.FieldDevices,
		Specifications:           r.Specifications,
		BacnetObjects:            r.BacnetObjects,
	}
}

func (r Repositories) ReferenceModule() serviceReference.Repositories {
	return serviceReference.Repositories{
		SystemTypes:         r.SystemTypes,
		SystemParts:         r.SystemParts,
		Apparats:            r.Apparats,
		ObjectData:          r.ObjectData,
		StateTexts:          r.StateTexts,
		NotificationClasses: r.NotificationClasses,
	}
}

func (r Repositories) AlarmModule() serviceAlarm.Repositories {
	return serviceAlarm.Repositories{
		AlarmDefinitions:        r.AlarmDefinitions,
		Units:                   r.Units,
		AlarmFields:             r.AlarmFields,
		AlarmTypes:              r.AlarmTypes,
		AlarmTypeFields:         r.AlarmTypeFields,
		BacnetObjectAlarmValues: r.BacnetObjectAlarmValues,
		BacnetObjects:           r.BacnetObjects,
	}
}

// Services bundles all facility services.
type Services struct {
	HierarchyCopier         *HierarchyCopier
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
	Unit                    *UnitService
	AlarmField              *AlarmFieldService
	AlarmTypeField          *AlarmTypeFieldService
	BacnetAlarmValue        *BacnetAlarmValueService
	BacnetReferenceUsage    *BacnetReferenceUsageService
}

// NewServices creates facility services using a factory-style constructor.
func NewServices(repos Repositories, cfgs ...Config) *Services {
	var cfg Config
	if len(cfgs) > 0 {
		cfg = cfgs[0]
	}
	tx := newTxCoordinator(cfg)
	hierarchyRepos := repos.HierarchyModule()
	referenceRepos := repos.ReferenceModule()
	fieldDeviceRepos := repos.FieldDeviceModule()
	objectDataRepos := repos.ObjectDataModule()
	alarmRepos := repos.AlarmModule()

	hierarchyCopier := NewHierarchyCopier(
		hierarchyRepos.ControlCabinets,
		hierarchyRepos.Buildings,
		hierarchyRepos.SPSControllers,
		hierarchyRepos.SystemTypes,
		hierarchyRepos.SPSControllerSystemTypes,
		hierarchyRepos.FieldDevices,
		hierarchyRepos.Specifications,
		hierarchyRepos.BacnetObjects,
	)
	hierarchyCopier.bindTransactions(tx)

	fieldDeviceService := NewFieldDeviceService(
		fieldDeviceRepos.FieldDevices,
		fieldDeviceRepos.SPSControllerSystemTypes,
		fieldDeviceRepos.SystemTypes,
		fieldDeviceRepos.Apparats,
		fieldDeviceRepos.SystemParts,
		fieldDeviceRepos.Specifications,
		fieldDeviceRepos.BacnetObjects,
		fieldDeviceRepos.ObjectData,
		fieldDeviceRepos.AlarmTypes,
		fieldDeviceRepos.BacnetObjectAlarmValues,
	)
	fieldDeviceService.bindTransactions(tx)
	fieldDeviceService.bindChangeRecorder(cfg.ChangeRecorder)
	objectDataService := NewObjectDataService(
		objectDataRepos.ObjectData,
		objectDataRepos.BacnetObjects,
		objectDataRepos.ObjectDataBacnetObjects,
		objectDataRepos.Apparats,
		objectDataRepos.AlarmDefinitions,
		objectDataRepos.AlarmTypes,
		repos.BacnetReferenceUsages,
	)
	objectDataService.bindTransactions(tx)
	bacnetObjectService := NewBacnetObjectService(
		objectDataRepos.BacnetObjects,
		objectDataRepos.FieldDevices,
		objectDataRepos.ObjectData,
		objectDataRepos.ObjectDataBacnetObjects,
		objectDataRepos.AlarmDefinitions,
		objectDataRepos.AlarmTypes,
	)
	bacnetObjectService.bindTransactions(tx)
	spsControllerService := NewSPSControllerService(
		hierarchyRepos.SPSControllers,
		hierarchyRepos.ControlCabinets,
		hierarchyRepos.Buildings,
		hierarchyRepos.SystemTypes,
		hierarchyRepos.SPSControllerSystemTypes,
		hierarchyRepos.FieldDevices,
		hierarchyCopier,
	)
	spsControllerService.bindTransactions(tx)
	controlCabinetService := NewControlCabinetService(
		hierarchyRepos.ControlCabinets,
		hierarchyRepos.Buildings,
		hierarchyRepos.SPSControllers,
		hierarchyRepos.SPSControllerSystemTypes,
		hierarchyRepos.FieldDevices,
		hierarchyRepos.BacnetObjects,
		hierarchyRepos.Specifications,
		hierarchyCopier,
	)
	controlCabinetService.bindTransactions(tx)

	return &Services{
		HierarchyCopier:   hierarchyCopier,
		Building:          NewBuildingService(hierarchyRepos.Buildings),
		SystemType:        NewSystemTypeService(referenceRepos.SystemTypes, repos.BacnetReferenceUsages),
		SystemPart:        NewSystemPartService(referenceRepos.SystemParts, repos.BacnetReferenceUsages),
		Apparat:           NewApparatService(referenceRepos.Apparats, referenceRepos.SystemParts, referenceRepos.ObjectData, repos.BacnetReferenceUsages),
		ControlCabinet:    controlCabinetService,
		FieldDevice:       fieldDeviceService,
		BacnetObject:      bacnetObjectService,
		SPSController:     spsControllerService,
		StateText:         NewStateTextService(referenceRepos.StateTexts, repos.BacnetReferenceUsages),
		NotificationClass: NewNotificationClassService(referenceRepos.NotificationClasses, repos.BacnetReferenceUsages),
		AlarmDefinition:   NewAlarmDefinitionService(alarmRepos.AlarmDefinitions, repos.BacnetReferenceUsages),
		ObjectData:        objectDataService,
		Unit:              NewUnitService(alarmRepos.Units),
		AlarmField:        NewAlarmFieldService(alarmRepos.AlarmFields),
		AlarmTypeField:    NewAlarmTypeFieldService(alarmRepos.AlarmTypeFields),
		SPSControllerSystemType: NewSPSControllerSystemTypeService(
			hierarchyRepos.SPSControllerSystemTypes,
			hierarchyCopier,
		),
		AlarmType: NewAlarmTypeService(alarmRepos.AlarmTypes, repos.BacnetReferenceUsages),
		BacnetAlarmValue: NewBacnetAlarmValueService(
			alarmRepos.BacnetObjectAlarmValues,
			alarmRepos.AlarmTypes,
			alarmRepos.BacnetObjects,
		),
		BacnetReferenceUsage: NewBacnetReferenceUsageService(repos.BacnetReferenceUsages),
	}
}
