package project

import (
	"encoding/json"
	"testing"
	"time"

	domainProject "github.com/besart951/go_infra_link/backend/internal/domain/project"
	dto "github.com/besart951/go_infra_link/backend/internal/handler/dto/project"
	"github.com/google/uuid"
)

func TestApplyProjectUpdatePreservesOmittedFields(t *testing.T) {
	startDate := time.Date(2026, 4, 22, 8, 0, 0, 0, time.UTC)
	phaseID := uuid.New()
	project := &domainProject.Project{
		Name:        "Existing",
		Description: "Existing description",
		Status:      domainProject.StatusPlanned,
		StartDate:   &startDate,
		PhaseID:     phaseID,
	}

	ApplyProjectUpdate(project, dto.UpdateProjectRequest{})

	if project.Name != "Existing" {
		t.Fatalf("expected name to stay unchanged, got %q", project.Name)
	}
	if project.Description != "Existing description" {
		t.Fatalf("expected description to stay unchanged, got %q", project.Description)
	}
	if project.Status != domainProject.StatusPlanned {
		t.Fatalf("expected status to stay unchanged, got %q", project.Status)
	}
	if project.StartDate == nil || !project.StartDate.Equal(startDate) {
		t.Fatalf("expected start date to stay unchanged, got %+v", project.StartDate)
	}
	if project.PhaseID != phaseID {
		t.Fatalf("expected phase id to stay unchanged, got %s", project.PhaseID)
	}
}

func TestApplyProjectUpdateAppliesExplicitEmptyAndNullValues(t *testing.T) {
	startDate := time.Date(2026, 4, 22, 8, 0, 0, 0, time.UTC)
	project := &domainProject.Project{
		Name:        "Existing",
		Description: "Existing description",
		Status:      domainProject.StatusPlanned,
		StartDate:   &startDate,
		PhaseID:     uuid.New(),
	}

	var req dto.UpdateProjectRequest
	if err := json.Unmarshal([]byte(`{"name":"","description":"","start_date":null}`), &req); err != nil {
		t.Fatalf("expected update request to decode, got %v", err)
	}
	ApplyProjectUpdate(project, req)

	if project.Name != "" {
		t.Fatalf("expected name to be cleared, got %q", project.Name)
	}
	if project.Description != "" {
		t.Fatalf("expected description to be cleared, got %q", project.Description)
	}
	if project.StartDate != nil {
		t.Fatalf("expected start date to be cleared, got %+v", project.StartDate)
	}
	if project.Status != domainProject.StatusPlanned {
		t.Fatalf("expected omitted status to stay unchanged, got %q", project.Status)
	}
}

func TestApplyProjectUpdateReplacesPresentValues(t *testing.T) {
	project := &domainProject.Project{
		Name:        "Existing",
		Description: "Existing description",
		Status:      domainProject.StatusPlanned,
		PhaseID:     uuid.New(),
	}
	newPhaseID := uuid.New()

	var req dto.UpdateProjectRequest
	if err := json.Unmarshal([]byte(`{
		"name":"Replacement",
		"description":"Replacement description",
		"status":"ongoing",
		"start_date":"2026-04-23",
		"phase_id":"`+newPhaseID.String()+`"
	}`), &req); err != nil {
		t.Fatalf("expected update request to decode, got %v", err)
	}
	ApplyProjectUpdate(project, req)

	if project.Name != "Replacement" {
		t.Fatalf("expected replaced name, got %q", project.Name)
	}
	if project.Description != "Replacement description" {
		t.Fatalf("expected replaced description, got %q", project.Description)
	}
	if project.Status != domainProject.StatusOngoing {
		t.Fatalf("expected replaced status, got %q", project.Status)
	}
	if project.StartDate == nil {
		t.Fatal("expected start date to be set")
	}
	if project.StartDate.Format("2006-01-02") != "2026-04-23" {
		t.Fatalf("expected replaced start date, got %s", project.StartDate.Format(time.RFC3339))
	}
	if project.PhaseID != newPhaseID {
		t.Fatalf("expected replaced phase id %s, got %s", newPhaseID, project.PhaseID)
	}
}
