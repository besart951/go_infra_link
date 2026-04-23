package facility

import (
	"context"

	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
)

type AlarmDefinitionService struct {
	baseService[domainFacility.AlarmDefinition]
}

func NewAlarmDefinitionService(repo domainFacility.AlarmDefinitionRepository) *AlarmDefinitionService {
	return &AlarmDefinitionService{baseService: newBase(repo, 10)}
}

func (s *AlarmDefinitionService) Create(ctx context.Context, ad *domainFacility.AlarmDefinition) error {
	return s.repo.Create(ctx, ad)
}

func (s *AlarmDefinitionService) Update(ctx context.Context, ad *domainFacility.AlarmDefinition) error {
	return s.repo.Update(ctx, ad)
}
