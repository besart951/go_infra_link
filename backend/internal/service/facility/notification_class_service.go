package facility

import (
	"context"

	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
)

type NotificationClassService struct {
	baseService[domainFacility.NotificationClass]
}

func NewNotificationClassService(repo domainFacility.NotificationClassRepository) *NotificationClassService {
	return &NotificationClassService{baseService: newBase[domainFacility.NotificationClass](repo, 10)}
}

func (s *NotificationClassService) Create(ctx context.Context, nc *domainFacility.NotificationClass) error {
	return s.repo.Create(ctx, nc)
}

func (s *NotificationClassService) Update(ctx context.Context, nc *domainFacility.NotificationClass) error {
	return s.repo.Update(ctx, nc)
}
