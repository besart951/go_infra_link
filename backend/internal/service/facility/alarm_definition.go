package facility

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
)

func (s *Service) GetAlarmDefinitionByIds(ids []uuid.UUID) ([]*domainFacility.AlarmDefinition, error) {
	return s.repo.GetAlarmDefinitionByIds(ids)
}

func (s *Service) CreateAlarmDefinition(entity *domainFacility.AlarmDefinition) error {
	return s.repo.CreateAlarmDefinition(entity)
}

func (s *Service) UpdateAlarmDefinition(entity *domainFacility.AlarmDefinition) error {
	return s.repo.UpdateAlarmDefinition(entity)
}

func (s *Service) DeleteAlarmDefinitionByIds(ids []uuid.UUID) error {
	return s.repo.DeleteAlarmDefinitionByIds(ids)
}

func (s *Service) GetPaginatedAlarmDefinitions(params domain.PaginationParams) (*domain.PaginatedList[domainFacility.AlarmDefinition], error) {
	return s.repo.GetPaginatedAlarmDefinitions(params)
}
