package facility

import (
	"context"
	"encoding/json"
	"errors"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestPersistedTaskRunsAfterManagerRestartWithoutResubmission(t *testing.T) {
	db := openCopyJobTestDB(t)
	ownerID := uuid.New()
	jobID := uuid.New()
	first := NewCopyJobManagerWithDB(nil, db)
	if _, err := first.SubmitTask(t.Context(), CopyJob{
		ID: jobID, OwnerID: ownerID, Kind: CopyJobKindFieldDevice,
		Class: FacilityJobClassExport, Type: FacilityJobTypeExport,
		Task: "test.export.v1", Payload: json.RawMessage(`{"value":42}`),
	}); err != nil {
		t.Fatalf("SubmitTask() error = %v", err)
	}
	first.Close()

	second := NewCopyJobManagerWithDB(nil, db)
	t.Cleanup(second.Close)
	runs := make(chan int, 1)
	second.RegisterTask("test.export.v1", func(_ context.Context, job CopyJob, report func(FacilityJobProgress)) (FacilityJobTaskResult, error) {
		var payload struct {
			Value int `json:"value"`
		}
		if err := json.Unmarshal(job.Payload, &payload); err != nil {
			return FacilityJobTaskResult{}, err
		}
		total := int64(1)
		report(FacilityJobProgress{Progress: 50, Stage: "generating", Processed: 1, Total: &total})
		runs <- payload.Value
		return FacilityJobTaskResult{Result: json.RawMessage(`{"done":true}`)}, nil
	})

	select {
	case got := <-runs:
		if got != 42 {
			t.Fatalf("payload value = %d, want 42", got)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("persisted task was not dispatched after restart")
	}
	completed := waitForCopyJobStatus(t, second, ownerID, jobID, CopyJobStatusCompleted)
	if completed.Processed != 1 || string(completed.Result) != `{"done":true}` {
		t.Fatalf("completed job = %#v", completed)
	}
}

func TestExportConcurrencyLimitAndRetry(t *testing.T) {
	db := openCopyJobTestDB(t)
	manager := NewCopyJobManagerWithDB(nil, db)
	t.Cleanup(manager.Close)
	ownerID := uuid.New()
	for range 2 {
		if _, err := manager.SubmitTask(t.Context(), CopyJob{
			ID: uuid.New(), OwnerID: ownerID, Kind: CopyJobKindFieldDevice,
			Class: FacilityJobClassExport, Type: FacilityJobTypeExport, Task: "unregistered.v1",
		}); err != nil {
			t.Fatalf("SubmitTask() error = %v", err)
		}
	}
	if _, err := manager.SubmitTask(t.Context(), CopyJob{
		ID: uuid.New(), OwnerID: ownerID, Kind: CopyJobKindFieldDevice,
		Class: FacilityJobClassExport, Type: FacilityJobTypeExport, Task: "unregistered.v1",
	}); !errors.Is(err, ErrFacilityJobLimit) {
		t.Fatalf("third export error = %v, want %v", err, ErrFacilityJobLimit)
	}
}

func TestDurableCopyJobSurvivesManagerRestart(t *testing.T) {
	db := openCopyJobTestDB(t)
	ownerID := uuid.New()
	jobID := uuid.New()

	first := NewCopyJobManagerWithDB(nil, db)
	if _, err := first.Start(ownerID, jobID, CopyJobKindControlCabinet, func(context.Context) error {
		return nil
	}); err != nil {
		t.Fatalf("start durable job: %v", err)
	}
	waitForCopyJobStatus(t, first, ownerID, jobID, CopyJobStatusCompleted)
	first.Close()

	second := NewCopyJobManagerWithDB(nil, db)
	t.Cleanup(second.Close)
	job, err := second.Get(ownerID, jobID)
	if err != nil {
		t.Fatalf("get job after restart: %v", err)
	}
	if job.Status != CopyJobStatusCompleted || job.Progress != 100 {
		t.Fatalf("persisted job = %#v, want completed at 100%%", job)
	}

	var duplicateRuns atomic.Int32
	duplicate, err := second.Start(ownerID, jobID, CopyJobKindControlCabinet, func(context.Context) error {
		duplicateRuns.Add(1)
		return nil
	})
	if err != nil {
		t.Fatalf("retry completed job: %v", err)
	}
	if duplicate.ID != jobID || duplicate.Status != CopyJobStatusCompleted {
		t.Fatalf("retry returned %#v", duplicate)
	}
	time.Sleep(20 * time.Millisecond)
	if duplicateRuns.Load() != 0 {
		t.Fatal("completed idempotent job ran twice")
	}
}

func TestDurableCopyJobCanBeReclaimedAfterLeaseExpires(t *testing.T) {
	db := openCopyJobTestDB(t)
	ownerID := uuid.New()
	jobID := uuid.New()
	started := make(chan struct{})

	first := NewCopyJobManagerWithDB(nil, db)
	if _, err := first.Start(ownerID, jobID, CopyJobKindSPSController, func(ctx context.Context) error {
		close(started)
		<-ctx.Done()
		return ctx.Err()
	}); err != nil {
		t.Fatalf("start durable job: %v", err)
	}
	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("first worker did not start")
	}
	first.Close()

	if err := db.Model(&copyJobRecord{}).
		Where("owner_id = ? AND id = ?", ownerID, jobID).
		Update("lease_until", time.Now().UTC().Add(-time.Minute)).Error; err != nil {
		t.Fatalf("expire lease: %v", err)
	}

	second := NewCopyJobManagerWithDB(nil, db)
	t.Cleanup(second.Close)
	resumed := make(chan struct{})
	if _, err := second.Start(ownerID, jobID, CopyJobKindSPSController, func(context.Context) error {
		close(resumed)
		return nil
	}); err != nil {
		t.Fatalf("resume durable job: %v", err)
	}
	select {
	case <-resumed:
	case <-time.After(time.Second):
		t.Fatal("expired job was not reclaimed")
	}
	waitForCopyJobStatus(t, second, ownerID, jobID, CopyJobStatusCompleted)

	var record copyJobRecord
	if err := db.Where("owner_id = ? AND id = ?", ownerID, jobID).First(&record).Error; err != nil {
		t.Fatalf("load reclaimed job: %v", err)
	}
	if record.Attempts != 2 {
		t.Fatalf("attempts = %d, want 2", record.Attempts)
	}
}

func TestMigrateCopyJobsCreatesChunkAndMappingTables(t *testing.T) {
	db := openCopyJobTestDB(t)
	for _, table := range []any{copyJobRecord{}, facilityJobItemRecord{}, facilityJobIDMappingRecord{}} {
		if !db.Migrator().HasTable(table) {
			t.Fatalf("missing migrated table for %T", table)
		}
	}
	sharedJobID := uuid.New()
	now := time.Now().UTC()
	for _, ownerID := range []uuid.UUID{uuid.New(), uuid.New()} {
		if err := db.Create(&facilityJobItemRecord{
			OwnerID: ownerID, JobID: sharedJobID, Ordinal: 0, Status: "queued", Input: []byte(`{}`), CreatedAt: now, UpdatedAt: now,
		}).Error; err != nil {
			t.Fatalf("create owner-scoped job item: %v", err)
		}
		if err := db.Create(&facilityJobIDMappingRecord{
			OwnerID: ownerID, JobID: sharedJobID, EntityType: "field_device", SourceID: uuid.New(), TargetID: uuid.New(), CreatedAt: now,
		}).Error; err != nil {
			t.Fatalf("create owner-scoped job mapping: %v", err)
		}
	}
}

func openCopyJobTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "facility-jobs.db")), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("access sqlite connection: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
	if err := MigrateCopyJobs(db); err != nil {
		t.Fatalf("migrate copy jobs: %v", err)
	}
	return db
}

func waitForCopyJobStatus(t *testing.T, manager *CopyJobManager, ownerID, jobID uuid.UUID, want CopyJobStatus) CopyJob {
	t.Helper()
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		job, err := manager.Get(ownerID, jobID)
		if err == nil && job.Status == want {
			return job
		}
		time.Sleep(5 * time.Millisecond)
	}
	job, err := manager.Get(ownerID, jobID)
	t.Fatalf("job status = %#v, error = %v; want %s", job, err, want)
	return CopyJob{}
}
