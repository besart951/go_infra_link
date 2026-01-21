package facility

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
)

type SpecificationService struct {
	repo domainFacility.SpecificationStore
}

func NewSpecificationService(repo domainFacility.SpecificationStore) *SpecificationService {
	return &SpecificationService{repo: repo}
}

func (s *SpecificationService) Create(specification *domainFacility.Specification) error {
	if specification.FieldDeviceID == nil {
		return domain.ErrInvalidArgument
	}
	// Enforce 1:1 relationship: each field_device can have at most one active specification.
	existing, err := s.repo.GetByFieldDeviceIDs([]uuid.UUID{*specification.FieldDeviceID})
	if err != nil {
		return err
	}
	if len(existing) > 0 {
		return domain.ErrConflict
	}
	return s.repo.Create(specification)
}

func (s *SpecificationService) GetByID(id uuid.UUID) (*domainFacility.Specification, error) {
	specifications, err := s.repo.GetByIds([]uuid.UUID{id})
	if err != nil {
		return nil, err
	}
	if len(specifications) == 0 {
		return nil, domain.ErrNotFound
	}
	return specifications[0], nil
}

func (s *SpecificationService) List(page, limit int, search string) (*domain.PaginatedList[domainFacility.Specification], error) {
	page, limit = normalizePagination(page, limit)
	return s.repo.GetPaginatedList(domain.PaginationParams{
		Page:   page,
		Limit:  limit,
		Search: search,
	})
}

func (s *SpecificationService) Update(specification *domainFacility.Specification) error {
	return s.repo.Update(specification)
}

func (s *SpecificationService) DeleteByIds(ids []uuid.UUID) error {
	return s.repo.DeleteByIds(ids)
}
