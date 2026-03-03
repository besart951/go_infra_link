package facility

import (
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
)

type StateTextService struct {
	baseService[domainFacility.StateText]
}

func NewStateTextService(repo domainFacility.StateTextRepository) *StateTextService {
	return &StateTextService{baseService: newBase[domainFacility.StateText](repo, 10)}
}

func (s *StateTextService) Create(stateText *domainFacility.StateText) error {
	return s.repo.Create(stateText)
}

func (s *StateTextService) Update(stateText *domainFacility.StateText) error {
	return s.repo.Update(stateText)
}
