package facilitycache

import (
	"context"
	"testing"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
)

func TestReferenceDataCache_CachesApparatReadsAndClonesResults(t *testing.T) {
	ctx := context.Background()
	apparatID := uuid.New()
	systemPartID := uuid.New()
	desc := "cached"

	apparatRepo := &fakeApparatRepository{
		items: map[uuid.UUID]*domainFacility.Apparat{
			apparatID: {
				Base:        domain.Base{ID: apparatID},
				ShortName:   "PMP",
				Name:        "Pump",
				Description: &desc,
				SystemParts: []*domainFacility.SystemPart{
					{Base: domain.Base{ID: systemPartID}, ShortName: "AIR", Name: "Air"},
				},
			},
		},
	}
	systemPartRepo := &fakeSystemPartRepository{}
	cachedApparats, _ := WrapReferenceData(apparatRepo, systemPartRepo)

	first, err := cachedApparats.GetByIds(ctx, []uuid.UUID{apparatID})
	if err != nil {
		t.Fatalf("first read failed: %v", err)
	}
	first[0].Name = "mutated"
	first[0].SystemParts[0].Name = "mutated"

	second, err := cachedApparats.GetByIds(ctx, []uuid.UUID{apparatID})
	if err != nil {
		t.Fatalf("second read failed: %v", err)
	}

	if apparatRepo.getByIDsCalls != 1 {
		t.Fatalf("expected one backing apparat read, got %d", apparatRepo.getByIDsCalls)
	}
	if second[0].Name != "Pump" {
		t.Fatalf("expected cached apparat clone to keep original name, got %q", second[0].Name)
	}
	if second[0].SystemParts[0].Name != "Air" {
		t.Fatalf("expected cached system part clone to keep original name, got %q", second[0].SystemParts[0].Name)
	}
}

func TestReferenceDataCache_ApparatWriteInvalidatesApparatAndSystemPartReads(t *testing.T) {
	ctx := context.Background()
	apparatID := uuid.New()
	systemPartID := uuid.New()

	apparatRepo := &fakeApparatRepository{
		items: map[uuid.UUID]*domainFacility.Apparat{
			apparatID: {Base: domain.Base{ID: apparatID}, ShortName: "PMP", Name: "Pump"},
		},
	}
	systemPartRepo := &fakeSystemPartRepository{
		items: map[uuid.UUID]*domainFacility.SystemPart{
			systemPartID: {Base: domain.Base{ID: systemPartID}, ShortName: "AIR", Name: "Air"},
		},
	}
	cachedApparats, cachedSystemParts := WrapReferenceData(apparatRepo, systemPartRepo)

	if _, err := cachedApparats.GetByIds(ctx, []uuid.UUID{apparatID}); err != nil {
		t.Fatalf("apparat read failed: %v", err)
	}
	if _, err := cachedSystemParts.GetByIds(ctx, []uuid.UUID{systemPartID}); err != nil {
		t.Fatalf("system part read failed: %v", err)
	}
	if err := cachedApparats.Update(ctx, &domainFacility.Apparat{Base: domain.Base{ID: apparatID}, ShortName: "PMP", Name: "Pump updated"}); err != nil {
		t.Fatalf("apparat update failed: %v", err)
	}
	if _, err := cachedApparats.GetByIds(ctx, []uuid.UUID{apparatID}); err != nil {
		t.Fatalf("apparat read after update failed: %v", err)
	}
	if _, err := cachedSystemParts.GetByIds(ctx, []uuid.UUID{systemPartID}); err != nil {
		t.Fatalf("system part read after apparat update failed: %v", err)
	}

	if apparatRepo.getByIDsCalls != 2 {
		t.Fatalf("expected apparat cache invalidation, got %d backing reads", apparatRepo.getByIDsCalls)
	}
	if systemPartRepo.getByIDsCalls != 2 {
		t.Fatalf("expected system part cache invalidation, got %d backing reads", systemPartRepo.getByIDsCalls)
	}
}

func TestReferenceDataCache_CachesPaginatedSystemPartReads(t *testing.T) {
	ctx := context.Background()
	systemPartID := uuid.New()
	systemPartRepo := &fakeSystemPartRepository{
		page: &domain.PaginatedList[domainFacility.SystemPart]{
			Items:      []domainFacility.SystemPart{{Base: domain.Base{ID: systemPartID}, ShortName: "AIR", Name: "Air"}},
			Total:      1,
			Page:       1,
			TotalPages: 1,
		},
	}
	_, cachedSystemParts := WrapReferenceData(&fakeApparatRepository{}, systemPartRepo)
	params := domain.PaginationParams{Page: 1, Limit: 10, Search: ""}

	first, err := cachedSystemParts.GetPaginatedList(ctx, params)
	if err != nil {
		t.Fatalf("first page read failed: %v", err)
	}
	first.Items[0].Name = "mutated"
	second, err := cachedSystemParts.GetPaginatedList(ctx, params)
	if err != nil {
		t.Fatalf("second page read failed: %v", err)
	}

	if systemPartRepo.pageCalls != 1 {
		t.Fatalf("expected one backing page read, got %d", systemPartRepo.pageCalls)
	}
	if second.Items[0].Name != "Air" {
		t.Fatalf("expected cached page clone to keep original name, got %q", second.Items[0].Name)
	}
}

type fakeApparatRepository struct {
	items         map[uuid.UUID]*domainFacility.Apparat
	page          *domain.PaginatedList[domainFacility.Apparat]
	filterPage    *domain.PaginatedList[domainFacility.Apparat]
	getByIDsCalls int
	pageCalls     int
	filterCalls   int
}

func (r *fakeApparatRepository) GetByIds(_ context.Context, ids []uuid.UUID) ([]*domainFacility.Apparat, error) {
	r.getByIDsCalls++
	out := make([]*domainFacility.Apparat, 0, len(ids))
	for _, id := range ids {
		if item, ok := r.items[id]; ok {
			out = append(out, cloneApparat(item))
		}
	}
	return out, nil
}

func (r *fakeApparatRepository) Create(_ context.Context, entity *domainFacility.Apparat) error {
	if r.items == nil {
		r.items = make(map[uuid.UUID]*domainFacility.Apparat)
	}
	r.items[entity.ID] = cloneApparat(entity)
	return nil
}

func (r *fakeApparatRepository) Update(_ context.Context, entity *domainFacility.Apparat) error {
	if r.items == nil {
		r.items = make(map[uuid.UUID]*domainFacility.Apparat)
	}
	r.items[entity.ID] = cloneApparat(entity)
	return nil
}

func (r *fakeApparatRepository) DeleteByIds(_ context.Context, ids []uuid.UUID) error {
	for _, id := range ids {
		delete(r.items, id)
	}
	return nil
}

func (r *fakeApparatRepository) GetPaginatedList(context.Context, domain.PaginationParams) (*domain.PaginatedList[domainFacility.Apparat], error) {
	r.pageCalls++
	return cloneApparatPage(r.page), nil
}

func (r *fakeApparatRepository) ExistsShortName(context.Context, string, *uuid.UUID) (bool, error) {
	return false, nil
}

func (r *fakeApparatRepository) ExistsName(context.Context, string, *uuid.UUID) (bool, error) {
	return false, nil
}

func (r *fakeApparatRepository) GetPaginatedListWithFilters(context.Context, domain.PaginationParams, domainFacility.ApparatFilterParams) (*domain.PaginatedList[domainFacility.Apparat], error) {
	r.filterCalls++
	return cloneApparatPage(r.filterPage), nil
}

type fakeSystemPartRepository struct {
	items         map[uuid.UUID]*domainFacility.SystemPart
	page          *domain.PaginatedList[domainFacility.SystemPart]
	getByIDsCalls int
	pageCalls     int
}

func (r *fakeSystemPartRepository) GetByIds(_ context.Context, ids []uuid.UUID) ([]*domainFacility.SystemPart, error) {
	r.getByIDsCalls++
	out := make([]*domainFacility.SystemPart, 0, len(ids))
	for _, id := range ids {
		if item, ok := r.items[id]; ok {
			out = append(out, cloneSystemPart(item))
		}
	}
	return out, nil
}

func (r *fakeSystemPartRepository) Create(_ context.Context, entity *domainFacility.SystemPart) error {
	if r.items == nil {
		r.items = make(map[uuid.UUID]*domainFacility.SystemPart)
	}
	r.items[entity.ID] = cloneSystemPart(entity)
	return nil
}

func (r *fakeSystemPartRepository) Update(_ context.Context, entity *domainFacility.SystemPart) error {
	if r.items == nil {
		r.items = make(map[uuid.UUID]*domainFacility.SystemPart)
	}
	r.items[entity.ID] = cloneSystemPart(entity)
	return nil
}

func (r *fakeSystemPartRepository) DeleteByIds(_ context.Context, ids []uuid.UUID) error {
	for _, id := range ids {
		delete(r.items, id)
	}
	return nil
}

func (r *fakeSystemPartRepository) GetPaginatedList(context.Context, domain.PaginationParams) (*domain.PaginatedList[domainFacility.SystemPart], error) {
	r.pageCalls++
	return cloneSystemPartPage(r.page), nil
}

func (r *fakeSystemPartRepository) ExistsShortName(context.Context, string, *uuid.UUID) (bool, error) {
	return false, nil
}

func (r *fakeSystemPartRepository) ExistsName(context.Context, string, *uuid.UUID) (bool, error) {
	return false, nil
}
