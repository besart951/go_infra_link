package facility

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
)

type AlarmDefinitionService struct {
	repo domainFacility.AlarmDefinitionRepository
}

func NewAlarmDefinitionService(repo domainFacility.AlarmDefinitionRepository) *AlarmDefinitionService {
	return &AlarmDefinitionService{repo: repo}
}

func (s *AlarmDefinitionService) Create(alarmDefinition *domainFacility.AlarmDefinition) error {
	return s.repo.Create(alarmDefinition)
}

func (s *AlarmDefinitionService) List(page, limit int, search string) (*domain.PaginatedList[domainFacility.AlarmDefinition], error) {
	page, limit = domain.NormalizePagination(page, limit, 10)
	return s.repo.GetPaginatedList(domain.PaginationParams{
		Page:   page,
		Limit:  limit,
		Search: search,
	})
}

func (s *AlarmDefinitionService) GetByID(id uuid.UUID) (*domainFacility.AlarmDefinition, error) {
	return domain.GetByID(s.repo, id)
}

func (s *AlarmDefinitionService) Update(alarmDefinition *domainFacility.AlarmDefinition) error {
	return s.repo.Update(alarmDefinition)
}

func (s *AlarmDefinitionService) DeleteByID(id uuid.UUID) error {
	return s.repo.DeleteByIds([]uuid.UUID{id})
}
