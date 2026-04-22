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
	svc := NewApparatService(repo)

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
	svc := NewApparatService(repo)

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
	svc := NewApparatService(repo)

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

type apparatServiceRepoFake struct {
	items                   map[uuid.UUID]*domainFacility.Apparat
	existsShortName         bool
	existsName              bool
	existsShortNameSequence []bool
	existsNameSequence      []bool
	createErr               error
	updateErr               error
	createCalls             int
	updateCalls             int
	existsShortNameCalls    int
	existsNameCalls         int
	getPaginatedListCalls   int
	lastShortNameExcludeID  *uuid.UUID
	lastNameExcludeID       *uuid.UUID
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

func cloneUUIDPtr(id *uuid.UUID) *uuid.UUID {
	if id == nil {
		return nil
	}
	clone := *id
	return &clone
}

var _ domainFacility.ApparatRepository = (*apparatServiceRepoFake)(nil)

func TestApparatService_Update_UsesExcludeIDForExactExistsQueries(t *testing.T) {
	itemID := uuid.New()
	repo := &apparatServiceRepoFake{}
	svc := NewApparatService(repo)

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
	svc := NewApparatService(repo)

	err := svc.Create(context.Background(), &domainFacility.Apparat{
		Base:      domain.Base{ID: uuid.New()},
		ShortName: "PMP",
		Name:      "Pump",
	})
	if err == nil || err.Error() != "boom" {
		t.Fatalf("expected original error, got %v", err)
	}
}
