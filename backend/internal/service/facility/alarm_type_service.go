package facility

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
)

type AlarmTypeService struct {
	repo domainFacility.AlarmTypeRepository
}

func NewAlarmTypeService(repo domainFacility.AlarmTypeRepository) *AlarmTypeService {
	return &AlarmTypeService{repo: repo}
}

func (s *AlarmTypeService) Create(alarmType *domainFacility.AlarmType) error {
	return s.repo.Create(alarmType)
}

func (s *AlarmTypeService) List(page, limit int, search string) (*domain.PaginatedList[domainFacility.AlarmType], error) {
	page, limit = domain.NormalizePagination(page, limit, 20)
	return s.repo.GetPaginatedList(domain.PaginationParams{
		Page:   page,
		Limit:  limit,
		Search: search,
	})
}

func (s *AlarmTypeService) GetByID(id uuid.UUID) (*domainFacility.AlarmType, error) {
	return domain.GetByID(s.repo, id)
}

func (s *AlarmTypeService) GetWithFields(id uuid.UUID) (*domainFacility.AlarmType, error) {
	return s.repo.GetWithFields(id)
}

func (s *AlarmTypeService) Update(alarmType *domainFacility.AlarmType) error {
	return s.repo.Update(alarmType)
}

func (s *AlarmTypeService) DeleteByID(id uuid.UUID) error {
	return s.repo.DeleteByIds([]uuid.UUID{id})
}
