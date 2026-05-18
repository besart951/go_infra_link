package facility

import (
	"context"

	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
)

type NotificationClassService struct {
	baseService[domainFacility.NotificationClass]
	deleteGuard bacnetReferenceDeleteGuard
}

func NewNotificationClassService(repo domainFacility.NotificationClassRepository, usageRepos ...domainFacility.BacnetReferenceUsageRepository) *NotificationClassService {
	return &NotificationClassService{
		baseService: newBase(repo, 10),
		deleteGuard: newBacnetReferenceDeleteGuard(domainFacility.BacnetReferenceResourceNotificationClass, usageRepos...),
	}
}

func (s *NotificationClassService) Create(ctx context.Context, nc *domainFacility.NotificationClass) error {
	return s.repo.Create(ctx, nc)
}

func (s *NotificationClassService) Update(ctx context.Context, nc *domainFacility.NotificationClass) error {
	return s.repo.Update(ctx, nc)
}

func (s *NotificationClassService) DeleteByID(ctx context.Context, id uuid.UUID) error {
	if err := s.deleteGuard.ensureDeleteAllowed(ctx, id); err != nil {
		return err
	}
	return s.baseService.DeleteByID(ctx, id)
}
