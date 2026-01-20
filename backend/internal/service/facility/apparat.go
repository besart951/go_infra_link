package facility

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
)

func (s *Service) GetApparatByIds(ids []uuid.UUID) ([]*domainFacility.Apparat, error) {
	return s.repo.GetApparatByIds(ids)
}

func (s *Service) CreateApparat(entity *domainFacility.Apparat) error {
	return s.repo.CreateApparat(entity)
}

func (s *Service) UpdateApparat(entity *domainFacility.Apparat) error {
	return s.repo.UpdateApparat(entity)
}

func (s *Service) DeleteApparatByIds(ids []uuid.UUID) error {
	return s.repo.DeleteApparatByIds(ids)
}

func (s *Service) GetPaginatedApparats(params domain.PaginationParams) (*domain.PaginatedList[domainFacility.Apparat], error) {
	return s.repo.GetPaginatedApparats(params)
}
