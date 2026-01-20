package facility

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
)

func (s *Service) GetNotificationClassByIds(ids []uuid.UUID) ([]*domainFacility.NotificationClass, error) {
	return s.NotificationClasses.GetByIds(ids)
}

func (s *Service) CreateNotificationClass(entity *domainFacility.NotificationClass) error {
	return s.NotificationClasses.Create(entity)
}

func (s *Service) UpdateNotificationClass(entity *domainFacility.NotificationClass) error {
	return s.NotificationClasses.Update(entity)
}

func (s *Service) DeleteNotificationClassByIds(ids []uuid.UUID) error {
	return s.NotificationClasses.DeleteByIds(ids)
}

func (s *Service) GetPaginatedNotificationClasses(params domain.PaginationParams) (*domain.PaginatedList[domainFacility.NotificationClass], error) {
	return s.NotificationClasses.GetPaginatedList(params)
}
