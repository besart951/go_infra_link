package facility

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
)

func (s *Service) GetStateTextByIds(ids []uuid.UUID) ([]*domainFacility.StateText, error) {
	return s.StateTexts.GetByIds(ids)
}

func (s *Service) CreateStateText(entity *domainFacility.StateText) error {
	return s.StateTexts.Create(entity)
}

func (s *Service) UpdateStateText(entity *domainFacility.StateText) error {
	return s.StateTexts.Update(entity)
}

func (s *Service) DeleteStateTextByIds(ids []uuid.UUID) error {
	return s.StateTexts.DeleteByIds(ids)
}

func (s *Service) GetPaginatedStateTexts(params domain.PaginationParams) (*domain.PaginatedList[domainFacility.StateText], error) {
	return s.StateTexts.GetPaginatedList(params)
}
