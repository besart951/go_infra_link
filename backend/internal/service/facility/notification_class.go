package facility

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
)

func (s *Service) GetNotificationClassByIds(ids []uuid.UUID) ([]*domainFacility.NotificationClass, error) {
	return s.repo.GetNotificationClassByIds(ids)
}

func (s *Service) CreateNotificationClass(entity *domainFacility.NotificationClass) error {
	return s.repo.CreateNotificationClass(entity)
}

func (s *Service) UpdateNotificationClass(entity *domainFacility.NotificationClass) error {
	return s.repo.UpdateNotificationClass(entity)
}

func (s *Service) DeleteNotificationClassByIds(ids []uuid.UUID) error {
	return s.repo.DeleteNotificationClassByIds(ids)
}

func (s *Service) GetPaginatedNotificationClasses(params domain.PaginationParams) (*domain.PaginatedList[domainFacility.NotificationClass], error) {
	return s.repo.GetPaginatedNotificationClasses(params)
}
