package facility

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
)

type NotificationClassService struct {
	repo domainFacility.NotificationClassRepository
}

func NewNotificationClassService(repo domainFacility.NotificationClassRepository) *NotificationClassService {
	return &NotificationClassService{repo: repo}
}

func (s *NotificationClassService) List(page, limit int, search string) (*domain.PaginatedList[domainFacility.NotificationClass], error) {
	return s.repo.GetPaginatedList(domain.PaginationParams{
		Page:   page,
		Limit:  limit,
		Search: search,
	})
}

func (s *NotificationClassService) GetByID(id uuid.UUID) (*domainFacility.NotificationClass, error) {
	items, err := s.repo.GetByIds([]uuid.UUID{id})
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, domain.ErrNotFound
	}
	return items[0], nil
}
