package facility

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
)

type AlarmTypeFieldService struct {
	repo domainFacility.AlarmTypeFieldRepository
}

func NewAlarmTypeFieldService(repo domainFacility.AlarmTypeFieldRepository) *AlarmTypeFieldService {
	return &AlarmTypeFieldService{repo: repo}
}

func (s *AlarmTypeFieldService) Create(item *domainFacility.AlarmTypeField) error {
	if err := validateAlarmTypeField(item); err != nil {
		return err
	}
	return s.repo.Create(item)
}

func (s *AlarmTypeFieldService) GetByID(id uuid.UUID) (*domainFacility.AlarmTypeField, error) {
	return domain.GetByID(s.repo, id)
}

func (s *AlarmTypeFieldService) Update(item *domainFacility.AlarmTypeField) error {
	if err := validateAlarmTypeField(item); err != nil {
		return err
	}
	return s.repo.Update(item)
}

func (s *AlarmTypeFieldService) DeleteByID(id uuid.UUID) error {
	return s.repo.DeleteByIds([]uuid.UUID{id})
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
