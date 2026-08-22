package facility

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestPersistedTaskRunsAfterManagerRestartWithoutResubmission(t *testing.T) {
	db := openFacilityJobTestDB(t)
	ownerID := uuid.New()
	jobID := uuid.New()
	first := NewFacilityJobManagerWithDB(nil, db)
	if _, err := first.SubmitTask(t.Context(), FacilityJob{
		ID: jobID, OwnerID: ownerID, Kind: FacilityJobKindFieldDevice,
		Class: FacilityJobClassExport, Type: FacilityJobTypeExport,
		Task: "test.export.v1", Payload: json.RawMessage(`{"value":42}`),
	}); err != nil {
		t.Fatalf("SubmitTask() error = %v", err)
	}
	first.Close()

	second := NewFacilityJobManagerWithDB(nil, db)
	t.Cleanup(second.Close)
	runs := make(chan int, 1)
	second.RegisterTask("test.export.v1", FacilityJobHandlerFunc(func(_ context.Context, execution FacilityJobExecution) (FacilityJobTaskResult, error) {
		job, report := execution.Job, execution.Reporter.Report
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
	}))

	select {
	case got := <-runs:
		if got != 42 {
			t.Fatalf("payload value = %d, want 42", got)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("persisted task was not dispatched after restart")
	}
	completed := waitForFacilityJobStatus(t, second, ownerID, jobID, FacilityJobStatusCompleted)
	if completed.Processed != 1 || string(completed.Result) != `{"done":true}` {
		t.Fatalf("completed job = %#v", completed)
	}
}

func TestExportConcurrencyLimitAndRetry(t *testing.T) {
	db := openFacilityJobTestDB(t)
	manager := NewFacilityJobManagerWithDB(nil, db)
	t.Cleanup(manager.Close)
	ownerID := uuid.New()
	for range 2 {
		if _, err := manager.SubmitTask(t.Context(), FacilityJob{
			ID: uuid.New(), OwnerID: ownerID, Kind: FacilityJobKindFieldDevice,
			Class: FacilityJobClassExport, Type: FacilityJobTypeExport, Task: "unregistered.v1",
		}); err != nil {
			t.Fatalf("SubmitTask() error = %v", err)
		}
	}
	if _, err := manager.SubmitTask(t.Context(), FacilityJob{
		ID: uuid.New(), OwnerID: ownerID, Kind: FacilityJobKindFieldDevice,
		Class: FacilityJobClassExport, Type: FacilityJobTypeExport, Task: "unregistered.v1",
	}); !errors.Is(err, ErrFacilityJobLimit) {
		t.Fatalf("third export error = %v, want %v", err, ErrFacilityJobLimit)
	}
}

func TestMigrateFacilityJobsCreatesChunkAndMappingTables(t *testing.T) {
	db := openFacilityJobTestDB(t)
	for _, table := range []any{facilityJobRecord{}, facilityJobItemRecord{}, facilityJobIDMappingRecord{}, facilityAggregateLifecycleRecord{}} {
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

func TestFacilityJobJSONFieldsPersistAsJSONBPostgres(t *testing.T) {
	dsn := os.Getenv("FACILITY_BENCHMARK_DSN")
	if dsn == "" {
		t.Skip("FACILITY_BENCHMARK_DSN is not configured")
	}
	database, err := gorm.Open(postgres.New(postgres.Config{DSN: dsn, PreferSimpleProtocol: true}), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	manager := NewFacilityJobManagerWithDB(nil, database)
	t.Cleanup(manager.Close)
	job := FacilityJob{
		OwnerID: uuid.New(), ID: uuid.New(), Kind: FacilityJobKindFieldDevice,
		Class: FacilityJobClassExport, Type: FacilityJobTypeExport, Task: "test.export.v1",
		Payload: json.RawMessage(`{"search":"pump"}`),
	}
	if _, err := manager.SubmitTask(t.Context(), job); err != nil {
		t.Fatalf("persist JSONB job fields: %v", err)
	}
}

func TestDeleteJobAdmissionLocksAggregateAtomically(t *testing.T) {
	db := openFacilityJobTestDB(t)
	if err := db.Exec("CREATE TABLE control_cabinets (id text PRIMARY KEY, version integer NOT NULL)").Error; err != nil {
		t.Fatalf("create aggregate table: %v", err)
	}
	resourceID := uuid.New()
	if err := db.Exec("INSERT INTO control_cabinets (id, version) VALUES (?, 1)", resourceID).Error; err != nil {
		t.Fatalf("seed aggregate: %v", err)
	}

	manager := NewFacilityJobManagerWithDB(nil, db)
	t.Cleanup(manager.Close)
	job := admittedDeleteJob(uuid.New(), uuid.New(), resourceID)
	if _, err := manager.SubmitTask(t.Context(), job); err != nil {
		t.Fatalf("submit admitted job: %v", err)
	}
	if _, err := manager.SubmitTask(t.Context(), job); err != nil {
		t.Fatalf("repeat idempotent admission: %v", err)
	}

	conflicting := admittedDeleteJob(uuid.New(), uuid.New(), resourceID)
	if _, err := manager.SubmitTask(t.Context(), conflicting); !errors.Is(err, ErrAggregateLocked) {
		t.Fatalf("conflicting admission error = %v, want %v", err, ErrAggregateLocked)
	}
	var count int64
	if err := db.Model(&facilityAggregateLifecycleRecord{}).Count(&count).Error; err != nil || count != 1 {
		t.Fatalf("lifecycle locks = %d, error = %v; want one", count, err)
	}
}

func TestDeleteJobAdmissionRejectsMissingAggregateWithoutCreatingJob(t *testing.T) {
	db := openFacilityJobTestDB(t)
	if err := db.Exec("CREATE TABLE control_cabinets (id text PRIMARY KEY, version integer NOT NULL)").Error; err != nil {
		t.Fatalf("create aggregate table: %v", err)
	}
	manager := NewFacilityJobManagerWithDB(nil, db)
	t.Cleanup(manager.Close)

	job := admittedDeleteJob(uuid.New(), uuid.New(), uuid.New())
	if _, err := manager.SubmitTask(t.Context(), job); !errors.Is(err, ErrAggregateNotFound) {
		t.Fatalf("missing aggregate error = %v, want %v", err, ErrAggregateNotFound)
	}
	var count int64
	if err := db.Model(&facilityJobRecord{}).Count(&count).Error; err != nil || count != 0 {
		t.Fatalf("jobs = %d, error = %v; want zero", count, err)
	}
}

func admittedDeleteJob(ownerID, jobID, resourceID uuid.UUID) FacilityJob {
	return FacilityJob{
		ID: jobID, OwnerID: ownerID, Kind: FacilityJobKindControlCabinet,
		Class: FacilityJobClassMutation, Type: FacilityJobTypeDelete,
		Task: FacilityJobTaskDeleteControlCabinet,
		Admission: &FacilityAggregateAdmission{
			ResourceID: resourceID, BaseVersion: 1, State: FacilityAggregateStateDeleting,
		},
	}
}

func openFacilityJobTestDB(t *testing.T) *gorm.DB {
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
	if err := MigrateFacilityJobs(db); err != nil {
		t.Fatalf("migrate facility jobs: %v", err)
	}
	return db
}

func waitForFacilityJobStatus(t *testing.T, manager *FacilityJobManager, ownerID, jobID uuid.UUID, want FacilityJobStatus) FacilityJob {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		job, err := manager.Get(ownerID, jobID)
		if err == nil && job.Status == want {
			return job
		}
		time.Sleep(5 * time.Millisecond)
	}
	job, err := manager.Get(ownerID, jobID)
	t.Fatalf("job status = %#v, error = %v; want %s", job, err, want)
	return FacilityJob{}
}
