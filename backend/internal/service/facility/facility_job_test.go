package facility

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"testing"
	"time"

	apprealtime "github.com/besart951/go_infra_link/backend/internal/application/realtime"
	cursorcodec "github.com/besart951/go_infra_link/backend/internal/cursor"
	"github.com/google/uuid"
)

type facilityJobPublisherFake struct {
	mu     sync.Mutex
	events []apprealtime.FacilityJobProgressEvent
}

func (f *facilityJobPublisherFake) BroadcastFacilityJobProgress(_ context.Context, event apprealtime.FacilityJobProgressEvent) {
	f.mu.Lock()
	f.events = append(f.events, event)
	f.mu.Unlock()
}

func TestFacilityJobManagerPublishesPersistedTaskProgress(t *testing.T) {
	db := openFacilityJobTestDB(t)
	publisher := &facilityJobPublisherFake{}
	manager := NewFacilityJobManagerWithDB(publisher, db)
	t.Cleanup(manager.Close)
	manager.RegisterTask("test.progress.v1", FacilityJobHandlerFunc(progressTask))

	job := newTestFacilityJob("test.progress.v1")
	if _, err := manager.SubmitTask(t.Context(), job); err != nil {
		t.Fatalf("SubmitTask() error = %v", err)
	}
	waitForFacilityJobStatus(t, manager, job.OwnerID, job.ID, FacilityJobStatusCompleted)

	publisher.mu.Lock()
	defer publisher.mu.Unlock()
	if !hasProgress(publisher.events, 47) || !hasProgress(publisher.events, 100) {
		t.Fatalf("published progress = %#v", publisher.events)
	}
}

func TestFacilityJobManagerListCursorIsOwnerScoped(t *testing.T) {
	db := openFacilityJobTestDB(t)
	manager := NewFacilityJobManagerWithDB(nil, db)
	t.Cleanup(manager.Close)
	ownerID := uuid.New()
	for index := range 4 {
		job := newTestFacilityJob("unregistered.v1")
		job.OwnerID = ownerID
		job.UpdatedAt = time.Now().UTC().Add(-time.Duration(index) * time.Minute)
		job.CreatedAt = job.UpdatedAt
		record := facilityJobRecordFromDomain(job)
		if err := db.Create(&record).Error; err != nil {
			t.Fatalf("create job: %v", err)
		}
	}

	page, err := manager.ListPage(ownerID, 2, "")
	if err != nil || len(page.Items) != 2 || page.NextCursor == "" {
		t.Fatalf("first page = %#v, error = %v", page, err)
	}
	if _, err := manager.ListPage(uuid.New(), 2, page.NextCursor); !errors.Is(err, cursorcodec.ErrInvalid) {
		t.Fatalf("cross-owner cursor error = %v", err)
	}
}

func progressTask(_ context.Context, execution FacilityJobExecution) (FacilityJobTaskResult, error) {
	execution.Reporter.Report(FacilityJobProgress{Progress: 47, Stage: "working"})
	return FacilityJobTaskResult{Result: json.RawMessage(`{"done":true}`)}, nil
}

func newTestFacilityJob(task string) FacilityJob {
	now := time.Now().UTC()
	return FacilityJob{
		ID: uuid.New(), OwnerID: uuid.New(), Kind: FacilityJobKindFieldDevice,
		Class: FacilityJobClassExport, Type: FacilityJobTypeExport, Task: task,
		Status: FacilityJobStatusQueued, Stage: facilityJobStageQueued,
		CreatedAt: now, UpdatedAt: now,
	}
}

func hasProgress(events []apprealtime.FacilityJobProgressEvent, expected int) bool {
	for _, event := range events {
		if event.Progress == expected {
			return true
		}
	}
	return false
}
