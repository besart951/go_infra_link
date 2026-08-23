package facility

import (
	"context"
	"encoding/json"
	"errors"
	"slices"
	"sort"
	"strings"
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
	FacilityJobStatusCancelled FacilityJobStatus = "cancelled"
)

type FacilityJobClass string

const (
	FacilityJobClassMutation FacilityJobClass = "mutation"
	FacilityJobClassExport   FacilityJobClass = "export"
)

type FacilityJobType string

type FacilityJobTaskID string

func NewFacilityJobTaskID(raw string) (FacilityJobTaskID, error) {
	task := FacilityJobTaskID(strings.TrimSpace(raw))
	if task == "" {
		return "", errors.New("facility job task ID is required")
	}
	return task, nil
}

type FacilityJobStage string

const (
	FacilityJobTypeCopy    FacilityJobType = "copy"
	FacilityJobTypeExport  FacilityJobType = "export"
	FacilityJobTypeBulk    FacilityJobType = "bulk"
	FacilityJobTypeDelete  FacilityJobType = "delete"
	FacilityJobTypeRestore FacilityJobType = "restore"
)

const (
	facilityJobStageQueued              FacilityJobStage = "queued"
	facilityJobStagePreparing           FacilityJobStage = "preparing"
	facilityJobStageCopyingRoot         FacilityJobStage = "copying_root"
	facilityJobStageCopyingControllers  FacilityJobStage = "copying_controllers"
	facilityJobStageCopyingSystemTypes  FacilityJobStage = "copying_system_types"
	facilityJobStageCopyingFieldDevices FacilityJobStage = "copying_field_devices"
	facilityJobStageFinalizing          FacilityJobStage = "finalizing"
	facilityJobStageCompleted           FacilityJobStage = "completed"
	facilityJobStageFailed              FacilityJobStage = "failed"
	facilityJobStageCancelled           FacilityJobStage = "cancelled"
	facilityJobRetention                                 = 90 * 24 * time.Hour
	facilityJobLeaseDuration                             = 30 * time.Second
	facilityJobHeartbeatInterval                         = 10 * time.Second
)

var (
	ErrFacilityJobNotFound       = errors.New("facility job not found")
	ErrFacilityJobLimit          = errors.New("facility job concurrency limit reached")
	ErrFacilityJobNotRetryable   = errors.New("facility job is not retryable")
	ErrFacilityJobNotCancellable = errors.New("facility job is not cancellable")
	ErrFacilityJobNotRemovable   = errors.New("facility job is not removable")
	errFacilityJobWorkerClosed   = errors.New("facility job worker is closed")
)

// FacilityJob is a durable, user-scoped asynchronous facility operation. The
// operation ID is supplied by the browser, making request retries idempotent.
type FacilityJob struct {
	ID          uuid.UUID
	OwnerID     uuid.UUID
	Kind        FacilityJobKind
	Class       FacilityJobClass
	Type        FacilityJobType
	Task        FacilityJobTaskID
	Payload     json.RawMessage
	Checkpoint  json.RawMessage
	Status      FacilityJobStatus
	Progress    int
	Stage       FacilityJobStage
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
	return j.Status == FacilityJobStatusCompleted || j.Status == FacilityJobStatusFailed || j.Status == FacilityJobStatusCancelled
}

type FacilityJobProgress struct {
	Progress   int
	Stage      FacilityJobStage
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
	FacilityJobTaskCopyControlCabinet          FacilityJobTaskID = "controlcabinet.copy.v1"
	FacilityJobTaskCopySPSController           FacilityJobTaskID = "spscontroller.copy.v1"
	FacilityJobTaskCopySPSControllerSystemType FacilityJobTaskID = "spscontrollersystemtype.copy.v1"
	FacilityJobTaskCopyFieldDevice             FacilityJobTaskID = "fielddevice.copy.v1"
	FacilityJobTaskCopyObjectData              FacilityJobTaskID = "objectdata.copy.v1"
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

type TaskRegistry struct {
	mu     sync.RWMutex
	tasks  map[FacilityJobTaskID]FacilityJobHandler
	signal func()
}

func newTaskRegistry() *TaskRegistry {
	return &TaskRegistry{tasks: make(map[FacilityJobTaskID]FacilityJobHandler)}
}

func (r *TaskRegistry) Register(task FacilityJobTaskID, handler FacilityJobHandler) {
	if r == nil || task == "" || handler == nil {
		return
	}
	r.mu.Lock()
	r.tasks[task] = handler
	r.mu.Unlock()
	if r.signal != nil {
		r.signal()
	}
}

func (r *TaskRegistry) handler(task FacilityJobTaskID) FacilityJobHandler {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.tasks[task]
}

func (r *TaskRegistry) IDs() []FacilityJobTaskID {
	r.mu.RLock()
	defer r.mu.RUnlock()
	tasks := make([]FacilityJobTaskID, 0, len(r.tasks))
	for task := range r.tasks {
		tasks = append(tasks, task)
	}
	sort.Slice(tasks, func(i, j int) bool { return tasks[i] < tasks[j] })
	return tasks
}

type Worker struct {
	mu        sync.Mutex
	jobs      map[facilityJobKey]FacilityJob
	store     JobStore
	tasks     *TaskRegistry
	workerID  string
	publisher apprealtime.FacilityJobProgressPublisher
	ctx       context.Context
	cancel    context.CancelFunc
	wg        sync.WaitGroup
	closeOnce sync.Once
	closed    bool
	cancels   map[facilityJobKey]context.CancelFunc
	wake      chan struct{}
}

type JobService struct {
	store     JobStore
	worker    *Worker
	publisher apprealtime.FacilityJobProgressPublisher
}

type FacilityJobs struct {
	Service *JobService
	Tasks   *TaskRegistry
	Worker  *Worker
}

func NewFacilityJobs(publisher apprealtime.FacilityJobProgressPublisher) *FacilityJobs {
	return newFacilityJobs(publisher, nil)
}

func NewDurableFacilityJobs(publisher apprealtime.FacilityJobProgressPublisher, db *gorm.DB) *FacilityJobs {
	return newFacilityJobs(publisher, newSQLFacilityJobStore(db))
}

func newFacilityJobs(publisher apprealtime.FacilityJobProgressPublisher, store JobStore) *FacilityJobs {
	ctx, cancel := context.WithCancel(context.Background())
	registry := newTaskRegistry()
	worker := &Worker{
		jobs:      make(map[facilityJobKey]FacilityJob),
		store:     store,
		tasks:     registry,
		workerID:  uuid.NewString(),
		publisher: publisher,
		ctx:       ctx,
		cancel:    cancel,
		cancels:   make(map[facilityJobKey]context.CancelFunc),
		wake:      make(chan struct{}, 1),
	}
	registry.signal = worker.signal
	service := &JobService{store: store, worker: worker, publisher: publisher}
	if store != nil {
		for range 2 {
			worker.wg.Add(2)
			go worker.dispatch(FacilityJobClassMutation)
			go worker.dispatch(FacilityJobClassExport)
		}
	}
	return &FacilityJobs{Service: service, Tasks: registry, Worker: worker}
}

func (s *JobService) SupportsDurableTasks() bool {
	return s != nil && s.store != nil
}

// SubmitTask persists all information required to execute a job after a
// process restart.
func (s *JobService) SubmitTask(ctx context.Context, job FacilityJob) (FacilityJob, error) {
	if s == nil || s.store == nil {
		return FacilityJob{}, errors.New("durable facility job store is unavailable")
	}
	closed := s.worker.isClosed()
	if closed {
		return FacilityJob{}, errFacilityJobWorkerClosed
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
	_ = s.store.Prune(ctx, now.Add(-facilityJobRetention))
	job.Status = FacilityJobStatusQueued
	job.Progress = 0
	job.Stage = facilityJobStageQueued
	job.CreatedAt = now
	job.UpdatedAt = now
	job.Retryable = true

	selected, created, err := s.store.CreateOrGetActive(ctx, job)
	if err != nil {
		return FacilityJob{}, err
	}
	if created {
		s.publish(selected)
	}
	s.worker.signal()
	return selected, nil
}

func (w *Worker) signal() {
	select {
	case w.wake <- struct{}{}:
	default:
	}
}

// Close cancels local workers and waits until their goroutines exit. Durable
// jobs deliberately remain running until their lease expires, allowing a new
// process to claim and resume them.
func (w *Worker) Close() {
	if w == nil {
		return
	}
	w.closeOnce.Do(func() {
		w.mu.Lock()
		w.closed = true
		w.mu.Unlock()
		w.cancel()
		w.wg.Wait()
	})
}

func (w *Worker) isClosed() bool {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.closed
}

func (s *JobService) Get(ownerID, jobID uuid.UUID) (FacilityJob, error) {
	if s == nil || s.store == nil {
		return FacilityJob{}, ErrFacilityJobNotFound
	}
	return s.store.Get(context.Background(), ownerID, jobID)
}

func (s *JobService) List(ownerID uuid.UUID, limit int) ([]FacilityJob, error) {
	if s == nil || s.store == nil {
		return nil, ErrFacilityJobNotFound
	}
	return s.store.List(context.Background(), ownerID, limit)
}

func (s *JobService) ListPage(ownerID uuid.UUID, limit int, encodedCursor string) (FacilityJobPage, error) {
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
	if s == nil || s.store == nil {
		return FacilityJobPage{}, ErrFacilityJobNotFound
	}
	items, err := s.store.ListPage(context.Background(), ownerID, after, limit+1)
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

func (w *Worker) startHeartbeat(key facilityJobKey) func() {
	if w.store == nil {
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
				_ = w.store.Heartbeat(context.Background(), key.ownerID, key.jobID, w.workerID, now, now.Add(facilityJobLeaseDuration))
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

func (w *Worker) fail(key facilityJobKey, cause error) {
	message := "job_failed"
	if cause != nil {
		message = cause.Error()
	}
	w.update(key, FacilityJobStatusFailed, 100, facilityJobStageFailed, message)
}

func (w *Worker) update(key facilityJobKey, status FacilityJobStatus, progress int, stage FacilityJobStage, failure string) {
	w.mu.Lock()
	job, ok := w.jobs[key]
	if !ok {
		w.mu.Unlock()
		return
	}
	job.Status = status
	job.Progress = progress
	job.Stage = stage
	job.Error = failure
	job.UpdatedAt = time.Now().UTC()
	w.jobs[key] = job
	w.mu.Unlock()
	if err := w.persistAndPublish(job); err != nil {
		return
	}
	if job.IsTerminal() {
		w.mu.Lock()
		delete(w.jobs, key)
		w.mu.Unlock()
	}
}

func (w *Worker) persistAndPublish(job FacilityJob) error {
	if w.store != nil {
		if err := w.store.Save(context.Background(), job, w.workerID); err != nil {
			return err
		}
	}
	publishFacilityJob(w.publisher, job)
	return nil
}

func (w *Worker) dispatch(class FacilityJobClass) {
	defer w.wg.Done()
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-w.ctx.Done():
			return
		case <-w.wake:
		case <-ticker.C:
		}

		for {
			tasks := w.tasks.IDs()
			if len(tasks) == 0 {
				break
			}
			now := time.Now().UTC()
			job, claimed, err := w.store.ClaimNext(w.ctx, class, tasks, w.workerID, now, now.Add(facilityJobLeaseDuration))
			if err != nil || !claimed {
				break
			}
			w.runPersistedTask(job)
		}
	}
}

func (w *Worker) runPersistedTask(job FacilityJob) {
	key := facilityJobKey{ownerID: job.OwnerID, jobID: job.ID}
	w.mu.Lock()
	handler := w.tasks.handler(job.Task)
	job.Status = FacilityJobStatusRunning
	job.Stage = facilityJobStagePreparing
	job.Progress = max(1, job.Progress)
	job.UpdatedAt = time.Now().UTC()
	w.jobs[key] = job
	w.mu.Unlock()
	if handler == nil {
		return
	}
	if err := w.persistAndPublish(job); err != nil {
		w.fail(key, err)
		return
	}
	stopHeartbeat := w.startHeartbeat(key)
	defer stopHeartbeat()
	taskCtx, cancel := context.WithCancel(w.ctx)
	w.mu.Lock()
	w.cancels[key] = cancel
	w.mu.Unlock()
	defer func() {
		w.mu.Lock()
		delete(w.cancels, key)
		w.mu.Unlock()
		cancel()
	}()

	execution := FacilityJobExecution{Job: job, Reporter: facilityJobReporter{report: func(progress FacilityJobProgress) {
		w.reportTask(key, progress)
	}}}
	result, err := handler.Execute(taskCtx, execution)
	if err != nil {
		if errors.Is(err, context.Canceled) && taskCtx.Err() != nil {
			return
		}
		w.fail(key, err)
		return
	}
	w.mu.Lock()
	completed := w.jobs[key]
	completed.Result = result.Result
	w.jobs[key] = completed
	w.mu.Unlock()
	w.update(key, FacilityJobStatusCompleted, 100, facilityJobStageCompleted, "")
}

func (w *Worker) reportTask(key facilityJobKey, progress FacilityJobProgress) {
	w.mu.Lock()
	job, ok := w.jobs[key]
	if !ok || job.IsTerminal() {
		w.mu.Unlock()
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
	w.jobs[key] = job
	w.mu.Unlock()
	if err := w.persistAndPublish(job); err != nil {
		w.fail(key, err)
	}
}

// Retry requeues a failed, retryable job while preserving its checkpoint.
func (s *JobService) Retry(ctx context.Context, ownerID, jobID uuid.UUID) (FacilityJob, error) {
	if s == nil || s.store == nil {
		return FacilityJob{}, ErrFacilityJobNotFound
	}
	job, err := s.store.Retry(ctx, ownerID, jobID, time.Now().UTC())
	if err != nil {
		return FacilityJob{}, err
	}
	s.worker.signal()
	s.publish(job)
	return job, nil
}

// Cancel marks an owned queued or running job as cancelled. A local worker is
// signalled through its task context; workers in other processes cannot write
// a later terminal state because the durable store only accepts running jobs.
func (s *JobService) Cancel(ctx context.Context, ownerID, jobID uuid.UUID) (FacilityJob, error) {
	if s == nil || s.store == nil {
		return FacilityJob{}, ErrFacilityJobNotFound
	}
	job, err := s.store.Cancel(ctx, ownerID, jobID, time.Now().UTC())
	if err != nil {
		return FacilityJob{}, err
	}
	key := facilityJobKey{ownerID: ownerID, jobID: jobID}
	s.worker.cancelJob(key)
	s.publish(job)
	return job, nil
}

// Remove permanently clears an owned, terminal job and its execution metadata.
// It never affects the facility resource produced by a completed copy job.
func (s *JobService) Remove(ctx context.Context, ownerID, jobID uuid.UUID) error {
	if s == nil || s.store == nil {
		return ErrFacilityJobNotFound
	}
	return s.store.Delete(ctx, ownerID, jobID)
}

func (s *JobService) publish(job FacilityJob) {
	publishFacilityJob(s.publisher, job)
}

func (w *Worker) cancelJob(key facilityJobKey) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if cancel := w.cancels[key]; cancel != nil {
		cancel()
	}
	delete(w.jobs, key)
}

func publishFacilityJob(publisher apprealtime.FacilityJobProgressPublisher, job FacilityJob) {
	if publisher == nil {
		return
	}
	publisher.BroadcastFacilityJobProgress(context.Background(), apprealtime.FacilityJobProgressEvent{
		JobID: job.ID, OwnerID: job.OwnerID, Kind: string(job.Kind), Status: string(job.Status),
		JobType: string(job.Type), Class: string(job.Class), Progress: job.Progress, Stage: string(job.Stage),
		Error: job.Error, Processed: job.Processed, Total: job.Total, Succeeded: job.Succeeded, Failed: job.Failed, UpdatedAt: job.UpdatedAt,
	})
}
