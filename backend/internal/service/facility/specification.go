package facility

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
)

func (s *Service) GetSpecificationByIds(ids []uuid.UUID) ([]*domainFacility.Specification, error) {
	return s.repo.GetSpecificationByIds(ids)
}

func (s *Service) CreateSpecification(entity *domainFacility.Specification) error {
	return s.repo.CreateSpecification(entity)
}

func (s *Service) UpdateSpecification(entity *domainFacility.Specification) error {
	return s.repo.UpdateSpecification(entity)
}

func (s *Service) DeleteSpecificationByIds(ids []uuid.UUID) error {
	return s.repo.DeleteSpecificationByIds(ids)
}

func (s *Service) GetPaginatedSpecifications(params domain.PaginationParams) (*domain.PaginatedList[domainFacility.Specification], error) {
	return s.repo.GetPaginatedSpecifications(params)
}
