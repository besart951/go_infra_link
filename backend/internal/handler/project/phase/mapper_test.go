package phase

import (
	"encoding/json"
	"testing"

	domainProject "github.com/besart951/go_infra_link/backend/internal/domain/project"
	dto "github.com/besart951/go_infra_link/backend/internal/handler/dto/project"
)

func TestApplyUpdatePreservesOmittedFields(t *testing.T) {
	phase := &domainProject.Phase{Name: "Existing"}

	applyUpdate(phase, dto.UpdatePhaseRequest{})

	if phase.Name != "Existing" {
		t.Fatalf("expected name to stay unchanged, got %q", phase.Name)
	}
}

func TestApplyUpdateAppliesExplicitEmptyValue(t *testing.T) {
	phase := &domainProject.Phase{Name: "Existing"}

	var req dto.UpdatePhaseRequest
	if err := json.Unmarshal([]byte(`{"name":""}`), &req); err != nil {
		t.Fatalf("expected update request to decode, got %v", err)
	}
	applyUpdate(phase, req)

	if phase.Name != "" {
		t.Fatalf("expected name to be cleared, got %q", phase.Name)
	}
}

func TestApplyUpdateReplacesPresentValue(t *testing.T) {
	phase := &domainProject.Phase{Name: "Existing"}

	replacement := "Replacement"
	applyUpdate(phase, dto.UpdatePhaseRequest{Name: &replacement})

	if phase.Name != replacement {
		t.Fatalf("expected name to be replaced, got %q", phase.Name)
	}
}
