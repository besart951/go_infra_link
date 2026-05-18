package facility

import (
	"context"

	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
)

type StateTextService struct {
	baseService[domainFacility.StateText]
	deleteGuard bacnetReferenceDeleteGuard
}

func NewStateTextService(repo domainFacility.StateTextRepository, usageRepos ...domainFacility.BacnetReferenceUsageRepository) *StateTextService {
	return &StateTextService{
		baseService: newBase(repo, 10),
		deleteGuard: newBacnetReferenceDeleteGuard(domainFacility.BacnetReferenceResourceStateText, usageRepos...),
	}
}

func (s *StateTextService) Create(ctx context.Context, stateText *domainFacility.StateText) error {
	return s.repo.Create(ctx, stateText)
}

func (s *StateTextService) Update(ctx context.Context, stateText *domainFacility.StateText) error {
	return s.repo.Update(ctx, stateText)
}

func (s *StateTextService) DeleteByID(ctx context.Context, id uuid.UUID) error {
	if err := s.deleteGuard.ensureDeleteAllowed(ctx, id); err != nil {
		return err
	}
	return s.baseService.DeleteByID(ctx, id)
}
