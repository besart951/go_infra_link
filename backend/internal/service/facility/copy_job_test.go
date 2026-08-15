package facility

import (
	"context"
	"sync"
	"testing"
	"time"

	apprealtime "github.com/besart951/go_infra_link/backend/internal/application/realtime"
	"github.com/google/uuid"
)

type copyJobProgressPublisherFake struct {
	mu     sync.Mutex
	events []apprealtime.CopyJobProgressEvent
}

func (f *copyJobProgressPublisherFake) BroadcastCopyJobProgress(_ context.Context, event apprealtime.CopyJobProgressEvent) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.events = append(f.events, event)
}

func (f *copyJobProgressPublisherFake) progressValues() []int {
	f.mu.Lock()
	defer f.mu.Unlock()
	values := make([]int, len(f.events))
	for i := range f.events {
		values[i] = f.events[i].Progress
	}
	return values
}

func TestCopyJobManagerReportsProgressAndPreventsDuplicateStart(t *testing.T) {
	publisher := &copyJobProgressPublisherFake{}
	manager := NewCopyJobManager(publisher)
	ownerID := uuid.New()
	operationID := uuid.New()
	started := make(chan struct{})
	finish := make(chan struct{})

	job := manager.Start(ownerID, operationID, CopyJobKindSPSController, func(ctx context.Context) error {
		reportCopyProgress(ctx, 47, copyJobStageCopyingSystemTypes)
		close(started)
		<-finish
		return nil
	})
	if job.Status != CopyJobStatusQueued || job.Progress != 0 {
		t.Fatalf("expected queued job at 0%%, got %#v", job)
	}

	<-started
	duplicate := manager.Start(ownerID, operationID, CopyJobKindSPSController, func(context.Context) error {
		t.Error("an existing operation ID must not start a second copy")
		return nil
	})
	if duplicate.ID != operationID || duplicate.Progress != 47 {
		t.Fatalf("expected the existing job, got %#v", duplicate)
	}

	otherOperation := manager.Start(ownerID, uuid.New(), CopyJobKindControlCabinet, func(context.Context) error {
		t.Error("an active job must prevent a second copy for the same user")
		return nil
	})
	if otherOperation.ID != operationID {
		t.Fatalf("expected active job %s, got %s", operationID, otherOperation.ID)
	}

	close(finish)
	completed := waitForCopyJob(t, manager, ownerID, operationID)
	if completed.Status != CopyJobStatusCompleted || completed.Progress != 100 {
		t.Fatalf("expected completed job at 100%%, got %#v", completed)
	}

	if got := publisher.progressValues(); !containsProgress(got, 0) || !containsProgress(got, 47) || !containsProgress(got, 100) {
		t.Fatalf("expected publish sequence to include 0, 47 and 100, got %v", got)
	}
}

func TestCopyJobManagerScopesIdempotencyKeysToTheirOwner(t *testing.T) {
	manager := NewCopyJobManager(nil)
	operationID := uuid.New()
	firstOwnerID := uuid.New()
	secondOwnerID := uuid.New()
	firstFinished := make(chan struct{})
	secondFinished := make(chan struct{})

	first := manager.Start(firstOwnerID, operationID, CopyJobKindControlCabinet, func(context.Context) error {
		<-firstFinished
		return nil
	})
	second := manager.Start(secondOwnerID, operationID, CopyJobKindSPSController, func(context.Context) error {
		<-secondFinished
		return nil
	})

	if first.ID != operationID || second.ID != operationID || first.OwnerID == second.OwnerID {
		t.Fatalf("jobs = %#v, %#v; want separate jobs with the same operation ID", first, second)
	}
	if _, err := manager.Get(firstOwnerID, operationID); err != nil {
		t.Fatalf("get first owner job: %v", err)
	}
	if _, err := manager.Get(secondOwnerID, operationID); err != nil {
		t.Fatalf("get second owner job: %v", err)
	}

	close(firstFinished)
	close(secondFinished)
}

func waitForCopyJob(t *testing.T, manager *CopyJobManager, ownerID, operationID uuid.UUID) CopyJob {
	t.Helper()
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		job, err := manager.Get(ownerID, operationID)
		if err != nil {
			t.Fatalf("get job: %v", err)
		}
		if job.IsTerminal() {
			return job
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatal("copy job did not complete")
	return CopyJob{}
}

func containsProgress(values []int, expected int) bool {
	for _, value := range values {
		if value == expected {
			return true
		}
	}
	return false
}
