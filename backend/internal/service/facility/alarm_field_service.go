package facility

import (
	"strings"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
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
	repo domainFacility.AlarmFieldRepository
}

func NewAlarmFieldService(repo domainFacility.AlarmFieldRepository) *AlarmFieldService {
	return &AlarmFieldService{repo: repo}
}

func (s *AlarmFieldService) Create(field *domainFacility.AlarmField) error {
	if err := validateAlarmField(field); err != nil {
		return err
	}
	return s.repo.Create(field)
}

func (s *AlarmFieldService) List(page, limit int, search string) (*domain.PaginatedList[domainFacility.AlarmField], error) {
	page, limit = domain.NormalizePagination(page, limit, 20)
	return s.repo.GetPaginatedList(domain.PaginationParams{
		Page:   page,
		Limit:  limit,
		Search: search,
	})
}

func (s *AlarmFieldService) GetByID(id uuid.UUID) (*domainFacility.AlarmField, error) {
	return domain.GetByID(s.repo, id)
}

func (s *AlarmFieldService) Update(field *domainFacility.AlarmField) error {
	if err := validateAlarmField(field); err != nil {
		return err
	}
	return s.repo.Update(field)
}

func (s *AlarmFieldService) DeleteByID(id uuid.UUID) error {
	return s.repo.DeleteByIds([]uuid.UUID{id})
}

func validateAlarmField(field *domainFacility.AlarmField) error {
	ve := domain.NewValidationError()
	if strings.TrimSpace(field.Key) == "" {
		ve.Add("key", "required")
	}
	if strings.TrimSpace(field.Label) == "" {
		ve.Add("label", "required")
	}
	if strings.TrimSpace(field.DataType) == "" {
		ve.Add("data_type", "required")
	} else {
		if _, ok := alarmFieldDataTypes[field.DataType]; !ok {
			ve.Add("data_type", "invalid")
		}
	}
	if len(ve.Fields) > 0 {
		return ve
	}
	return nil
}
