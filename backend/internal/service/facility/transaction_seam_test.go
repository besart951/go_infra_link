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

type txSPSControllerSystemTypeRepo struct {
	*fakeSpsControllerSystemTypeRepo
	deleteCalls int
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
