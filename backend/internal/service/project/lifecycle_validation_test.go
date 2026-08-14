package project

import (
	"context"
	"testing"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainProject "github.com/besart951/go_infra_link/backend/internal/domain/project"
	"github.com/google/uuid"
)

func TestProjectLifecycleCreateMapsMissingPhaseToField(t *testing.T) {
	services := NewServices(Dependencies{
		Projects: newProjectRepo(),
		Phases:   newProjectPhaseRepo(),
	})

	err := services.Lifecycle.Create(context.Background(), &domainProject.Project{
		Name:      "Plant upgrade",
		Status:    domainProject.StatusPlanned,
		PhaseID:   uuid.New(),
		CreatorID: uuid.New(),
	})
	validationErr, ok := domain.AsValidationError(err)
	if !ok {
		t.Fatalf("error = %T, want *domain.ValidationError", err)
	}
	if got := validationErr.Codes["project.phase_id"]; got != "invalid_reference" {
		t.Fatalf("phase_id code = %q, want invalid_reference", got)
	}
}
