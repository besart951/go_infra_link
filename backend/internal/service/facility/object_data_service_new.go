package facility

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
)

type ObjectDataService struct {
	repo domainFacility.ObjectDataStore
}

func NewObjectDataService(repo domainFacility.ObjectDataStore) *ObjectDataService {
	return &ObjectDataService{repo: repo}
}

func (s *ObjectDataService) List(page, limit int, search string) (*domain.PaginatedList[domainFacility.ObjectData], error) {
	return s.repo.GetPaginatedList(domain.PaginationParams{
		Page:   page,
		Limit:  limit,
		Search: search,
	})
}

func (s *ObjectDataService) GetByID(id uuid.UUID) (*domainFacility.ObjectData, error) {
	items, err := s.repo.GetByIds([]uuid.UUID{id})
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, domain.ErrNotFound
	}
	return items[0], nil
}
