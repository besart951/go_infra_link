package facility_test

import (
	"context"
	"errors"
	"testing"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/besart951/go_infra_link/backend/internal/service/facility"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type txFieldDeviceStore struct {
	*fakeFieldDeviceStore
	getIDsErr error
}

func (r *txFieldDeviceStore) GetIDsBySPSControllerSystemTypeIDs(ctx context.Context, ids []uuid.UUID) ([]uuid.UUID, error) {
	if r.getIDsErr != nil {
		return nil, r.getIDsErr
	}
	return r.fakeFieldDeviceStore.GetIDsBySPSControllerSystemTypeIDs(ctx, ids)
}

type txObjectDataStore struct {
	*fakeObjectDataStore
	bacnetObjectIDs map[uuid.UUID][]uuid.UUID
}

func (r *txObjectDataStore) GetBacnetObjectIDs(_ context.Context, objectDataID uuid.UUID) ([]uuid.UUID, error) {
	ids := r.bacnetObjectIDs[objectDataID]
	return append([]uuid.UUID(nil), ids...), nil
}

type fakeBacnetObjectStore struct {
	items                    map[uuid.UUID]*domainFacility.BacnetObject
	failCreate               error
	failBulkCreate           error
	failUpdate               error
	deleteByFieldDeviceCalls int
	createdCount             int
	updatedCount             int
}

func (r *fakeBacnetObjectStore) GetByIds(_ context.Context, ids []uuid.UUID) ([]*domainFacility.BacnetObject, error) {
	out := make([]*domainFacility.BacnetObject, 0, len(ids))
	for _, id := range ids {
		if item, ok := r.items[id]; ok {
			clone := *item
			out = append(out, &clone)
		}
	}
	return out, nil
}

func (r *fakeBacnetObjectStore) Create(_ context.Context, entity *domainFacility.BacnetObject) error {
	if r.failCreate != nil {
		return r.failCreate
	}
	if r.items == nil {
		r.items = make(map[uuid.UUID]*domainFacility.BacnetObject)
	}
	clone := *entity
	if clone.ID == uuid.Nil {
		clone.ID = uuid.New()
		entity.ID = clone.ID
	}
	r.items[clone.ID] = &clone
	r.createdCount++
	return nil
}

func (r *fakeBacnetObjectStore) BulkCreate(_ context.Context, entities []*domainFacility.BacnetObject, batchSize int) error {
	if r.failBulkCreate != nil {
		return r.failBulkCreate
	}
	for _, entity := range entities {
		if err := r.Create(context.Background(), entity); err != nil {
			return err
		}
	}
	return nil
}

func (r *fakeBacnetObjectStore) Update(_ context.Context, entity *domainFacility.BacnetObject) error {
	if r.failUpdate != nil {
		return r.failUpdate
	}
	if r.items == nil {
		r.items = make(map[uuid.UUID]*domainFacility.BacnetObject)
	}
	clone := *entity
	r.items[clone.ID] = &clone
	r.updatedCount++
	return nil
}

func (r *fakeBacnetObjectStore) DeleteByIds(_ context.Context, ids []uuid.UUID) error {
	for _, id := range ids {
		delete(r.items, id)
	}
	return nil
}

func (r *fakeBacnetObjectStore) GetPaginatedList(_ context.Context, params domain.PaginationParams) (*domain.PaginatedList[domainFacility.BacnetObject], error) {
	items := make([]domainFacility.BacnetObject, 0, len(r.items))
	for _, item := range r.items {
		items = append(items, *item)
	}
	return &domain.PaginatedList[domainFacility.BacnetObject]{
		Items:      items,
		Total:      int64(len(items)),
		Page:       1,
		TotalPages: 1,
	}, nil
}

func (r *fakeBacnetObjectStore) GetByFieldDeviceIDs(_ context.Context, ids []uuid.UUID) ([]*domainFacility.BacnetObject, error) {
	idSet := make(map[uuid.UUID]struct{}, len(ids))
	for _, id := range ids {
		idSet[id] = struct{}{}
	}
	out := make([]*domainFacility.BacnetObject, 0)
	for _, item := range r.items {
		if item.FieldDeviceID == nil {
			continue
		}
		if _, ok := idSet[*item.FieldDeviceID]; !ok {
			continue
		}
		clone := *item
		out = append(out, &clone)
	}
	return out, nil
}

func (r *fakeBacnetObjectStore) DeleteByFieldDeviceIDs(_ context.Context, ids []uuid.UUID) error {
	r.deleteByFieldDeviceCalls++
	idSet := make(map[uuid.UUID]struct{}, len(ids))
	for _, id := range ids {
		idSet[id] = struct{}{}
	}
	for id, item := range r.items {
		if item.FieldDeviceID == nil {
			continue
		}
		if _, ok := idSet[*item.FieldDeviceID]; ok {
			delete(r.items, id)
		}
	}
	return nil
}

type fakeObjectDataBacnetObjectStore struct {
	addErr                error
	addCalls              int
	deleteByObjectDataIDs []uuid.UUID
	links                 map[uuid.UUID][]uuid.UUID
}

func (r *fakeObjectDataBacnetObjectStore) Add(_ context.Context, objectDataID uuid.UUID, bacnetObjectID uuid.UUID) error {
	r.addCalls++
	if r.addErr != nil {
		return r.addErr
	}
	if r.links == nil {
		r.links = make(map[uuid.UUID][]uuid.UUID)
	}
	r.links[objectDataID] = append(r.links[objectDataID], bacnetObjectID)
	return nil
}

func (r *fakeObjectDataBacnetObjectStore) Delete(_ context.Context, objectDataID uuid.UUID, bacnetObjectID uuid.UUID) error {
	ids := r.links[objectDataID]
	filtered := ids[:0]
	for _, id := range ids {
		if id != bacnetObjectID {
			filtered = append(filtered, id)
		}
	}
	r.links[objectDataID] = filtered
	return nil
}

func (r *fakeObjectDataBacnetObjectStore) DeleteByObjectDataID(_ context.Context, objectDataID uuid.UUID) error {
	r.deleteByObjectDataIDs = append(r.deleteByObjectDataIDs, objectDataID)
	delete(r.links, objectDataID)
	return nil
}

type fakeAlarmTypeRepo struct {
	items map[uuid.UUID]*domainFacility.AlarmType
}

func (r *fakeAlarmTypeRepo) GetByIds(_ context.Context, ids []uuid.UUID) ([]*domainFacility.AlarmType, error) {
	out := make([]*domainFacility.AlarmType, 0, len(ids))
	for _, id := range ids {
		if item, ok := r.items[id]; ok {
			clone := *item
			clone.Fields = append([]domainFacility.AlarmTypeField(nil), item.Fields...)
			out = append(out, &clone)
		}
	}
	return out, nil
}

func (r *fakeAlarmTypeRepo) Create(_ context.Context, entity *domainFacility.AlarmType) error {
	if r.items == nil {
		r.items = make(map[uuid.UUID]*domainFacility.AlarmType)
	}
	clone := *entity
	clone.Fields = append([]domainFacility.AlarmTypeField(nil), entity.Fields...)
	r.items[entity.ID] = &clone
	return nil
}

func (r *fakeAlarmTypeRepo) Update(_ context.Context, entity *domainFacility.AlarmType) error {
	return r.Create(context.Background(), entity)
}

func (r *fakeAlarmTypeRepo) DeleteByIds(_ context.Context, ids []uuid.UUID) error {
	for _, id := range ids {
		delete(r.items, id)
	}
	return nil
}

func (r *fakeAlarmTypeRepo) GetPaginatedList(_ context.Context, params domain.PaginationParams) (*domain.PaginatedList[domainFacility.AlarmType], error) {
	items := make([]domainFacility.AlarmType, 0, len(r.items))
	for _, item := range r.items {
		clone := *item
		clone.Fields = append([]domainFacility.AlarmTypeField(nil), item.Fields...)
		items = append(items, clone)
	}
	return &domain.PaginatedList[domainFacility.AlarmType]{
		Items:      items,
		Total:      int64(len(items)),
		Page:       1,
		TotalPages: 1,
	}, nil
}

func (r *fakeAlarmTypeRepo) GetWithFields(_ context.Context, id uuid.UUID) (*domainFacility.AlarmType, error) {
	if item, ok := r.items[id]; ok {
		clone := *item
		clone.Fields = append([]domainFacility.AlarmTypeField(nil), item.Fields...)
		return &clone, nil
	}
	return nil, nil
}

func (r *fakeAlarmTypeRepo) ListWithFields(_ context.Context, params domain.PaginationParams) (*domain.PaginatedList[domainFacility.AlarmType], error) {
	return r.GetPaginatedList(context.Background(), params)
}

type fakeBacnetObjectAlarmValueRepo struct {
	items           map[uuid.UUID]*domainFacility.BacnetObjectAlarmValue
	failBulkCreate  error
	bulkCreateCalls int
}

func (r *fakeBacnetObjectAlarmValueRepo) GetByIds(_ context.Context, ids []uuid.UUID) ([]*domainFacility.BacnetObjectAlarmValue, error) {
	out := make([]*domainFacility.BacnetObjectAlarmValue, 0, len(ids))
	for _, id := range ids {
		if item, ok := r.items[id]; ok {
			clone := *item
			out = append(out, &clone)
		}
	}
	return out, nil
}

func (r *fakeBacnetObjectAlarmValueRepo) Create(_ context.Context, entity *domainFacility.BacnetObjectAlarmValue) error {
	if r.items == nil {
		r.items = make(map[uuid.UUID]*domainFacility.BacnetObjectAlarmValue)
	}
	clone := *entity
	if clone.ID == uuid.Nil {
		clone.ID = uuid.New()
		entity.ID = clone.ID
	}
	r.items[clone.ID] = &clone
	return nil
}

func (r *fakeBacnetObjectAlarmValueRepo) Update(_ context.Context, entity *domainFacility.BacnetObjectAlarmValue) error {
	return r.Create(context.Background(), entity)
}

func (r *fakeBacnetObjectAlarmValueRepo) DeleteByIds(_ context.Context, ids []uuid.UUID) error {
	for _, id := range ids {
		delete(r.items, id)
	}
	return nil
}

func (r *fakeBacnetObjectAlarmValueRepo) GetPaginatedList(_ context.Context, params domain.PaginationParams) (*domain.PaginatedList[domainFacility.BacnetObjectAlarmValue], error) {
	items := make([]domainFacility.BacnetObjectAlarmValue, 0, len(r.items))
	for _, item := range r.items {
		items = append(items, *item)
	}
	return &domain.PaginatedList[domainFacility.BacnetObjectAlarmValue]{
		Items:      items,
		Total:      int64(len(items)),
		Page:       1,
		TotalPages: 1,
	}, nil
}

func (r *fakeBacnetObjectAlarmValueRepo) GetByBacnetObjectID(_ context.Context, bacnetObjectID uuid.UUID) ([]domainFacility.BacnetObjectAlarmValue, error) {
	out := make([]domainFacility.BacnetObjectAlarmValue, 0)
	for _, item := range r.items {
		if item.BacnetObjectID == bacnetObjectID {
			out = append(out, *item)
		}
	}
	return out, nil
}

func (r *fakeBacnetObjectAlarmValueRepo) BulkCreate(_ context.Context, values []*domainFacility.BacnetObjectAlarmValue, batchSize int) error {
	r.bulkCreateCalls++
	if r.failBulkCreate != nil {
		return r.failBulkCreate
	}
	for _, value := range values {
		if err := r.Create(context.Background(), value); err != nil {
			return err
		}
	}
	return nil
}

func (r *fakeBacnetObjectAlarmValueRepo) ReplaceForBacnetObject(_ context.Context, bacnetObjectID uuid.UUID, values []domainFacility.BacnetObjectAlarmValue) error {
	for id, item := range r.items {
		if item.BacnetObjectID == bacnetObjectID {
			delete(r.items, id)
		}
	}
	for i := range values {
		value := values[i]
		if value.BacnetObjectID != bacnetObjectID {
			continue
		}
		if err := r.Create(context.Background(), &value); err != nil {
			return err
		}
	}
	return nil
}

type txSPSControllerSystemTypeRepo struct {
	*fakeSpsControllerSystemTypeRepo
	deleteCalls                int
	deleteBySPSControllerCalls int
}

func (r *txSPSControllerSystemTypeRepo) Create(_ context.Context, entity *domainFacility.SPSControllerSystemType) error {
	clone := *entity
	if clone.ID == uuid.Nil {
		clone.ID = uuid.New()
		entity.ID = clone.ID
	}
	r.items[clone.ID] = &clone
	return nil
}

func (r *txSPSControllerSystemTypeRepo) DeleteByIds(ctx context.Context, ids []uuid.UUID) error {
	r.deleteCalls++
	return r.fakeSpsControllerSystemTypeRepo.DeleteByIds(ctx, ids)
}

func (r *txSPSControllerSystemTypeRepo) DeleteBySPSControllerIDs(ctx context.Context, ids []uuid.UUID) error {
	r.deleteBySPSControllerCalls++
	return r.fakeSpsControllerSystemTypeRepo.DeleteBySPSControllerIDs(ctx, ids)
}

type fakeHierarchyControlCabinetRepo struct {
	items       map[uuid.UUID]*domainFacility.ControlCabinet
	createCalls int
	deleteCalls int
}

type fakeHierarchyBuildingRepo struct {
	items map[uuid.UUID]*domainFacility.Building
}

func (r *fakeHierarchyBuildingRepo) GetByIds(_ context.Context, ids []uuid.UUID) ([]*domainFacility.Building, error) {
	out := make([]*domainFacility.Building, 0, len(ids))
	for _, id := range ids {
		if item, ok := r.items[id]; ok {
			clone := *item
			out = append(out, &clone)
		}
	}
	return out, nil
}

func (r *fakeHierarchyBuildingRepo) Create(_ context.Context, entity *domainFacility.Building) error {
	if r.items == nil {
		r.items = make(map[uuid.UUID]*domainFacility.Building)
	}
	clone := *entity
	if clone.ID == uuid.Nil {
		clone.ID = uuid.New()
		entity.ID = clone.ID
	}
	r.items[clone.ID] = &clone
	return nil
}

func (r *fakeHierarchyBuildingRepo) Update(_ context.Context, entity *domainFacility.Building) error {
	clone := *entity
	r.items[clone.ID] = &clone
	return nil
}

func (r *fakeHierarchyBuildingRepo) DeleteByIds(_ context.Context, ids []uuid.UUID) error {
	for _, id := range ids {
		delete(r.items, id)
	}
	return nil
}

func (r *fakeHierarchyBuildingRepo) GetPaginatedList(_ context.Context, params domain.PaginationParams) (*domain.PaginatedList[domainFacility.Building], error) {
	items := make([]domainFacility.Building, 0, len(r.items))
	for _, item := range r.items {
		items = append(items, *item)
	}
	return &domain.PaginatedList[domainFacility.Building]{Items: items, Total: int64(len(items)), Page: 1, TotalPages: 1}, nil
}

func (r *fakeHierarchyBuildingRepo) ExistsIWSCodeGroup(_ context.Context, iwsCode string, buildingGroup int, excludeID *uuid.UUID) (bool, error) {
	return false, nil
}

func (r *fakeHierarchyControlCabinetRepo) GetByIds(_ context.Context, ids []uuid.UUID) ([]*domainFacility.ControlCabinet, error) {
	out := make([]*domainFacility.ControlCabinet, 0, len(ids))
	for _, id := range ids {
		if item, ok := r.items[id]; ok {
			clone := *item
			out = append(out, &clone)
		}
	}
	return out, nil
}

func (r *fakeHierarchyControlCabinetRepo) Create(_ context.Context, entity *domainFacility.ControlCabinet) error {
	if r.items == nil {
		r.items = make(map[uuid.UUID]*domainFacility.ControlCabinet)
	}
	clone := *entity
	if clone.ID == uuid.Nil {
		clone.ID = uuid.New()
		entity.ID = clone.ID
	}
	r.items[clone.ID] = &clone
	r.createCalls++
	return nil
}

func (r *fakeHierarchyControlCabinetRepo) Update(_ context.Context, entity *domainFacility.ControlCabinet) error {
	clone := *entity
	r.items[clone.ID] = &clone
	return nil
}

func (r *fakeHierarchyControlCabinetRepo) DeleteByIds(_ context.Context, ids []uuid.UUID) error {
	r.deleteCalls++
	for _, id := range ids {
		delete(r.items, id)
	}
	return nil
}

func (r *fakeHierarchyControlCabinetRepo) GetPaginatedList(_ context.Context, params domain.PaginationParams) (*domain.PaginatedList[domainFacility.ControlCabinet], error) {
	items := make([]domainFacility.ControlCabinet, 0, len(r.items))
	for _, item := range r.items {
		items = append(items, *item)
	}
	return &domain.PaginatedList[domainFacility.ControlCabinet]{Items: items, Total: int64(len(items)), Page: 1, TotalPages: 1}, nil
}

func (r *fakeHierarchyControlCabinetRepo) GetPaginatedListByBuildingID(_ context.Context, buildingID uuid.UUID, params domain.PaginationParams) (*domain.PaginatedList[domainFacility.ControlCabinet], error) {
	items := make([]domainFacility.ControlCabinet, 0)
	for _, item := range r.items {
		if item.BuildingID == buildingID {
			items = append(items, *item)
		}
	}
	return &domain.PaginatedList[domainFacility.ControlCabinet]{Items: items, Total: int64(len(items)), Page: 1, TotalPages: 1}, nil
}

func (r *fakeHierarchyControlCabinetRepo) GetIDsByBuildingID(_ context.Context, buildingID uuid.UUID) ([]uuid.UUID, error) {
	ids := make([]uuid.UUID, 0)
	for id, item := range r.items {
		if item.BuildingID == buildingID {
			ids = append(ids, id)
		}
	}
	return ids, nil
}

func (r *fakeHierarchyControlCabinetRepo) ExistsControlCabinetNr(_ context.Context, buildingID uuid.UUID, controlCabinetNr string, excludeID *uuid.UUID) (bool, error) {
	for id, item := range r.items {
		if excludeID != nil && id == *excludeID {
			continue
		}
		if item.BuildingID != buildingID || item.ControlCabinetNr == nil {
			continue
		}
		if *item.ControlCabinetNr == controlCabinetNr {
			return true, nil
		}
	}
	return false, nil
}

type fakeHierarchySPSControllerRepo struct {
	items       map[uuid.UUID]*domainFacility.SPSController
	createCalls int
	deleteCalls int
}

func (r *fakeHierarchySPSControllerRepo) GetByIds(_ context.Context, ids []uuid.UUID) ([]*domainFacility.SPSController, error) {
	out := make([]*domainFacility.SPSController, 0, len(ids))
	for _, id := range ids {
		if item, ok := r.items[id]; ok {
			clone := *item
			out = append(out, &clone)
		}
	}
	return out, nil
}

func (r *fakeHierarchySPSControllerRepo) Create(_ context.Context, entity *domainFacility.SPSController) error {
	if r.items == nil {
		r.items = make(map[uuid.UUID]*domainFacility.SPSController)
	}
	clone := *entity
	if clone.ID == uuid.Nil {
		clone.ID = uuid.New()
		entity.ID = clone.ID
	}
	r.items[clone.ID] = &clone
	r.createCalls++
	return nil
}

func (r *fakeHierarchySPSControllerRepo) Update(_ context.Context, entity *domainFacility.SPSController) error {
	clone := *entity
	r.items[clone.ID] = &clone
	return nil
}

func (r *fakeHierarchySPSControllerRepo) DeleteByIds(_ context.Context, ids []uuid.UUID) error {
	r.deleteCalls++
	for _, id := range ids {
		delete(r.items, id)
	}
	return nil
}

func (r *fakeHierarchySPSControllerRepo) GetPaginatedList(_ context.Context, params domain.PaginationParams) (*domain.PaginatedList[domainFacility.SPSController], error) {
	items := make([]domainFacility.SPSController, 0, len(r.items))
	for _, item := range r.items {
		items = append(items, *item)
	}
	return &domain.PaginatedList[domainFacility.SPSController]{Items: items, Total: int64(len(items)), Page: 1, TotalPages: 1}, nil
}

func (r *fakeHierarchySPSControllerRepo) GetPaginatedListByControlCabinetID(_ context.Context, controlCabinetID uuid.UUID, params domain.PaginationParams) (*domain.PaginatedList[domainFacility.SPSController], error) {
	items := make([]domainFacility.SPSController, 0)
	for _, item := range r.items {
		if item.ControlCabinetID == controlCabinetID {
			items = append(items, *item)
		}
	}
	return &domain.PaginatedList[domainFacility.SPSController]{Items: items, Total: int64(len(items)), Page: 1, TotalPages: 1}, nil
}

func (r *fakeHierarchySPSControllerRepo) GetIDsByControlCabinetID(_ context.Context, controlCabinetID uuid.UUID) ([]uuid.UUID, error) {
	ids := make([]uuid.UUID, 0)
	for id, item := range r.items {
		if item.ControlCabinetID == controlCabinetID {
			ids = append(ids, id)
		}
	}
	return ids, nil
}

func (r *fakeHierarchySPSControllerRepo) GetIDsByControlCabinetIDs(_ context.Context, controlCabinetIDs []uuid.UUID) ([]uuid.UUID, error) {
	set := make(map[uuid.UUID]struct{}, len(controlCabinetIDs))
	for _, id := range controlCabinetIDs {
		set[id] = struct{}{}
	}
	ids := make([]uuid.UUID, 0)
	for id, item := range r.items {
		if _, ok := set[item.ControlCabinetID]; ok {
			ids = append(ids, id)
		}
	}
	return ids, nil
}

func (r *fakeHierarchySPSControllerRepo) ListGADevicesByControlCabinetID(_ context.Context, controlCabinetID uuid.UUID) ([]string, error) {
	devices := make([]string, 0)
	for _, item := range r.items {
		if item.ControlCabinetID != controlCabinetID || item.GADevice == nil {
			continue
		}
		devices = append(devices, *item.GADevice)
	}
	return devices, nil
}

func (r *fakeHierarchySPSControllerRepo) ExistsDeviceName(_ context.Context, controlCabinetID uuid.UUID, deviceName string, excludeID *uuid.UUID) (bool, error) {
	for id, item := range r.items {
		if excludeID != nil && id == *excludeID {
			continue
		}
		if item.ControlCabinetID == controlCabinetID && item.DeviceName == deviceName {
			return true, nil
		}
	}
	return false, nil
}

func (r *fakeHierarchySPSControllerRepo) ExistsGADevice(_ context.Context, controlCabinetID uuid.UUID, gaDevice string, excludeID *uuid.UUID) (bool, error) {
	for id, item := range r.items {
		if excludeID != nil && id == *excludeID {
			continue
		}
		if item.ControlCabinetID == controlCabinetID && item.GADevice != nil && *item.GADevice == gaDevice {
			return true, nil
		}
	}
	return false, nil
}

func (r *fakeHierarchySPSControllerRepo) ExistsIPAddressVlan(_ context.Context, ipAddress string, vlan string, excludeID *uuid.UUID) (bool, error) {
	return false, nil
}

func (r *fakeHierarchySPSControllerRepo) GetByIdsForExport(_ context.Context, ids []uuid.UUID) ([]domainFacility.SPSController, error) {
	out := make([]domainFacility.SPSController, 0, len(ids))
	for _, id := range ids {
		if item, ok := r.items[id]; ok {
			out = append(out, *item)
		}
	}
	return out, nil
}

func newTxServices(baseRepos, txRepos facility.Repositories, runnerCalls *int) *facility.Services {
	return facility.NewServices(baseRepos, facility.Config{
		TxRunner: func(run func(tx *gorm.DB) error) error {
			*runnerCalls++
			return run(nil)
		},
		TxFactory: func(tx *gorm.DB) (*facility.Services, error) {
			return facility.NewServices(txRepos), nil
		},
	})
}

func TestFacilityTransaction_FieldDeviceCreateFailureDoesNotCleanupOutsideTransaction(t *testing.T) {
	templateID := uuid.New()
	objectDataID := uuid.New()
	fieldDeviceID := uuid.New()
	spsSystemTypeID := uuid.New()
	systemTypeID := uuid.New()
	apparatID := uuid.New()
	systemPartID := uuid.New()
	cloneErr := errors.New("clone failed")

	baseFieldDevices := &fakeFieldDeviceStore{items: map[uuid.UUID]*domainFacility.FieldDevice{}}
	txFieldDevices := &fakeFieldDeviceStore{items: map[uuid.UUID]*domainFacility.FieldDevice{}}
	sharedSystemTypes := &fakeSystemTypeRepo{items: map[uuid.UUID]*domainFacility.SystemType{
		systemTypeID: {Base: domain.Base{ID: systemTypeID}, NumberMin: 1, NumberMax: 99},
	}}
	sharedSystemParts := &fakeSystemPartRepo{items: map[uuid.UUID]*domainFacility.SystemPart{
		systemPartID: {Base: domain.Base{ID: systemPartID}, ShortName: "AIR", Name: "Air"},
	}}
	sharedApparats := &fakeApparatRepo{items: map[uuid.UUID]*domainFacility.Apparat{
		apparatID: {Base: domain.Base{ID: apparatID}, ShortName: "PMP", Name: "Pump"},
	}}
	sharedSPSSystemTypes := &fakeSpsControllerSystemTypeRepo{items: map[uuid.UUID]*domainFacility.SPSControllerSystemType{
		spsSystemTypeID: {Base: domain.Base{ID: spsSystemTypeID}, SystemTypeID: systemTypeID},
	}}
	txObjectDatas := &txObjectDataStore{
		fakeObjectDataStore: &fakeObjectDataStore{items: map[uuid.UUID]*domainFacility.ObjectData{
			objectDataID: {Base: domain.Base{ID: objectDataID}, IsActive: true},
		}},
		bacnetObjectIDs: map[uuid.UUID][]uuid.UUID{objectDataID: {templateID}},
	}
	txBacnetObjects := &fakeBacnetObjectStore{
		items: map[uuid.UUID]*domainFacility.BacnetObject{
			templateID: {
				Base:           domain.Base{ID: templateID},
				TextFix:        "AI1",
				SoftwareType:   domainFacility.BacnetSoftwareTypeAI,
				SoftwareNumber: 1,
			},
		},
		failBulkCreate: cloneErr,
	}

	baseRepos := facility.Repositories{
		FieldDevices:             baseFieldDevices,
		SPSControllerSystemTypes: sharedSPSSystemTypes,
		SystemTypes:              sharedSystemTypes,
		Apparats:                 sharedApparats,
		SystemParts:              sharedSystemParts,
	}
	txRepos := facility.Repositories{
		FieldDevices:             txFieldDevices,
		SPSControllerSystemTypes: sharedSPSSystemTypes,
		SystemTypes:              sharedSystemTypes,
		Apparats:                 sharedApparats,
		SystemParts:              sharedSystemParts,
		ObjectData:               txObjectDatas,
		BacnetObjects:            txBacnetObjects,
	}

	runnerCalls := 0
	services := newTxServices(baseRepos, txRepos, &runnerCalls)

	err := services.FieldDevice.CreateWithBacnetObjects(
		context.Background(),
		newFieldDevice(fieldDeviceID, spsSystemTypeID, apparatID, systemPartID, 11),
		&objectDataID,
		nil,
	)
	if !errors.Is(err, cloneErr) {
		t.Fatalf("expected clone error, got %v", err)
	}
	if runnerCalls != 1 {
		t.Fatalf("expected one transaction run, got %d", runnerCalls)
	}
	if len(baseFieldDevices.items) != 0 {
		t.Fatalf("expected base repository to remain unchanged, got %+v", baseFieldDevices.items)
	}
	if txFieldDevices.deleteCalls != 0 {
		t.Fatalf("expected no compensating delete inside transaction, got %d", txFieldDevices.deleteCalls)
	}
}

func TestFacilityTransaction_FieldDeviceCreateWithBacnetObjectsFailureDoesNotEscapeTransaction(t *testing.T) {
	fieldDeviceID := uuid.New()
	spsSystemTypeID := uuid.New()
	systemTypeID := uuid.New()
	apparatID := uuid.New()
	systemPartID := uuid.New()
	alarmTypeID := uuid.New()
	alarmTypeFieldID := uuid.New()
	alarmValueErr := errors.New("alarm value write failed")

	baseFieldDevices := &fakeFieldDeviceStore{items: map[uuid.UUID]*domainFacility.FieldDevice{}}
	txFieldDevices := &fakeFieldDeviceStore{items: map[uuid.UUID]*domainFacility.FieldDevice{}}
	sharedSystemTypes := &fakeSystemTypeRepo{items: map[uuid.UUID]*domainFacility.SystemType{
		systemTypeID: {Base: domain.Base{ID: systemTypeID}, NumberMin: 1, NumberMax: 99},
	}}
	sharedSystemParts := &fakeSystemPartRepo{items: map[uuid.UUID]*domainFacility.SystemPart{
		systemPartID: {Base: domain.Base{ID: systemPartID}, ShortName: "AIR", Name: "Air"},
	}}
	sharedApparats := &fakeApparatRepo{items: map[uuid.UUID]*domainFacility.Apparat{
		apparatID: {Base: domain.Base{ID: apparatID}, ShortName: "PMP", Name: "Pump"},
	}}
	sharedSPSSystemTypes := &fakeSpsControllerSystemTypeRepo{items: map[uuid.UUID]*domainFacility.SPSControllerSystemType{
		spsSystemTypeID: {Base: domain.Base{ID: spsSystemTypeID}, SystemTypeID: systemTypeID},
	}}
	txBacnetObjects := &fakeBacnetObjectStore{items: map[uuid.UUID]*domainFacility.BacnetObject{}}
	txAlarmTypes := &fakeAlarmTypeRepo{items: map[uuid.UUID]*domainFacility.AlarmType{
		alarmTypeID: {
			Base: domain.Base{ID: alarmTypeID},
			Fields: []domainFacility.AlarmTypeField{{
				Base:         domain.Base{ID: alarmTypeFieldID},
				AlarmTypeID:  alarmTypeID,
				AlarmFieldID: uuid.New(),
				AlarmField:   &domainFacility.AlarmField{DataType: "number"},
			}},
		},
	}}
	txAlarmValues := &fakeBacnetObjectAlarmValueRepo{failBulkCreate: alarmValueErr}

	baseRepos := facility.Repositories{
		FieldDevices:             baseFieldDevices,
		SPSControllerSystemTypes: sharedSPSSystemTypes,
		SystemTypes:              sharedSystemTypes,
		Apparats:                 sharedApparats,
		SystemParts:              sharedSystemParts,
	}
	txRepos := facility.Repositories{
		FieldDevices:             txFieldDevices,
		SPSControllerSystemTypes: sharedSPSSystemTypes,
		SystemTypes:              sharedSystemTypes,
		Apparats:                 sharedApparats,
		SystemParts:              sharedSystemParts,
		BacnetObjects:            txBacnetObjects,
		AlarmTypes:               txAlarmTypes,
		BacnetObjectAlarmValues:  txAlarmValues,
	}

	runnerCalls := 0
	services := newTxServices(baseRepos, txRepos, &runnerCalls)
	bacnetObjects := []domainFacility.BacnetObject{{
		TextFix:        "AI1",
		SoftwareType:   domainFacility.BacnetSoftwareTypeAI,
		SoftwareNumber: 1,
		AlarmTypeID:    &alarmTypeID,
	}}

	err := services.FieldDevice.CreateWithBacnetObjects(
		context.Background(),
		newFieldDevice(fieldDeviceID, spsSystemTypeID, apparatID, systemPartID, 11),
		nil,
		bacnetObjects,
	)
	if !errors.Is(err, alarmValueErr) {
		t.Fatalf("expected alarm value error, got %v", err)
	}
	if runnerCalls != 1 {
		t.Fatalf("expected one transaction run, got %d", runnerCalls)
	}
	if len(baseFieldDevices.items) != 0 {
		t.Fatalf("expected base field devices to remain unchanged, got %+v", baseFieldDevices.items)
	}
	if len(txBacnetObjects.items) == 0 || txAlarmValues.bulkCreateCalls != 1 {
		t.Fatalf("expected mid-flow bacnet writes before failure, got objects=%d alarmCalls=%d", len(txBacnetObjects.items), txAlarmValues.bulkCreateCalls)
	}
	if txFieldDevices.deleteCalls != 0 {
		t.Fatalf("expected no compensating field-device delete inside transaction, got %d", txFieldDevices.deleteCalls)
	}
}

func TestFacilityTransaction_FieldDeviceUpdateWithBacnetObjectsFailureDoesNotEscapeTransaction(t *testing.T) {
	fieldDeviceID := uuid.New()
	oldBacnetObjectID := uuid.New()
	spsSystemTypeID := uuid.New()
	systemTypeID := uuid.New()
	apparatID := uuid.New()
	systemPartID := uuid.New()
	alarmTypeID := uuid.New()
	alarmTypeFieldID := uuid.New()
	alarmValueErr := errors.New("alarm value write failed")

	baseFieldDevice := newFieldDevice(fieldDeviceID, spsSystemTypeID, apparatID, systemPartID, 11)
	txFieldDevice := newFieldDevice(fieldDeviceID, spsSystemTypeID, apparatID, systemPartID, 11)
	updatedFieldDevice := newFieldDevice(fieldDeviceID, spsSystemTypeID, apparatID, systemPartID, 22)
	oldFieldDeviceID := fieldDeviceID
	baseFieldDevices := &fakeFieldDeviceStore{items: map[uuid.UUID]*domainFacility.FieldDevice{fieldDeviceID: baseFieldDevice}}
	txFieldDevices := &fakeFieldDeviceStore{items: map[uuid.UUID]*domainFacility.FieldDevice{fieldDeviceID: txFieldDevice}}
	sharedSystemTypes := &fakeSystemTypeRepo{items: map[uuid.UUID]*domainFacility.SystemType{
		systemTypeID: {Base: domain.Base{ID: systemTypeID}, NumberMin: 1, NumberMax: 99},
	}}
	sharedSystemParts := &fakeSystemPartRepo{items: map[uuid.UUID]*domainFacility.SystemPart{
		systemPartID: {Base: domain.Base{ID: systemPartID}, ShortName: "AIR", Name: "Air"},
	}}
	sharedApparats := &fakeApparatRepo{items: map[uuid.UUID]*domainFacility.Apparat{
		apparatID: {Base: domain.Base{ID: apparatID}, ShortName: "PMP", Name: "Pump"},
	}}
	sharedSPSSystemTypes := &fakeSpsControllerSystemTypeRepo{items: map[uuid.UUID]*domainFacility.SPSControllerSystemType{
		spsSystemTypeID: {Base: domain.Base{ID: spsSystemTypeID}, SystemTypeID: systemTypeID},
	}}
	baseBacnetObjects := &fakeBacnetObjectStore{items: map[uuid.UUID]*domainFacility.BacnetObject{
		oldBacnetObjectID: {Base: domain.Base{ID: oldBacnetObjectID}, FieldDeviceID: &oldFieldDeviceID, TextFix: "OLD", SoftwareType: domainFacility.BacnetSoftwareTypeAI, SoftwareNumber: 1},
	}}
	txBacnetObjects := &fakeBacnetObjectStore{items: map[uuid.UUID]*domainFacility.BacnetObject{
		oldBacnetObjectID: {Base: domain.Base{ID: oldBacnetObjectID}, FieldDeviceID: &oldFieldDeviceID, TextFix: "OLD", SoftwareType: domainFacility.BacnetSoftwareTypeAI, SoftwareNumber: 1},
	}}
	txAlarmTypes := &fakeAlarmTypeRepo{items: map[uuid.UUID]*domainFacility.AlarmType{
		alarmTypeID: {
			Base: domain.Base{ID: alarmTypeID},
			Fields: []domainFacility.AlarmTypeField{{
				Base:         domain.Base{ID: alarmTypeFieldID},
				AlarmTypeID:  alarmTypeID,
				AlarmFieldID: uuid.New(),
				AlarmField:   &domainFacility.AlarmField{DataType: "number"},
			}},
		},
	}}
	txAlarmValues := &fakeBacnetObjectAlarmValueRepo{failBulkCreate: alarmValueErr}

	baseRepos := facility.Repositories{
		FieldDevices:             baseFieldDevices,
		SPSControllerSystemTypes: sharedSPSSystemTypes,
		SystemTypes:              sharedSystemTypes,
		Apparats:                 sharedApparats,
		SystemParts:              sharedSystemParts,
		BacnetObjects:            baseBacnetObjects,
	}
	txRepos := facility.Repositories{
		FieldDevices:             txFieldDevices,
		SPSControllerSystemTypes: sharedSPSSystemTypes,
		SystemTypes:              sharedSystemTypes,
		Apparats:                 sharedApparats,
		SystemParts:              sharedSystemParts,
		BacnetObjects:            txBacnetObjects,
		AlarmTypes:               txAlarmTypes,
		BacnetObjectAlarmValues:  txAlarmValues,
	}

	runnerCalls := 0
	services := newTxServices(baseRepos, txRepos, &runnerCalls)
	bacnetObjects := []domainFacility.BacnetObject{{
		TextFix:        "NEW",
		SoftwareType:   domainFacility.BacnetSoftwareTypeAO,
		SoftwareNumber: 2,
		AlarmTypeID:    &alarmTypeID,
	}}

	err := services.FieldDevice.UpdateWithBacnetObjects(context.Background(), updatedFieldDevice, nil, &bacnetObjects)
	if !errors.Is(err, alarmValueErr) {
		t.Fatalf("expected alarm value error, got %v", err)
	}
	if runnerCalls != 1 {
		t.Fatalf("expected one transaction run, got %d", runnerCalls)
	}
	if got := baseFieldDevices.items[fieldDeviceID].ApparatNr; got != 11 {
		t.Fatalf("expected base field device to keep apparat_nr 11, got %d", got)
	}
	if got := baseBacnetObjects.items[oldBacnetObjectID].TextFix; got != "OLD" {
		t.Fatalf("expected base bacnet object to remain unchanged, got %q", got)
	}
	if txBacnetObjects.deleteByFieldDeviceCalls != 1 || txAlarmValues.bulkCreateCalls != 1 {
		t.Fatalf("expected delete-and-recreate path before failure, got deleteCalls=%d alarmCalls=%d", txBacnetObjects.deleteByFieldDeviceCalls, txAlarmValues.bulkCreateCalls)
	}
}

func TestFacilityTransaction_FieldDeviceReplaceFromObjectDataFailureDoesNotEscapeTransaction(t *testing.T) {
	templateID := uuid.New()
	fieldDeviceID := uuid.New()
	oldBacnetObjectID := uuid.New()
	objectDataID := uuid.New()
	spsSystemTypeID := uuid.New()
	systemTypeID := uuid.New()
	apparatID := uuid.New()
	systemPartID := uuid.New()
	alarmTypeID := uuid.New()
	alarmTypeFieldID := uuid.New()
	alarmValueErr := errors.New("alarm value write failed")

	baseFieldDevice := newFieldDevice(fieldDeviceID, spsSystemTypeID, apparatID, systemPartID, 11)
	txFieldDevice := newFieldDevice(fieldDeviceID, spsSystemTypeID, apparatID, systemPartID, 11)
	updatedFieldDevice := newFieldDevice(fieldDeviceID, spsSystemTypeID, apparatID, systemPartID, 33)
	oldFieldDeviceID := fieldDeviceID
	baseFieldDevices := &fakeFieldDeviceStore{items: map[uuid.UUID]*domainFacility.FieldDevice{fieldDeviceID: baseFieldDevice}}
	txFieldDevices := &fakeFieldDeviceStore{items: map[uuid.UUID]*domainFacility.FieldDevice{fieldDeviceID: txFieldDevice}}
	sharedSystemTypes := &fakeSystemTypeRepo{items: map[uuid.UUID]*domainFacility.SystemType{
		systemTypeID: {Base: domain.Base{ID: systemTypeID}, NumberMin: 1, NumberMax: 99},
	}}
	sharedSystemParts := &fakeSystemPartRepo{items: map[uuid.UUID]*domainFacility.SystemPart{
		systemPartID: {Base: domain.Base{ID: systemPartID}, ShortName: "AIR", Name: "Air"},
	}}
	sharedApparats := &fakeApparatRepo{items: map[uuid.UUID]*domainFacility.Apparat{
		apparatID: {Base: domain.Base{ID: apparatID}, ShortName: "PMP", Name: "Pump"},
	}}
	sharedSPSSystemTypes := &fakeSpsControllerSystemTypeRepo{items: map[uuid.UUID]*domainFacility.SPSControllerSystemType{
		spsSystemTypeID: {Base: domain.Base{ID: spsSystemTypeID}, SystemTypeID: systemTypeID},
	}}
	baseBacnetObjects := &fakeBacnetObjectStore{items: map[uuid.UUID]*domainFacility.BacnetObject{
		oldBacnetObjectID: {Base: domain.Base{ID: oldBacnetObjectID}, FieldDeviceID: &oldFieldDeviceID, TextFix: "OLD", SoftwareType: domainFacility.BacnetSoftwareTypeAI, SoftwareNumber: 1},
	}}
	txBacnetObjects := &fakeBacnetObjectStore{items: map[uuid.UUID]*domainFacility.BacnetObject{
		oldBacnetObjectID: {Base: domain.Base{ID: oldBacnetObjectID}, FieldDeviceID: &oldFieldDeviceID, TextFix: "OLD", SoftwareType: domainFacility.BacnetSoftwareTypeAI, SoftwareNumber: 1},
		templateID: {
			Base:           domain.Base{ID: templateID},
			TextFix:        "TPL",
			SoftwareType:   domainFacility.BacnetSoftwareTypeAV,
			SoftwareNumber: 7,
			AlarmTypeID:    &alarmTypeID,
		},
	}}
	txObjectDatas := &txObjectDataStore{
		fakeObjectDataStore: &fakeObjectDataStore{items: map[uuid.UUID]*domainFacility.ObjectData{
			objectDataID: {Base: domain.Base{ID: objectDataID}, IsActive: true},
		}},
		bacnetObjectIDs: map[uuid.UUID][]uuid.UUID{objectDataID: {templateID}},
	}
	txAlarmTypes := &fakeAlarmTypeRepo{items: map[uuid.UUID]*domainFacility.AlarmType{
		alarmTypeID: {
			Base: domain.Base{ID: alarmTypeID},
			Fields: []domainFacility.AlarmTypeField{{
				Base:         domain.Base{ID: alarmTypeFieldID},
				AlarmTypeID:  alarmTypeID,
				AlarmFieldID: uuid.New(),
				AlarmField:   &domainFacility.AlarmField{DataType: "number"},
			}},
		},
	}}
	txAlarmValues := &fakeBacnetObjectAlarmValueRepo{failBulkCreate: alarmValueErr}

	baseRepos := facility.Repositories{
		FieldDevices:             baseFieldDevices,
		SPSControllerSystemTypes: sharedSPSSystemTypes,
		SystemTypes:              sharedSystemTypes,
		Apparats:                 sharedApparats,
		SystemParts:              sharedSystemParts,
		BacnetObjects:            baseBacnetObjects,
	}
	txRepos := facility.Repositories{
		FieldDevices:             txFieldDevices,
		SPSControllerSystemTypes: sharedSPSSystemTypes,
		SystemTypes:              sharedSystemTypes,
		Apparats:                 sharedApparats,
		SystemParts:              sharedSystemParts,
		BacnetObjects:            txBacnetObjects,
		ObjectData:               txObjectDatas,
		AlarmTypes:               txAlarmTypes,
		BacnetObjectAlarmValues:  txAlarmValues,
	}

	runnerCalls := 0
	services := newTxServices(baseRepos, txRepos, &runnerCalls)

	err := services.FieldDevice.UpdateWithBacnetObjects(context.Background(), updatedFieldDevice, &objectDataID, nil)
	if !errors.Is(err, alarmValueErr) {
		t.Fatalf("expected alarm value error, got %v", err)
	}
	if runnerCalls != 1 {
		t.Fatalf("expected one transaction run, got %d", runnerCalls)
	}
	if got := baseFieldDevices.items[fieldDeviceID].ApparatNr; got != 11 {
		t.Fatalf("expected base field device to keep apparat_nr 11, got %d", got)
	}
	if got := baseBacnetObjects.items[oldBacnetObjectID].TextFix; got != "OLD" {
		t.Fatalf("expected base bacnet object to remain unchanged, got %q", got)
	}
	if txBacnetObjects.deleteByFieldDeviceCalls != 1 || txAlarmValues.bulkCreateCalls != 1 {
		t.Fatalf("expected object-data replacement path before failure, got deleteCalls=%d alarmCalls=%d", txBacnetObjects.deleteByFieldDeviceCalls, txAlarmValues.bulkCreateCalls)
	}
}

func TestFacilityTransaction_FieldDeviceMultiCreateFailureDoesNotUseCompensatingDelete(t *testing.T) {
	fieldDeviceID := uuid.New()
	spsSystemTypeID := uuid.New()
	systemTypeID := uuid.New()
	apparatID := uuid.New()
	systemPartID := uuid.New()
	alarmTypeID := uuid.New()
	alarmTypeFieldID := uuid.New()
	alarmValueErr := errors.New("alarm value write failed")

	baseFieldDevices := &fakeFieldDeviceStore{items: map[uuid.UUID]*domainFacility.FieldDevice{}}
	txFieldDevices := &fakeFieldDeviceStore{items: map[uuid.UUID]*domainFacility.FieldDevice{}}
	sharedSystemTypes := &fakeSystemTypeRepo{items: map[uuid.UUID]*domainFacility.SystemType{
		systemTypeID: {Base: domain.Base{ID: systemTypeID}, NumberMin: 1, NumberMax: 99},
	}}
	sharedSystemParts := &fakeSystemPartRepo{items: map[uuid.UUID]*domainFacility.SystemPart{
		systemPartID: {Base: domain.Base{ID: systemPartID}, ShortName: "AIR", Name: "Air"},
	}}
	sharedApparats := &fakeApparatRepo{items: map[uuid.UUID]*domainFacility.Apparat{
		apparatID: {Base: domain.Base{ID: apparatID}, ShortName: "PMP", Name: "Pump"},
	}}
	sharedSPSSystemTypes := &fakeSpsControllerSystemTypeRepo{items: map[uuid.UUID]*domainFacility.SPSControllerSystemType{
		spsSystemTypeID: {Base: domain.Base{ID: spsSystemTypeID}, SystemTypeID: systemTypeID},
	}}
	txBacnetObjects := &fakeBacnetObjectStore{items: map[uuid.UUID]*domainFacility.BacnetObject{}}
	txAlarmTypes := &fakeAlarmTypeRepo{items: map[uuid.UUID]*domainFacility.AlarmType{
		alarmTypeID: {
			Base: domain.Base{ID: alarmTypeID},
			Fields: []domainFacility.AlarmTypeField{{
				Base:         domain.Base{ID: alarmTypeFieldID},
				AlarmTypeID:  alarmTypeID,
				AlarmFieldID: uuid.New(),
				AlarmField:   &domainFacility.AlarmField{DataType: "number"},
			}},
		},
	}}
	txAlarmValues := &fakeBacnetObjectAlarmValueRepo{failBulkCreate: alarmValueErr}

	baseRepos := facility.Repositories{
		FieldDevices:             baseFieldDevices,
		SPSControllerSystemTypes: sharedSPSSystemTypes,
		SystemTypes:              sharedSystemTypes,
		Apparats:                 sharedApparats,
		SystemParts:              sharedSystemParts,
	}
	txRepos := facility.Repositories{
		FieldDevices:             txFieldDevices,
		SPSControllerSystemTypes: sharedSPSSystemTypes,
		SystemTypes:              sharedSystemTypes,
		Apparats:                 sharedApparats,
		SystemParts:              sharedSystemParts,
		BacnetObjects:            txBacnetObjects,
		AlarmTypes:               txAlarmTypes,
		BacnetObjectAlarmValues:  txAlarmValues,
	}

	runnerCalls := 0
	services := newTxServices(baseRepos, txRepos, &runnerCalls)
	result := services.FieldDevice.MultiCreate(context.Background(), []domainFacility.FieldDeviceCreateItem{{
		FieldDevice: newFieldDevice(fieldDeviceID, spsSystemTypeID, apparatID, systemPartID, 11),
		BacnetObjects: []domainFacility.BacnetObject{{
			TextFix:        "AI1",
			SoftwareType:   domainFacility.BacnetSoftwareTypeAI,
			SoftwareNumber: 1,
			AlarmTypeID:    &alarmTypeID,
		}},
	}})

	if result.SuccessCount != 0 || result.FailureCount != 1 {
		t.Fatalf("expected one failure and no successes, got %+v", result)
	}
	if result.Results[0].Error != alarmValueErr.Error() {
		t.Fatalf("expected alarm value error to surface, got %+v", result.Results[0])
	}
	if runnerCalls != 1 {
		t.Fatalf("expected one transaction run, got %d", runnerCalls)
	}
	if txFieldDevices.deleteCalls != 0 {
		t.Fatalf("expected no compensating delete inside multi-create, got %d", txFieldDevices.deleteCalls)
	}
	if len(baseFieldDevices.items) != 0 {
		t.Fatalf("expected base repository to remain unchanged, got %+v", baseFieldDevices.items)
	}
}

func TestFacilityTransaction_BacnetObjectCreateFailureDoesNotEscapeTransaction(t *testing.T) {
	objectDataID := uuid.New()
	linkErr := errors.New("attach failed")

	baseBacnetObjects := &fakeBacnetObjectStore{items: map[uuid.UUID]*domainFacility.BacnetObject{}}
	txBacnetObjects := &fakeBacnetObjectStore{items: map[uuid.UUID]*domainFacility.BacnetObject{}}
	baseObjectDatas := &fakeObjectDataStore{items: map[uuid.UUID]*domainFacility.ObjectData{
		objectDataID: {Base: domain.Base{ID: objectDataID}, IsActive: true},
	}}
	txObjectDatas := &fakeObjectDataStore{items: map[uuid.UUID]*domainFacility.ObjectData{
		objectDataID: {Base: domain.Base{ID: objectDataID}, IsActive: true},
	}}
	txLinks := &fakeObjectDataBacnetObjectStore{addErr: linkErr}

	baseRepos := facility.Repositories{
		BacnetObjects: baseBacnetObjects,
		ObjectData:    baseObjectDatas,
	}
	txRepos := facility.Repositories{
		BacnetObjects:           txBacnetObjects,
		ObjectData:              txObjectDatas,
		ObjectDataBacnetObjects: txLinks,
	}

	runnerCalls := 0
	services := newTxServices(baseRepos, txRepos, &runnerCalls)
	bacnetObject := &domainFacility.BacnetObject{
		TextFix:        "AI1",
		SoftwareType:   domainFacility.BacnetSoftwareTypeAI,
		SoftwareNumber: 7,
	}

	err := services.BacnetObject.CreateWithParent(context.Background(), bacnetObject, nil, &objectDataID)
	if !errors.Is(err, linkErr) {
		t.Fatalf("expected link error, got %v", err)
	}
	if runnerCalls != 1 {
		t.Fatalf("expected one transaction run, got %d", runnerCalls)
	}
	if len(baseBacnetObjects.items) != 0 {
		t.Fatalf("expected base bacnet objects to remain unchanged, got %+v", baseBacnetObjects.items)
	}
	if txLinks.addCalls != 1 {
		t.Fatalf("expected one link attempt, got %d", txLinks.addCalls)
	}
}

func TestFacilityTransaction_BacnetObjectReplaceForObjectDataFailureDoesNotEscapeTransaction(t *testing.T) {
	objectDataID := uuid.New()
	oldBacnetObjectID := uuid.New()
	linkErr := errors.New("attach failed")

	baseBacnetObjects := &fakeBacnetObjectStore{items: map[uuid.UUID]*domainFacility.BacnetObject{
		oldBacnetObjectID: {Base: domain.Base{ID: oldBacnetObjectID}, TextFix: "OLD", SoftwareType: domainFacility.BacnetSoftwareTypeAI, SoftwareNumber: 1},
	}}
	txBacnetObjects := &fakeBacnetObjectStore{items: map[uuid.UUID]*domainFacility.BacnetObject{}}
	baseObjectDatas := &fakeObjectDataStore{items: map[uuid.UUID]*domainFacility.ObjectData{
		objectDataID: {Base: domain.Base{ID: objectDataID}, IsActive: true},
	}}
	txObjectDatas := &fakeObjectDataStore{items: map[uuid.UUID]*domainFacility.ObjectData{
		objectDataID: {Base: domain.Base{ID: objectDataID}, IsActive: true},
	}}
	baseLinks := &fakeObjectDataBacnetObjectStore{links: map[uuid.UUID][]uuid.UUID{objectDataID: {oldBacnetObjectID}}}
	txLinks := &fakeObjectDataBacnetObjectStore{addErr: linkErr, links: map[uuid.UUID][]uuid.UUID{objectDataID: {oldBacnetObjectID}}}

	baseRepos := facility.Repositories{
		BacnetObjects:           baseBacnetObjects,
		ObjectData:              baseObjectDatas,
		ObjectDataBacnetObjects: baseLinks,
	}
	txRepos := facility.Repositories{
		BacnetObjects:           txBacnetObjects,
		ObjectData:              txObjectDatas,
		ObjectDataBacnetObjects: txLinks,
	}

	runnerCalls := 0
	services := newTxServices(baseRepos, txRepos, &runnerCalls)
	inputs := []domainFacility.BacnetObject{{
		TextFix:        "NEW",
		SoftwareType:   domainFacility.BacnetSoftwareTypeAO,
		SoftwareNumber: 2,
	}}

	err := services.BacnetObject.ReplaceForObjectData(context.Background(), objectDataID, inputs)
	if !errors.Is(err, linkErr) {
		t.Fatalf("expected link error, got %v", err)
	}
	if runnerCalls != 1 {
		t.Fatalf("expected one transaction run, got %d", runnerCalls)
	}
	if len(baseLinks.links[objectDataID]) != 1 || baseLinks.links[objectDataID][0] != oldBacnetObjectID {
		t.Fatalf("expected base object-data links to remain unchanged, got %+v", baseLinks.links[objectDataID])
	}
	if txBacnetObjects.createdCount != 1 || txLinks.addCalls != 1 || len(txLinks.deleteByObjectDataIDs) != 1 {
		t.Fatalf("expected replace-for-object-data path before failure, got created=%d addCalls=%d deleteCalls=%d", txBacnetObjects.createdCount, txLinks.addCalls, len(txLinks.deleteByObjectDataIDs))
	}
}

func TestFacilityTransaction_HierarchyCopySPSControllerFailureDoesNotUseRollbackDeletes(t *testing.T) {
	originalControllerID := uuid.New()
	controlCabinetID := uuid.New()
	buildingID := uuid.New()
	systemTypeID := uuid.New()
	originalSystemTypeID := uuid.New()
	copyErr := errors.New("copy children failed")
	gaDevice := "BBB"
	number := 1

	baseBuildings := &fakeHierarchyBuildingRepo{items: map[uuid.UUID]*domainFacility.Building{
		buildingID: {Base: domain.Base{ID: buildingID}, IWSCode: "BLD"},
	}}
	txBuildings := &fakeHierarchyBuildingRepo{items: map[uuid.UUID]*domainFacility.Building{
		buildingID: {Base: domain.Base{ID: buildingID}, IWSCode: "BLD"},
	}}
	baseCabinets := &fakeHierarchyControlCabinetRepo{items: map[uuid.UUID]*domainFacility.ControlCabinet{
		controlCabinetID: {Base: domain.Base{ID: controlCabinetID}, BuildingID: buildingID},
	}}
	txCabinets := &fakeHierarchyControlCabinetRepo{items: map[uuid.UUID]*domainFacility.ControlCabinet{
		controlCabinetID: {Base: domain.Base{ID: controlCabinetID}, BuildingID: buildingID},
	}}
	baseControllers := &fakeHierarchySPSControllerRepo{items: map[uuid.UUID]*domainFacility.SPSController{
		originalControllerID: {Base: domain.Base{ID: originalControllerID}, ControlCabinetID: controlCabinetID, GADevice: &gaDevice, DeviceName: "OLD"},
	}}
	txControllers := &fakeHierarchySPSControllerRepo{items: map[uuid.UUID]*domainFacility.SPSController{
		originalControllerID: {Base: domain.Base{ID: originalControllerID}, ControlCabinetID: controlCabinetID, GADevice: &gaDevice, DeviceName: "OLD"},
	}}
	baseSystemTypes := &fakeSystemTypeRepo{items: map[uuid.UUID]*domainFacility.SystemType{
		systemTypeID: {Base: domain.Base{ID: systemTypeID}, NumberMin: 1, NumberMax: 10},
	}}
	txSystemTypes := &fakeSystemTypeRepo{items: map[uuid.UUID]*domainFacility.SystemType{
		systemTypeID: {Base: domain.Base{ID: systemTypeID}, NumberMin: 1, NumberMax: 10},
	}}
	baseSPSSystemTypes := &fakeSpsControllerSystemTypeRepo{items: map[uuid.UUID]*domainFacility.SPSControllerSystemType{
		originalSystemTypeID: {Base: domain.Base{ID: originalSystemTypeID}, SPSControllerID: originalControllerID, SystemTypeID: systemTypeID, Number: &number},
	}}
	txSPSSystemTypes := &txSPSControllerSystemTypeRepo{fakeSpsControllerSystemTypeRepo: &fakeSpsControllerSystemTypeRepo{items: map[uuid.UUID]*domainFacility.SPSControllerSystemType{
		originalSystemTypeID: {Base: domain.Base{ID: originalSystemTypeID}, SPSControllerID: originalControllerID, SystemTypeID: systemTypeID, Number: &number},
	}}}
	txFieldDevices := &txFieldDeviceStore{
		fakeFieldDeviceStore: &fakeFieldDeviceStore{items: map[uuid.UUID]*domainFacility.FieldDevice{}},
		getIDsErr:            copyErr,
	}

	baseRepos := facility.Repositories{
		Buildings:                baseBuildings,
		ControlCabinets:          baseCabinets,
		SPSControllers:           baseControllers,
		SystemTypes:              baseSystemTypes,
		SPSControllerSystemTypes: baseSPSSystemTypes,
	}
	txRepos := facility.Repositories{
		Buildings:                txBuildings,
		ControlCabinets:          txCabinets,
		SPSControllers:           txControllers,
		SystemTypes:              txSystemTypes,
		SPSControllerSystemTypes: txSPSSystemTypes,
		FieldDevices:             txFieldDevices,
	}

	runnerCalls := 0
	services := newTxServices(baseRepos, txRepos, &runnerCalls)

	_, err := services.SPSController.CopyByID(context.Background(), originalControllerID)
	if !errors.Is(err, copyErr) {
		t.Fatalf("expected copy error, got %v", err)
	}
	if runnerCalls != 1 {
		t.Fatalf("expected one transaction run, got %d", runnerCalls)
	}
	if txControllers.createCalls != 1 || txSPSSystemTypes.deleteBySPSControllerCalls != 0 {
		t.Fatalf("expected mid-flow writes without rollback cleanup, got controllerCreates=%d deleteByControllerCalls=%d", txControllers.createCalls, txSPSSystemTypes.deleteBySPSControllerCalls)
	}
	if txControllers.deleteCalls != 0 || txSPSSystemTypes.deleteCalls != 0 {
		t.Fatalf("expected no rollback delete calls, got controllerDeletes=%d systemTypeDeletes=%d", txControllers.deleteCalls, txSPSSystemTypes.deleteCalls)
	}
	if len(baseControllers.items) != 1 {
		t.Fatalf("expected base repository to remain unchanged, got %+v", baseControllers.items)
	}
}

func TestFacilityTransaction_HierarchyCopyControlCabinetFailureDoesNotUseRollbackDeletes(t *testing.T) {
	originalCabinetID := uuid.New()
	originalControllerID := uuid.New()
	buildingID := uuid.New()
	systemTypeID := uuid.New()
	originalSystemTypeID := uuid.New()
	copyErr := errors.New("copy children failed")
	gaDevice := "BBB"
	number := 1

	baseBuildings := &fakeHierarchyBuildingRepo{items: map[uuid.UUID]*domainFacility.Building{
		buildingID: {Base: domain.Base{ID: buildingID}, IWSCode: "BLD"},
	}}
	txBuildings := &fakeHierarchyBuildingRepo{items: map[uuid.UUID]*domainFacility.Building{
		buildingID: {Base: domain.Base{ID: buildingID}, IWSCode: "BLD"},
	}}
	baseCabinets := &fakeHierarchyControlCabinetRepo{items: map[uuid.UUID]*domainFacility.ControlCabinet{
		originalCabinetID: {Base: domain.Base{ID: originalCabinetID}, BuildingID: buildingID},
	}}
	txCabinets := &fakeHierarchyControlCabinetRepo{items: map[uuid.UUID]*domainFacility.ControlCabinet{
		originalCabinetID: {Base: domain.Base{ID: originalCabinetID}, BuildingID: buildingID},
	}}
	baseControllers := &fakeHierarchySPSControllerRepo{items: map[uuid.UUID]*domainFacility.SPSController{
		originalControllerID: {Base: domain.Base{ID: originalControllerID}, ControlCabinetID: originalCabinetID, GADevice: &gaDevice, DeviceName: "OLD"},
	}}
	txControllers := &fakeHierarchySPSControllerRepo{items: map[uuid.UUID]*domainFacility.SPSController{
		originalControllerID: {Base: domain.Base{ID: originalControllerID}, ControlCabinetID: originalCabinetID, GADevice: &gaDevice, DeviceName: "OLD"},
	}}
	baseSystemTypes := &fakeSystemTypeRepo{items: map[uuid.UUID]*domainFacility.SystemType{
		systemTypeID: {Base: domain.Base{ID: systemTypeID}, NumberMin: 1, NumberMax: 10},
	}}
	txSystemTypes := &fakeSystemTypeRepo{items: map[uuid.UUID]*domainFacility.SystemType{
		systemTypeID: {Base: domain.Base{ID: systemTypeID}, NumberMin: 1, NumberMax: 10},
	}}
	baseSPSSystemTypes := &fakeSpsControllerSystemTypeRepo{items: map[uuid.UUID]*domainFacility.SPSControllerSystemType{
		originalSystemTypeID: {Base: domain.Base{ID: originalSystemTypeID}, SPSControllerID: originalControllerID, SystemTypeID: systemTypeID, Number: &number},
	}}
	txSPSSystemTypes := &txSPSControllerSystemTypeRepo{fakeSpsControllerSystemTypeRepo: &fakeSpsControllerSystemTypeRepo{items: map[uuid.UUID]*domainFacility.SPSControllerSystemType{
		originalSystemTypeID: {Base: domain.Base{ID: originalSystemTypeID}, SPSControllerID: originalControllerID, SystemTypeID: systemTypeID, Number: &number},
	}}}
	txFieldDevices := &txFieldDeviceStore{
		fakeFieldDeviceStore: &fakeFieldDeviceStore{items: map[uuid.UUID]*domainFacility.FieldDevice{}},
		getIDsErr:            copyErr,
	}

	baseRepos := facility.Repositories{
		Buildings:                baseBuildings,
		ControlCabinets:          baseCabinets,
		SPSControllers:           baseControllers,
		SystemTypes:              baseSystemTypes,
		SPSControllerSystemTypes: baseSPSSystemTypes,
	}
	txRepos := facility.Repositories{
		Buildings:                txBuildings,
		ControlCabinets:          txCabinets,
		SPSControllers:           txControllers,
		SystemTypes:              txSystemTypes,
		SPSControllerSystemTypes: txSPSSystemTypes,
		FieldDevices:             txFieldDevices,
	}

	runnerCalls := 0
	services := newTxServices(baseRepos, txRepos, &runnerCalls)

	_, err := services.ControlCabinet.CopyByID(context.Background(), originalCabinetID)
	if !errors.Is(err, copyErr) {
		t.Fatalf("expected copy error, got %v", err)
	}
	if runnerCalls != 1 {
		t.Fatalf("expected one transaction run, got %d", runnerCalls)
	}
	if txCabinets.createCalls != 1 || txControllers.createCalls != 1 {
		t.Fatalf("expected cabinet and controller copies before failure, got cabinetCreates=%d controllerCreates=%d", txCabinets.createCalls, txControllers.createCalls)
	}
	if txCabinets.deleteCalls != 0 || txControllers.deleteCalls != 0 || txSPSSystemTypes.deleteBySPSControllerCalls != 0 {
		t.Fatalf("expected no rollback delete calls, got cabinetDeletes=%d controllerDeletes=%d deleteByControllerCalls=%d", txCabinets.deleteCalls, txControllers.deleteCalls, txSPSSystemTypes.deleteBySPSControllerCalls)
	}
	if len(baseCabinets.items) != 1 {
		t.Fatalf("expected base repository to remain unchanged, got %+v", baseCabinets.items)
	}
}

func TestFacilityTransaction_HierarchyCopyFailureDoesNotUseRollbackDeletes(t *testing.T) {
	originalID := uuid.New()
	spsControllerID := uuid.New()
	systemTypeID := uuid.New()
	copyErr := errors.New("copy children failed")
	one := 1

	baseSystemTypes := &fakeSystemTypeRepo{items: map[uuid.UUID]*domainFacility.SystemType{
		systemTypeID: {Base: domain.Base{ID: systemTypeID}, NumberMin: 1, NumberMax: 10},
	}}
	baseSPSSystemTypes := &fakeSpsControllerSystemTypeRepo{items: map[uuid.UUID]*domainFacility.SPSControllerSystemType{
		originalID: {
			Base:            domain.Base{ID: originalID},
			SPSControllerID: spsControllerID,
			SystemTypeID:    systemTypeID,
			Number:          &one,
		},
	}}
	txSPSSystemTypes := &txSPSControllerSystemTypeRepo{fakeSpsControllerSystemTypeRepo: &fakeSpsControllerSystemTypeRepo{items: map[uuid.UUID]*domainFacility.SPSControllerSystemType{
		originalID: {
			Base:            domain.Base{ID: originalID},
			SPSControllerID: spsControllerID,
			SystemTypeID:    systemTypeID,
			Number:          &one,
		},
	}}}
	txFieldDevices := &txFieldDeviceStore{
		fakeFieldDeviceStore: &fakeFieldDeviceStore{items: map[uuid.UUID]*domainFacility.FieldDevice{}},
		getIDsErr:            copyErr,
	}

	baseRepos := facility.Repositories{
		SystemTypes:              baseSystemTypes,
		SPSControllerSystemTypes: baseSPSSystemTypes,
	}
	txRepos := facility.Repositories{
		SystemTypes:              baseSystemTypes,
		SPSControllerSystemTypes: txSPSSystemTypes,
		FieldDevices:             txFieldDevices,
	}

	runnerCalls := 0
	services := newTxServices(baseRepos, txRepos, &runnerCalls)

	_, err := services.SPSControllerSystemType.CopyByID(context.Background(), originalID)
	if !errors.Is(err, copyErr) {
		t.Fatalf("expected copy error, got %v", err)
	}
	if runnerCalls != 1 {
		t.Fatalf("expected one transaction run, got %d", runnerCalls)
	}
	if txSPSSystemTypes.deleteCalls != 0 {
		t.Fatalf("expected no rollback delete calls, got %d", txSPSSystemTypes.deleteCalls)
	}
	if len(baseSPSSystemTypes.items) != 1 {
		t.Fatalf("expected base repository to remain unchanged, got %+v", baseSPSSystemTypes.items)
	}
}
