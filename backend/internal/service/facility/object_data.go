package facility

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
)

func (s *Service) GetObjectDataByIds(ids []uuid.UUID) ([]*domainFacility.ObjectData, error) {
	return s.repo.GetObjectDataByIds(ids)
}

func (s *Service) CreateObjectData(entity *domainFacility.ObjectData) error {
	return s.repo.CreateObjectData(entity)
}

func (s *Service) UpdateObjectData(entity *domainFacility.ObjectData) error {
	return s.repo.UpdateObjectData(entity)
}

func (s *Service) DeleteObjectDataByIds(ids []uuid.UUID) error {
	return s.repo.DeleteObjectDataByIds(ids)
}

func (s *Service) GetPaginatedObjectData(params domain.PaginationParams) (*domain.PaginatedList[domainFacility.ObjectData], error) {
	return s.repo.GetPaginatedObjectData(params)
}
