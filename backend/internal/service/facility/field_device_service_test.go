package facility_test

import (
	"testing"
	"time"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/besart951/go_infra_link/backend/internal/service/facility"
	"github.com/google/uuid"
)

type fakeFieldDeviceStore struct {
	items map[uuid.UUID]*domainFacility.FieldDevice
}

func (r *fakeFieldDeviceStore) GetByIds(ids []uuid.UUID) ([]*domainFacility.FieldDevice, error) {
	out := make([]*domainFacility.FieldDevice, 0, len(ids))
	for _, id := range ids {
		if item, ok := r.items[id]; ok {
			clone := *item
			out = append(out, &clone)
		}
	}
	return out, nil
}

func (r *fakeFieldDeviceStore) Create(entity *domainFacility.FieldDevice) error {
	clone := *entity
	r.items[entity.ID] = &clone
	return nil
}

func (r *fakeFieldDeviceStore) Update(entity *domainFacility.FieldDevice) error {
	clone := *entity
	r.items[entity.ID] = &clone
	return nil
}

func (r *fakeFieldDeviceStore) DeleteByIds(ids []uuid.UUID) error {
	for _, id := range ids {
		delete(r.items, id)
	}
	return nil
}

func (r *fakeFieldDeviceStore) GetPaginatedList(params domain.PaginationParams) (*domain.PaginatedList[domainFacility.FieldDevice], error) {
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

func (r *fakeFieldDeviceStore) GetIDsBySPSControllerSystemTypeIDs(ids []uuid.UUID) ([]uuid.UUID, error) {
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
	spsControllerSystemTypeID uuid.UUID,
	systemPartID *uuid.UUID,
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
		if systemPartID != nil && item.SystemPartID != *systemPartID {
			continue
		}
		return true, nil
	}
	return false, nil
}

func (r *fakeFieldDeviceStore) GetUsedApparatNumbers(
	spsControllerSystemTypeID uuid.UUID,
	systemPartID *uuid.UUID,
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
		if systemPartID != nil && item.SystemPartID != *systemPartID {
			continue
		}
		out = append(out, item.ApparatNr)
	}
	return out, nil
}

type fakeSpsControllerSystemTypeRepo struct {
	items map[uuid.UUID]*domainFacility.SPSControllerSystemType
}

func (r *fakeSpsControllerSystemTypeRepo) GetByIds(ids []uuid.UUID) ([]*domainFacility.SPSControllerSystemType, error) {
	out := make([]*domainFacility.SPSControllerSystemType, 0, len(ids))
	for _, id := range ids {
		if item, ok := r.items[id]; ok {
			clone := *item
			out = append(out, &clone)
		}
	}
	return out, nil
}

func (r *fakeSpsControllerSystemTypeRepo) Create(entity *domainFacility.SPSControllerSystemType) error {
	clone := *entity
	r.items[entity.ID] = &clone
	return nil
}

func (r *fakeSpsControllerSystemTypeRepo) Update(entity *domainFacility.SPSControllerSystemType) error {
	clone := *entity
	r.items[entity.ID] = &clone
	return nil
}

func (r *fakeSpsControllerSystemTypeRepo) DeleteByIds(ids []uuid.UUID) error {
	for _, id := range ids {
		delete(r.items, id)
	}
	return nil
}

func (r *fakeSpsControllerSystemTypeRepo) GetPaginatedList(params domain.PaginationParams) (*domain.PaginatedList[domainFacility.SPSControllerSystemType], error) {
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

func (r *fakeSpsControllerSystemTypeRepo) GetIDsBySPSControllerIDs(ids []uuid.UUID) ([]uuid.UUID, error) {
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

func (r *fakeSpsControllerSystemTypeRepo) SoftDeleteBySPSControllerIDs(ids []uuid.UUID) error {
	if len(ids) == 0 {
		return nil
	}
	idSet := make(map[uuid.UUID]struct{}, len(ids))
	for _, id := range ids {
		idSet[id] = struct{}{}
	}
	for id, item := range r.items {
		if _, ok := idSet[item.SPSControllerID]; ok {
			clone := *item
			now := time.Now().UTC()
			clone.DeletedAt = &now
			r.items[id] = &clone
		}
	}
	return nil
}

type fakeSystemTypeRepo struct {
	items map[uuid.UUID]*domainFacility.SystemType
}

func (r *fakeSystemTypeRepo) GetByIds(ids []uuid.UUID) ([]*domainFacility.SystemType, error) {
	out := make([]*domainFacility.SystemType, 0, len(ids))
	for _, id := range ids {
		if item, ok := r.items[id]; ok {
			clone := *item
			out = append(out, &clone)
		}
	}
	return out, nil
}

func (r *fakeSystemTypeRepo) Create(entity *domainFacility.SystemType) error {
	clone := *entity
	r.items[entity.ID] = &clone
	return nil
}

func (r *fakeSystemTypeRepo) Update(entity *domainFacility.SystemType) error {
	clone := *entity
	r.items[entity.ID] = &clone
	return nil
}

func (r *fakeSystemTypeRepo) DeleteByIds(ids []uuid.UUID) error {
	for _, id := range ids {
		delete(r.items, id)
	}
	return nil
}

func (r *fakeSystemTypeRepo) GetPaginatedList(params domain.PaginationParams) (*domain.PaginatedList[domainFacility.SystemType], error) {
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

func (r *fakeSystemTypeRepo) ExistsName(name string, excludeID *uuid.UUID) (bool, error) {
	return false, nil
}

func (r *fakeSystemTypeRepo) ExistsOverlappingRange(numberMin, numberMax int, excludeID *uuid.UUID) (bool, error) {
	return false, nil
}

type fakeApparatRepo struct {
	items map[uuid.UUID]*domainFacility.Apparat
}

func (r *fakeApparatRepo) GetByIds(ids []uuid.UUID) ([]*domainFacility.Apparat, error) {
	out := make([]*domainFacility.Apparat, 0, len(ids))
	for _, id := range ids {
		if item, ok := r.items[id]; ok {
			clone := *item
			out = append(out, &clone)
		}
	}
	return out, nil
}

func (r *fakeApparatRepo) Create(entity *domainFacility.Apparat) error {
	clone := *entity
	r.items[entity.ID] = &clone
	return nil
}

func (r *fakeApparatRepo) Update(entity *domainFacility.Apparat) error {
	clone := *entity
	r.items[entity.ID] = &clone
	return nil
}

func (r *fakeApparatRepo) DeleteByIds(ids []uuid.UUID) error {
	for _, id := range ids {
		delete(r.items, id)
	}
	return nil
}

func (r *fakeApparatRepo) GetPaginatedList(params domain.PaginationParams) (*domain.PaginatedList[domainFacility.Apparat], error) {
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

type fakeSystemPartRepo struct {
	items map[uuid.UUID]*domainFacility.SystemPart
}

func (r *fakeSystemPartRepo) GetByIds(ids []uuid.UUID) ([]*domainFacility.SystemPart, error) {
	out := make([]*domainFacility.SystemPart, 0, len(ids))
	for _, id := range ids {
		if item, ok := r.items[id]; ok {
			clone := *item
			out = append(out, &clone)
		}
	}
	return out, nil
}

func (r *fakeSystemPartRepo) Create(entity *domainFacility.SystemPart) error {
	clone := *entity
	r.items[entity.ID] = &clone
	return nil
}

func (r *fakeSystemPartRepo) Update(entity *domainFacility.SystemPart) error {
	clone := *entity
	r.items[entity.ID] = &clone
	return nil
}

func (r *fakeSystemPartRepo) DeleteByIds(ids []uuid.UUID) error {
	for _, id := range ids {
		delete(r.items, id)
	}
	return nil
}

func (r *fakeSystemPartRepo) GetPaginatedList(params domain.PaginationParams) (*domain.PaginatedList[domainFacility.SystemPart], error) {
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
		nil,
		nil,
		systemTypeRepo,
		nil,
		apparatRepo,
		systemPartRepo,
		nil,
		nil,
		nil,
	)

	updates := []domainFacility.BulkFieldDeviceUpdate{
		{ID: fd1ID, ApparatNr: intPtr(2)},
		{ID: fd2ID, ApparatNr: intPtr(1)},
	}
	result := svc.BulkUpdate(updates)

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
		nil,
		nil,
		systemTypeRepo,
		nil,
		apparatRepo,
		systemPartRepo,
		nil,
		nil,
		nil,
	)

	updates := []domainFacility.BulkFieldDeviceUpdate{
		{ID: fd1ID, ApparatNr: intPtr(3)},
	}
	result := svc.BulkUpdate(updates)

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
		nil,
		nil,
		systemTypeRepo,
		nil,
		apparatRepo,
		systemPartRepo,
		nil,
		nil,
		nil,
	)

	updates := []domainFacility.BulkFieldDeviceUpdate{
		{ID: fd1ID, ApparatNr: intPtr(2)},
	}

	// Act
	result := svc.BulkUpdate(updates)

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
		nil,
		nil,
		systemTypeRepo,
		nil,
		apparatRepo,
		systemPartRepo,
		nil,
		nil,
		nil,
	)

	updates := []domainFacility.BulkFieldDeviceUpdate{
		{ID: fd1ID, ApparatNr: intPtr(2)},
	}

	// Act
	result := svc.BulkUpdate(updates)

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
		nil,
		nil,
		systemTypeRepo,
		nil,
		apparatRepo,
		systemPartRepo,
		nil,
		nil,
		nil,
	)

	updates := []domainFacility.BulkFieldDeviceUpdate{
		{ID: fd1ID, ApparatNr: intPtr(2)},
	}

	// Act
	result := svc.BulkUpdate(updates)

	// Assert
	if result.FailureCount != 0 {
		t.Fatalf("expected 0 failures, got %d", result.FailureCount)
	}
	if fieldDeviceRepo.items[fd1ID].ApparatNr != 2 {
		t.Fatalf("expected apparat_nr to update to 2, got %d", fieldDeviceRepo.items[fd1ID].ApparatNr)
	}
}
