package facility

import (
	"context"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
)

type AlarmTypeFieldService struct {
	baseService[domainFacility.AlarmTypeField]
}

func NewAlarmTypeFieldService(repo domainFacility.AlarmTypeFieldRepository) *AlarmTypeFieldService {
	return &AlarmTypeFieldService{baseService: newBase[domainFacility.AlarmTypeField](repo, 20)}
}

func (s *AlarmTypeFieldService) Create(ctx context.Context, item *domainFacility.AlarmTypeField) error {
	if err := validateAlarmTypeField(item); err != nil {
		return err
	}
	return s.repo.Create(ctx, item)
}

func (s *AlarmTypeFieldService) Update(ctx context.Context, item *domainFacility.AlarmTypeField) error {
	if err := validateAlarmTypeField(item); err != nil {
		return err
	}
	return s.repo.Update(ctx, item)
}

func validateAlarmTypeField(item *domainFacility.AlarmTypeField) error {
	ve := domain.NewValidationError()
	if item.AlarmTypeID == uuid.Nil {
		ve = ve.Add("alarm_type_id", "required")
	}
	if item.AlarmFieldID == uuid.Nil {
		ve = ve.Add("alarm_field_id", "required")
	}
	if item.DisplayOrder < 0 {
		ve = ve.Add("display_order", "invalid")
	}
	if len(ve.Fields) > 0 {
		return ve
	}
	return nil
}
