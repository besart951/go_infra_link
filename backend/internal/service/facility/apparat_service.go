package facility

import (
	"context"
	"errors"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
)

type ApparatService struct {
	baseService[domainFacility.Apparat]
	extRepo          domainFacility.ApparatRepository
	systemPartReader domain.Reader[domainFacility.SystemPart]
	objectDataReader domain.Reader[domainFacility.ObjectData]
}

func NewApparatService(
	repo domainFacility.ApparatRepository,
	systemPartReader domain.Reader[domainFacility.SystemPart],
	objectDataReader domain.Reader[domainFacility.ObjectData],
) *ApparatService {
	return &ApparatService{
		baseService:      newBase(repo, 10),
		extRepo:          repo,
		systemPartReader: systemPartReader,
		objectDataReader: objectDataReader,
	}
}

func (s *ApparatService) Create(ctx context.Context, apparat *domainFacility.Apparat) error {
	if err := s.Validate(ctx, apparat, nil); err != nil {
		return err
	}
	if err := s.repo.Create(ctx, apparat); err != nil {
		return s.mapWriteConflict(ctx, apparat, nil, err)
	}
	return nil
}

func (s *ApparatService) GetByIDs(ctx context.Context, ids []uuid.UUID) ([]*domainFacility.Apparat, error) {
	return s.extRepo.GetByIds(ctx, ids)
}

func (s *ApparatService) CreateWithSystemPartIDs(ctx context.Context, apparat *domainFacility.Apparat, systemPartIDs []uuid.UUID) error {
	if len(systemPartIDs) > 0 {
		systemParts, err := s.loadSystemParts(ctx, systemPartIDs)
		if err != nil {
			return err
		}
		apparat.SystemParts = systemParts
	}

	return s.Create(ctx, apparat)
}

func (s *ApparatService) ListWithFilters(ctx context.Context, params domain.PaginationParams, filters domainFacility.ApparatFilterParams) (*domain.PaginatedList[domainFacility.Apparat], error) {
	page, limit := domain.NormalizePagination(params.Page, params.Limit, s.defaultLimit)
	params.Page = page
	params.Limit = limit

	if filters.ObjectDataID == nil && filters.SystemPartID == nil {
		return s.repo.GetPaginatedList(ctx, params)
	}

	if filters.ObjectDataID != nil {
		if err := validateChecks(referenceExists(ctx, s.objectDataReader, *filters.ObjectDataID)); err != nil {
			return nil, err
		}
	}
	if filters.SystemPartID != nil {
		if err := validateChecks(referenceExists(ctx, s.systemPartReader, *filters.SystemPartID)); err != nil {
			return nil, err
		}
	}

	return s.extRepo.GetPaginatedListWithFilters(ctx, params, filters)
}

func (s *ApparatService) Update(ctx context.Context, apparat *domainFacility.Apparat) error {
	if err := s.Validate(ctx, apparat, &apparat.ID); err != nil {
		return err
	}
	if err := s.repo.Update(ctx, apparat); err != nil {
		return s.mapWriteConflict(ctx, apparat, &apparat.ID, err)
	}
	return nil
}

func (s *ApparatService) UpdateWithSystemPartIDs(ctx context.Context, apparat *domainFacility.Apparat, systemPartIDs *[]uuid.UUID) error {
	if systemPartIDs != nil {
		if len(*systemPartIDs) == 0 {
			apparat.SystemParts = []*domainFacility.SystemPart{}
		} else {
			systemParts, err := s.loadSystemParts(ctx, *systemPartIDs)
			if err != nil {
				return err
			}
			apparat.SystemParts = systemParts
		}
	}

	return s.Update(ctx, apparat)
}

func (s *ApparatService) GetSystemPartIDs(ctx context.Context, id uuid.UUID) ([]uuid.UUID, error) {
	apparat, err := s.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return extractIDs(apparat.SystemParts, func(sp *domainFacility.SystemPart) uuid.UUID { return sp.ID }), nil
}

func (s *ApparatService) loadSystemParts(ctx context.Context, ids []uuid.UUID) ([]*domainFacility.SystemPart, error) {
	uniqueIDs := uniqueUUIDs(ids)
	if len(uniqueIDs) == 0 {
		return []*domainFacility.SystemPart{}, nil
	}

	systemParts, err := s.systemPartReader.GetByIds(ctx, uniqueIDs)
	if err != nil {
		return nil, err
	}

	found := make(map[uuid.UUID]struct{}, len(systemParts))
	for _, systemPart := range systemParts {
		if systemPart != nil {
			found[systemPart.ID] = struct{}{}
		}
	}
	for _, id := range uniqueIDs {
		if _, ok := found[id]; !ok {
			return nil, domain.ErrNotFound
		}
	}

	return systemParts, nil
}

func (s *ApparatService) Validate(ctx context.Context, apparat *domainFacility.Apparat, excludeID *uuid.UUID) error {
	if err := s.validateRequiredFields(apparat); err != nil {
		return err
	}
	return s.ensureUnique(ctx, apparat, excludeID)
}

func (s *ApparatService) validateRequiredFields(apparat *domainFacility.Apparat) error {
	return validateRules(
		requiredTrimmedExact(apparatShortNameField, apparat.ShortName, 3),
		requiredTrimmed(apparatNameField, apparat.Name),
	)
}

func (s *ApparatService) ensureUnique(ctx context.Context, apparat *domainFacility.Apparat, excludeID *uuid.UUID) error {
	ve, err := s.uniqueValidationError(ctx, apparat, excludeID)
	if err != nil {
		return err
	}
	if ve != nil {
		return ve
	}
	return nil
}

func (s *ApparatService) uniqueValidationError(ctx context.Context, apparat *domainFacility.Apparat, excludeID *uuid.UUID) (*domain.ValidationError, error) {
	err := validateChecks(
		uniqueIfPresent(apparatShortNameField, apparat.ShortName, func() (bool, error) {
			return s.extRepo.ExistsShortName(ctx, apparat.ShortName, excludeID)
		}),
		uniqueIfPresent(apparatNameField, apparat.Name, func() (bool, error) {
			return s.extRepo.ExistsName(ctx, apparat.Name, excludeID)
		}),
	)
	if err == nil {
		return nil, nil
	}
	ve, ok := domain.AsValidationError(err)
	if !ok {
		return nil, err
	}
	return ve, nil
}

func (s *ApparatService) mapWriteConflict(ctx context.Context, apparat *domainFacility.Apparat, excludeID *uuid.UUID, err error) error {
	if !errors.Is(err, domain.ErrConflict) {
		return err
	}

	ve, checkErr := s.uniqueValidationError(ctx, apparat, excludeID)
	if checkErr != nil {
		return checkErr
	}
	if ve != nil {
		return ve
	}

	return err
}

func uniqueUUIDs(ids []uuid.UUID) []uuid.UUID {
	seen := make(map[uuid.UUID]struct{}, len(ids))
	unique := make([]uuid.UUID, 0, len(ids))
	for _, id := range ids {
		if id == uuid.Nil {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		unique = append(unique, id)
	}
	return unique
}
