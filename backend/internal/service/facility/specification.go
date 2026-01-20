package facility

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
)

func (s *Service) GetSpecificationByIds(ids []uuid.UUID) ([]*domainFacility.Specification, error) {
	return s.Specifications.GetByIds(ids)
}

func (s *Service) CreateSpecification(entity *domainFacility.Specification) error {
	return s.Specifications.Create(entity)
}

func (s *Service) UpdateSpecification(entity *domainFacility.Specification) error {
	return s.Specifications.Update(entity)
}

func (s *Service) DeleteSpecificationByIds(ids []uuid.UUID) error {
	return s.Specifications.DeleteByIds(ids)
}

func (s *Service) GetPaginatedSpecifications(params domain.PaginationParams) (*domain.PaginatedList[domainFacility.Specification], error) {
	return s.Specifications.GetPaginatedList(params)
}
