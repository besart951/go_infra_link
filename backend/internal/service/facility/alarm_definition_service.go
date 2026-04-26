package facility

import (
	"context"
	"strings"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
)

type AlarmDefinitionService struct {
	baseService[domainFacility.AlarmDefinition]
}

func NewAlarmDefinitionService(repo domainFacility.AlarmDefinitionRepository) *AlarmDefinitionService {
	return &AlarmDefinitionService{baseService: newBase(repo, 10)}
}

func (s *AlarmDefinitionService) Create(ctx context.Context, ad *domainFacility.AlarmDefinition) error {
	if err := validateAlarmDefinition(ad); err != nil {
		return err
	}
	return s.repo.Create(ctx, ad)
}

func (s *AlarmDefinitionService) Update(ctx context.Context, ad *domainFacility.AlarmDefinition) error {
	if err := validateAlarmDefinition(ad); err != nil {
		return err
	}
	return s.repo.Update(ctx, ad)
}

func validateAlarmDefinition(ad *domainFacility.AlarmDefinition) error {
	ve := domain.NewValidationError()
	if ad == nil {
		return domain.ErrInvalidArgument
	}
	if strings.TrimSpace(ad.Name) == "" {
		ve = ve.Add("alarmdefinition.name", "name is required")
	}
	if ad.AlarmTypeID == nil {
		ve = ve.Add("alarmdefinition.alarm_type_id", "alarm_type_id is required")
	}
	if len(ve.Fields) > 0 {
		return ve
	}
	return nil
}
