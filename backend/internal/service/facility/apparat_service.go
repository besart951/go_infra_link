package facility

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
)

type ApparatService struct {
	repo domainFacility.ApparatRepository
}

func NewApparatService(repo domainFacility.ApparatRepository) *ApparatService {
	return &ApparatService{repo: repo}
}

func (s *ApparatService) Create(apparat *domainFacility.Apparat) error {
	return s.repo.Create(apparat)
}

func (s *ApparatService) GetByID(id uuid.UUID) (*domainFacility.Apparat, error) {
	apparats, err := s.repo.GetByIds([]uuid.UUID{id})
	if err != nil {
		return nil, err
	}
	if len(apparats) == 0 {
		return nil, domain.ErrNotFound
	}
	return apparats[0], nil
}

func (s *ApparatService) List(page, limit int, search string) (*domain.PaginatedList[domainFacility.Apparat], error) {
	page, limit = domain.NormalizePagination(page, limit, 10)
	return s.repo.GetPaginatedList(domain.PaginationParams{
		Page:   page,
		Limit:  limit,
		Search: search,
	})
}

func (s *ApparatService) Update(apparat *domainFacility.Apparat) error {
	return s.repo.Update(apparat)
}

func (s *ApparatService) DeleteByIds(ids []uuid.UUID) error {
	return s.repo.DeleteByIds(ids)
}
