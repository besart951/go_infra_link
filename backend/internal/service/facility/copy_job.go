package facility

import (
	"context"
	"encoding/json"
	"errors"
	"slices"
	"sort"
	"sync"
	"time"

	apprealtime "github.com/besart951/go_infra_link/backend/internal/application/realtime"
	cursorcodec "github.com/besart951/go_infra_link/backend/internal/cursor"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CopyJobKind string

const (
	CopyJobKindControlCabinet          CopyJobKind = "control_cabinet"
	CopyJobKindSPSController           CopyJobKind = "sps_controller"
	CopyJobKindSPSControllerSystemType CopyJobKind = "sps_controller_system_type"
	CopyJobKindFieldDevice             CopyJobKind = "field_device"
	CopyJobKindObjectData              CopyJobKind = "object_data"
)

type CopyJobStatus string

const (
	CopyJobStatusQueued    CopyJobStatus = "queued"
	CopyJobStatusRunning   CopyJobStatus = "running"
	CopyJobStatusCompleted CopyJobStatus = "completed"
	CopyJobStatusFailed    CopyJobStatus = "failed"
)

type FacilityJobClass string

const (
	FacilityJobClassMutation FacilityJobClass = "mutation"
	FacilityJobClassExport   FacilityJobClass = "export"
)

type FacilityJobType string

const (
	FacilityJobTypeCopy    FacilityJobType = "copy"
	FacilityJobTypeExport  FacilityJobType = "export"
	FacilityJobTypeBulk    FacilityJobType = "bulk"
	FacilityJobTypeDelete  FacilityJobType = "delete"
	FacilityJobTypeRestore FacilityJobType = "restore"
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
	copyJobRetention                = 90 * 24 * time.Hour
	copyJobLeaseDuration            = 30 * time.Second
	copyJobHeartbeatInterval        = 10 * time.Second
)

var (
	ErrCopyJobNotFound         = errors.New("copy job not found")
	ErrFacilityJobLimit        = errors.New("facility job concurrency limit reached")
	ErrFacilityJobNotRetryable = errors.New("facility job is not retryable")
	errCopyJobManagerClosed    = errors.New("copy job manager is closed")
)

// CopyJob is a durable, user-scoped asynchronous facility operation. The
// operation ID is supplied by the browser, making request retries idempotent.
type CopyJob struct {
	ID          uuid.UUID
	OwnerID     uuid.UUID
	Kind        CopyJobKind
	Class       FacilityJobClass
	Type        FacilityJobType
	Task        string
	Payload     json.RawMessage
	Checkpoint  json.RawMessage
	Status      CopyJobStatus
	Progress    int
	Stage       string
	Error       string
	Attempts    int
	Processed   int64
	Total       *int64
	Succeeded   int64
	Failed      int64
	Retryable   bool
	Result      json.RawMessage
	CreatedAt   time.Time
	UpdatedAt   time.Time
	CompletedAt *time.Time
}

type copyJobKey struct {
	ownerID uuid.UUID
	jobID   uuid.UUID
}

func (j CopyJob) IsTerminal() bool {
	return j.Status == CopyJobStatusCompleted || j.Status == CopyJobStatusFailed
}

type FacilityJobProgress struct {
	Progress   int
	Stage      string
	Processed  int64
	Total      *int64
	Succeeded  int64
	Failed     int64
	Checkpoint json.RawMessage
}

type FacilityJobTaskResult struct {
	Result json.RawMessage
}

type FacilityJobTask func(context.Context, CopyJob, func(FacilityJobProgress)) (FacilityJobTaskResult, error)

const (
	FacilityJobTaskCopyControlCabinet          = "controlcabinet.copy.v1"
	FacilityJobTaskCopySPSController           = "spscontroller.copy.v1"
	FacilityJobTaskCopySPSControllerSystemType = "spscontrollersystemtype.copy.v1"
	FacilityJobTaskCopyFieldDevice             = "fielddevice.copy.v1"
	FacilityJobTaskCopyObjectData              = "objectdata.copy.v1"
)

type FacilityCopyTaskPayload struct {
	SourceID uuid.UUID `json:"source_id"`
}

type FacilityJobPage struct {
	Items          []CopyJob
	NextCursor     string
	PreviousCursor string
}

type facilityJobCursor struct {
	UpdatedAt time.Time `json:"updated_at"`
	ID        uuid.UUID `json:"id"`
	Direction string    `json:"direction,omitempty"`
	OwnerID   uuid.UUID `json:"owner_id"`
}

type CopyJobManager struct {
	mu           sync.Mutex
	jobs         map[copyJobKey]CopyJob
	activeByUser map[uuid.UUID]uuid.UUID
	running      map[copyJobKey]struct{}
	store        copyJobStore
	workerID     string
	publisher    apprealtime.CopyJobProgressPublisher
	ctx          context.Context
	cancel       context.CancelFunc
	wg           sync.WaitGroup
	closeOnce    sync.Once
	closed       bool
	tasks        map[string]FacilityJobTask
	wake         chan struct{}
}

func NewCopyJobManager(publisher apprealtime.CopyJobProgressPublisher) *CopyJobManager {
	return newCopyJobManager(publisher, nil)
}

func NewCopyJobManagerWithDB(publisher apprealtime.CopyJobProgressPublisher, db *gorm.DB) *CopyJobManager {
	return newCopyJobManager(publisher, newSQLCopyJobStore(db))
}

func newCopyJobManager(publisher apprealtime.CopyJobProgressPublisher, store copyJobStore) *CopyJobManager {
	ctx, cancel := context.WithCancel(context.Background())
	m := &CopyJobManager{
		jobs:         make(map[copyJobKey]CopyJob),
		activeByUser: make(map[uuid.UUID]uuid.UUID),
		running:      make(map[copyJobKey]struct{}),
		store:        store,
		workerID:     uuid.NewString(),
		publisher:    publisher,
		ctx:          ctx,
		cancel:       cancel,
		tasks:        make(map[string]FacilityJobTask),
		wake:         make(chan struct{}, 1),
	}
	if store != nil {
		for range 2 {
			m.wg.Add(2)
			go m.dispatch(FacilityJobClassMutation)
			go m.dispatch(FacilityJobClassExport)
		}
	}
	return m
}

// RegisterTask connects a versioned persisted task name to executable domain
// logic. Registration is intentionally explicit so the worker can safely
// ignore jobs created by a newer application version.
func (m *CopyJobManager) RegisterTask(task string, handler FacilityJobTask) {
	if m == nil || task == "" || handler == nil {
		return
	}
	m.mu.Lock()
	m.tasks[task] = handler
	m.mu.Unlock()
	m.signalWorkers()
}

func (m *CopyJobManager) SupportsDurableTasks() bool {
	return m != nil && m.store != nil
}

// SubmitTask persists all information required to execute a job after a
// process restart. Unlike Start, it never captures an HTTP callback.
func (m *CopyJobManager) SubmitTask(ctx context.Context, job CopyJob) (CopyJob, error) {
	if m == nil || m.store == nil {
		return CopyJob{}, errors.New("durable facility job store is unavailable")
	}
	if job.ID == uuid.Nil {
		job.ID = uuid.New()
	}
	if job.OwnerID == uuid.Nil || job.Task == "" {
		return CopyJob{}, errors.New("facility job owner and task are required")
	}
	if job.Class == "" {
		job.Class = FacilityJobClassMutation
	}
	if job.Type == "" {
		job.Type = FacilityJobTypeCopy
	}
	now := time.Now().UTC()
	job.Status = CopyJobStatusQueued
	job.Progress = 0
	job.Stage = copyJobStageQueued
	job.CreatedAt = now
	job.UpdatedAt = now
	job.Retryable = true

	selected, created, err := m.store.CreateOrGetActive(ctx, job)
	if err != nil {
		return CopyJob{}, err
	}
	if created {
		m.publish(selected)
	}
	m.signalWorkers()
	return selected, nil
}

func (m *CopyJobManager) signalWorkers() {
	select {
	case m.wake <- struct{}{}:
	default:
	}
}

// Start creates a job or returns the user's active job. Only one hierarchy
// operation runs for a user at a time. With a store configured, status and
// leases survive process restarts; re-submitting the same operation attaches
// the executable command to the persisted job again.
func (m *CopyJobManager) Start(
	ownerID uuid.UUID,
	operationID uuid.UUID,
	kind CopyJobKind,
	work func(context.Context) error,
) (CopyJob, error) {
	m.mu.Lock()
	if m.closed {
		m.mu.Unlock()
		return CopyJob{}, errCopyJobManagerClosed
	}
	m.mu.Unlock()

	now := time.Now().UTC()
	candidate := CopyJob{
		ID:        operationID,
		OwnerID:   ownerID,
		Kind:      kind,
		Class:     FacilityJobClassMutation,
		Type:      FacilityJobTypeCopy,
		Status:    CopyJobStatusQueued,
		Progress:  0,
		Stage:     copyJobStageQueued,
		CreatedAt: now,
		UpdatedAt: now,
	}

	job := candidate
	created := true
	if m.store != nil {
		_ = m.store.Prune(m.ctx, now.Add(-copyJobRetention))
		var err error
		job, created, err = m.store.CreateOrGetActive(m.ctx, candidate)
		if err != nil {
			return CopyJob{}, err
		}
	}

	key := copyJobKey{ownerID: job.OwnerID, jobID: job.ID}
	m.mu.Lock()
	if m.closed {
		m.mu.Unlock()
		return CopyJob{}, errCopyJobManagerClosed
	}
	if m.store == nil {
		m.pruneCompletedLocked(now)
		if existing, ok := m.jobs[copyJobKey{ownerID: ownerID, jobID: operationID}]; ok {
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
		job = candidate
		key = copyJobKey{ownerID: ownerID, jobID: operationID}
	}

	m.jobs[key] = job
	if !job.IsTerminal() {
		m.activeByUser[job.OwnerID] = job.ID
	}
	_, alreadyRunning := m.running[key]
	shouldStart := !job.IsTerminal() && job.ID == operationID && !alreadyRunning
	if shouldStart {
		m.running[key] = struct{}{}
		m.wg.Add(1)
	}
	m.mu.Unlock()

	if created {
		m.publish(job)
	}
	if shouldStart {
		go func() {
			defer m.wg.Done()
			defer m.removeLocalRun(key)
			m.run(key, work)
		}()
	}
	return job, nil
}

// Close cancels local workers and waits until their goroutines exit. Durable
// jobs deliberately remain running until their lease expires, allowing a new
// process to claim and resume them.
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
	if m.store != nil {
		return m.store.Get(context.Background(), ownerID, jobID)
	}
	m.mu.Lock()
	m.pruneCompletedLocked(time.Now().UTC())
	job, ok := m.jobs[copyJobKey{ownerID: ownerID, jobID: jobID}]
	m.mu.Unlock()
	if !ok {
		return CopyJob{}, ErrCopyJobNotFound
	}
	return job, nil
}

func (m *CopyJobManager) List(ownerID uuid.UUID, limit int) ([]CopyJob, error) {
	if m.store != nil {
		return m.store.List(context.Background(), ownerID, limit)
	}
	if limit <= 0 || limit > 200 {
		limit = 100
	}
	m.mu.Lock()
	m.pruneCompletedLocked(time.Now().UTC())
	jobs := make([]CopyJob, 0, len(m.jobs))
	for _, job := range m.jobs {
		if job.OwnerID == ownerID {
			jobs = append(jobs, job)
		}
	}
	m.mu.Unlock()
	sort.Slice(jobs, func(i, j int) bool {
		if jobs[i].UpdatedAt.Equal(jobs[j].UpdatedAt) {
			return jobs[i].ID.String() > jobs[j].ID.String()
		}
		return jobs[i].UpdatedAt.After(jobs[j].UpdatedAt)
	})
	if len(jobs) > limit {
		jobs = jobs[:limit]
	}
	return jobs, nil
}

func (m *CopyJobManager) ListPage(ownerID uuid.UUID, limit int, encodedCursor string) (FacilityJobPage, error) {
	if limit <= 0 {
		limit = 50
	}
	limit = min(limit, 100)
	var after *facilityJobCursor
	if encodedCursor != "" {
		var decoded facilityJobCursor
		if err := cursorcodec.Decode(encodedCursor, "facility_jobs", &decoded); err != nil ||
			decoded.ID == uuid.Nil || decoded.UpdatedAt.IsZero() || decoded.OwnerID != ownerID ||
			(decoded.Direction != "" && decoded.Direction != "next" && decoded.Direction != "previous") {
			return FacilityJobPage{}, cursorcodec.ErrInvalid
		}
		if decoded.Direction == "" {
			decoded.Direction = "next"
		}
		after = &decoded
	}
	if m.store == nil {
		items, err := m.List(ownerID, 200)
		if err != nil {
			return FacilityJobPage{}, err
		}
		items = filterInMemoryJobPage(items, after, limit+1)
		return buildFacilityJobPage(ownerID, items, limit, after)
	}
	items, err := m.store.ListPage(context.Background(), ownerID, after, limit+1)
	if err != nil {
		return FacilityJobPage{}, err
	}
	return buildFacilityJobPage(ownerID, items, limit, after)
}

func filterInMemoryJobPage(items []CopyJob, after *facilityJobCursor, limit int) []CopyJob {
	filtered := make([]CopyJob, 0, min(len(items), limit))
	if after != nil && after.Direction == "previous" {
		for i := len(items) - 1; i >= 0; i-- {
			job := items[i]
			if job.UpdatedAt.Before(after.UpdatedAt) || (job.UpdatedAt.Equal(after.UpdatedAt) && job.ID.String() <= after.ID.String()) {
				continue
			}
			filtered = append(filtered, job)
			if len(filtered) == limit {
				break
			}
		}
		return filtered
	}
	for _, job := range items {
		if after != nil && (job.UpdatedAt.After(after.UpdatedAt) || (job.UpdatedAt.Equal(after.UpdatedAt) && job.ID.String() >= after.ID.String())) {
			continue
		}
		filtered = append(filtered, job)
		if len(filtered) == limit {
			break
		}
	}
	return filtered
}

func buildFacilityJobPage(ownerID uuid.UUID, items []CopyJob, limit int, pageCursor *facilityJobCursor) (FacilityJobPage, error) {
	page := FacilityJobPage{}
	hasMore := len(items) > limit
	if hasMore {
		items = items[:limit]
	}
	if pageCursor != nil && pageCursor.Direction == "previous" {
		slices.Reverse(items)
	}
	if len(items) == 0 {
		return page, nil
	}
	first := items[0]
	last := items[len(items)-1]
	if pageCursor == nil || pageCursor.Direction != "previous" {
		if hasMore {
			next, err := cursorcodec.Encode("facility_jobs", facilityJobCursor{UpdatedAt: last.UpdatedAt, ID: last.ID, Direction: "next", OwnerID: ownerID})
			if err != nil {
				return FacilityJobPage{}, err
			}
			page.NextCursor = next
		}
		if pageCursor != nil {
			previous, err := cursorcodec.Encode("facility_jobs", facilityJobCursor{UpdatedAt: first.UpdatedAt, ID: first.ID, Direction: "previous", OwnerID: ownerID})
			if err != nil {
				return FacilityJobPage{}, err
			}
			page.PreviousCursor = previous
		}
	} else {
		if hasMore {
			previous, err := cursorcodec.Encode("facility_jobs", facilityJobCursor{UpdatedAt: first.UpdatedAt, ID: first.ID, Direction: "previous", OwnerID: ownerID})
			if err != nil {
				return FacilityJobPage{}, err
			}
			page.PreviousCursor = previous
		}
		next, err := cursorcodec.Encode("facility_jobs", facilityJobCursor{UpdatedAt: last.UpdatedAt, ID: last.ID, Direction: "next", OwnerID: ownerID})
		if err != nil {
			return FacilityJobPage{}, err
		}
		page.NextCursor = next
	}
	page.Items = items
	return page, nil
}

func (m *CopyJobManager) pruneCompletedLocked(now time.Time) {
	for key, job := range m.jobs {
		if job.IsTerminal() && now.Sub(job.UpdatedAt) > copyJobRetention {
			delete(m.jobs, key)
		}
	}
}

func (m *CopyJobManager) run(key copyJobKey, work func(context.Context) error) {
	if m.store != nil {
		now := time.Now().UTC()
		claimed, err := m.store.Claim(m.ctx, key.ownerID, key.jobID, m.workerID, now, now.Add(copyJobLeaseDuration))
		if err != nil || !claimed {
			return
		}
	}

	m.update(key, CopyJobStatusRunning, 1, copyJobStagePreparing, "")
	stopHeartbeat := m.startHeartbeat(key)
	defer stopHeartbeat()
	ctx := withCopyProgressReporter(m.ctx, func(progress int, stage string) {
		m.report(key, progress, stage)
	})

	if err := work(ctx); err != nil {
		if m.store != nil && errors.Is(err, context.Canceled) && m.ctx.Err() != nil {
			return
		}
		m.fail(key, err)
		return
	}
	m.update(key, CopyJobStatusCompleted, 100, copyJobStageCompleted, "")
}

func (m *CopyJobManager) startHeartbeat(key copyJobKey) func() {
	if m.store == nil {
		return func() {}
	}
	stop := make(chan struct{})
	done := make(chan struct{})
	go func() {
		defer close(done)
		ticker := time.NewTicker(copyJobHeartbeatInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				now := time.Now().UTC()
				_ = m.store.Heartbeat(context.Background(), key.ownerID, key.jobID, m.workerID, now, now.Add(copyJobLeaseDuration))
			case <-stop:
				return
			}
		}
	}()
	return func() {
		close(stop)
		<-done
	}
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
	m.persistAndPublish(job)
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
	m.persistAndPublish(job)
}

func (m *CopyJobManager) persistAndPublish(job CopyJob) {
	if m.store != nil {
		if err := m.store.Save(context.Background(), job, m.workerID); err != nil {
			return
		}
	}
	m.publish(job)
}

func (m *CopyJobManager) dispatch(class FacilityJobClass) {
	defer m.wg.Done()
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-m.ctx.Done():
			return
		case <-m.wake:
		case <-ticker.C:
		}

		for {
			tasks := m.registeredTasks()
			if len(tasks) == 0 {
				break
			}
			now := time.Now().UTC()
			job, claimed, err := m.store.ClaimNext(m.ctx, class, tasks, m.workerID, now, now.Add(copyJobLeaseDuration))
			if err != nil || !claimed {
				break
			}
			m.runPersistedTask(job)
		}
	}
}

func (m *CopyJobManager) registeredTasks() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	tasks := make([]string, 0, len(m.tasks))
	for task := range m.tasks {
		tasks = append(tasks, task)
	}
	sort.Strings(tasks)
	return tasks
}

func (m *CopyJobManager) runPersistedTask(job CopyJob) {
	key := copyJobKey{ownerID: job.OwnerID, jobID: job.ID}
	m.mu.Lock()
	handler := m.tasks[job.Task]
	job.Status = CopyJobStatusRunning
	job.Stage = copyJobStagePreparing
	job.Progress = max(1, job.Progress)
	job.UpdatedAt = time.Now().UTC()
	m.jobs[key] = job
	m.mu.Unlock()
	if handler == nil {
		return
	}
	m.persistAndPublish(job)
	stopHeartbeat := m.startHeartbeat(key)
	defer stopHeartbeat()

	result, err := handler(m.ctx, job, func(progress FacilityJobProgress) {
		m.reportTask(key, progress)
	})
	if err != nil {
		if errors.Is(err, context.Canceled) && m.ctx.Err() != nil {
			return
		}
		m.fail(key, err)
		return
	}
	m.mu.Lock()
	completed := m.jobs[key]
	completed.Result = result.Result
	m.jobs[key] = completed
	m.mu.Unlock()
	m.update(key, CopyJobStatusCompleted, 100, copyJobStageCompleted, "")
}

func (m *CopyJobManager) reportTask(key copyJobKey, progress FacilityJobProgress) {
	m.mu.Lock()
	job, ok := m.jobs[key]
	if !ok || job.IsTerminal() {
		m.mu.Unlock()
		return
	}
	if progress.Progress >= job.Progress {
		job.Progress = min(progress.Progress, 99)
	}
	if progress.Stage != "" {
		job.Stage = progress.Stage
	}
	job.Processed = progress.Processed
	job.Total = progress.Total
	job.Succeeded = progress.Succeeded
	job.Failed = progress.Failed
	if len(progress.Checkpoint) > 0 {
		job.Checkpoint = progress.Checkpoint
	}
	job.UpdatedAt = time.Now().UTC()
	m.jobs[key] = job
	m.mu.Unlock()
	m.persistAndPublish(job)
}

// Retry requeues a failed, retryable job while preserving its checkpoint.
func (m *CopyJobManager) Retry(ctx context.Context, ownerID, jobID uuid.UUID) (CopyJob, error) {
	if m == nil || m.store == nil {
		return CopyJob{}, ErrCopyJobNotFound
	}
	job, err := m.store.Retry(ctx, ownerID, jobID, time.Now().UTC())
	if err != nil {
		return CopyJob{}, err
	}
	m.signalWorkers()
	m.publish(job)
	return job, nil
}

func (m *CopyJobManager) removeLocalRun(key copyJobKey) {
	m.mu.Lock()
	delete(m.running, key)
	m.mu.Unlock()
}

func (m *CopyJobManager) publish(job CopyJob) {
	if m.publisher == nil {
		return
	}
	m.publisher.BroadcastCopyJobProgress(context.Background(), apprealtime.CopyJobProgressEvent{
		JobID: job.ID, OwnerID: job.OwnerID, Kind: string(job.Kind), Status: string(job.Status),
		JobType: string(job.Type), Class: string(job.Class), Progress: job.Progress, Stage: job.Stage,
		Error: job.Error, Processed: job.Processed, Total: job.Total, Succeeded: job.Succeeded, Failed: job.Failed, UpdatedAt: job.UpdatedAt,
	})
}
