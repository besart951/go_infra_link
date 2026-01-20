package facility

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
)

func (s *Service) GetAlarmDefinitionByIds(ids []uuid.UUID) ([]*domainFacility.AlarmDefinition, error) {
	return s.AlarmDefinitions.GetByIds(ids)
}

func (s *Service) CreateAlarmDefinition(entity *domainFacility.AlarmDefinition) error {
	return s.AlarmDefinitions.Create(entity)
}

func (s *Service) UpdateAlarmDefinition(entity *domainFacility.AlarmDefinition) error {
	return s.AlarmDefinitions.Update(entity)
}

func (s *Service) DeleteAlarmDefinitionByIds(ids []uuid.UUID) error {
	return s.AlarmDefinitions.DeleteByIds(ids)
}

func (s *Service) GetPaginatedAlarmDefinitions(params domain.PaginationParams) (*domain.PaginatedList[domainFacility.AlarmDefinition], error) {
	return s.AlarmDefinitions.GetPaginatedList(params)
}
