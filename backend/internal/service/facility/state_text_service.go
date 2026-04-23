package facility

import (
	"context"

	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
)

type StateTextService struct {
	baseService[domainFacility.StateText]
}

func NewStateTextService(repo domainFacility.StateTextRepository) *StateTextService {
	return &StateTextService{baseService: newBase(repo, 10)}
}

func (s *StateTextService) Create(ctx context.Context, stateText *domainFacility.StateText) error {
	return s.repo.Create(ctx, stateText)
}

func (s *StateTextService) Update(ctx context.Context, stateText *domainFacility.StateText) error {
	return s.repo.Update(ctx, stateText)
}
