package facility

import (
	"context"
	"errors"
	"testing"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
)

func TestApparatService_Create_DuplicateShortNameUsesExactExistsQuery(t *testing.T) {
	repo := &apparatServiceRepoFake{existsShortName: true}
	svc := newTestApparatService(repo, nil, nil)

	err := svc.Create(context.Background(), &domainFacility.Apparat{
		Base:      domain.Base{ID: uuid.New()},
		ShortName: "PMP",
		Name:      "Pump",
	})

	ve, ok := domain.AsValidationError(err)
	if !ok {
		t.Fatalf("expected validation error, got %v", err)
	}
	if ve.Fields["apparat.short_name"] != "short_name must be unique" {
		t.Fatalf("expected short name uniqueness error, got %+v", ve.Fields)
	}
	if repo.existsShortNameCalls != 1 {
		t.Fatalf("expected one short-name existence check, got %d", repo.existsShortNameCalls)
	}
	if repo.existsNameCalls != 1 {
		t.Fatalf("expected name existence check to still run, got %d", repo.existsNameCalls)
	}
	if repo.getPaginatedListCalls != 0 {
		t.Fatalf("expected no paginated uniqueness query, got %d", repo.getPaginatedListCalls)
	}
}

func TestApparatService_Create_DuplicateNameUsesExactExistsQuery(t *testing.T) {
	repo := &apparatServiceRepoFake{existsName: true}
	svc := newTestApparatService(repo, nil, nil)

	err := svc.Create(context.Background(), &domainFacility.Apparat{
		Base:      domain.Base{ID: uuid.New()},
		ShortName: "PMP",
		Name:      "Pump",
	})

	ve, ok := domain.AsValidationError(err)
	if !ok {
		t.Fatalf("expected validation error, got %v", err)
	}
	if ve.Fields["apparat.name"] != "name must be unique" {
		t.Fatalf("expected name uniqueness error, got %+v", ve.Fields)
	}
	if repo.existsShortNameCalls != 1 {
		t.Fatalf("expected one short-name existence check, got %d", repo.existsShortNameCalls)
	}
	if repo.existsNameCalls != 1 {
		t.Fatalf("expected one name existence check, got %d", repo.existsNameCalls)
	}
	if repo.getPaginatedListCalls != 0 {
		t.Fatalf("expected no paginated uniqueness query, got %d", repo.getPaginatedListCalls)
	}
}

func TestApparatService_Create_MapsRepositoryConflictIntoValidationError(t *testing.T) {
	repo := &apparatServiceRepoFake{
		createErr:               domain.ErrConflict,
		existsShortNameSequence: []bool{false, true},
	}
	svc := newTestApparatService(repo, nil, nil)

	err := svc.Create(context.Background(), &domainFacility.Apparat{
		Base:      domain.Base{ID: uuid.New()},
		ShortName: "PMP",
		Name:      "Pump",
	})

	ve, ok := domain.AsValidationError(err)
	if !ok {
		t.Fatalf("expected validation error, got %v", err)
	}
	if ve.Fields["apparat.short_name"] != "short_name must be unique" {
		t.Fatalf("expected short-name conflict to map into validation error, got %+v", ve.Fields)
	}
	if repo.getPaginatedListCalls != 0 {
		t.Fatalf("expected no paginated uniqueness query, got %d", repo.getPaginatedListCalls)
	}
	if repo.createCalls != 1 {
		t.Fatalf("expected one create call, got %d", repo.createCalls)
	}
	if repo.existsShortNameCalls != 2 {
		t.Fatalf("expected short-name existence to be rechecked after conflict, got %d", repo.existsShortNameCalls)
	}
}

func TestApparatService_Update_UsesExcludeIDForExactExistsQueries(t *testing.T) {
	itemID := uuid.New()
	repo := &apparatServiceRepoFake{}
	svc := newTestApparatService(repo, nil, nil)

	err := svc.Update(context.Background(), &domainFacility.Apparat{
		Base:      domain.Base{ID: itemID},
		ShortName: "PMP",
		Name:      "Pump",
	})
	if err != nil {
		t.Fatalf("expected update to succeed, got %v", err)
	}
	if repo.lastShortNameExcludeID == nil || *repo.lastShortNameExcludeID != itemID {
		t.Fatalf("expected short-name exclude id %s, got %v", itemID, repo.lastShortNameExcludeID)
	}
	if repo.lastNameExcludeID == nil || *repo.lastNameExcludeID != itemID {
		t.Fatalf("expected name exclude id %s, got %v", itemID, repo.lastNameExcludeID)
	}
	if repo.getPaginatedListCalls != 0 {
		t.Fatalf("expected no paginated uniqueness query, got %d", repo.getPaginatedListCalls)
	}
}

func TestApparatService_Create_PreservesNonConflictErrors(t *testing.T) {
	repo := &apparatServiceRepoFake{createErr: errors.New("boom")}
	svc := newTestApparatService(repo, nil, nil)

	err := svc.Create(context.Background(), &domainFacility.Apparat{
		Base:      domain.Base{ID: uuid.New()},
		ShortName: "PMP",
		Name:      "Pump",
	})
	if err == nil || err.Error() != "boom" {
		t.Fatalf("expected original error, got %v", err)
	}
}

func TestApparatService_CreateWithSystemPartIDs_ResolvesSystemParts(t *testing.T) {
	systemPartID := uuid.New()
	repo := &apparatServiceRepoFake{}
	systemPartReader := &apparatServiceSystemPartReaderFake{
		items: map[uuid.UUID]*domainFacility.SystemPart{
			systemPartID: {Base: domain.Base{ID: systemPartID}, ShortName: "PMP", Name: "Pump"},
		},
	}
	svc := newTestApparatService(repo, systemPartReader, nil)

	apparat := &domainFacility.Apparat{
		Base:      domain.Base{ID: uuid.New()},
		ShortName: "APP",
		Name:      "Apparat",
	}

	if err := svc.CreateWithSystemPartIDs(context.Background(), apparat, []uuid.UUID{systemPartID}); err != nil {
		t.Fatalf("expected create to succeed, got %v", err)
	}
	if repo.createCalls != 1 {
		t.Fatalf("expected one create call, got %d", repo.createCalls)
	}
	if len(apparat.SystemParts) != 1 || apparat.SystemParts[0] == nil || apparat.SystemParts[0].ID != systemPartID {
		t.Fatalf("expected resolved system part %s, got %+v", systemPartID, apparat.SystemParts)
	}
	if len(systemPartReader.lastIDs) != 1 || systemPartReader.lastIDs[0] != systemPartID {
		t.Fatalf("expected system-part reader call with %s, got %v", systemPartID, systemPartReader.lastIDs)
	}
}

func TestApparatService_CreateWithSystemPartIDs_RejectsMissingSystemParts(t *testing.T) {
	missingID := uuid.New()
	repo := &apparatServiceRepoFake{}
	svc := newTestApparatService(repo, &apparatServiceSystemPartReaderFake{}, nil)

	err := svc.CreateWithSystemPartIDs(context.Background(), &domainFacility.Apparat{
		Base:      domain.Base{ID: uuid.New()},
		ShortName: "APP",
		Name:      "Apparat",
	}, []uuid.UUID{missingID})
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected not found, got %v", err)
	}
	if repo.createCalls != 0 {
		t.Fatalf("expected create to be skipped on missing system parts, got %d", repo.createCalls)
	}
}

func TestApparatService_UpdateWithSystemPartIDs_ClearsSystemPartsOnEmptySlice(t *testing.T) {
	repo := &apparatServiceRepoFake{}
	svc := newTestApparatService(repo, nil, nil)

	apparat := &domainFacility.Apparat{
		Base:        domain.Base{ID: uuid.New()},
		ShortName:   "APP",
		Name:        "Apparat",
		SystemParts: []*domainFacility.SystemPart{{Base: domain.Base{ID: uuid.New()}, ShortName: "OLD", Name: "Old"}},
	}
	empty := []uuid.UUID{}

	if err := svc.UpdateWithSystemPartIDs(context.Background(), apparat, &empty); err != nil {
		t.Fatalf("expected update to succeed, got %v", err)
	}
	if repo.updateCalls != 1 {
		t.Fatalf("expected one update call, got %d", repo.updateCalls)
	}
	if len(apparat.SystemParts) != 0 {
		t.Fatalf("expected system parts to be cleared, got %+v", apparat.SystemParts)
	}
}

func TestApparatService_ListWithFilters_EnsuresReferencesExistAndDelegates(t *testing.T) {
	objectDataID := uuid.New()
	systemPartID := uuid.New()
	repo := &apparatServiceRepoFake{
		filteredResult: &domain.PaginatedList[domainFacility.Apparat]{
			Items:      []domainFacility.Apparat{{Base: domain.Base{ID: uuid.New()}, ShortName: "APP", Name: "Apparat"}},
			Total:      1,
			Page:       2,
			TotalPages: 3,
		},
	}
	objectDataReader := &apparatServiceObjectDataReaderFake{
		items: map[uuid.UUID]*domainFacility.ObjectData{
			objectDataID: {Base: domain.Base{ID: objectDataID}},
		},
	}
	systemPartReader := &apparatServiceSystemPartReaderFake{
		items: map[uuid.UUID]*domainFacility.SystemPart{
			systemPartID: {Base: domain.Base{ID: systemPartID}, ShortName: "PMP", Name: "Pump"},
		},
	}
	svc := newTestApparatService(repo, systemPartReader, objectDataReader)

	params := domain.PaginationParams{Page: 2, Limit: 5, Search: "pump"}
	filters := domainFacility.ApparatFilterParams{ObjectDataID: &objectDataID, SystemPartID: &systemPartID}
	result, err := svc.ListWithFilters(context.Background(), params, filters)
	if err != nil {
		t.Fatalf("expected list with filters to succeed, got %v", err)
	}
	if result.Total != 1 || result.Page != 2 || result.TotalPages != 3 {
		t.Fatalf("expected filtered result to round-trip, got %+v", result)
	}
	if repo.getPaginatedListWithFiltersCalls != 1 {
		t.Fatalf("expected filtered repo path once, got %d", repo.getPaginatedListWithFiltersCalls)
	}
	if repo.lastFilterParams.Search != "pump" || repo.lastFilterParams.Page != 2 || repo.lastFilterParams.Limit != 5 {
		t.Fatalf("expected normalized params to be forwarded, got %+v", repo.lastFilterParams)
	}
	if repo.lastFilters.ObjectDataID == nil || *repo.lastFilters.ObjectDataID != objectDataID {
		t.Fatalf("expected object data filter %s, got %+v", objectDataID, repo.lastFilters)
	}
	if repo.lastFilters.SystemPartID == nil || *repo.lastFilters.SystemPartID != systemPartID {
		t.Fatalf("expected system part filter %s, got %+v", systemPartID, repo.lastFilters)
	}
}

func TestApparatService_ListWithFilters_RejectsMissingReferencesBeforeQuery(t *testing.T) {
	objectDataID := uuid.New()
	repo := &apparatServiceRepoFake{}
	svc := newTestApparatService(repo, nil, &apparatServiceObjectDataReaderFake{})

	_, err := svc.ListWithFilters(context.Background(), domain.PaginationParams{Page: 1, Limit: 10}, domainFacility.ApparatFilterParams{ObjectDataID: &objectDataID})
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected not found, got %v", err)
	}
	if repo.getPaginatedListWithFiltersCalls != 0 {
		t.Fatalf("expected filtered repo path to be skipped, got %d", repo.getPaginatedListWithFiltersCalls)
	}
}

func newTestApparatService(
	repo domainFacility.ApparatRepository,
	systemPartReader domain.Reader[domainFacility.SystemPart],
	objectDataReader domain.Reader[domainFacility.ObjectData],
) *ApparatService {
	if systemPartReader == nil {
		systemPartReader = &apparatServiceSystemPartReaderFake{}
	}
	if objectDataReader == nil {
		objectDataReader = &apparatServiceObjectDataReaderFake{}
	}
	return NewApparatService(repo, systemPartReader, objectDataReader)
}

type apparatServiceRepoFake struct {
	items                            map[uuid.UUID]*domainFacility.Apparat
	existsShortName                  bool
	existsName                       bool
	existsShortNameSequence          []bool
	existsNameSequence               []bool
	createErr                        error
	updateErr                        error
	filteredErr                      error
	filteredResult                   *domain.PaginatedList[domainFacility.Apparat]
	createCalls                      int
	updateCalls                      int
	existsShortNameCalls             int
	existsNameCalls                  int
	getPaginatedListCalls            int
	getPaginatedListWithFiltersCalls int
	lastShortNameExcludeID           *uuid.UUID
	lastNameExcludeID                *uuid.UUID
	lastFilterParams                 domain.PaginationParams
	lastFilters                      domainFacility.ApparatFilterParams
}

func (r *apparatServiceRepoFake) GetByIds(_ context.Context, ids []uuid.UUID) ([]*domainFacility.Apparat, error) {
	items := make([]*domainFacility.Apparat, 0, len(ids))
	for _, id := range ids {
		if item, ok := r.items[id]; ok {
			clone := *item
			items = append(items, &clone)
		}
	}
	return items, nil
}

func (r *apparatServiceRepoFake) Create(_ context.Context, entity *domainFacility.Apparat) error {
	r.createCalls++
	if r.createErr != nil {
		return r.createErr
	}
	if r.items == nil {
		r.items = map[uuid.UUID]*domainFacility.Apparat{}
	}
	clone := *entity
	r.items[entity.ID] = &clone
	return nil
}

func (r *apparatServiceRepoFake) Update(_ context.Context, entity *domainFacility.Apparat) error {
	r.updateCalls++
	if r.updateErr != nil {
		return r.updateErr
	}
	if r.items == nil {
		r.items = map[uuid.UUID]*domainFacility.Apparat{}
	}
	clone := *entity
	r.items[entity.ID] = &clone
	return nil
}

func (r *apparatServiceRepoFake) DeleteByIds(_ context.Context, ids []uuid.UUID) error {
	for _, id := range ids {
		delete(r.items, id)
	}
	return nil
}

func (r *apparatServiceRepoFake) GetPaginatedList(_ context.Context, _ domain.PaginationParams) (*domain.PaginatedList[domainFacility.Apparat], error) {
	r.getPaginatedListCalls++
	return &domain.PaginatedList[domainFacility.Apparat]{}, nil
}

func (r *apparatServiceRepoFake) GetPaginatedListWithFilters(_ context.Context, params domain.PaginationParams, filters domainFacility.ApparatFilterParams) (*domain.PaginatedList[domainFacility.Apparat], error) {
	r.getPaginatedListWithFiltersCalls++
	r.lastFilterParams = params
	r.lastFilters = filters
	if r.filteredErr != nil {
		return nil, r.filteredErr
	}
	if r.filteredResult != nil {
		return r.filteredResult, nil
	}
	return &domain.PaginatedList[domainFacility.Apparat]{}, nil
}

func (r *apparatServiceRepoFake) ExistsShortName(_ context.Context, _ string, excludeID *uuid.UUID) (bool, error) {
	r.existsShortNameCalls++
	r.lastShortNameExcludeID = cloneUUIDPtr(excludeID)
	if len(r.existsShortNameSequence) > 0 {
		result := r.existsShortNameSequence[0]
		r.existsShortNameSequence = r.existsShortNameSequence[1:]
		return result, nil
	}
	return r.existsShortName, nil
}

func (r *apparatServiceRepoFake) ExistsName(_ context.Context, _ string, excludeID *uuid.UUID) (bool, error) {
	r.existsNameCalls++
	r.lastNameExcludeID = cloneUUIDPtr(excludeID)
	if len(r.existsNameSequence) > 0 {
		result := r.existsNameSequence[0]
		r.existsNameSequence = r.existsNameSequence[1:]
		return result, nil
	}
	return r.existsName, nil
}

type apparatServiceSystemPartReaderFake struct {
	items   map[uuid.UUID]*domainFacility.SystemPart
	err     error
	lastIDs []uuid.UUID
}

func (r *apparatServiceSystemPartReaderFake) GetByIds(_ context.Context, ids []uuid.UUID) ([]*domainFacility.SystemPart, error) {
	r.lastIDs = append([]uuid.UUID(nil), ids...)
	if r.err != nil {
		return nil, r.err
	}
	items := make([]*domainFacility.SystemPart, 0, len(ids))
	for _, id := range ids {
		if item, ok := r.items[id]; ok {
			clone := *item
			items = append(items, &clone)
		}
	}
	return items, nil
}

type apparatServiceObjectDataReaderFake struct {
	items   map[uuid.UUID]*domainFacility.ObjectData
	err     error
	lastIDs []uuid.UUID
}

func (r *apparatServiceObjectDataReaderFake) GetByIds(_ context.Context, ids []uuid.UUID) ([]*domainFacility.ObjectData, error) {
	r.lastIDs = append([]uuid.UUID(nil), ids...)
	if r.err != nil {
		return nil, r.err
	}
	items := make([]*domainFacility.ObjectData, 0, len(ids))
	for _, id := range ids {
		if item, ok := r.items[id]; ok {
			clone := *item
			items = append(items, &clone)
		}
	}
	return items, nil
}

func cloneUUIDPtr(id *uuid.UUID) *uuid.UUID {
	if id == nil {
		return nil
	}
	clone := *id
	return &clone
}

var _ domainFacility.ApparatRepository = (*apparatServiceRepoFake)(nil)
