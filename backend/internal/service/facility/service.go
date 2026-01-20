package facility

import (
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
)

// Service is a thin layer over the facility repository.
// It mirrors the repository capabilities but gives you a stable place
// to add validation, business rules, auth checks, etc.
type Service struct {
	repo domainFacility.Repository
}

func New(repo domainFacility.Repository) *Service {
	return &Service{repo: repo}
}
