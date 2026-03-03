package facility

import (
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
)

type NotificationClassService struct {
	baseService[domainFacility.NotificationClass]
}

func NewNotificationClassService(repo domainFacility.NotificationClassRepository) *NotificationClassService {
	return &NotificationClassService{baseService: newBase[domainFacility.NotificationClass](repo, 10)}
}

func (s *NotificationClassService) Create(nc *domainFacility.NotificationClass) error {
	return s.repo.Create(nc)
}

func (s *NotificationClassService) Update(nc *domainFacility.NotificationClass) error {
	return s.repo.Update(nc)
}
