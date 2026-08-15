package facility

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	apprealtime "github.com/besart951/go_infra_link/backend/internal/application/realtime"
	"github.com/google/uuid"
)

type copyJobProgressPublisherFake struct {
	mu      sync.Mutex
	events  []apprealtime.CopyJobProgressEvent
	updates chan apprealtime.CopyJobProgressEvent
}

func newCopyJobProgressPublisherFake() *copyJobProgressPublisherFake {
	return &copyJobProgressPublisherFake{updates: make(chan apprealtime.CopyJobProgressEvent, 16)}
}

func (f *copyJobProgressPublisherFake) BroadcastCopyJobProgress(_ context.Context, event apprealtime.CopyJobProgressEvent) {
	f.mu.Lock()
	f.events = append(f.events, event)
	f.mu.Unlock()

	select {
	case f.updates <- event:
	default:
	}
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

func (f *copyJobProgressPublisherFake) waitForTerminal(t *testing.T, jobID uuid.UUID) apprealtime.CopyJobProgressEvent {
	t.Helper()
	timeout := time.NewTimer(time.Second)
	defer timeout.Stop()

	for {
		select {
		case event := <-f.updates:
			if event.JobID == jobID && (event.Status == string(CopyJobStatusCompleted) || event.Status == string(CopyJobStatusFailed)) {
				return event
			}
		case <-timeout.C:
			t.Fatal("copy job did not reach a terminal state")
		}
	}
}

func TestCopyJobManagerReportsProgressAndPreventsDuplicateStart(t *testing.T) {
	publisher := newCopyJobProgressPublisherFake()
	manager := NewCopyJobManager(publisher)
	t.Cleanup(manager.Close)
	ownerID := uuid.New()
	operationID := uuid.New()
	started := make(chan struct{})
	finish := make(chan struct{})
	var finishOnce sync.Once
	finishCopy := func() { finishOnce.Do(func() { close(finish) }) }
	t.Cleanup(finishCopy)

	job, err := manager.Start(ownerID, operationID, CopyJobKindSPSController, func(ctx context.Context) error {
		reportCopyProgress(ctx, 47, copyJobStageCopyingSystemTypes)
		close(started)
		<-finish
		return nil
	})
	if err != nil {
		t.Fatalf("start copy job: %v", err)
	}
	if job.Status != CopyJobStatusQueued || job.Progress != 0 {
		t.Fatalf("expected queued job at 0%%, got %#v", job)
	}

	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("copy job did not start")
	}
	duplicate, err := manager.Start(ownerID, operationID, CopyJobKindSPSController, func(context.Context) error {
		t.Error("an existing operation ID must not start a second copy")
		return nil
	})
	if err != nil {
		t.Fatalf("start duplicate copy job: %v", err)
	}
	if duplicate.ID != operationID || duplicate.Progress != 47 {
		t.Fatalf("expected the existing job, got %#v", duplicate)
	}

	otherOperation, err := manager.Start(ownerID, uuid.New(), CopyJobKindControlCabinet, func(context.Context) error {
		t.Error("an active job must prevent a second copy for the same user")
		return nil
	})
	if err != nil {
		t.Fatalf("start competing copy job: %v", err)
	}
	if otherOperation.ID != operationID {
		t.Fatalf("expected active job %s, got %s", operationID, otherOperation.ID)
	}

	finishCopy()
	completed := publisher.waitForTerminal(t, operationID)
	if completed.Status != string(CopyJobStatusCompleted) || completed.Progress != 100 {
		t.Fatalf("expected completed job at 100%%, got %#v", completed)
	}

	if got := publisher.progressValues(); !containsProgress(got, 0) || !containsProgress(got, 47) || !containsProgress(got, 100) {
		t.Fatalf("expected publish sequence to include 0, 47 and 100, got %v", got)
	}
}

func TestCopyJobManagerScopesIdempotencyKeysToTheirOwner(t *testing.T) {
	manager := NewCopyJobManager(nil)
	t.Cleanup(manager.Close)
	operationID := uuid.New()
	firstOwnerID := uuid.New()
	secondOwnerID := uuid.New()
	firstFinished := make(chan struct{})
	secondFinished := make(chan struct{})
	var finishedOnce sync.Once
	finishCopies := func() {
		finishedOnce.Do(func() {
			close(firstFinished)
			close(secondFinished)
		})
	}
	t.Cleanup(finishCopies)

	first, err := manager.Start(firstOwnerID, operationID, CopyJobKindControlCabinet, func(context.Context) error {
		<-firstFinished
		return nil
	})
	if err != nil {
		t.Fatalf("start first copy job: %v", err)
	}
	second, err := manager.Start(secondOwnerID, operationID, CopyJobKindSPSController, func(context.Context) error {
		<-secondFinished
		return nil
	})
	if err != nil {
		t.Fatalf("start second copy job: %v", err)
	}

	if first.ID != operationID || second.ID != operationID || first.OwnerID == second.OwnerID {
		t.Fatalf("jobs = %#v, %#v; want separate jobs with the same operation ID", first, second)
	}
	if _, err := manager.Get(firstOwnerID, operationID); err != nil {
		t.Fatalf("get first owner job: %v", err)
	}
	if _, err := manager.Get(secondOwnerID, operationID); err != nil {
		t.Fatalf("get second owner job: %v", err)
	}

	finishCopies()
	manager.Close()
}

func TestCopyJobManagerCloseCancelsActiveWorkAndRejectsNewJobs(t *testing.T) {
	manager := NewCopyJobManager(nil)
	ownerID := uuid.New()
	operationID := uuid.New()
	started := make(chan struct{})
	canceled := make(chan struct{})

	_, err := manager.Start(ownerID, operationID, CopyJobKindControlCabinet, func(ctx context.Context) error {
		close(started)
		<-ctx.Done()
		close(canceled)
		return ctx.Err()
	})
	if err != nil {
		t.Fatalf("start copy job: %v", err)
	}
	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("copy job did not start")
	}

	closed := make(chan struct{})
	go func() {
		manager.Close()
		close(closed)
	}()

	select {
	case <-canceled:
	case <-time.After(time.Second):
		t.Fatal("copy job did not receive shutdown cancellation")
	}
	select {
	case <-closed:
	case <-time.After(time.Second):
		t.Fatal("copy job manager did not wait for the worker")
	}

	if _, err := manager.Start(ownerID, uuid.New(), CopyJobKindControlCabinet, func(context.Context) error { return nil }); !errors.Is(err, errCopyJobManagerClosed) {
		t.Fatalf("start after close error = %v, want %v", err, errCopyJobManagerClosed)
	}
}

func containsProgress(values []int, expected int) bool {
	for _, value := range values {
		if value == expected {
			return true
		}
	}
	return false
}
