package facility

import (
	"errors"
	"testing"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	"github.com/google/uuid"
)

func TestObjectDataActivateForProject(t *testing.T) {
	projectID := uuid.New()
	objectData := &ObjectData{}

	if err := objectData.ActivateForProject(projectID); err != nil {
		t.Fatalf("activate ObjectData: %v", err)
	}
	if objectData.ProjectID == nil || *objectData.ProjectID != projectID || !objectData.IsActive {
		t.Fatalf("unexpected activated state: %+v", objectData)
	}
}

func TestObjectDataActivateForProjectRejectsAnotherOwner(t *testing.T) {
	ownerID := uuid.New()
	objectData := &ObjectData{ProjectID: &ownerID}

	err := objectData.ActivateForProject(uuid.New())

	if !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("expected conflict, got %v", err)
	}
	if objectData.ProjectID == nil || *objectData.ProjectID != ownerID {
		t.Fatalf("owner changed after conflict: %+v", objectData)
	}
}

func TestObjectDataDeactivateForProjectRetainsOwnership(t *testing.T) {
	projectID := uuid.New()
	objectData := &ObjectData{ProjectID: &projectID, IsActive: true}

	if err := objectData.DeactivateForProject(projectID); err != nil {
		t.Fatalf("deactivate ObjectData: %v", err)
	}
	if objectData.ProjectID == nil || *objectData.ProjectID != projectID || objectData.IsActive {
		t.Fatalf("unexpected deactivated state: %+v", objectData)
	}
}

func TestObjectDataDeactivateForProjectRejectsMissingOwnership(t *testing.T) {
	objectData := &ObjectData{IsActive: true}

	err := objectData.DeactivateForProject(uuid.New())

	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected not found, got %v", err)
	}
	if !objectData.IsActive {
		t.Fatal("state changed after rejected deactivation")
	}
}
