package historycapture

import (
	"context"
	"fmt"
	domainFieldDevice "github.com/besart951/go_infra_link/backend/internal/domain/facility/fielddevice"
	domainHierarchy "github.com/besart951/go_infra_link/backend/internal/domain/facility/hierarchy"
	domainObjectData "github.com/besart951/go_infra_link/backend/internal/domain/facility/objectdata"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	domainHistory "github.com/besart951/go_infra_link/backend/internal/domain/history"
	"github.com/google/uuid"
)

type repository[T any] struct {
	domain.Repository[T]
	audit audit[T]
}

func WrapRepository[T any](table string, next domain.Repository[T], store ChangeStore) domain.Repository[T] {
	if store == nil {
		return next
	}
	return &repository[T]{Repository: next, audit: newAudit[T](table, store)}
}

func (r *repository[T]) Create(ctx context.Context, entity *T) error {
	return r.audit.create(ctx, r.Repository, entity)
}

func (r *repository[T]) Update(ctx context.Context, entity *T) error {
	return r.audit.update(ctx, r.Repository, entity)
}

func (r *repository[T]) DeleteByIds(ctx context.Context, ids []uuid.UUID) error {
	return r.audit.deleteByIds(ctx, r.Repository, ids)
}

type BuildingRepository struct {
	domainFacility.BuildingRepository
	audit audit[domainFacility.Building]
}

func WrapBuilding(next domainFacility.BuildingRepository, store ChangeStore) domainFacility.BuildingRepository {
	return &BuildingRepository{BuildingRepository: next, audit: newAudit[domainFacility.Building]("buildings", store)}
}

func (r *BuildingRepository) Create(ctx context.Context, entity *domainFacility.Building) error {
	return r.audit.create(ctx, r.BuildingRepository, entity)
}
func (r *BuildingRepository) Update(ctx context.Context, entity *domainFacility.Building) error {
	return r.audit.update(ctx, r.BuildingRepository, entity)
}
func (r *BuildingRepository) DeleteByIds(ctx context.Context, ids []uuid.UUID) error {
	return r.audit.deleteByIds(ctx, r.BuildingRepository, ids)
}

type SystemTypeRepository struct {
	domainFacility.SystemTypeRepository
	audit audit[domainFacility.SystemType]
}

func WrapSystemType(next domainFacility.SystemTypeRepository, store ChangeStore) domainFacility.SystemTypeRepository {
	return &SystemTypeRepository{SystemTypeRepository: next, audit: newAudit[domainFacility.SystemType]("system_types", store)}
}

func (r *SystemTypeRepository) Create(ctx context.Context, entity *domainFacility.SystemType) error {
	return r.audit.create(ctx, r.SystemTypeRepository, entity)
}
func (r *SystemTypeRepository) Update(ctx context.Context, entity *domainFacility.SystemType) error {
	return r.audit.update(ctx, r.SystemTypeRepository, entity)
}
func (r *SystemTypeRepository) DeleteByIds(ctx context.Context, ids []uuid.UUID) error {
	return r.audit.deleteByIds(ctx, r.SystemTypeRepository, ids)
}

type SystemPartRepository struct {
	domainFacility.SystemPartRepository
	audit audit[domainFacility.SystemPart]
}

func WrapSystemPart(next domainFacility.SystemPartRepository, store ChangeStore) domainFacility.SystemPartRepository {
	return &SystemPartRepository{SystemPartRepository: next, audit: newAudit[domainFacility.SystemPart]("system_parts", store)}
}

func (r *SystemPartRepository) Create(ctx context.Context, entity *domainFacility.SystemPart) error {
	return r.audit.create(ctx, r.SystemPartRepository, entity)
}
func (r *SystemPartRepository) Update(ctx context.Context, entity *domainFacility.SystemPart) error {
	return r.audit.update(ctx, r.SystemPartRepository, entity)
}
func (r *SystemPartRepository) DeleteByIds(ctx context.Context, ids []uuid.UUID) error {
	return r.audit.deleteByIds(ctx, r.SystemPartRepository, ids)
}

type ApparatRepository struct {
	domainFacility.ApparatRepository
	audit audit[domainFacility.Apparat]
}

func WrapApparat(next domainFacility.ApparatRepository, store ChangeStore) domainFacility.ApparatRepository {
	return &ApparatRepository{ApparatRepository: next, audit: newAudit[domainFacility.Apparat]("apparats", store)}
}

func (r *ApparatRepository) Create(ctx context.Context, entity *domainFacility.Apparat) error {
	return r.audit.create(ctx, r.ApparatRepository, entity)
}
func (r *ApparatRepository) Update(ctx context.Context, entity *domainFacility.Apparat) error {
	return r.audit.update(ctx, r.ApparatRepository, entity)
}
func (r *ApparatRepository) DeleteByIds(ctx context.Context, ids []uuid.UUID) error {
	return r.audit.deleteByIds(ctx, r.ApparatRepository, ids)
}

type ControlCabinetRepository struct {
	domainFacility.ControlCabinetRepository
	audit audit[domainFacility.ControlCabinet]
}

func WrapControlCabinet(next domainFacility.ControlCabinetRepository, store ChangeStore) domainFacility.ControlCabinetRepository {
	return &ControlCabinetRepository{ControlCabinetRepository: next, audit: newAudit[domainFacility.ControlCabinet]("control_cabinets", store)}
}

func (r *ControlCabinetRepository) Create(ctx context.Context, entity *domainFacility.ControlCabinet) error {
	return r.audit.create(ctx, r.ControlCabinetRepository, entity)
}
func (r *ControlCabinetRepository) Update(ctx context.Context, entity *domainFacility.ControlCabinet) error {
	return r.audit.update(ctx, r.ControlCabinetRepository, entity)
}
func (r *ControlCabinetRepository) DeleteByIds(ctx context.Context, ids []uuid.UUID) error {
	return r.audit.deleteByIds(ctx, r.ControlCabinetRepository, ids)
}

type SPSControllerRepository struct {
	domainFacility.SPSControllerRepository
	audit audit[domainFacility.SPSController]
}

func WrapSPSController(next domainFacility.SPSControllerRepository, store ChangeStore) domainFacility.SPSControllerRepository {
	return &SPSControllerRepository{SPSControllerRepository: next, audit: newAudit[domainFacility.SPSController]("sps_controllers", store)}
}

func (r *SPSControllerRepository) Create(ctx context.Context, entity *domainFacility.SPSController) error {
	return r.audit.create(ctx, r.SPSControllerRepository, entity)
}
func (r *SPSControllerRepository) Update(ctx context.Context, entity *domainFacility.SPSController) error {
	return r.audit.update(ctx, r.SPSControllerRepository, entity)
}
func (r *SPSControllerRepository) DeleteByIds(ctx context.Context, ids []uuid.UUID) error {
	return r.audit.deleteByIds(ctx, r.SPSControllerRepository, ids)
}

type SPSControllerSystemTypeStore struct {
	domainHierarchy.SPSControllerSystemTypeStore
	audit audit[domainFacility.SPSControllerSystemType]
	store ChangeStore
}

func WrapSPSControllerSystemType(next domainHierarchy.SPSControllerSystemTypeStore, store ChangeStore) domainHierarchy.SPSControllerSystemTypeStore {
	return &SPSControllerSystemTypeStore{SPSControllerSystemTypeStore: next, audit: newAudit[domainFacility.SPSControllerSystemType]("sps_controller_system_types", store), store: store}
}

func (r *SPSControllerSystemTypeStore) Create(ctx context.Context, entity *domainFacility.SPSControllerSystemType) error {
	return r.audit.create(ctx, r.SPSControllerSystemTypeStore, entity)
}
func (r *SPSControllerSystemTypeStore) Update(ctx context.Context, entity *domainFacility.SPSControllerSystemType) error {
	return r.audit.update(ctx, r.SPSControllerSystemTypeStore, entity)
}
func (r *SPSControllerSystemTypeStore) DeleteByIds(ctx context.Context, ids []uuid.UUID) error {
	return r.audit.deleteByIds(ctx, r.SPSControllerSystemTypeStore, ids)
}
func (r *SPSControllerSystemTypeStore) DeleteBySPSControllerIDs(ctx context.Context, ids []uuid.UUID) error {
	if len(ids) == 0 {
		return nil
	}
	return r.audit.deleteRows(ctx,
		func(ctx context.Context) (map[uuid.UUID]domainHistory.JSONB, error) {
			return r.store.LoadRowsWhere(ctx, "sps_controller_system_types", "sps_controller_id IN ?", ids)
		},
		func(ctx context.Context) error {
			return r.SPSControllerSystemTypeStore.DeleteBySPSControllerIDs(ctx, ids)
		},
	)
}

type FieldDeviceStore struct {
	domainFieldDevice.FieldDeviceStore
	audit audit[domainFacility.FieldDevice]
	store ChangeStore
}

func WrapFieldDevice(next domainFieldDevice.FieldDeviceStore, store ChangeStore) domainFieldDevice.FieldDeviceStore {
	return &FieldDeviceStore{FieldDeviceStore: next, audit: newAudit[domainFacility.FieldDevice]("field_devices", store), store: store}
}

func (r *FieldDeviceStore) Create(ctx context.Context, entity *domainFacility.FieldDevice) error {
	return r.audit.create(ctx, r.FieldDeviceStore, entity)
}
func (r *FieldDeviceStore) Update(ctx context.Context, entity *domainFacility.FieldDevice) error {
	return r.audit.update(ctx, r.FieldDeviceStore, entity)
}
func (r *FieldDeviceStore) DeleteByIds(ctx context.Context, ids []uuid.UUID) error {
	return r.audit.deleteByIds(ctx, r.FieldDeviceStore, ids)
}
func (r *FieldDeviceStore) BulkCreate(ctx context.Context, entities []*domainFacility.FieldDevice, batchSize int) error {
	return r.audit.bulkCreate(ctx,
		func(ctx context.Context) error { return r.FieldDeviceStore.BulkCreate(ctx, entities, batchSize) },
		func() []uuid.UUID { return idsOf(entities) },
	)
}
func (r *FieldDeviceStore) AssignSpecificationIDs(ctx context.Context, assignments map[uuid.UUID]uuid.UUID) error {
	if len(assignments) == 0 {
		return nil
	}
	ids := make([]uuid.UUID, 0, len(assignments))
	for id := range assignments {
		ids = append(ids, id)
	}
	before, err := r.store.LoadRows(ctx, "field_devices", ids)
	if err != nil {
		return err
	}
	if err := r.FieldDeviceStore.AssignSpecificationIDs(ctx, assignments); err != nil {
		return err
	}
	return r.store.RecordUpdates(ctx, "field_devices", before)
}
func (r *FieldDeviceStore) DeleteBySPSControllerSystemTypeIDs(ctx context.Context, ids []uuid.UUID) error {
	if len(ids) == 0 {
		return nil
	}
	return r.audit.deleteRows(ctx,
		func(ctx context.Context) (map[uuid.UUID]domainHistory.JSONB, error) {
			return r.store.LoadRowsWhere(ctx, "field_devices", "sps_controller_system_type_id IN ?", ids)
		},
		func(ctx context.Context) error {
			deleter, ok := r.FieldDeviceStore.(interface {
				DeleteBySPSControllerSystemTypeIDs(context.Context, []uuid.UUID) error
			})
			if ok {
				return deleter.DeleteBySPSControllerSystemTypeIDs(ctx, ids)
			}
			fieldDeviceIDs, err := r.FieldDeviceStore.GetIDsBySPSControllerSystemTypeIDs(ctx, ids)
			if err != nil {
				return err
			}
			return r.FieldDeviceStore.DeleteByIds(ctx, fieldDeviceIDs)
		},
	)
}

type SpecificationStore struct {
	domainFieldDevice.SpecificationStore
	audit audit[domainFacility.Specification]
	store ChangeStore
}

func WrapSpecification(next domainFieldDevice.SpecificationStore, store ChangeStore) domainFieldDevice.SpecificationStore {
	return &SpecificationStore{SpecificationStore: next, audit: newAudit[domainFacility.Specification]("specifications", store), store: store}
}

func (r *SpecificationStore) Create(ctx context.Context, entity *domainFacility.Specification) error {
	return r.audit.create(ctx, r.SpecificationStore, entity)
}
func (r *SpecificationStore) Update(ctx context.Context, entity *domainFacility.Specification) error {
	return r.audit.update(ctx, r.SpecificationStore, entity)
}
func (r *SpecificationStore) DeleteByIds(ctx context.Context, ids []uuid.UUID) error {
	return r.audit.deleteByIds(ctx, r.SpecificationStore, ids)
}
func (r *SpecificationStore) BulkCreate(ctx context.Context, entities []*domainFacility.Specification, batchSize int) error {
	return r.audit.bulkCreate(ctx,
		func(ctx context.Context) error { return r.SpecificationStore.BulkCreate(ctx, entities, batchSize) },
		func() []uuid.UUID { return idsOf(entities) },
	)
}
func (r *SpecificationStore) DeleteByFieldDeviceIDs(ctx context.Context, ids []uuid.UUID) error {
	if len(ids) == 0 {
		return nil
	}
	return r.audit.deleteRows(ctx,
		func(ctx context.Context) (map[uuid.UUID]domainHistory.JSONB, error) {
			return r.store.LoadRowsWhere(ctx, "specifications", "field_device_id IN ?", ids)
		},
		func(ctx context.Context) error { return r.SpecificationStore.DeleteByFieldDeviceIDs(ctx, ids) },
	)
}
func (r *SpecificationStore) DeleteBySPSControllerSystemTypeIDs(ctx context.Context, ids []uuid.UUID) error {
	if len(ids) == 0 {
		return nil
	}
	return r.audit.deleteRows(ctx,
		func(ctx context.Context) (map[uuid.UUID]domainHistory.JSONB, error) {
			return r.store.LoadRowsWhere(ctx, "specifications", `
				field_device_id IN (
					SELECT id FROM field_devices WHERE sps_controller_system_type_id IN ?
				)`, ids)
		},
		func(ctx context.Context) error {
			deleter, ok := r.SpecificationStore.(interface {
				DeleteBySPSControllerSystemTypeIDs(context.Context, []uuid.UUID) error
			})
			if ok {
				return deleter.DeleteBySPSControllerSystemTypeIDs(ctx, ids)
			}
			return nil
		},
	)
}

type BacnetObjectStore struct {
	domainObjectData.BacnetObjectStore
	audit audit[domainFacility.BacnetObject]
	store ChangeStore
}

var _ domainObjectData.ObjectDataBacnetObjectCreator = (*BacnetObjectStore)(nil)

func WrapBacnetObject(next domainObjectData.BacnetObjectStore, store ChangeStore) domainObjectData.BacnetObjectStore {
	return &BacnetObjectStore{BacnetObjectStore: next, audit: newAudit[domainFacility.BacnetObject]("bacnet_objects", store), store: store}
}

func (r *BacnetObjectStore) Create(ctx context.Context, entity *domainFacility.BacnetObject) error {
	return r.audit.create(ctx, r.BacnetObjectStore, entity)
}
func (r *BacnetObjectStore) CreateForObjectData(
	ctx context.Context,
	objectDataID uuid.UUID,
	entity *domainFacility.BacnetObject,
) error {
	creator, ok := r.BacnetObjectStore.(domainObjectData.ObjectDataBacnetObjectCreator)
	if !ok {
		return fmt.Errorf("BACnet object store does not support ObjectData-owned create")
	}
	if err := creator.CreateForObjectData(ctx, objectDataID, entity); err != nil {
		return err
	}
	return r.audit.recordCreated(ctx, entity.ID)
}
func (r *BacnetObjectStore) Update(ctx context.Context, entity *domainFacility.BacnetObject) error {
	return r.audit.update(ctx, r.BacnetObjectStore, entity)
}
func (r *BacnetObjectStore) DeleteByIds(ctx context.Context, ids []uuid.UUID) error {
	return r.audit.deleteByIds(ctx, r.BacnetObjectStore, ids)
}
func (r *BacnetObjectStore) BulkCreate(ctx context.Context, entities []*domainFacility.BacnetObject, batchSize int) error {
	return r.audit.bulkCreate(ctx,
		func(ctx context.Context) error { return r.BacnetObjectStore.BulkCreate(ctx, entities, batchSize) },
		func() []uuid.UUID { return idsOf(entities) },
	)
}
func (r *BacnetObjectStore) AssignSoftwareReferenceIDs(ctx context.Context, assignments map[uuid.UUID]uuid.UUID) error {
	if len(assignments) == 0 {
		return nil
	}
	ids := make([]uuid.UUID, 0, len(assignments))
	for id := range assignments {
		ids = append(ids, id)
	}
	before, err := r.store.LoadRows(ctx, "bacnet_objects", ids)
	if err != nil {
		return err
	}
	if err := r.BacnetObjectStore.AssignSoftwareReferenceIDs(ctx, assignments); err != nil {
		return err
	}
	return r.store.RecordUpdates(ctx, "bacnet_objects", before)
}
func (r *BacnetObjectStore) DeleteByFieldDeviceIDs(ctx context.Context, ids []uuid.UUID) error {
	if len(ids) == 0 {
		return nil
	}
	return r.audit.deleteRows(ctx,
		func(ctx context.Context) (map[uuid.UUID]domainHistory.JSONB, error) {
			return r.store.LoadRowsWhere(ctx, "bacnet_objects", "field_device_id IN ?", ids)
		},
		func(ctx context.Context) error { return r.BacnetObjectStore.DeleteByFieldDeviceIDs(ctx, ids) },
	)
}
func (r *BacnetObjectStore) DeleteBySPSControllerSystemTypeIDs(ctx context.Context, ids []uuid.UUID) error {
	if len(ids) == 0 {
		return nil
	}
	return r.audit.deleteRows(ctx,
		func(ctx context.Context) (map[uuid.UUID]domainHistory.JSONB, error) {
			return r.store.LoadRowsWhere(ctx, "bacnet_objects", `
				field_device_id IN (
					SELECT id FROM field_devices WHERE sps_controller_system_type_id IN ?
				)`, ids)
		},
		func(ctx context.Context) error {
			deleter, ok := r.BacnetObjectStore.(interface {
				DeleteBySPSControllerSystemTypeIDs(context.Context, []uuid.UUID) error
			})
			if ok {
				return deleter.DeleteBySPSControllerSystemTypeIDs(ctx, ids)
			}
			return nil
		},
	)
}

type ObjectDataStore struct {
	domainObjectData.ObjectDataStore
	audit audit[domainFacility.ObjectData]
}

func WrapObjectData(next domainObjectData.ObjectDataStore, store ChangeStore) domainObjectData.ObjectDataStore {
	return &ObjectDataStore{ObjectDataStore: next, audit: newAudit[domainFacility.ObjectData]("object_data", store)}
}

func (r *ObjectDataStore) Create(ctx context.Context, entity *domainFacility.ObjectData) error {
	return r.audit.create(ctx, r.ObjectDataStore, entity)
}
func (r *ObjectDataStore) Update(ctx context.Context, entity *domainFacility.ObjectData) error {
	return r.audit.update(ctx, r.ObjectDataStore, entity)
}
func (r *ObjectDataStore) DeleteByIds(ctx context.Context, ids []uuid.UUID) error {
	return r.audit.deleteByIds(ctx, r.ObjectDataStore, ids)
}

type AlarmDefinitionRepository struct {
	domainFacility.AlarmDefinitionRepository
	audit audit[domainFacility.AlarmDefinition]
}

func WrapAlarmDefinition(next domainFacility.AlarmDefinitionRepository, store ChangeStore) domainFacility.AlarmDefinitionRepository {
	return &AlarmDefinitionRepository{AlarmDefinitionRepository: next, audit: newAudit[domainFacility.AlarmDefinition]("alarm_definitions", store)}
}

func (r *AlarmDefinitionRepository) Create(ctx context.Context, entity *domainFacility.AlarmDefinition) error {
	return r.audit.create(ctx, r.AlarmDefinitionRepository, entity)
}
func (r *AlarmDefinitionRepository) Update(ctx context.Context, entity *domainFacility.AlarmDefinition) error {
	return r.audit.update(ctx, r.AlarmDefinitionRepository, entity)
}
func (r *AlarmDefinitionRepository) DeleteByIds(ctx context.Context, ids []uuid.UUID) error {
	return r.audit.deleteByIds(ctx, r.AlarmDefinitionRepository, ids)
}

type AlarmTypeRepository struct {
	domainFacility.AlarmTypeRepository
	audit audit[domainFacility.AlarmType]
}

func WrapAlarmType(next domainFacility.AlarmTypeRepository, store ChangeStore) domainFacility.AlarmTypeRepository {
	return &AlarmTypeRepository{AlarmTypeRepository: next, audit: newAudit[domainFacility.AlarmType]("alarm_types", store)}
}

func (r *AlarmTypeRepository) Create(ctx context.Context, entity *domainFacility.AlarmType) error {
	return r.audit.create(ctx, r.AlarmTypeRepository, entity)
}
func (r *AlarmTypeRepository) Update(ctx context.Context, entity *domainFacility.AlarmType) error {
	return r.audit.update(ctx, r.AlarmTypeRepository, entity)
}
func (r *AlarmTypeRepository) DeleteByIds(ctx context.Context, ids []uuid.UUID) error {
	return r.audit.deleteByIds(ctx, r.AlarmTypeRepository, ids)
}

type BacnetObjectAlarmValueRepository struct {
	domainFacility.BacnetObjectAlarmValueRepository
	audit audit[domainFacility.BacnetObjectAlarmValue]
}

func WrapBacnetObjectAlarmValue(next domainFacility.BacnetObjectAlarmValueRepository, store ChangeStore) domainFacility.BacnetObjectAlarmValueRepository {
	return &BacnetObjectAlarmValueRepository{BacnetObjectAlarmValueRepository: next, audit: newAudit[domainFacility.BacnetObjectAlarmValue]("bacnet_object_alarm_values", store)}
}

func (r *BacnetObjectAlarmValueRepository) Create(ctx context.Context, entity *domainFacility.BacnetObjectAlarmValue) error {
	return r.audit.create(ctx, r.BacnetObjectAlarmValueRepository, entity)
}
func (r *BacnetObjectAlarmValueRepository) Update(ctx context.Context, entity *domainFacility.BacnetObjectAlarmValue) error {
	return r.audit.update(ctx, r.BacnetObjectAlarmValueRepository, entity)
}
func (r *BacnetObjectAlarmValueRepository) DeleteByIds(ctx context.Context, ids []uuid.UUID) error {
	return r.audit.deleteByIds(ctx, r.BacnetObjectAlarmValueRepository, ids)
}
func (r *BacnetObjectAlarmValueRepository) BulkCreate(ctx context.Context, values []*domainFacility.BacnetObjectAlarmValue, batchSize int) error {
	return r.audit.bulkCreate(ctx,
		func(ctx context.Context) error {
			return r.BacnetObjectAlarmValueRepository.BulkCreate(ctx, values, batchSize)
		},
		func() []uuid.UUID { return idsOf(values) },
	)
}
func (r *BacnetObjectAlarmValueRepository) ReplaceForBacnetObject(ctx context.Context, bacnetObjectID uuid.UUID, values []domainFacility.BacnetObjectAlarmValue) error {
	return r.audit.deleteRows(ctx,
		func(ctx context.Context) (map[uuid.UUID]domainHistory.JSONB, error) {
			return r.audit.store.LoadRowsWhere(ctx, "bacnet_object_alarm_values", "bacnet_object_id = ?", bacnetObjectID)
		},
		func(ctx context.Context) error {
			if err := r.BacnetObjectAlarmValueRepository.ReplaceForBacnetObject(ctx, bacnetObjectID, values); err != nil {
				return err
			}
			ids := make([]uuid.UUID, 0, len(values))
			for i := range values {
				ids = append(ids, values[i].ID)
			}
			return r.audit.recordCreatedIDs(ctx, ids)
		},
	)
}
