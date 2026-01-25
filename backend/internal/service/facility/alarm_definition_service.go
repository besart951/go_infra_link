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

func (s *AlarmDefinitionService) List(page, limit int, search string) (*domain.PaginatedList[domainFacility.AlarmDefinition], error) {
	return s.repo.GetPaginatedList(domain.PaginationParams{
		Page:   page,
		Limit:  limit,
		Search: search,
	})
}

func (s *AlarmDefinitionService) GetByID(id uuid.UUID) (*domainFacility.AlarmDefinition, error) {
	items, err := s.repo.GetByIds([]uuid.UUID{id})
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, domain.ErrNotFound
	}
	return items[0], nil
}
