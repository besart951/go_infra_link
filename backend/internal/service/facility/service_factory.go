package facility

import (
	"context"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
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
	Units                    domainFacility.UnitRepository
	AlarmFields              domainFacility.AlarmFieldRepository
	AlarmTypes               domainFacility.AlarmTypeRepository
	AlarmTypeFields          domainFacility.AlarmTypeFieldRepository
	BacnetObjectAlarmValues  domainFacility.BacnetObjectAlarmValueRepository
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
}

// NewServices creates facility services using a factory-style constructor.
func NewServices(repos Repositories, cfgs ...Config) *Services {
	var cfg Config
	if len(cfgs) > 0 {
		cfg = cfgs[0]
	}
	tx := newTxCoordinator(cfg)

	hierarchyCopier := NewHierarchyCopier(
		repos.ControlCabinets,
		repos.Buildings,
		repos.SPSControllers,
		repos.SystemTypes,
		repos.SPSControllerSystemTypes,
		repos.FieldDevices,
		repos.Specifications,
		repos.BacnetObjects,
	)
	hierarchyCopier.bindTransactions(tx)

	fieldDeviceService := NewFieldDeviceService(
		repos.FieldDevices,
		repos.SPSControllerSystemTypes,
		repos.SystemTypes,
		repos.Apparats,
		repos.SystemParts,
		repos.Specifications,
		repos.BacnetObjects,
		repos.ObjectData,
		repos.AlarmTypes,
		repos.BacnetObjectAlarmValues,
	)
	fieldDeviceService.bindTransactions(tx)
	fieldDeviceService.bindChangeRecorder(cfg.ChangeRecorder)
	objectDataService := NewObjectDataService(
		repos.ObjectData,
		repos.BacnetObjects,
		repos.ObjectDataBacnetObjects,
		repos.Apparats,
		repos.AlarmDefinitions,
		repos.AlarmTypes,
	)
	objectDataService.bindTransactions(tx)
	bacnetObjectService := NewBacnetObjectService(
		repos.BacnetObjects,
		repos.FieldDevices,
		repos.ObjectData,
		repos.ObjectDataBacnetObjects,
		repos.AlarmDefinitions,
		repos.AlarmTypes,
	)
	bacnetObjectService.bindTransactions(tx)
	spsControllerService := NewSPSControllerService(
		repos.SPSControllers,
		repos.ControlCabinets,
		repos.Buildings,
		repos.SystemTypes,
		repos.SPSControllerSystemTypes,
		repos.FieldDevices,
		hierarchyCopier,
	)
	spsControllerService.bindTransactions(tx)
	controlCabinetService := NewControlCabinetService(
		repos.ControlCabinets,
		repos.Buildings,
		repos.SPSControllers,
		repos.SPSControllerSystemTypes,
		repos.FieldDevices,
		repos.BacnetObjects,
		repos.Specifications,
		hierarchyCopier,
	)
	controlCabinetService.bindTransactions(tx)

	return &Services{
		HierarchyCopier: hierarchyCopier,
		Building:        NewBuildingService(repos.Buildings),
		SystemType:      NewSystemTypeService(repos.SystemTypes),
		SystemPart:      NewSystemPartService(repos.SystemParts),
		Apparat:         NewApparatService(repos.Apparats, repos.SystemParts, repos.ObjectData),
		ControlCabinet:   controlCabinetService,
		FieldDevice:       fieldDeviceService,
		BacnetObject:      bacnetObjectService,
		SPSController:     spsControllerService,
		StateText:         NewStateTextService(repos.StateTexts),
		NotificationClass: NewNotificationClassService(repos.NotificationClasses),
		AlarmDefinition:   NewAlarmDefinitionService(repos.AlarmDefinitions),
		ObjectData:        objectDataService,
		Unit:              NewUnitService(repos.Units),
		AlarmField:        NewAlarmFieldService(repos.AlarmFields),
		AlarmTypeField:    NewAlarmTypeFieldService(repos.AlarmTypeFields),
		SPSControllerSystemType: NewSPSControllerSystemTypeService(
			repos.SPSControllerSystemTypes,
			hierarchyCopier,
		),
		AlarmType: NewAlarmTypeService(repos.AlarmTypes),
		BacnetAlarmValue: NewBacnetAlarmValueService(
			repos.BacnetObjectAlarmValues,
			repos.AlarmTypes,
			repos.BacnetObjects,
		),
	}
}
