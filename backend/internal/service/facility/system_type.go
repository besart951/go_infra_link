package facility

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
)

func (s *Service) GetSystemTypeByIds(ids []uuid.UUID) ([]*domainFacility.SystemType, error) {
	return s.SystemTypes.GetByIds(ids)
}

func (s *Service) CreateSystemType(entity *domainFacility.SystemType) error {
	return s.SystemTypes.Create(entity)
}

func (s *Service) UpdateSystemType(entity *domainFacility.SystemType) error {
	return s.SystemTypes.Update(entity)
}

func (s *Service) DeleteSystemTypeByIds(ids []uuid.UUID) error {
	return s.SystemTypes.DeleteByIds(ids)
}

func (s *Service) GetPaginatedSystemTypes(params domain.PaginationParams) (*domain.PaginatedList[domainFacility.SystemType], error) {
	return s.SystemTypes.GetPaginatedList(params)
}
