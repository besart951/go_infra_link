package shared

import (
	"testing"
	"time"

	facilityservice "github.com/besart951/go_infra_link/backend/internal/service/facility"
	"github.com/google/uuid"
)

func TestToCopyJobResponsePreservesTransportState(t *testing.T) {
	now := time.Now().UTC()
	jobID := uuid.New()

	response := ToCopyJobResponse(facilityservice.CopyJob{
		ID:        jobID,
		Kind:      facilityservice.CopyJobKindSPSController,
		Status:    facilityservice.CopyJobStatusRunning,
		Progress:  42,
		Stage:     "copying_controllers",
		CreatedAt: now,
		UpdatedAt: now,
	})

	if response.JobID != jobID || response.Kind != "sps_controller" || response.Status != "running" {
		t.Fatalf("expected job identity and state to be preserved, got %#v", response)
	}
	if response.Progress != 42 || response.Stage != "copying_controllers" {
		t.Fatalf("expected progress transport state to be preserved, got %#v", response)
	}
}
