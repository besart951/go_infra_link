package facility

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
)

func (s *Service) GetBacnetObjectByIds(ids []uuid.UUID) ([]*domainFacility.BacnetObject, error) {
	return s.repo.GetBacnetObjectByIds(ids)
}

func (s *Service) CreateBacnetObject(entity *domainFacility.BacnetObject) error {
	return s.repo.CreateBacnetObject(entity)
}

func (s *Service) UpdateBacnetObject(entity *domainFacility.BacnetObject) error {
	return s.repo.UpdateBacnetObject(entity)
}

func (s *Service) DeleteBacnetObjectByIds(ids []uuid.UUID) error {
	return s.repo.DeleteBacnetObjectByIds(ids)
}

func (s *Service) GetPaginatedBacnetObjects(params domain.PaginationParams) (*domain.PaginatedList[domainFacility.BacnetObject], error) {
	return s.repo.GetPaginatedBacnetObjects(params)
}
