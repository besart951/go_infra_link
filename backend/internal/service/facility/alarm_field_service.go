package facility

import (
	"context"
	"strings"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
)

var alarmFieldDataTypes = map[string]struct{}{
	"number":    {},
	"integer":   {},
	"boolean":   {},
	"string":    {},
	"enum":      {},
	"duration":  {},
	"state_map": {},
	"json":      {},
}

type AlarmFieldService struct {
	baseService[domainFacility.AlarmField]
}

func NewAlarmFieldService(repo domainFacility.AlarmFieldRepository) *AlarmFieldService {
	return &AlarmFieldService{baseService: newBase[domainFacility.AlarmField](repo, 20)}
}

func (s *AlarmFieldService) Create(ctx context.Context, field *domainFacility.AlarmField) error {
	if err := validateAlarmField(field); err != nil {
		return err
	}
	return s.repo.Create(ctx, field)
}

func (s *AlarmFieldService) Update(ctx context.Context, field *domainFacility.AlarmField) error {
	if err := validateAlarmField(field); err != nil {
		return err
	}
	return s.repo.Update(ctx, field)
}

func validateAlarmField(field *domainFacility.AlarmField) error {
	ve := domain.NewValidationError()
	if strings.TrimSpace(field.Key) == "" {
		ve = ve.Add("key", "required")
	}
	if strings.TrimSpace(field.Label) == "" {
		ve = ve.Add("label", "required")
	}
	if strings.TrimSpace(field.DataType) == "" {
		ve = ve.Add("data_type", "required")
	} else {
		if _, ok := alarmFieldDataTypes[field.DataType]; !ok {
			ve = ve.Add("data_type", "invalid")
		}
	}
	if len(ve.Fields) > 0 {
		return ve
	}
	return nil
}
