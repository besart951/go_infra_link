package facility

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
)

func (s *Service) GetSystemPartByIds(ids []uuid.UUID) ([]*domainFacility.SystemPart, error) {
	return s.SystemParts.GetByIds(ids)
}

func (s *Service) CreateSystemPart(entity *domainFacility.SystemPart) error {
	return s.SystemParts.Create(entity)
}

func (s *Service) UpdateSystemPart(entity *domainFacility.SystemPart) error {
	return s.SystemParts.Update(entity)
}

func (s *Service) DeleteSystemPartByIds(ids []uuid.UUID) error {
	return s.SystemParts.DeleteByIds(ids)
}

func (s *Service) GetPaginatedSystemParts(params domain.PaginationParams) (*domain.PaginatedList[domainFacility.SystemPart], error) {
	return s.SystemParts.GetPaginatedList(params)
}
