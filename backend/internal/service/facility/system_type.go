package facility

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
)

func (s *Service) GetSystemTypeByIds(ids []uuid.UUID) ([]*domainFacility.SystemType, error) {
	return s.repo.GetSystemTypeByIds(ids)
}

func (s *Service) CreateSystemType(entity *domainFacility.SystemType) error {
	return s.repo.CreateSystemType(entity)
}

func (s *Service) UpdateSystemType(entity *domainFacility.SystemType) error {
	return s.repo.UpdateSystemType(entity)
}

func (s *Service) DeleteSystemTypeByIds(ids []uuid.UUID) error {
	return s.repo.DeleteSystemTypeByIds(ids)
}

func (s *Service) GetPaginatedSystemTypes(params domain.PaginationParams) (*domain.PaginatedList[domainFacility.SystemType], error) {
	return s.repo.GetPaginatedSystemTypes(params)
}
