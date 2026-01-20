package facility

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
)

func (s *Service) GetApparatByIds(ids []uuid.UUID) ([]*domainFacility.Apparat, error) {
	return s.Apparats.GetByIds(ids)
}

func (s *Service) CreateApparat(entity *domainFacility.Apparat) error {
	return s.Apparats.Create(entity)
}

func (s *Service) UpdateApparat(entity *domainFacility.Apparat) error {
	return s.Apparats.Update(entity)
}

func (s *Service) DeleteApparatByIds(ids []uuid.UUID) error {
	return s.Apparats.DeleteByIds(ids)
}

func (s *Service) GetPaginatedApparats(params domain.PaginationParams) (*domain.PaginatedList[domainFacility.Apparat], error) {
	return s.Apparats.GetPaginatedList(params)
}
