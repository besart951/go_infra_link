package facility

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
)

type ControlCabinetService struct {
	repo domainFacility.ControlCabinetRepository
}

func NewControlCabinetService(repo domainFacility.ControlCabinetRepository) *ControlCabinetService {
	return &ControlCabinetService{repo: repo}
}

func (s *ControlCabinetService) Create(controlCabinet *domainFacility.ControlCabinet) error {
	return s.repo.Create(controlCabinet)
}

func (s *ControlCabinetService) GetById(id uuid.UUID) (*domainFacility.ControlCabinet, error) {
	controlCabinets, err := s.repo.GetByIds([]uuid.UUID{id})
	if err != nil {
		return nil, err
	}
	if len(controlCabinets) == 0 {
		return nil, nil
	}
	return controlCabinets[0], nil
}

func (s *ControlCabinetService) List(page, limit int, search string) (*domain.PaginatedList[domainFacility.ControlCabinet], error) {
	page, limit = normalizePagination(page, limit)
	return s.repo.GetPaginatedList(domain.PaginationParams{
		Page:   page,
		Limit:  limit,
		Search: search,
	})
}

func (s *ControlCabinetService) Update(controlCabinet *domainFacility.ControlCabinet) error {
	return s.repo.Update(controlCabinet)
}

func (s *ControlCabinetService) DeleteByIds(ids []uuid.UUID) error {
	return s.repo.DeleteByIds(ids)
}
