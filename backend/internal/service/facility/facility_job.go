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
	apptransaction "github.com/besart951/go_infra_link/backend/internal/application/transaction"
	cursorcodec "github.com/besart951/go_infra_link/backend/internal/cursor"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type FacilityJobKind string

const (
	FacilityJobKindControlCabinet          FacilityJobKind = "control_cabinet"
	FacilityJobKindSPSController           FacilityJobKind = "sps_controller"
	FacilityJobKindSPSControllerSystemType FacilityJobKind = "sps_controller_system_type"
	FacilityJobKindFieldDevice             FacilityJobKind = "field_device"
	FacilityJobKindObjectData              FacilityJobKind = "object_data"
)

type FacilityJobStatus string

const (
	FacilityJobStatusQueued    FacilityJobStatus = "queued"
	FacilityJobStatusRunning   FacilityJobStatus = "running"
	FacilityJobStatusCompleted FacilityJobStatus = "completed"
	FacilityJobStatusFailed    FacilityJobStatus = "failed"
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
	facilityJobStageQueued              = "queued"
	facilityJobStagePreparing           = "preparing"
	facilityJobStageCopyingRoot         = "copying_root"
	facilityJobStageCopyingControllers  = "copying_controllers"
	facilityJobStageCopyingSystemTypes  = "copying_system_types"
	facilityJobStageCopyingFieldDevices = "copying_field_devices"
	facilityJobStageFinalizing          = "finalizing"
	facilityJobStageCompleted           = "completed"
	facilityJobStageFailed              = "failed"
	facilityJobRetention                = 90 * 24 * time.Hour
	facilityJobLeaseDuration            = 30 * time.Second
	facilityJobHeartbeatInterval        = 10 * time.Second
)

var (
	ErrFacilityJobNotFound      = errors.New("facility job not found")
	ErrFacilityJobLimit         = errors.New("facility job concurrency limit reached")
	ErrFacilityJobNotRetryable  = errors.New("facility job is not retryable")
	errFacilityJobManagerClosed = errors.New("facility job manager is closed")
)

// FacilityJob is a durable, user-scoped asynchronous facility operation. The
// operation ID is supplied by the browser, making request retries idempotent.
type FacilityJob struct {
	ID          uuid.UUID
	OwnerID     uuid.UUID
	Kind        FacilityJobKind
	Class       FacilityJobClass
	Type        FacilityJobType
	Task        string
	Payload     json.RawMessage
	Checkpoint  json.RawMessage
	Status      FacilityJobStatus
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
	Admission   *FacilityAggregateAdmission
}

type facilityJobKey struct {
	ownerID uuid.UUID
	jobID   uuid.UUID
}

func (j FacilityJob) IsTerminal() bool {
	return j.Status == FacilityJobStatusCompleted || j.Status == FacilityJobStatusFailed
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

type FacilityJobReporter interface {
	Report(FacilityJobProgress)
}

type FacilityJobExecution struct {
	Job        FacilityJob
	Reporter   FacilityJobReporter
	UnitOfWork apptransaction.UnitOfWork
}

type FacilityJobHandler interface {
	Execute(context.Context, FacilityJobExecution) (FacilityJobTaskResult, error)
}

type FacilityJobHandlerFunc func(context.Context, FacilityJobExecution) (FacilityJobTaskResult, error)

func (handler FacilityJobHandlerFunc) Execute(ctx context.Context, execution FacilityJobExecution) (FacilityJobTaskResult, error) {
	return handler(ctx, execution)
}

type facilityJobReporter struct {
	report func(FacilityJobProgress)
}

func (reporter facilityJobReporter) Report(progress FacilityJobProgress) {
	reporter.report(progress)
}

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
	Items          []FacilityJob
	NextCursor     string
	PreviousCursor string
}

type facilityJobCursor struct {
	UpdatedAt time.Time `json:"updated_at"`
	ID        uuid.UUID `json:"id"`
	Direction string    `json:"direction,omitempty"`
	OwnerID   uuid.UUID `json:"owner_id"`
}

type FacilityJobManager struct {
	mu        sync.Mutex
	jobs      map[facilityJobKey]FacilityJob
	store     facilityJobStore
	workerID  string
	publisher apprealtime.FacilityJobProgressPublisher
	ctx       context.Context
	cancel    context.CancelFunc
	wg        sync.WaitGroup
	closeOnce sync.Once
	closed    bool
	tasks     map[string]FacilityJobHandler
	wake      chan struct{}
}

func NewFacilityJobManager(publisher apprealtime.FacilityJobProgressPublisher) *FacilityJobManager {
	return newFacilityJobManager(publisher, nil)
}

func NewFacilityJobManagerWithDB(publisher apprealtime.FacilityJobProgressPublisher, db *gorm.DB) *FacilityJobManager {
	return newFacilityJobManager(publisher, newSQLFacilityJobStore(db))
}

func newFacilityJobManager(publisher apprealtime.FacilityJobProgressPublisher, store facilityJobStore) *FacilityJobManager {
	ctx, cancel := context.WithCancel(context.Background())
	m := &FacilityJobManager{
		jobs:      make(map[facilityJobKey]FacilityJob),
		store:     store,
		workerID:  uuid.NewString(),
		publisher: publisher,
		ctx:       ctx,
		cancel:    cancel,
		tasks:     make(map[string]FacilityJobHandler),
		wake:      make(chan struct{}, 1),
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
func (m *FacilityJobManager) RegisterTask(task string, handler FacilityJobHandler) {
	if m == nil || task == "" || handler == nil {
		return
	}
	m.mu.Lock()
	m.tasks[task] = handler
	m.mu.Unlock()
	m.signalWorkers()
}

func (m *FacilityJobManager) SupportsDurableTasks() bool {
	return m != nil && m.store != nil
}

// SubmitTask persists all information required to execute a job after a
// process restart.
func (m *FacilityJobManager) SubmitTask(ctx context.Context, job FacilityJob) (FacilityJob, error) {
	if m == nil || m.store == nil {
		return FacilityJob{}, errors.New("durable facility job store is unavailable")
	}
	m.mu.Lock()
	closed := m.closed
	m.mu.Unlock()
	if closed {
		return FacilityJob{}, errFacilityJobManagerClosed
	}
	if job.ID == uuid.Nil {
		job.ID = uuid.New()
	}
	if job.OwnerID == uuid.Nil || job.Task == "" {
		return FacilityJob{}, errors.New("facility job owner and task are required")
	}
	if job.Class == "" {
		job.Class = FacilityJobClassMutation
	}
	if job.Type == "" {
		job.Type = FacilityJobTypeCopy
	}
	now := time.Now().UTC()
	_ = m.store.Prune(ctx, now.Add(-facilityJobRetention))
	job.Status = FacilityJobStatusQueued
	job.Progress = 0
	job.Stage = facilityJobStageQueued
	job.CreatedAt = now
	job.UpdatedAt = now
	job.Retryable = true

	selected, created, err := m.store.CreateOrGetActive(ctx, job)
	if err != nil {
		return FacilityJob{}, err
	}
	if created {
		m.publish(selected)
	}
	m.signalWorkers()
	return selected, nil
}

func (m *FacilityJobManager) signalWorkers() {
	select {
	case m.wake <- struct{}{}:
	default:
	}
}

// Close cancels local workers and waits until their goroutines exit. Durable
// jobs deliberately remain running until their lease expires, allowing a new
// process to claim and resume them.
func (m *FacilityJobManager) Close() {
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

func (m *FacilityJobManager) Get(ownerID, jobID uuid.UUID) (FacilityJob, error) {
	if m == nil || m.store == nil {
		return FacilityJob{}, ErrFacilityJobNotFound
	}
	return m.store.Get(context.Background(), ownerID, jobID)
}

func (m *FacilityJobManager) List(ownerID uuid.UUID, limit int) ([]FacilityJob, error) {
	if m == nil || m.store == nil {
		return nil, ErrFacilityJobNotFound
	}
	return m.store.List(context.Background(), ownerID, limit)
}

func (m *FacilityJobManager) ListPage(ownerID uuid.UUID, limit int, encodedCursor string) (FacilityJobPage, error) {
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
	if m == nil || m.store == nil {
		return FacilityJobPage{}, ErrFacilityJobNotFound
	}
	items, err := m.store.ListPage(context.Background(), ownerID, after, limit+1)
	if err != nil {
		return FacilityJobPage{}, err
	}
	return buildFacilityJobPage(ownerID, items, limit, after)
}

func buildFacilityJobPage(ownerID uuid.UUID, items []FacilityJob, limit int, pageCursor *facilityJobCursor) (FacilityJobPage, error) {
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

func (m *FacilityJobManager) startHeartbeat(key facilityJobKey) func() {
	if m.store == nil {
		return func() {}
	}
	stop := make(chan struct{})
	done := make(chan struct{})
	go func() {
		defer close(done)
		ticker := time.NewTicker(facilityJobHeartbeatInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				now := time.Now().UTC()
				_ = m.store.Heartbeat(context.Background(), key.ownerID, key.jobID, m.workerID, now, now.Add(facilityJobLeaseDuration))
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

func (m *FacilityJobManager) fail(key facilityJobKey, cause error) {
	message := "job_failed"
	if cause != nil {
		message = cause.Error()
	}
	m.update(key, FacilityJobStatusFailed, 100, facilityJobStageFailed, message)
}

func (m *FacilityJobManager) update(key facilityJobKey, status FacilityJobStatus, progress int, stage, failure string) {
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
	m.mu.Unlock()
	m.persistAndPublish(job)
	if job.IsTerminal() {
		m.mu.Lock()
		delete(m.jobs, key)
		m.mu.Unlock()
	}
}

func (m *FacilityJobManager) persistAndPublish(job FacilityJob) {
	if m.store != nil {
		if err := m.store.Save(context.Background(), job, m.workerID); err != nil {
			return
		}
	}
	m.publish(job)
}

func (m *FacilityJobManager) dispatch(class FacilityJobClass) {
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
			job, claimed, err := m.store.ClaimNext(m.ctx, class, tasks, m.workerID, now, now.Add(facilityJobLeaseDuration))
			if err != nil || !claimed {
				break
			}
			m.runPersistedTask(job)
		}
	}
}

func (m *FacilityJobManager) registeredTasks() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	tasks := make([]string, 0, len(m.tasks))
	for task := range m.tasks {
		tasks = append(tasks, task)
	}
	sort.Strings(tasks)
	return tasks
}

func (m *FacilityJobManager) runPersistedTask(job FacilityJob) {
	key := facilityJobKey{ownerID: job.OwnerID, jobID: job.ID}
	m.mu.Lock()
	handler := m.tasks[job.Task]
	job.Status = FacilityJobStatusRunning
	job.Stage = facilityJobStagePreparing
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

	execution := FacilityJobExecution{Job: job, Reporter: facilityJobReporter{report: func(progress FacilityJobProgress) {
		m.reportTask(key, progress)
	}}}
	result, err := handler.Execute(m.ctx, execution)
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
	m.update(key, FacilityJobStatusCompleted, 100, facilityJobStageCompleted, "")
}

func (m *FacilityJobManager) reportTask(key facilityJobKey, progress FacilityJobProgress) {
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
func (m *FacilityJobManager) Retry(ctx context.Context, ownerID, jobID uuid.UUID) (FacilityJob, error) {
	if m == nil || m.store == nil {
		return FacilityJob{}, ErrFacilityJobNotFound
	}
	job, err := m.store.Retry(ctx, ownerID, jobID, time.Now().UTC())
	if err != nil {
		return FacilityJob{}, err
	}
	m.signalWorkers()
	m.publish(job)
	return job, nil
}

func (m *FacilityJobManager) publish(job FacilityJob) {
	if m.publisher == nil {
		return
	}
	m.publisher.BroadcastFacilityJobProgress(context.Background(), apprealtime.FacilityJobProgressEvent{
		JobID: job.ID, OwnerID: job.OwnerID, Kind: string(job.Kind), Status: string(job.Status),
		JobType: string(job.Type), Class: string(job.Class), Progress: job.Progress, Stage: job.Stage,
		Error: job.Error, Processed: job.Processed, Total: job.Total, Succeeded: job.Succeeded, Failed: job.Failed, UpdatedAt: job.UpdatedAt,
	})
}
