package facility

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
)

type SystemPartService struct {
	repo domainFacility.SystemPartRepository
}

func NewSystemPartService(repo domainFacility.SystemPartRepository) *SystemPartService {
	return &SystemPartService{repo: repo}
}

func (s *SystemPartService) Create(systemPart *domainFacility.SystemPart) error {
	return s.repo.Create(systemPart)
}

func (s *SystemPartService) GetByID(id uuid.UUID) (*domainFacility.SystemPart, error) {
	systemParts, err := s.repo.GetByIds([]uuid.UUID{id})
	if err != nil {
		return nil, err
	}
	if len(systemParts) == 0 {
		return nil, domain.ErrNotFound
	}
	return systemParts[0], nil
}

func (s *SystemPartService) List(page, limit int, search string) (*domain.PaginatedList[domainFacility.SystemPart], error) {
	page, limit = domain.NormalizePagination(page, limit, 10)
	return s.repo.GetPaginatedList(domain.PaginationParams{
		Page:   page,
		Limit:  limit,
		Search: search,
	})
}

func (s *SystemPartService) Update(systemPart *domainFacility.SystemPart) error {
	return s.repo.Update(systemPart)
}

func (s *SystemPartService) DeleteByID(id uuid.UUID) error {
	return s.repo.DeleteByIds([]uuid.UUID{id})
}
