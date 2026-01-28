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

func (s *NotificationClassService) Create(notificationClass *domainFacility.NotificationClass) error {
	return s.repo.Create(notificationClass)
}

func (s *NotificationClassService) List(page, limit int, search string) (*domain.PaginatedList[domainFacility.NotificationClass], error) {
	page, limit = domain.NormalizePagination(page, limit, 10)
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

func (s *NotificationClassService) Update(notificationClass *domainFacility.NotificationClass) error {
	return s.repo.Update(notificationClass)
}

func (s *NotificationClassService) DeleteByID(id uuid.UUID) error {
	return s.repo.DeleteByIds([]uuid.UUID{id})
}
