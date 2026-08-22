package wire

import (
	"context"
	"encoding/json"
	"testing"

	domainproject "github.com/besart951/go_infra_link/backend/internal/domain/project"
	facilityservice "github.com/besart951/go_infra_link/backend/internal/service/facility"
	projectservice "github.com/besart951/go_infra_link/backend/internal/service/project"
	"github.com/google/uuid"
)

func TestProjectCopyTaskResumesFromCompletedCheckpoint(t *testing.T) {
	resultID := uuid.New()
	checkpoint, err := json.Marshal(projectCopyCheckpoint{ResultID: resultID, ProjectChangeRecorded: true})
	if err != nil {
		t.Fatal(err)
	}
	payload, err := json.Marshal(domainproject.FacilityCopyJobPayload{ProjectID: uuid.New(), SourceID: uuid.New()})
	if err != nil {
		t.Fatal(err)
	}
	copyCalls := 0
	execution := projectCopyExecution{
		job: facilityservice.CopyJob{Checkpoint: checkpoint, Payload: payload},
		operation: projectCopyOperation{copy: func(context.Context, *projectservice.Services, projectCopyCommand) (uuid.UUID, error) {
			copyCalls++
			return uuid.New(), nil
		}},
		report: func(facilityservice.FacilityJobProgress) {},
	}

	result, err := (projectCopyTaskRegistrar{}).executeProjectCopy(t.Context(), execution)
	if err != nil {
		t.Fatal(err)
	}
	if copyCalls != 0 {
		t.Fatalf("copy called %d times", copyCalls)
	}
	var decoded map[string]string
	if err := json.Unmarshal(result.Result, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded["resource_id"] != resultID.String() {
		t.Fatalf("resource_id = %q", decoded["resource_id"])
	}
}
