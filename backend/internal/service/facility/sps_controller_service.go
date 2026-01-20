package facility

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
)

type SPSControllerService struct {
	repo domainFacility.SPSControllerRepository
}

func NewSPSControllerService(repo domainFacility.SPSControllerRepository) *SPSControllerService {
	return &SPSControllerService{repo: repo}
}

func (s *SPSControllerService) Create(spsController *domainFacility.SPSController) error {
	return s.repo.Create(spsController)
}

func (s *SPSControllerService) GetById(id uuid.UUID) (*domainFacility.SPSController, error) {
	spsControllers, err := s.repo.GetByIds([]uuid.UUID{id})
	if err != nil {
		return nil, err
	}
	if len(spsControllers) == 0 {
		return nil, nil
	}
	return spsControllers[0], nil
}

func (s *SPSControllerService) List(page, limit int, search string) (*domain.PaginatedList[domainFacility.SPSController], error) {
	page, limit = normalizePagination(page, limit)
	return s.repo.GetPaginatedList(domain.PaginationParams{
		Page:   page,
		Limit:  limit,
		Search: search,
	})
}

func (s *SPSControllerService) Update(spsController *domainFacility.SPSController) error {
	return s.repo.Update(spsController)
}

func (s *SPSControllerService) DeleteByIds(ids []uuid.UUID) error {
	return s.repo.DeleteByIds(ids)
}
