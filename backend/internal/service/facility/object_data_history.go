package facility

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
)

func (s *Service) GetObjectDataHistoryByIds(ids []uuid.UUID) ([]*domainFacility.ObjectDataHistory, error) {
	return s.ObjectDataHistory.GetByIds(ids)
}

func (s *Service) CreateObjectDataHistory(entity *domainFacility.ObjectDataHistory) error {
	return s.ObjectDataHistory.Create(entity)
}

func (s *Service) GetPaginatedObjectDataHistory(params domain.PaginationParams) (*domain.PaginatedList[domainFacility.ObjectDataHistory], error) {
	return s.ObjectDataHistory.GetPaginatedList(params)
}
