package facility

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
)

type StateTextService struct {
	repo domainFacility.StateTextRepository
}

func NewStateTextService(repo domainFacility.StateTextRepository) *StateTextService {
	return &StateTextService{repo: repo}
}

func (s *StateTextService) List(page, limit int, search string) (*domain.PaginatedList[domainFacility.StateText], error) {
	return s.repo.GetPaginatedList(domain.PaginationParams{
		Page:   page,
		Limit:  limit,
		Search: search,
	})
}

func (s *StateTextService) GetByID(id uuid.UUID) (*domainFacility.StateText, error) {
	items, err := s.repo.GetByIds([]uuid.UUID{id})
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, domain.ErrNotFound
	}
	return items[0], nil
}
