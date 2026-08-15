package facility

import (
	"context"
	"errors"
	"sync"
	"time"

	apprealtime "github.com/besart951/go_infra_link/backend/internal/application/realtime"
	"github.com/google/uuid"
)

type CopyJobKind string

const (
	CopyJobKindControlCabinet          CopyJobKind = "control_cabinet"
	CopyJobKindSPSController           CopyJobKind = "sps_controller"
	CopyJobKindSPSControllerSystemType CopyJobKind = "sps_controller_system_type"
)

type CopyJobStatus string

const (
	CopyJobStatusQueued    CopyJobStatus = "queued"
	CopyJobStatusRunning   CopyJobStatus = "running"
	CopyJobStatusCompleted CopyJobStatus = "completed"
	CopyJobStatusFailed    CopyJobStatus = "failed"
)

const (
	copyJobStageQueued              = "queued"
	copyJobStagePreparing           = "preparing"
	copyJobStageCopyingRoot         = "copying_root"
	copyJobStageCopyingControllers  = "copying_controllers"
	copyJobStageCopyingSystemTypes  = "copying_system_types"
	copyJobStageCopyingFieldDevices = "copying_field_devices"
	copyJobStageFinalizing          = "finalizing"
	copyJobStageCompleted           = "completed"
	copyJobStageFailed              = "failed"
	copyJobRetention                = 24 * time.Hour
)

var (
	ErrCopyJobNotFound      = errors.New("copy job not found")
	errCopyJobManagerClosed = errors.New("copy job manager is closed")
)

// CopyJob is an in-memory, user-scoped asynchronous facility copy. The
// operation ID is supplied by the browser, making retries after a connection
// loss idempotent.
type CopyJob struct {
	ID        uuid.UUID
	OwnerID   uuid.UUID
	Kind      CopyJobKind
	Status    CopyJobStatus
	Progress  int
	Stage     string
	Error     string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type copyJobKey struct {
	ownerID uuid.UUID
	jobID   uuid.UUID
}

func (j CopyJob) IsTerminal() bool {
	return j.Status == CopyJobStatusCompleted || j.Status == CopyJobStatusFailed
}

type CopyJobManager struct {
	mu           sync.Mutex
	jobs         map[copyJobKey]CopyJob
	activeByUser map[uuid.UUID]uuid.UUID
	publisher    apprealtime.CopyJobProgressPublisher
	ctx          context.Context
	cancel       context.CancelFunc
	wg           sync.WaitGroup
	closeOnce    sync.Once
	closed       bool
}

func NewCopyJobManager(publisher apprealtime.CopyJobProgressPublisher) *CopyJobManager {
	ctx, cancel := context.WithCancel(context.Background())
	return &CopyJobManager{
		jobs:         make(map[copyJobKey]CopyJob),
		activeByUser: make(map[uuid.UUID]uuid.UUID),
		publisher:    publisher,
		ctx:          ctx,
		cancel:       cancel,
	}
}

// Start creates a job or returns the user's active job. Only one hierarchy
// copy runs for a user at a time, protecting generated numbers from parallel
// duplicate clicks and allowing reconnects to recover the exact job.
func (m *CopyJobManager) Start(
	ownerID uuid.UUID,
	operationID uuid.UUID,
	kind CopyJobKind,
	work func(context.Context) error,
) (CopyJob, error) {
	key := copyJobKey{ownerID: ownerID, jobID: operationID}
	m.mu.Lock()
	if m.closed {
		m.mu.Unlock()
		return CopyJob{}, errCopyJobManagerClosed
	}
	m.pruneCompletedLocked(time.Now().UTC())
	if existing, ok := m.jobs[key]; ok {
		m.mu.Unlock()
		return existing, nil
	}
	if activeID, ok := m.activeByUser[ownerID]; ok {
		if active, exists := m.jobs[copyJobKey{ownerID: ownerID, jobID: activeID}]; exists {
			m.mu.Unlock()
			return active, nil
		}
		delete(m.activeByUser, ownerID)
	}

	now := time.Now().UTC()
	job := CopyJob{
		ID:        operationID,
		OwnerID:   ownerID,
		Kind:      kind,
		Status:    CopyJobStatusQueued,
		Progress:  0,
		Stage:     copyJobStageQueued,
		CreatedAt: now,
		UpdatedAt: now,
	}
	m.jobs[key] = job
	m.activeByUser[ownerID] = job.ID
	m.wg.Add(1)
	m.mu.Unlock()

	m.publish(job)
	go func() {
		defer m.wg.Done()
		m.run(key, work)
	}()
	return job, nil
}

// Close cancels active copies and waits until their worker goroutines exit.
// It is safe to call more than once.
func (m *CopyJobManager) Close() {
	if m == nil {
		return
	}
	m.closeOnce.Do(func() {
		m.mu.Lock()
		m.closed = true
		m.mu.Unlock()
		m.cancel()
		m.wg.Wait()
	})
}

func (m *CopyJobManager) Get(ownerID, jobID uuid.UUID) (CopyJob, error) {
	m.mu.Lock()
	m.pruneCompletedLocked(time.Now().UTC())
	job, ok := m.jobs[copyJobKey{ownerID: ownerID, jobID: jobID}]
	m.mu.Unlock()
	if !ok {
		return CopyJob{}, ErrCopyJobNotFound
	}
	return job, nil
}

func (m *CopyJobManager) pruneCompletedLocked(now time.Time) {
	for key, job := range m.jobs {
		if job.IsTerminal() && now.Sub(job.UpdatedAt) > copyJobRetention {
			delete(m.jobs, key)
		}
	}
}

func (m *CopyJobManager) run(key copyJobKey, work func(context.Context) error) {
	m.update(key, CopyJobStatusRunning, 1, copyJobStagePreparing, "")
	ctx := withCopyProgressReporter(m.ctx, func(progress int, stage string) {
		m.report(key, progress, stage)
	})

	if err := work(ctx); err != nil {
		m.fail(key, err)
		return
	}
	m.update(key, CopyJobStatusCompleted, 100, copyJobStageCompleted, "")
}

func (m *CopyJobManager) report(key copyJobKey, progress int, stage string) {
	m.mu.Lock()
	job, ok := m.jobs[key]
	if !ok || job.IsTerminal() {
		m.mu.Unlock()
		return
	}
	if progress < job.Progress {
		progress = job.Progress
	}
	if progress > 99 {
		progress = 99
	}
	job.Progress = progress
	if stage != "" {
		job.Stage = stage
	}
	job.UpdatedAt = time.Now().UTC()
	m.jobs[key] = job
	m.mu.Unlock()
	m.publish(job)
}

func (m *CopyJobManager) fail(key copyJobKey, cause error) {
	message := "copy_failed"
	if cause != nil {
		message = cause.Error()
	}
	m.update(key, CopyJobStatusFailed, 100, copyJobStageFailed, message)
}

func (m *CopyJobManager) update(key copyJobKey, status CopyJobStatus, progress int, stage, failure string) {
	m.mu.Lock()
	job, ok := m.jobs[key]
	if !ok {
		m.mu.Unlock()
		return
	}
	job.Status = status
	job.Progress = progress
	job.Stage = stage
	job.Error = failure
	job.UpdatedAt = time.Now().UTC()
	m.jobs[key] = job
	if job.IsTerminal() && m.activeByUser[job.OwnerID] == job.ID {
		delete(m.activeByUser, job.OwnerID)
	}
	m.mu.Unlock()
	m.publish(job)
}

func (m *CopyJobManager) publish(job CopyJob) {
	if m.publisher == nil {
		return
	}
	m.publisher.BroadcastCopyJobProgress(context.Background(), apprealtime.CopyJobProgressEvent{
		JobID: job.ID, OwnerID: job.OwnerID, Kind: string(job.Kind), Status: string(job.Status),
		Progress: job.Progress, Stage: job.Stage, Error: job.Error, UpdatedAt: job.UpdatedAt,
	})
}
