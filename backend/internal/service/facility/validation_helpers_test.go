package facility

import (
	"context"
	"errors"
	"testing"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
)

func TestValidateRules_AppliesSharedRequiredAndLengthPolicies(t *testing.T) {
	err := validateRules(
		requiredTrimmedExact(apparatShortNameField, "  APP1 ", 3),
		requiredTrimmedMax(systemTypeNameField, "  Alpha  ", 4),
		optionalMaxLength(fieldDeviceBMKField, new("TOO-LONG-BMK"), 10),
	)

	ve, ok := domain.AsValidationError(err)
	if !ok {
		t.Fatalf("expected validation error, got %v", err)
	}
	if ve.Fields["apparat.short_name"] != "short_name must be exactly 3 characters" {
		t.Fatalf("expected short-name exact length error, got %+v", ve.Fields)
	}
	if ve.Fields["systemtype.name"] != "name must be at most 4 characters" {
		t.Fatalf("expected name max-length error, got %+v", ve.Fields)
	}
	if ve.Fields["fielddevice.bmk"] != "bmk must be at most 10 characters" {
		t.Fatalf("expected optional max-length error, got %+v", ve.Fields)
	}
}

func TestValidateChecks_CombinesUniquenessPolicies(t *testing.T) {
	err := validateChecks(
		uniqueIfPresent(apparatShortNameField, "APP", func() (bool, error) { return true, nil }),
		uniqueWithinIfPresent(controlCabinetNumberField, buildingScope, "1A", func() (bool, error) { return true, nil }),
	)

	ve, ok := domain.AsValidationError(err)
	if !ok {
		t.Fatalf("expected validation error, got %v", err)
	}
	if ve.Fields["apparat.short_name"] != "short_name must be unique" {
		t.Fatalf("expected short-name uniqueness error, got %+v", ve.Fields)
	}
	if ve.Fields["controlcabinet.control_cabinet_nr"] != "control_cabinet_nr must be unique within the building" {
		t.Fatalf("expected scoped uniqueness error, got %+v", ve.Fields)
	}
}

func TestValidateChecks_ReferenceExistsPropagatesLookupErrors(t *testing.T) {
	err := validateChecks(referenceExists(context.Background(), validationReaderFake{}, uuid.New()))
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected not found, got %v", err)
	}
}

func TestBuildingServiceValidate_UsesSharedPolicies(t *testing.T) {
	svc := NewBuildingService(&validationBuildingRepoFake{existsIWSCodeGroup: true})

	err := svc.Validate(context.Background(), &domainFacility.Building{
		IWSCode:       " ABCD ",
		BuildingGroup: 2,
	}, nil)

	ve, ok := domain.AsValidationError(err)
	if !ok {
		t.Fatalf("expected validation error, got %v", err)
	}
	if ve.Fields["building.iws_code"] != "iws_code must be unique within the building group" {
		t.Fatalf("expected scoped uniqueness error, got %+v", ve.Fields)
	}
}

func TestControlCabinetServiceValidate_UsesSharedPolicies(t *testing.T) {
	buildingID := uuid.New()
	number := "CAB-123456789"
	svc := NewControlCabinetService(
		&validationControlCabinetRepoFake{},
		&validationBuildingRepoFake{items: map[uuid.UUID]*domainFacility.Building{
			buildingID: {Base: domain.Base{ID: buildingID}},
		}},
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
	)

	err := svc.Validate(context.Background(), &domainFacility.ControlCabinet{
		BuildingID:       buildingID,
		ControlCabinetNr: &number,
	}, nil)

	ve, ok := domain.AsValidationError(err)
	if !ok {
		t.Fatalf("expected validation error, got %v", err)
	}
	if ve.Fields["controlcabinet.control_cabinet_nr"] != "control_cabinet_nr must be at most 11 characters" {
		t.Fatalf("expected max-length error, got %+v", ve.Fields)
	}
}

type validationReaderFake struct{}

func (validationReaderFake) GetByIds(context.Context, []uuid.UUID) ([]*struct{}, error) {
	return nil, nil
}

type validationBuildingRepoFake struct {
	items              map[uuid.UUID]*domainFacility.Building
	existsIWSCodeGroup bool
}

func (f *validationBuildingRepoFake) GetByIds(_ context.Context, ids []uuid.UUID) ([]*domainFacility.Building, error) {
	out := make([]*domainFacility.Building, 0, len(ids))
	for _, id := range ids {
		if item, ok := f.items[id]; ok {
			clone := *item
			out = append(out, &clone)
		}
	}
	return out, nil
}

func (f *validationBuildingRepoFake) Create(context.Context, *domainFacility.Building) error {
	return nil
}

func (f *validationBuildingRepoFake) Update(context.Context, *domainFacility.Building) error {
	return nil
}

func (f *validationBuildingRepoFake) DeleteByIds(context.Context, []uuid.UUID) error {
	return nil
}

func (f *validationBuildingRepoFake) GetPaginatedList(context.Context, domain.PaginationParams) (*domain.PaginatedList[domainFacility.Building], error) {
	return &domain.PaginatedList[domainFacility.Building]{}, nil
}

func (f *validationBuildingRepoFake) ExistsIWSCodeGroup(context.Context, string, int, *uuid.UUID) (bool, error) {
	return f.existsIWSCodeGroup, nil
}

type validationControlCabinetRepoFake struct{}

func (f *validationControlCabinetRepoFake) GetByIds(context.Context, []uuid.UUID) ([]*domainFacility.ControlCabinet, error) {
	return nil, nil
}

func (f *validationControlCabinetRepoFake) Create(context.Context, *domainFacility.ControlCabinet) error {
	return nil
}

func (f *validationControlCabinetRepoFake) Update(context.Context, *domainFacility.ControlCabinet) error {
	return nil
}

func (f *validationControlCabinetRepoFake) DeleteByIds(context.Context, []uuid.UUID) error {
	return nil
}

func (f *validationControlCabinetRepoFake) GetPaginatedList(context.Context, domain.PaginationParams) (*domain.PaginatedList[domainFacility.ControlCabinet], error) {
	return &domain.PaginatedList[domainFacility.ControlCabinet]{}, nil
}

func (f *validationControlCabinetRepoFake) GetPaginatedListByBuildingID(context.Context, uuid.UUID, domain.PaginationParams) (*domain.PaginatedList[domainFacility.ControlCabinet], error) {
	return &domain.PaginatedList[domainFacility.ControlCabinet]{}, nil
}

func (f *validationControlCabinetRepoFake) GetIDsByBuildingID(context.Context, uuid.UUID) ([]uuid.UUID, error) {
	return nil, nil
}

func (f *validationControlCabinetRepoFake) ExistsControlCabinetNr(context.Context, uuid.UUID, string, *uuid.UUID) (bool, error) {
	return false, nil
}

var _ domainFacility.BuildingRepository = (*validationBuildingRepoFake)(nil)
var _ domainFacility.ControlCabinetRepository = (*validationControlCabinetRepoFake)(nil)
