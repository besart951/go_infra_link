package facility_test

import (
	"context"
	"sort"
	"strings"
	"testing"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/besart951/go_infra_link/backend/internal/service/facility"
	"github.com/google/uuid"
)

type fakeFieldDeviceStore struct {
	items          map[uuid.UUID]*domainFacility.FieldDevice
	deleteCalls    int
	deletedBatches [][]uuid.UUID
}

func (r *fakeFieldDeviceStore) GetByIds(_ context.Context, ids []uuid.UUID) ([]*domainFacility.FieldDevice, error) {
	out := make([]*domainFacility.FieldDevice, 0, len(ids))
	for _, id := range ids {
		if item, ok := r.items[id]; ok {
			clone := *item
			out = append(out, &clone)
		}
	}
	return out, nil
}

func (r *fakeFieldDeviceStore) Create(_ context.Context, entity *domainFacility.FieldDevice) error {
	clone := *entity
	r.items[entity.ID] = &clone
	return nil
}

func (r *fakeFieldDeviceStore) BulkCreate(_ context.Context, entities []*domainFacility.FieldDevice, batchSize int) error {
	for _, entity := range entities {
		if err := r.Create(context.Background(), entity); err != nil {
			return err
		}
	}
	return nil
}

func (r *fakeFieldDeviceStore) Update(_ context.Context, entity *domainFacility.FieldDevice) error {
	clone := *entity
	r.items[entity.ID] = &clone
	return nil
}

func (r *fakeFieldDeviceStore) DeleteByIds(_ context.Context, ids []uuid.UUID) error {
	r.deleteCalls++
	batch := append([]uuid.UUID(nil), ids...)
	r.deletedBatches = append(r.deletedBatches, batch)
	for _, id := range ids {
		delete(r.items, id)
	}
	return nil
}

func (r *fakeFieldDeviceStore) GetPaginatedList(_ context.Context, params domain.PaginationParams) (*domain.PaginatedList[domainFacility.FieldDevice], error) {
	items := make([]domainFacility.FieldDevice, 0, len(r.items))
	for _, item := range r.items {
		items = append(items, *item)
	}
	return &domain.PaginatedList[domainFacility.FieldDevice]{
		Items:      items,
		Total:      int64(len(items)),
		Page:       1,
		TotalPages: 1,
	}, nil
}

func (r *fakeFieldDeviceStore) GetPaginatedListWithFilters(
	_ context.Context,
	params domain.PaginationParams,
	filters domainFacility.FieldDeviceFilterParams,
) (*domain.PaginatedList[domainFacility.FieldDevice], error) {
	return &domain.PaginatedList[domainFacility.FieldDevice]{
		Items:      []domainFacility.FieldDevice{},
		Total:      0,
		Page:       1,
		TotalPages: 1,
	}, nil
}

func (r *fakeFieldDeviceStore) GetIDsBySPSControllerSystemTypeIDs(_ context.Context, ids []uuid.UUID) ([]uuid.UUID, error) {
	idSet := make(map[uuid.UUID]struct{}, len(ids))
	for _, id := range ids {
		idSet[id] = struct{}{}
	}
	out := make([]uuid.UUID, 0)
	for id, item := range r.items {
		if _, ok := idSet[item.SPSControllerSystemTypeID]; ok {
			out = append(out, id)
		}
	}
	return out, nil
}

func (r *fakeFieldDeviceStore) ExistsApparatNrConflict(
	_ context.Context,
	spsControllerSystemTypeID uuid.UUID,
	systemPartID uuid.UUID,
	apparatID uuid.UUID,
	apparatNr int,
	excludeIDs []uuid.UUID,
) (bool, error) {
	exclude := make(map[uuid.UUID]struct{}, len(excludeIDs))
	for _, id := range excludeIDs {
		exclude[id] = struct{}{}
	}
	for id, item := range r.items {
		if _, ok := exclude[id]; ok {
			continue
		}
		if item.SPSControllerSystemTypeID != spsControllerSystemTypeID {
			continue
		}
		if item.ApparatID != apparatID {
			continue
		}
		if item.ApparatNr != apparatNr {
			continue
		}
		if item.SystemPartID != systemPartID {
			continue
		}
		return true, nil
	}
	return false, nil
}

func (r *fakeFieldDeviceStore) GetUsedApparatNumbers(
	_ context.Context,
	spsControllerSystemTypeID uuid.UUID,
	systemPartID uuid.UUID,
	apparatID uuid.UUID,
) ([]int, error) {
	out := make([]int, 0)
	for _, item := range r.items {
		if item.SPSControllerSystemTypeID != spsControllerSystemTypeID {
			continue
		}
		if item.ApparatID != apparatID {
			continue
		}
		if item.SystemPartID != systemPartID {
			continue
		}
		out = append(out, item.ApparatNr)
	}
	return out, nil
}

type fakeSpecificationStore struct {
	items map[uuid.UUID]*domainFacility.Specification
}

func (r *fakeSpecificationStore) GetByIds(_ context.Context, ids []uuid.UUID) ([]*domainFacility.Specification, error) {
	out := make([]*domainFacility.Specification, 0, len(ids))
	for _, id := range ids {
		if item, ok := r.items[id]; ok {
			clone := *item
			out = append(out, &clone)
		}
	}
	return out, nil
}

func (r *fakeSpecificationStore) Create(_ context.Context, entity *domainFacility.Specification) error {
	clone := *entity
	if clone.ID == uuid.Nil {
		clone.ID = uuid.New()
	}
	r.items[clone.ID] = &clone
	entity.ID = clone.ID
	return nil
}

func (r *fakeSpecificationStore) BulkCreate(_ context.Context, entities []*domainFacility.Specification, batchSize int) error {
	for _, entity := range entities {
		if err := r.Create(context.Background(), entity); err != nil {
			return err
		}
	}
	return nil
}

func (r *fakeSpecificationStore) Update(_ context.Context, entity *domainFacility.Specification) error {
	clone := *entity
	r.items[entity.ID] = &clone
	return nil
}

func (r *fakeSpecificationStore) DeleteByIds(_ context.Context, ids []uuid.UUID) error {
	for _, id := range ids {
		delete(r.items, id)
	}
	return nil
}

func (r *fakeSpecificationStore) GetPaginatedList(_ context.Context, params domain.PaginationParams) (*domain.PaginatedList[domainFacility.Specification], error) {
	items := make([]domainFacility.Specification, 0, len(r.items))
	for _, item := range r.items {
		items = append(items, *item)
	}
	return &domain.PaginatedList[domainFacility.Specification]{
		Items:      items,
		Total:      int64(len(items)),
		Page:       1,
		TotalPages: 1,
	}, nil
}

func (r *fakeSpecificationStore) GetByFieldDeviceIDs(_ context.Context, fieldDeviceIDs []uuid.UUID) ([]*domainFacility.Specification, error) {
	set := make(map[uuid.UUID]struct{}, len(fieldDeviceIDs))
	for _, id := range fieldDeviceIDs {
		set[id] = struct{}{}
	}

	out := make([]*domainFacility.Specification, 0)
	for _, item := range r.items {
		if item.FieldDeviceID == nil {
			continue
		}
		if _, ok := set[*item.FieldDeviceID]; !ok {
			continue
		}
		clone := *item
		out = append(out, &clone)
	}
	return out, nil
}

func (r *fakeSpecificationStore) DeleteByFieldDeviceIDs(_ context.Context, fieldDeviceIDs []uuid.UUID) error {
	set := make(map[uuid.UUID]struct{}, len(fieldDeviceIDs))
	for _, id := range fieldDeviceIDs {
		set[id] = struct{}{}
	}
	for id, item := range r.items {
		if item.FieldDeviceID == nil {
			continue
		}
		if _, ok := set[*item.FieldDeviceID]; ok {
			delete(r.items, id)
		}
	}
	return nil
}

type fakeSpsControllerSystemTypeRepo struct {
	items map[uuid.UUID]*domainFacility.SPSControllerSystemType
}

func (r *fakeSpsControllerSystemTypeRepo) GetByIds(_ context.Context, ids []uuid.UUID) ([]*domainFacility.SPSControllerSystemType, error) {
	out := make([]*domainFacility.SPSControllerSystemType, 0, len(ids))
	for _, id := range ids {
		if item, ok := r.items[id]; ok {
			clone := *item
			out = append(out, &clone)
		}
	}
	return out, nil
}

func (r *fakeSpsControllerSystemTypeRepo) Create(_ context.Context, entity *domainFacility.SPSControllerSystemType) error {
	clone := *entity
	r.items[entity.ID] = &clone
	return nil
}

func (r *fakeSpsControllerSystemTypeRepo) Update(_ context.Context, entity *domainFacility.SPSControllerSystemType) error {
	clone := *entity
	r.items[entity.ID] = &clone
	return nil
}

func (r *fakeSpsControllerSystemTypeRepo) DeleteByIds(_ context.Context, ids []uuid.UUID) error {
	for _, id := range ids {
		delete(r.items, id)
	}
	return nil
}

func (r *fakeSpsControllerSystemTypeRepo) GetPaginatedList(_ context.Context, params domain.PaginationParams) (*domain.PaginatedList[domainFacility.SPSControllerSystemType], error) {
	items := make([]domainFacility.SPSControllerSystemType, 0, len(r.items))
	for _, item := range r.items {
		items = append(items, *item)
	}
	return &domain.PaginatedList[domainFacility.SPSControllerSystemType]{
		Items:      items,
		Total:      int64(len(items)),
		Page:       1,
		TotalPages: 1,
	}, nil
}

func (r *fakeSpsControllerSystemTypeRepo) GetPaginatedListBySPSControllerID(
	_ context.Context,
	spsControllerID uuid.UUID,
	params domain.PaginationParams,
) (*domain.PaginatedList[domainFacility.SPSControllerSystemType], error) {
	items := make([]domainFacility.SPSControllerSystemType, 0)
	for _, item := range r.items {
		if item.SPSControllerID == spsControllerID {
			items = append(items, *item)
		}
	}
	return &domain.PaginatedList[domainFacility.SPSControllerSystemType]{
		Items:      items,
		Total:      int64(len(items)),
		Page:       1,
		TotalPages: 1,
	}, nil
}

func (r *fakeSpsControllerSystemTypeRepo) GetPaginatedListByProjectID(
	_ context.Context,
	projectID uuid.UUID,
	params domain.PaginationParams,
) (*domain.PaginatedList[domainFacility.SPSControllerSystemType], error) {
	return &domain.PaginatedList[domainFacility.SPSControllerSystemType]{
		Items:      []domainFacility.SPSControllerSystemType{},
		Total:      0,
		Page:       1,
		TotalPages: 1,
	}, nil
}

func (r *fakeSpsControllerSystemTypeRepo) ListBySPSControllerID(
	_ context.Context,
	spsControllerID uuid.UUID,
) ([]*domainFacility.SPSControllerSystemType, error) {
	out := make([]*domainFacility.SPSControllerSystemType, 0)
	for _, item := range r.items {
		if item.SPSControllerID != spsControllerID {
			continue
		}
		clone := *item
		out = append(out, &clone)
	}
	return out, nil
}

func (r *fakeSpsControllerSystemTypeRepo) GetIDsBySPSControllerIDs(_ context.Context, ids []uuid.UUID) ([]uuid.UUID, error) {
	if len(ids) == 0 {
		return []uuid.UUID{}, nil
	}
	idSet := make(map[uuid.UUID]struct{}, len(ids))
	for _, id := range ids {
		idSet[id] = struct{}{}
	}
	out := make([]uuid.UUID, 0)
	for id, item := range r.items {
		if _, ok := idSet[item.SPSControllerID]; ok {
			out = append(out, id)
		}
	}
	return out, nil
}

func (r *fakeSpsControllerSystemTypeRepo) DeleteBySPSControllerIDs(_ context.Context, ids []uuid.UUID) error {
	if len(ids) == 0 {
		return nil
	}
	idSet := make(map[uuid.UUID]struct{}, len(ids))
	for _, id := range ids {
		idSet[id] = struct{}{}
	}
	for id, item := range r.items {
		if _, ok := idSet[item.SPSControllerID]; ok {
			delete(r.items, id)
		}
	}
	return nil
}

type fakeSystemTypeRepo struct {
	items map[uuid.UUID]*domainFacility.SystemType
}

func (r *fakeSystemTypeRepo) GetByIds(_ context.Context, ids []uuid.UUID) ([]*domainFacility.SystemType, error) {
	out := make([]*domainFacility.SystemType, 0, len(ids))
	for _, id := range ids {
		if item, ok := r.items[id]; ok {
			clone := *item
			out = append(out, &clone)
		}
	}
	return out, nil
}

func (r *fakeSystemTypeRepo) Create(_ context.Context, entity *domainFacility.SystemType) error {
	clone := *entity
	r.items[entity.ID] = &clone
	return nil
}

func (r *fakeSystemTypeRepo) Update(_ context.Context, entity *domainFacility.SystemType) error {
	clone := *entity
	r.items[entity.ID] = &clone
	return nil
}

func (r *fakeSystemTypeRepo) DeleteByIds(_ context.Context, ids []uuid.UUID) error {
	for _, id := range ids {
		delete(r.items, id)
	}
	return nil
}

func (r *fakeSystemTypeRepo) GetPaginatedList(_ context.Context, params domain.PaginationParams) (*domain.PaginatedList[domainFacility.SystemType], error) {
	items := make([]domainFacility.SystemType, 0, len(r.items))
	for _, item := range r.items {
		items = append(items, *item)
	}
	return &domain.PaginatedList[domainFacility.SystemType]{
		Items:      items,
		Total:      int64(len(items)),
		Page:       1,
		TotalPages: 1,
	}, nil
}

func (r *fakeSystemTypeRepo) ExistsName(_ context.Context, name string, excludeID *uuid.UUID) (bool, error) {
	return false, nil
}

func (r *fakeSystemTypeRepo) ExistsOverlappingRange(_ context.Context, numberMin, numberMax int, excludeID *uuid.UUID) (bool, error) {
	return false, nil
}

type fakeApparatRepo struct {
	items map[uuid.UUID]*domainFacility.Apparat
}

func (r *fakeApparatRepo) ExistsShortName(_ context.Context, shortName string, excludeID *uuid.UUID) (bool, error) {
	for _, item := range r.items {
		if excludeID != nil && item.ID == *excludeID {
			continue
		}
		if strings.EqualFold(item.ShortName, shortName) {
			return true, nil
		}
	}
	return false, nil
}

func (r *fakeApparatRepo) ExistsName(_ context.Context, name string, excludeID *uuid.UUID) (bool, error) {
	for _, item := range r.items {
		if excludeID != nil && item.ID == *excludeID {
			continue
		}
		if strings.EqualFold(item.Name, name) {
			return true, nil
		}
	}
	return false, nil
}

func (r *fakeApparatRepo) GetByIds(_ context.Context, ids []uuid.UUID) ([]*domainFacility.Apparat, error) {
	out := make([]*domainFacility.Apparat, 0, len(ids))
	for _, id := range ids {
		if item, ok := r.items[id]; ok {
			out = append(out, cloneApparat(item))
		}
	}
	sort.Slice(out, func(i, j int) bool {
		leftShort := strings.ToLower(out[i].ShortName)
		rightShort := strings.ToLower(out[j].ShortName)
		if leftShort != rightShort {
			return leftShort < rightShort
		}
		return strings.ToLower(out[i].Name) < strings.ToLower(out[j].Name)
	})
	return out, nil
}

func (r *fakeApparatRepo) Create(_ context.Context, entity *domainFacility.Apparat) error {
	clone := *entity
	r.items[entity.ID] = &clone
	return nil
}

func (r *fakeApparatRepo) Update(_ context.Context, entity *domainFacility.Apparat) error {
	clone := *entity
	r.items[entity.ID] = &clone
	return nil
}

func (r *fakeApparatRepo) DeleteByIds(_ context.Context, ids []uuid.UUID) error {
	for _, id := range ids {
		delete(r.items, id)
	}
	return nil
}

func (r *fakeApparatRepo) GetPaginatedList(_ context.Context, params domain.PaginationParams) (*domain.PaginatedList[domainFacility.Apparat], error) {
	items := make([]domainFacility.Apparat, 0, len(r.items))
	for _, item := range r.items {
		items = append(items, *item)
	}
	return &domain.PaginatedList[domainFacility.Apparat]{
		Items:      items,
		Total:      int64(len(items)),
		Page:       1,
		TotalPages: 1,
	}, nil
}

func (r *fakeApparatRepo) GetPaginatedListWithFilters(_ context.Context, params domain.PaginationParams, _ domainFacility.ApparatFilterParams) (*domain.PaginatedList[domainFacility.Apparat], error) {
	return r.GetPaginatedList(context.Background(), params)
}

type fakeSystemPartRepo struct {
	items map[uuid.UUID]*domainFacility.SystemPart
}

func (r *fakeSystemPartRepo) ExistsShortName(_ context.Context, shortName string, excludeID *uuid.UUID) (bool, error) {
	for _, item := range r.items {
		if excludeID != nil && item.ID == *excludeID {
			continue
		}
		if strings.EqualFold(item.ShortName, shortName) {
			return true, nil
		}
	}
	return false, nil
}

func (r *fakeSystemPartRepo) ExistsName(_ context.Context, name string, excludeID *uuid.UUID) (bool, error) {
	for _, item := range r.items {
		if excludeID != nil && item.ID == *excludeID {
			continue
		}
		if strings.EqualFold(item.Name, name) {
			return true, nil
		}
	}
	return false, nil
}

func (r *fakeSystemPartRepo) GetByIds(_ context.Context, ids []uuid.UUID) ([]*domainFacility.SystemPart, error) {
	out := make([]*domainFacility.SystemPart, 0, len(ids))
	for _, id := range ids {
		if item, ok := r.items[id]; ok {
			clone := *item
			out = append(out, &clone)
		}
	}
	return out, nil
}

func (r *fakeSystemPartRepo) Create(_ context.Context, entity *domainFacility.SystemPart) error {
	clone := *entity
	r.items[entity.ID] = &clone
	return nil
}

func (r *fakeSystemPartRepo) Update(_ context.Context, entity *domainFacility.SystemPart) error {
	clone := *entity
	r.items[entity.ID] = &clone
	return nil
}

func (r *fakeSystemPartRepo) DeleteByIds(_ context.Context, ids []uuid.UUID) error {
	for _, id := range ids {
		delete(r.items, id)
	}
	return nil
}

func (r *fakeSystemPartRepo) GetPaginatedList(_ context.Context, params domain.PaginationParams) (*domain.PaginatedList[domainFacility.SystemPart], error) {
	items := make([]domainFacility.SystemPart, 0, len(r.items))
	for _, item := range r.items {
		items = append(items, *item)
	}
	return &domain.PaginatedList[domainFacility.SystemPart]{
		Items:      items,
		Total:      int64(len(items)),
		Page:       1,
		TotalPages: 1,
	}, nil
}

func intPtr(value int) *int {
	return &value
}

func stringPtr(value string) *string {
	return &value
}

type fakeObjectDataStore struct {
	templates        []*domainFacility.ObjectData
	projectTemplates map[uuid.UUID][]*domainFacility.ObjectData
	items            map[uuid.UUID]*domainFacility.ObjectData
}

func (r *fakeObjectDataStore) GetByIds(_ context.Context, ids []uuid.UUID) ([]*domainFacility.ObjectData, error) {
	out := make([]*domainFacility.ObjectData, 0, len(ids))
	for _, id := range ids {
		if item, ok := r.items[id]; ok {
			out = append(out, cloneObjectData(item))
		}
	}
	return out, nil
}

func (r *fakeObjectDataStore) Create(_ context.Context, entity *domainFacility.ObjectData) error {
	if r.items == nil {
		r.items = make(map[uuid.UUID]*domainFacility.ObjectData)
	}
	r.items[entity.ID] = cloneObjectData(entity)
	return nil
}

func (r *fakeObjectDataStore) Update(_ context.Context, entity *domainFacility.ObjectData) error {
	if r.items == nil {
		r.items = make(map[uuid.UUID]*domainFacility.ObjectData)
	}
	r.items[entity.ID] = cloneObjectData(entity)
	return nil
}

func (r *fakeObjectDataStore) DeleteByIds(_ context.Context, ids []uuid.UUID) error {
	for _, id := range ids {
		delete(r.items, id)
	}
	return nil
}

func (r *fakeObjectDataStore) GetPaginatedList(_ context.Context, params domain.PaginationParams) (*domain.PaginatedList[domainFacility.ObjectData], error) {
	items := make([]domainFacility.ObjectData, 0, len(r.items))
	for _, item := range r.items {
		items = append(items, *cloneObjectData(item))
	}
	return &domain.PaginatedList[domainFacility.ObjectData]{
		Items:      items,
		Total:      int64(len(items)),
		Page:       1,
		TotalPages: 1,
	}, nil
}

func (r *fakeObjectDataStore) GetBacnetObjectIDs(_ context.Context, objectDataID uuid.UUID) ([]uuid.UUID, error) {
	return []uuid.UUID{}, nil
}

func (r *fakeObjectDataStore) ExistsByDescription(_ context.Context, projectID *uuid.UUID, description string, excludeID *uuid.UUID) (bool, error) {
	return false, nil
}

func (r *fakeObjectDataStore) GetTemplates(_ context.Context) ([]*domainFacility.ObjectData, error) {
	return sortedObjectDataSlice(r.templates), nil
}

func (r *fakeObjectDataStore) GetTemplatesLite(_ context.Context) ([]*domainFacility.ObjectData, error) {
	return sortedObjectDataSlice(r.templates), nil
}

func (r *fakeObjectDataStore) GetForProject(_ context.Context, projectID uuid.UUID) ([]*domainFacility.ObjectData, error) {
	return sortedObjectDataSlice(r.projectTemplates[projectID]), nil
}

func (r *fakeObjectDataStore) GetForProjectLite(_ context.Context, projectID uuid.UUID) ([]*domainFacility.ObjectData, error) {
	return sortedObjectDataSlice(r.projectTemplates[projectID]), nil
}

func (r *fakeObjectDataStore) GetPaginatedListForProject(_ context.Context, projectID uuid.UUID, params domain.PaginationParams) (*domain.PaginatedList[domainFacility.ObjectData], error) {
	items := derefObjectDatas(sortedObjectDataSlice(r.projectTemplates[projectID]))
	return &domain.PaginatedList[domainFacility.ObjectData]{
		Items:      items,
		Total:      int64(len(items)),
		Page:       1,
		TotalPages: 1,
	}, nil
}

func (r *fakeObjectDataStore) GetPaginatedListByApparatID(_ context.Context, apparatID uuid.UUID, params domain.PaginationParams) (*domain.PaginatedList[domainFacility.ObjectData], error) {
	return &domain.PaginatedList[domainFacility.ObjectData]{Items: []domainFacility.ObjectData{}, Total: 0, Page: 1, TotalPages: 1}, nil
}

func (r *fakeObjectDataStore) GetPaginatedListBySystemPartID(_ context.Context, systemPartID uuid.UUID, params domain.PaginationParams) (*domain.PaginatedList[domainFacility.ObjectData], error) {
	return &domain.PaginatedList[domainFacility.ObjectData]{Items: []domainFacility.ObjectData{}, Total: 0, Page: 1, TotalPages: 1}, nil
}

func (r *fakeObjectDataStore) GetPaginatedListByApparatAndSystemPartID(_ context.Context, apparatID, systemPartID uuid.UUID, params domain.PaginationParams) (*domain.PaginatedList[domainFacility.ObjectData], error) {
	return &domain.PaginatedList[domainFacility.ObjectData]{Items: []domainFacility.ObjectData{}, Total: 0, Page: 1, TotalPages: 1}, nil
}

func (r *fakeObjectDataStore) GetPaginatedListForProjectByApparatID(_ context.Context, projectID, apparatID uuid.UUID, params domain.PaginationParams) (*domain.PaginatedList[domainFacility.ObjectData], error) {
	return &domain.PaginatedList[domainFacility.ObjectData]{Items: []domainFacility.ObjectData{}, Total: 0, Page: 1, TotalPages: 1}, nil
}

func (r *fakeObjectDataStore) GetPaginatedListForProjectBySystemPartID(_ context.Context, projectID, systemPartID uuid.UUID, params domain.PaginationParams) (*domain.PaginatedList[domainFacility.ObjectData], error) {
	return &domain.PaginatedList[domainFacility.ObjectData]{Items: []domainFacility.ObjectData{}, Total: 0, Page: 1, TotalPages: 1}, nil
}

func (r *fakeObjectDataStore) GetPaginatedListForProjectByApparatAndSystemPartID(_ context.Context, projectID, apparatID, systemPartID uuid.UUID, params domain.PaginationParams) (*domain.PaginatedList[domainFacility.ObjectData], error) {
	return &domain.PaginatedList[domainFacility.ObjectData]{Items: []domainFacility.ObjectData{}, Total: 0, Page: 1, TotalPages: 1}, nil
}

func cloneApparat(item *domainFacility.Apparat) *domainFacility.Apparat {
	clone := *item
	if item.SystemParts != nil {
		clone.SystemParts = make([]*domainFacility.SystemPart, 0, len(item.SystemParts))
		for _, systemPart := range item.SystemParts {
			if systemPart == nil {
				continue
			}
			systemPartClone := *systemPart
			clone.SystemParts = append(clone.SystemParts, &systemPartClone)
		}
		sort.Slice(clone.SystemParts, func(i, j int) bool {
			leftShort := strings.ToLower(clone.SystemParts[i].ShortName)
			rightShort := strings.ToLower(clone.SystemParts[j].ShortName)
			if leftShort != rightShort {
				return leftShort < rightShort
			}
			return strings.ToLower(clone.SystemParts[i].Name) < strings.ToLower(clone.SystemParts[j].Name)
		})
	}
	return &clone
}

func cloneObjectData(item *domainFacility.ObjectData) *domainFacility.ObjectData {
	clone := *item
	if item.Apparats != nil {
		clone.Apparats = make([]*domainFacility.Apparat, 0, len(item.Apparats))
		for _, apparat := range item.Apparats {
			if apparat == nil {
				continue
			}
			clone.Apparats = append(clone.Apparats, cloneApparat(apparat))
		}
		sort.Slice(clone.Apparats, func(i, j int) bool {
			leftShort := strings.ToLower(clone.Apparats[i].ShortName)
			rightShort := strings.ToLower(clone.Apparats[j].ShortName)
			if leftShort != rightShort {
				return leftShort < rightShort
			}
			return strings.ToLower(clone.Apparats[i].Name) < strings.ToLower(clone.Apparats[j].Name)
		})
	}
	return &clone
}

func sortedObjectDataSlice(items []*domainFacility.ObjectData) []*domainFacility.ObjectData {
	out := make([]*domainFacility.ObjectData, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		out = append(out, cloneObjectData(item))
	}
	sort.Slice(out, func(i, j int) bool {
		leftDescription := strings.ToLower(out[i].Description)
		rightDescription := strings.ToLower(out[j].Description)
		if leftDescription != rightDescription {
			return leftDescription < rightDescription
		}
		return strings.ToLower(out[i].Version) < strings.ToLower(out[j].Version)
	})
	return out
}

func derefObjectDatas(items []*domainFacility.ObjectData) []domainFacility.ObjectData {
	out := make([]domainFacility.ObjectData, 0, len(items))
	for _, item := range items {
		if item != nil {
			out = append(out, *item)
		}
	}
	return out
}

func TestFieldDeviceService_DeleteByIDs_UsesSingleRepositoryDelete(t *testing.T) {
	fdID1 := uuid.New()
	fdID2 := uuid.New()
	apparatID := uuid.New()
	systemPartID := uuid.New()
	spsSystemTypeID := uuid.New()

	fieldDeviceRepo := &fakeFieldDeviceStore{
		items: map[uuid.UUID]*domainFacility.FieldDevice{
			fdID1: newFieldDevice(fdID1, spsSystemTypeID, apparatID, systemPartID, 1),
			fdID2: newFieldDevice(fdID2, spsSystemTypeID, apparatID, systemPartID, 2),
		},
	}

	svc := facility.NewFieldDeviceService(
		fieldDeviceRepo,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
	)

	if err := svc.DeleteByIDs(context.Background(), []uuid.UUID{fdID1, fdID2}); err != nil {
		t.Fatalf("expected delete to succeed, got %v", err)
	}
	if fieldDeviceRepo.deleteCalls != 1 {
		t.Fatalf("expected one repository delete call, got %d", fieldDeviceRepo.deleteCalls)
	}
	if len(fieldDeviceRepo.deletedBatches) != 1 || len(fieldDeviceRepo.deletedBatches[0]) != 2 {
		t.Fatalf("expected one delete batch with 2 ids, got %+v", fieldDeviceRepo.deletedBatches)
	}
	if len(fieldDeviceRepo.items) != 0 {
		t.Fatalf("expected all field devices to be deleted, got %d remaining", len(fieldDeviceRepo.items))
	}
}

func TestFieldDeviceService_GetFieldDeviceOptions_PreservesRepositoryOrdering(t *testing.T) {
	projectID := uuid.New()
	objectDataZuluID := uuid.New()
	objectDataAlphaV2ID := uuid.New()
	objectDataAlphaV10ID := uuid.New()
	apparatZuluID := uuid.New()
	apparatAlphaID := uuid.New()
	apparatBetaID := uuid.New()
	systemPartZuluID := uuid.New()
	systemPartAlphaID := uuid.New()
	systemPartBetaID := uuid.New()

	objectDataRepo := &fakeObjectDataStore{
		templates: []*domainFacility.ObjectData{
			{
				Base:        domain.Base{ID: objectDataZuluID},
				Description: "Zulu Object",
				Version:     "2",
				IsActive:    true,
				Apparats: []*domainFacility.Apparat{
					{Base: domain.Base{ID: apparatZuluID}, ShortName: "Zulu", Name: "Pump"},
					{Base: domain.Base{ID: apparatAlphaID}, ShortName: "Alpha", Name: "Valve"},
				},
			},
			{
				Base:        domain.Base{ID: objectDataAlphaV10ID},
				Description: "Alpha Object",
				Version:     "10",
				IsActive:    true,
				Apparats: []*domainFacility.Apparat{
					{Base: domain.Base{ID: apparatZuluID}, ShortName: "Zulu", Name: "Pump"},
				},
			},
			{
				Base:        domain.Base{ID: objectDataAlphaV2ID},
				Description: "Alpha Object",
				Version:     "2",
				IsActive:    true,
				Apparats: []*domainFacility.Apparat{
					{Base: domain.Base{ID: apparatBetaID}, ShortName: "Beta", Name: "Damper"},
					{Base: domain.Base{ID: apparatAlphaID}, ShortName: "Alpha", Name: "Valve"},
				},
			},
		},
		projectTemplates: map[uuid.UUID][]*domainFacility.ObjectData{
			projectID: {
				{
					Base:        domain.Base{ID: objectDataZuluID},
					Description: "Zulu Object",
					Version:     "2",
					IsActive:    true,
					ProjectID:   &projectID,
					Apparats: []*domainFacility.Apparat{
						{Base: domain.Base{ID: apparatZuluID}, ShortName: "Zulu", Name: "Pump"},
						{Base: domain.Base{ID: apparatAlphaID}, ShortName: "Alpha", Name: "Valve"},
					},
				},
				{
					Base:        domain.Base{ID: objectDataAlphaV2ID},
					Description: "Alpha Object",
					Version:     "2",
					IsActive:    true,
					ProjectID:   &projectID,
					Apparats: []*domainFacility.Apparat{
						{Base: domain.Base{ID: apparatBetaID}, ShortName: "Beta", Name: "Damper"},
						{Base: domain.Base{ID: apparatAlphaID}, ShortName: "Alpha", Name: "Valve"},
					},
				},
			},
		},
	}

	apparatRepo := &fakeApparatRepo{
		items: map[uuid.UUID]*domainFacility.Apparat{
			apparatZuluID: {
				Base:      domain.Base{ID: apparatZuluID},
				ShortName: "Zulu",
				Name:      "Pump",
				SystemParts: []*domainFacility.SystemPart{
					{Base: domain.Base{ID: systemPartZuluID}, ShortName: "Zulu", Name: "Zone"},
					{Base: domain.Base{ID: systemPartAlphaID}, ShortName: "Alpha", Name: "Air"},
				},
			},
			apparatAlphaID: {
				Base:      domain.Base{ID: apparatAlphaID},
				ShortName: "Alpha",
				Name:      "Valve",
				SystemParts: []*domainFacility.SystemPart{
					{Base: domain.Base{ID: systemPartBetaID}, ShortName: "Beta", Name: "Heat"},
					{Base: domain.Base{ID: systemPartAlphaID}, ShortName: "Alpha", Name: "Air"},
				},
			},
			apparatBetaID: {
				Base:      domain.Base{ID: apparatBetaID},
				ShortName: "Beta",
				Name:      "Damper",
				SystemParts: []*domainFacility.SystemPart{
					{Base: domain.Base{ID: systemPartBetaID}, ShortName: "Beta", Name: "Heat"},
				},
			},
		},
	}

	svc := facility.NewFieldDeviceService(
		nil,
		nil,
		nil,
		apparatRepo,
		nil,
		nil,
		nil,
		objectDataRepo,
		nil,
		nil,
	)

	options, err := svc.GetFieldDeviceOptions(context.Background())
	if err != nil {
		t.Fatalf("expected options to load, got %v", err)
	}

	if got := []string{options.ObjectDatas[0].Description + ":" + options.ObjectDatas[0].Version, options.ObjectDatas[1].Description + ":" + options.ObjectDatas[1].Version, options.ObjectDatas[2].Description + ":" + options.ObjectDatas[2].Version}; got[0] != "Alpha Object:10" || got[1] != "Alpha Object:2" || got[2] != "Zulu Object:2" {
		t.Fatalf("expected object datas in repository order, got %#v", got)
	}

	if got := []string{options.Apparats[0].ShortName, options.Apparats[1].ShortName, options.Apparats[2].ShortName}; got[0] != "Alpha" || got[1] != "Beta" || got[2] != "Zulu" {
		t.Fatalf("expected apparats in repository order, got %#v", got)
	}

	if got := []string{options.SystemParts[0].ShortName, options.SystemParts[1].ShortName, options.SystemParts[2].ShortName}; got[0] != "Alpha" || got[1] != "Beta" || got[2] != "Zulu" {
		t.Fatalf("expected system parts in repository order, got %#v", got)
	}

	projectOptions, err := svc.GetFieldDeviceOptionsForProject(context.Background(), projectID)
	if err != nil {
		t.Fatalf("expected project options to load, got %v", err)
	}

	if got := []string{projectOptions.ObjectDatas[0].Description + ":" + projectOptions.ObjectDatas[0].Version, projectOptions.ObjectDatas[1].Description + ":" + projectOptions.ObjectDatas[1].Version}; got[0] != "Alpha Object:2" || got[1] != "Zulu Object:2" {
		t.Fatalf("expected project object datas in repository order, got %#v", got)
	}

	if got := []string{projectOptions.Apparats[0].ShortName, projectOptions.Apparats[1].ShortName, projectOptions.Apparats[2].ShortName}; got[0] != "Alpha" || got[1] != "Beta" || got[2] != "Zulu" {
		t.Fatalf("expected project apparats in repository order, got %#v", got)
	}
}

func newFieldDevice(
	id uuid.UUID,
	spsControllerSystemTypeID uuid.UUID,
	apparatID uuid.UUID,
	systemPartID uuid.UUID,
	apparatNr int,
) *domainFacility.FieldDevice {
	return &domainFacility.FieldDevice{
		Base:                      domain.Base{ID: id},
		SPSControllerSystemTypeID: spsControllerSystemTypeID,
		ApparatID:                 apparatID,
		SystemPartID:              systemPartID,
		ApparatNr:                 apparatNr,
	}
}

func TestFieldDeviceService_MultiCreate_CharacterizesPartialSuccessAndBatchConflicts(t *testing.T) {
	fd1ID := uuid.New()
	fd2ID := uuid.New()
	apparatID := uuid.New()
	systemPartID := uuid.New()
	spsSystemTypeID := uuid.New()
	systemTypeID := uuid.New()

	fieldDeviceRepo := &fakeFieldDeviceStore{items: map[uuid.UUID]*domainFacility.FieldDevice{}}
	spsSystemTypeRepo := &fakeSpsControllerSystemTypeRepo{
		items: map[uuid.UUID]*domainFacility.SPSControllerSystemType{
			spsSystemTypeID: {
				Base:         domain.Base{ID: spsSystemTypeID},
				SystemTypeID: systemTypeID,
			},
		},
	}
	systemTypeRepo := &fakeSystemTypeRepo{
		items: map[uuid.UUID]*domainFacility.SystemType{
			systemTypeID: {Base: domain.Base{ID: systemTypeID}, NumberMin: 1, NumberMax: 99},
		},
	}
	apparatRepo := &fakeApparatRepo{
		items: map[uuid.UUID]*domainFacility.Apparat{
			apparatID: {Base: domain.Base{ID: apparatID}, ShortName: "PMP", Name: "Pump"},
		},
	}
	systemPartRepo := &fakeSystemPartRepo{
		items: map[uuid.UUID]*domainFacility.SystemPart{
			systemPartID: {Base: domain.Base{ID: systemPartID}, ShortName: "AIR", Name: "Air"},
		},
	}

	svc := facility.NewFieldDeviceService(
		fieldDeviceRepo,
		spsSystemTypeRepo,
		systemTypeRepo,
		apparatRepo,
		systemPartRepo,
		nil,
		nil,
		nil,
		nil,
		nil,
	)

	result := svc.MultiCreate(context.Background(), []domainFacility.FieldDeviceCreateItem{
		{FieldDevice: newFieldDevice(fd1ID, spsSystemTypeID, apparatID, systemPartID, 11)},
		{FieldDevice: newFieldDevice(fd2ID, spsSystemTypeID, apparatID, systemPartID, 11)},
		{},
	})

	if result.TotalRequests != 3 || result.SuccessCount != 1 || result.FailureCount != 2 {
		t.Fatalf("expected one success and two failures, got %+v", result)
	}
	if !result.Results[0].Success || result.Results[0].FieldDevice == nil || result.Results[0].FieldDevice.ID != fd1ID {
		t.Fatalf("expected first item to succeed with fd1, got %+v", result.Results[0])
	}
	if result.Results[1].Success || result.Results[1].ErrorField != "fielddevice.apparat_nr" || result.Results[1].Error != "apparatnummer ist bereits vergeben" {
		t.Fatalf("expected second item to fail on in-batch apparat_nr conflict, got %+v", result.Results[1])
	}
	if result.Results[2].Success || result.Results[2].ErrorField != "fielddevice" || result.Results[2].Error != "field device is required" {
		t.Fatalf("expected third item to fail as missing field device, got %+v", result.Results[2])
	}
	if _, ok := fieldDeviceRepo.items[fd1ID]; !ok || len(fieldDeviceRepo.items) != 1 {
		t.Fatalf("expected only successful field device to be persisted, got %+v", fieldDeviceRepo.items)
	}
}

func TestFieldDeviceService_BulkUpdate_AllowsSwapApparatNr(t *testing.T) {
	fd1ID := uuid.New()
	fd2ID := uuid.New()
	apparatID := uuid.New()
	systemPartID := uuid.New()
	spsSystemTypeID := uuid.New()
	systemTypeID := uuid.New()

	fieldDeviceRepo := &fakeFieldDeviceStore{
		items: map[uuid.UUID]*domainFacility.FieldDevice{
			fd1ID: newFieldDevice(fd1ID, spsSystemTypeID, apparatID, systemPartID, 1),
			fd2ID: newFieldDevice(fd2ID, spsSystemTypeID, apparatID, systemPartID, 2),
		},
	}
	spsSystemTypeRepo := &fakeSpsControllerSystemTypeRepo{
		items: map[uuid.UUID]*domainFacility.SPSControllerSystemType{
			spsSystemTypeID: {
				Base:         domain.Base{ID: spsSystemTypeID},
				SystemTypeID: systemTypeID,
			},
		},
	}
	systemTypeRepo := &fakeSystemTypeRepo{
		items: map[uuid.UUID]*domainFacility.SystemType{
			systemTypeID: {Base: domain.Base{ID: systemTypeID}},
		},
	}
	apparatRepo := &fakeApparatRepo{
		items: map[uuid.UUID]*domainFacility.Apparat{
			apparatID: {Base: domain.Base{ID: apparatID}},
		},
	}
	systemPartRepo := &fakeSystemPartRepo{
		items: map[uuid.UUID]*domainFacility.SystemPart{
			systemPartID: {Base: domain.Base{ID: systemPartID}},
		},
	}

	svc := facility.NewFieldDeviceService(
		fieldDeviceRepo,
		spsSystemTypeRepo,
		systemTypeRepo,
		apparatRepo,
		systemPartRepo,
		nil,
		nil,
		nil,
		nil,
		nil,
	)

	updates := []domainFacility.BulkFieldDeviceUpdate{
		{ID: fd1ID, ApparatNr: intPtr(2)},
		{ID: fd2ID, ApparatNr: intPtr(1)},
	}
	result := svc.BulkUpdate(context.Background(), updates)

	if result.FailureCount != 0 {
		t.Fatalf("expected 0 failures, got %d", result.FailureCount)
	}
	if result.SuccessCount != 2 {
		t.Fatalf("expected 2 successes, got %d", result.SuccessCount)
	}

	if fieldDeviceRepo.items[fd1ID].ApparatNr != 2 {
		t.Fatalf("expected fd1 apparat_nr=2, got %d", fieldDeviceRepo.items[fd1ID].ApparatNr)
	}
	if fieldDeviceRepo.items[fd2ID].ApparatNr != 1 {
		t.Fatalf("expected fd2 apparat_nr=1, got %d", fieldDeviceRepo.items[fd2ID].ApparatNr)
	}
}

func TestFieldDeviceService_BulkUpdate_DetectsApparatNrConflict(t *testing.T) {
	fd1ID := uuid.New()
	fd2ID := uuid.New()
	apparatID := uuid.New()
	systemPartID := uuid.New()
	spsSystemTypeID := uuid.New()
	systemTypeID := uuid.New()

	fieldDeviceRepo := &fakeFieldDeviceStore{
		items: map[uuid.UUID]*domainFacility.FieldDevice{
			fd1ID: newFieldDevice(fd1ID, spsSystemTypeID, apparatID, systemPartID, 1),
			fd2ID: newFieldDevice(fd2ID, spsSystemTypeID, apparatID, systemPartID, 3),
		},
	}
	spsSystemTypeRepo := &fakeSpsControllerSystemTypeRepo{
		items: map[uuid.UUID]*domainFacility.SPSControllerSystemType{
			spsSystemTypeID: {
				Base:         domain.Base{ID: spsSystemTypeID},
				SystemTypeID: systemTypeID,
			},
		},
	}
	systemTypeRepo := &fakeSystemTypeRepo{
		items: map[uuid.UUID]*domainFacility.SystemType{
			systemTypeID: {Base: domain.Base{ID: systemTypeID}},
		},
	}
	apparatRepo := &fakeApparatRepo{
		items: map[uuid.UUID]*domainFacility.Apparat{
			apparatID: {Base: domain.Base{ID: apparatID}},
		},
	}
	systemPartRepo := &fakeSystemPartRepo{
		items: map[uuid.UUID]*domainFacility.SystemPart{
			systemPartID: {Base: domain.Base{ID: systemPartID}},
		},
	}

	svc := facility.NewFieldDeviceService(
		fieldDeviceRepo,
		spsSystemTypeRepo,
		systemTypeRepo,
		apparatRepo,
		systemPartRepo,
		nil,
		nil,
		nil,
		nil,
		nil,
	)

	updates := []domainFacility.BulkFieldDeviceUpdate{
		{ID: fd1ID, ApparatNr: intPtr(3)},
	}
	result := svc.BulkUpdate(context.Background(), updates)

	if result.FailureCount != 1 {
		t.Fatalf("expected 1 failure, got %d", result.FailureCount)
	}
	if result.Results[0].Fields["fielddevice.apparat_nr"] == "" {
		t.Fatalf("expected fielddevice.apparat_nr error, got %+v", result.Results[0].Fields)
	}

	if fieldDeviceRepo.items[fd1ID].ApparatNr != 1 {
		t.Fatalf("expected fd1 apparat_nr unchanged, got %d", fieldDeviceRepo.items[fd1ID].ApparatNr)
	}
}

func TestFieldDeviceService_BulkUpdate_PartialUpdate_ApparatNr_Succeeds(t *testing.T) {
	// Arrange
	fd1ID := uuid.New()
	apparatID := uuid.New()
	systemPartID := uuid.New()
	spsSystemTypeID := uuid.New()
	systemTypeID := uuid.New()

	fieldDeviceRepo := &fakeFieldDeviceStore{
		items: map[uuid.UUID]*domainFacility.FieldDevice{
			fd1ID: newFieldDevice(fd1ID, spsSystemTypeID, apparatID, systemPartID, 1),
		},
	}
	spsSystemTypeRepo := &fakeSpsControllerSystemTypeRepo{
		items: map[uuid.UUID]*domainFacility.SPSControllerSystemType{
			spsSystemTypeID: {
				Base:         domain.Base{ID: spsSystemTypeID},
				SystemTypeID: systemTypeID,
			},
		},
	}
	systemTypeRepo := &fakeSystemTypeRepo{
		items: map[uuid.UUID]*domainFacility.SystemType{
			systemTypeID: {Base: domain.Base{ID: systemTypeID}},
		},
	}
	apparatRepo := &fakeApparatRepo{
		items: map[uuid.UUID]*domainFacility.Apparat{
			apparatID: {Base: domain.Base{ID: apparatID}},
		},
	}
	systemPartRepo := &fakeSystemPartRepo{
		items: map[uuid.UUID]*domainFacility.SystemPart{
			systemPartID: {Base: domain.Base{ID: systemPartID}},
		},
	}

	svc := facility.NewFieldDeviceService(
		fieldDeviceRepo,
		spsSystemTypeRepo,
		systemTypeRepo,
		apparatRepo,
		systemPartRepo,
		nil,
		nil,
		nil,
		nil,
		nil,
	)

	updates := []domainFacility.BulkFieldDeviceUpdate{
		{ID: fd1ID, ApparatNr: intPtr(2)},
	}

	// Act
	result := svc.BulkUpdate(context.Background(), updates)

	// Assert
	if result.FailureCount != 0 {
		t.Fatalf("expected 0 failures, got %d", result.FailureCount)
	}
	if fieldDeviceRepo.items[fd1ID].ApparatNr != 2 {
		t.Fatalf("expected apparat_nr to update to 2, got %d", fieldDeviceRepo.items[fd1ID].ApparatNr)
	}
}

func TestFieldDeviceService_BulkUpdate_PartialUpdate_ApparatNr_Conflict(t *testing.T) {
	// Arrange
	fd1ID := uuid.New()
	fd2ID := uuid.New()
	apparatID := uuid.New()
	systemPartID := uuid.New()
	spsSystemTypeID := uuid.New()
	systemTypeID := uuid.New()

	fieldDeviceRepo := &fakeFieldDeviceStore{
		items: map[uuid.UUID]*domainFacility.FieldDevice{
			fd1ID: newFieldDevice(fd1ID, spsSystemTypeID, apparatID, systemPartID, 1),
			fd2ID: newFieldDevice(fd2ID, spsSystemTypeID, apparatID, systemPartID, 2),
		},
	}
	spsSystemTypeRepo := &fakeSpsControllerSystemTypeRepo{
		items: map[uuid.UUID]*domainFacility.SPSControllerSystemType{
			spsSystemTypeID: {
				Base:         domain.Base{ID: spsSystemTypeID},
				SystemTypeID: systemTypeID,
			},
		},
	}
	systemTypeRepo := &fakeSystemTypeRepo{
		items: map[uuid.UUID]*domainFacility.SystemType{
			systemTypeID: {Base: domain.Base{ID: systemTypeID}},
		},
	}
	apparatRepo := &fakeApparatRepo{
		items: map[uuid.UUID]*domainFacility.Apparat{
			apparatID: {Base: domain.Base{ID: apparatID}},
		},
	}
	systemPartRepo := &fakeSystemPartRepo{
		items: map[uuid.UUID]*domainFacility.SystemPart{
			systemPartID: {Base: domain.Base{ID: systemPartID}},
		},
	}

	svc := facility.NewFieldDeviceService(
		fieldDeviceRepo,
		spsSystemTypeRepo,
		systemTypeRepo,
		apparatRepo,
		systemPartRepo,
		nil,
		nil,
		nil,
		nil,
		nil,
	)

	updates := []domainFacility.BulkFieldDeviceUpdate{
		{ID: fd1ID, ApparatNr: intPtr(2)},
	}

	// Act
	result := svc.BulkUpdate(context.Background(), updates)

	// Assert
	if result.FailureCount != 1 {
		t.Fatalf("expected 1 failure, got %d", result.FailureCount)
	}
	if result.Results[0].Fields["fielddevice.apparat_nr"] == "" {
		t.Fatalf("expected fielddevice.apparat_nr error, got %+v", result.Results[0].Fields)
	}
	if fieldDeviceRepo.items[fd1ID].ApparatNr != 1 {
		t.Fatalf("expected apparat_nr unchanged, got %d", fieldDeviceRepo.items[fd1ID].ApparatNr)
	}
}

func TestFieldDeviceService_BulkUpdate_PartialUpdate_ApparatNr_DifferentSystemPart_Allows(t *testing.T) {
	// Arrange
	fd1ID := uuid.New()
	fd2ID := uuid.New()
	apparatID := uuid.New()
	systemPartID1 := uuid.New()
	systemPartID2 := uuid.New()
	spsSystemTypeID := uuid.New()
	systemTypeID := uuid.New()

	fieldDeviceRepo := &fakeFieldDeviceStore{
		items: map[uuid.UUID]*domainFacility.FieldDevice{
			fd1ID: newFieldDevice(fd1ID, spsSystemTypeID, apparatID, systemPartID1, 1),
			fd2ID: newFieldDevice(fd2ID, spsSystemTypeID, apparatID, systemPartID2, 2),
		},
	}
	spsSystemTypeRepo := &fakeSpsControllerSystemTypeRepo{
		items: map[uuid.UUID]*domainFacility.SPSControllerSystemType{
			spsSystemTypeID: {
				Base:         domain.Base{ID: spsSystemTypeID},
				SystemTypeID: systemTypeID,
			},
		},
	}
	systemTypeRepo := &fakeSystemTypeRepo{
		items: map[uuid.UUID]*domainFacility.SystemType{
			systemTypeID: {Base: domain.Base{ID: systemTypeID}},
		},
	}
	apparatRepo := &fakeApparatRepo{
		items: map[uuid.UUID]*domainFacility.Apparat{
			apparatID: {Base: domain.Base{ID: apparatID}},
		},
	}
	systemPartRepo := &fakeSystemPartRepo{
		items: map[uuid.UUID]*domainFacility.SystemPart{
			systemPartID1: {Base: domain.Base{ID: systemPartID1}},
			systemPartID2: {Base: domain.Base{ID: systemPartID2}},
		},
	}

	svc := facility.NewFieldDeviceService(
		fieldDeviceRepo,
		spsSystemTypeRepo,
		systemTypeRepo,
		apparatRepo,
		systemPartRepo,
		nil,
		nil,
		nil,
		nil,
		nil,
	)

	updates := []domainFacility.BulkFieldDeviceUpdate{
		{ID: fd1ID, ApparatNr: intPtr(2)},
	}

	// Act
	result := svc.BulkUpdate(context.Background(), updates)

	// Assert
	if result.FailureCount != 0 {
		t.Fatalf("expected 0 failures, got %d", result.FailureCount)
	}
	if fieldDeviceRepo.items[fd1ID].ApparatNr != 2 {
		t.Fatalf("expected apparat_nr to update to 2, got %d", fieldDeviceRepo.items[fd1ID].ApparatNr)
	}
}

func TestFieldDeviceService_BulkUpdate_TextIndividuellOnly_Succeeds(t *testing.T) {
	fd1ID := uuid.New()
	fd2ID := uuid.New()
	apparatID := uuid.New()
	systemPartID := uuid.New()
	spsSystemTypeID := uuid.New()
	systemTypeID := uuid.New()

	fieldDeviceRepo := &fakeFieldDeviceStore{
		items: map[uuid.UUID]*domainFacility.FieldDevice{
			fd1ID: newFieldDevice(fd1ID, spsSystemTypeID, apparatID, systemPartID, 1),
			fd2ID: newFieldDevice(fd2ID, spsSystemTypeID, apparatID, systemPartID, 2),
		},
	}
	spsSystemTypeRepo := &fakeSpsControllerSystemTypeRepo{
		items: map[uuid.UUID]*domainFacility.SPSControllerSystemType{
			spsSystemTypeID: {
				Base:         domain.Base{ID: spsSystemTypeID},
				SystemTypeID: systemTypeID,
			},
		},
	}
	systemTypeRepo := &fakeSystemTypeRepo{
		items: map[uuid.UUID]*domainFacility.SystemType{
			systemTypeID: {Base: domain.Base{ID: systemTypeID}},
		},
	}
	apparatRepo := &fakeApparatRepo{
		items: map[uuid.UUID]*domainFacility.Apparat{
			apparatID: {Base: domain.Base{ID: apparatID}},
		},
	}
	systemPartRepo := &fakeSystemPartRepo{
		items: map[uuid.UUID]*domainFacility.SystemPart{
			systemPartID: {Base: domain.Base{ID: systemPartID}},
		},
	}

	svc := facility.NewFieldDeviceService(
		fieldDeviceRepo,
		spsSystemTypeRepo,
		systemTypeRepo,
		apparatRepo,
		systemPartRepo,
		nil,
		nil,
		nil,
		nil,
		nil,
	)

	updates := []domainFacility.BulkFieldDeviceUpdate{
		{ID: fd1ID, TextIndividuell: stringPtr("klj")},
		{ID: fd2ID, TextIndividuell: stringPtr("kj")},
	}

	result := svc.BulkUpdate(context.Background(), updates)

	if result.FailureCount != 0 {
		t.Fatalf("expected 0 failures, got %d (results=%+v)", result.FailureCount, result.Results)
	}
	if result.SuccessCount != 2 {
		t.Fatalf("expected 2 successes, got %d", result.SuccessCount)
	}

	if fieldDeviceRepo.items[fd1ID].TextIndividuell == nil || *fieldDeviceRepo.items[fd1ID].TextIndividuell != "klj" {
		t.Fatalf("expected fd1 text_fix=klj, got %+v", fieldDeviceRepo.items[fd1ID].TextIndividuell)
	}
	if fieldDeviceRepo.items[fd2ID].TextIndividuell == nil || *fieldDeviceRepo.items[fd2ID].TextIndividuell != "kj" {
		t.Fatalf("expected fd2 text_fix=kj, got %+v", fieldDeviceRepo.items[fd2ID].TextIndividuell)
	}
}

func TestFieldDeviceService_BulkUpdate_AllowsClearingOptionalTextFields(t *testing.T) {
	fdID := uuid.New()
	apparatID := uuid.New()
	systemPartID := uuid.New()
	spsSystemTypeID := uuid.New()
	systemTypeID := uuid.New()

	device := newFieldDevice(fdID, spsSystemTypeID, apparatID, systemPartID, 1)
	device.BMK = stringPtr("BMK-1")
	device.Description = stringPtr("Description")
	device.TextIndividuell = stringPtr("TextFix")

	fieldDeviceRepo := &fakeFieldDeviceStore{
		items: map[uuid.UUID]*domainFacility.FieldDevice{
			fdID: device,
		},
	}
	spsSystemTypeRepo := &fakeSpsControllerSystemTypeRepo{
		items: map[uuid.UUID]*domainFacility.SPSControllerSystemType{
			spsSystemTypeID: {
				Base:         domain.Base{ID: spsSystemTypeID},
				SystemTypeID: systemTypeID,
			},
		},
	}
	systemTypeRepo := &fakeSystemTypeRepo{
		items: map[uuid.UUID]*domainFacility.SystemType{
			systemTypeID: {Base: domain.Base{ID: systemTypeID}},
		},
	}
	apparatRepo := &fakeApparatRepo{
		items: map[uuid.UUID]*domainFacility.Apparat{
			apparatID: {Base: domain.Base{ID: apparatID}},
		},
	}
	systemPartRepo := &fakeSystemPartRepo{
		items: map[uuid.UUID]*domainFacility.SystemPart{
			systemPartID: {Base: domain.Base{ID: systemPartID}},
		},
	}

	svc := facility.NewFieldDeviceService(
		fieldDeviceRepo,
		spsSystemTypeRepo,
		systemTypeRepo,
		apparatRepo,
		systemPartRepo,
		nil,
		nil,
		nil,
		nil,
		nil,
	)

	result := svc.BulkUpdate(context.Background(), []domainFacility.BulkFieldDeviceUpdate{
		{
			ID:                 fdID,
			HasBMK:             true,
			HasDescription:     true,
			HasTextIndividuell: true,
		},
	})

	if result.FailureCount != 0 {
		t.Fatalf("expected 0 failures, got %d (results=%+v)", result.FailureCount, result.Results)
	}

	updated := fieldDeviceRepo.items[fdID]
	if updated.BMK != nil {
		t.Fatalf("expected bmk to be cleared, got %+v", updated.BMK)
	}
	if updated.Description != nil {
		t.Fatalf("expected description to be cleared, got %+v", updated.Description)
	}
	if updated.TextIndividuell != nil {
		t.Fatalf("expected text_fix to be cleared, got %+v", updated.TextIndividuell)
	}
}

func TestFieldDeviceService_BulkUpdate_ClearsExistingSpecificationFields(t *testing.T) {
	fdID := uuid.New()
	apparatID := uuid.New()
	systemPartID := uuid.New()
	spsSystemTypeID := uuid.New()
	systemTypeID := uuid.New()
	specID := uuid.New()

	device := newFieldDevice(fdID, spsSystemTypeID, apparatID, systemPartID, 1)
	device.SpecificationID = &specID

	specStore := &fakeSpecificationStore{
		items: map[uuid.UUID]*domainFacility.Specification{
			specID: {
				Base:                  domain.Base{ID: specID},
				FieldDeviceID:         &fdID,
				SpecificationSupplier: stringPtr("Supplier"),
				AdditionalInfoSize:    intPtr(12),
			},
		},
	}

	fieldDeviceRepo := &fakeFieldDeviceStore{
		items: map[uuid.UUID]*domainFacility.FieldDevice{
			fdID: device,
		},
	}
	spsSystemTypeRepo := &fakeSpsControllerSystemTypeRepo{
		items: map[uuid.UUID]*domainFacility.SPSControllerSystemType{
			spsSystemTypeID: {
				Base:         domain.Base{ID: spsSystemTypeID},
				SystemTypeID: systemTypeID,
			},
		},
	}
	systemTypeRepo := &fakeSystemTypeRepo{
		items: map[uuid.UUID]*domainFacility.SystemType{
			systemTypeID: {Base: domain.Base{ID: systemTypeID}},
		},
	}
	apparatRepo := &fakeApparatRepo{
		items: map[uuid.UUID]*domainFacility.Apparat{
			apparatID: {Base: domain.Base{ID: apparatID}},
		},
	}
	systemPartRepo := &fakeSystemPartRepo{
		items: map[uuid.UUID]*domainFacility.SystemPart{
			systemPartID: {Base: domain.Base{ID: systemPartID}},
		},
	}

	svc := facility.NewFieldDeviceService(
		fieldDeviceRepo,
		spsSystemTypeRepo,
		systemTypeRepo,
		apparatRepo,
		systemPartRepo,
		specStore,
		nil,
		nil,
		nil,
		nil,
	)

	result := svc.BulkUpdate(context.Background(), []domainFacility.BulkFieldDeviceUpdate{
		{
			ID: fdID,
			Specification: &domainFacility.SpecificationPatch{
				HasSpecificationSupplier: true,
				HasAdditionalInfoSize:    true,
			},
		},
	})

	if result.FailureCount != 0 {
		t.Fatalf("expected 0 failures, got %d (results=%+v)", result.FailureCount, result.Results)
	}

	updatedSpec := specStore.items[specID]
	if updatedSpec.SpecificationSupplier != nil {
		t.Fatalf("expected specification_supplier to be cleared, got %+v", updatedSpec.SpecificationSupplier)
	}
	if updatedSpec.AdditionalInfoSize != nil {
		t.Fatalf("expected additional_info_size to be cleared, got %+v", updatedSpec.AdditionalInfoSize)
	}
}

func TestFieldDeviceService_BulkUpdate_ClearOnlyPatchDoesNotCreateEmptySpecification(t *testing.T) {
	fdID := uuid.New()
	apparatID := uuid.New()
	systemPartID := uuid.New()
	spsSystemTypeID := uuid.New()
	systemTypeID := uuid.New()

	fieldDeviceRepo := &fakeFieldDeviceStore{
		items: map[uuid.UUID]*domainFacility.FieldDevice{
			fdID: newFieldDevice(fdID, spsSystemTypeID, apparatID, systemPartID, 1),
		},
	}
	spsSystemTypeRepo := &fakeSpsControllerSystemTypeRepo{
		items: map[uuid.UUID]*domainFacility.SPSControllerSystemType{
			spsSystemTypeID: {
				Base:         domain.Base{ID: spsSystemTypeID},
				SystemTypeID: systemTypeID,
			},
		},
	}
	systemTypeRepo := &fakeSystemTypeRepo{
		items: map[uuid.UUID]*domainFacility.SystemType{
			systemTypeID: {Base: domain.Base{ID: systemTypeID}},
		},
	}
	apparatRepo := &fakeApparatRepo{
		items: map[uuid.UUID]*domainFacility.Apparat{
			apparatID: {Base: domain.Base{ID: apparatID}},
		},
	}
	systemPartRepo := &fakeSystemPartRepo{
		items: map[uuid.UUID]*domainFacility.SystemPart{
			systemPartID: {Base: domain.Base{ID: systemPartID}},
		},
	}
	specStore := &fakeSpecificationStore{items: map[uuid.UUID]*domainFacility.Specification{}}

	svc := facility.NewFieldDeviceService(
		fieldDeviceRepo,
		spsSystemTypeRepo,
		systemTypeRepo,
		apparatRepo,
		systemPartRepo,
		specStore,
		nil,
		nil,
		nil,
		nil,
	)

	result := svc.BulkUpdate(context.Background(), []domainFacility.BulkFieldDeviceUpdate{
		{
			ID: fdID,
			Specification: &domainFacility.SpecificationPatch{
				HasSpecificationSupplier: true,
			},
		},
	})

	if result.FailureCount != 0 {
		t.Fatalf("expected 0 failures, got %d (results=%+v)", result.FailureCount, result.Results)
	}
	if len(specStore.items) != 0 {
		t.Fatalf("expected no specification to be created, got %d item(s)", len(specStore.items))
	}
}
