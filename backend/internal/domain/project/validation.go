package project

import (
	"strings"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	"github.com/google/uuid"
)

// Validate checks Project invariants independently of any transport or database.
func (p Project) Validate() error {
	validation := domain.NewValidationError()
	if strings.TrimSpace(p.Name) == "" {
		validation.AddCode("project.name", "required", "name is required")
	} else if len(p.Name) > 255 {
		validation.AddCode("project.name", "max", "name must be at most 255 characters")
	}
	if !p.Status.Valid() {
		validation.AddCode("project.status", "oneof", "status must be one of: planned ongoing completed")
	}
	if p.PhaseID == uuid.Nil {
		validation.AddCode("project.phase_id", "required", "phase_id is required")
	}
	if p.CreatorID == uuid.Nil {
		validation.AddCode("project.creator_id", "required", "creator_id is required")
	}
	if len(validation.Fields) == 0 {
		return nil
	}
	return validation
}

func (s ProjectStatus) Valid() bool {
	switch s {
	case StatusPlanned, StatusOngoing, StatusCompleted:
		return true
	default:
		return false
	}
}
