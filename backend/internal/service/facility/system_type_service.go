package facility

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
)

type SystemTypeService struct {
	repo domainFacility.SystemTypeRepository
}

func NewSystemTypeService(repo domainFacility.SystemTypeRepository) *SystemTypeService {
	return &SystemTypeService{repo: repo}
}

func (s *SystemTypeService) Create(systemType *domainFacility.SystemType) error {
	return s.repo.Create(systemType)
}

func (s *SystemTypeService) GetByID(id uuid.UUID) (*domainFacility.SystemType, error) {
	systemTypes, err := s.repo.GetByIds([]uuid.UUID{id})
	if err != nil {
		return nil, err
	}
	if len(systemTypes) == 0 {
		return nil, domain.ErrNotFound
	}
	return systemTypes[0], nil
}

func (s *SystemTypeService) List(page, limit int, search string) (*domain.PaginatedList[domainFacility.SystemType], error) {
	page, limit = domain.NormalizePagination(page, limit, 10)
	return s.repo.GetPaginatedList(domain.PaginationParams{
		Page:   page,
		Limit:  limit,
		Search: search,
	})
}

func (s *SystemTypeService) Update(systemType *domainFacility.SystemType) error {
	return s.repo.Update(systemType)
}

func (s *SystemTypeService) DeleteByIds(ids []uuid.UUID) error {
	return s.repo.DeleteByIds(ids)
}
