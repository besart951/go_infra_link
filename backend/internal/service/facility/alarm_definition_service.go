package facility

import (
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
)

type AlarmDefinitionService struct {
	baseService[domainFacility.AlarmDefinition]
}

func NewAlarmDefinitionService(repo domainFacility.AlarmDefinitionRepository) *AlarmDefinitionService {
	return &AlarmDefinitionService{baseService: newBase[domainFacility.AlarmDefinition](repo, 10)}
}

func (s *AlarmDefinitionService) Create(ad *domainFacility.AlarmDefinition) error {
	return s.repo.Create(ad)
}

func (s *AlarmDefinitionService) Update(ad *domainFacility.AlarmDefinition) error {
	return s.repo.Update(ad)
}
