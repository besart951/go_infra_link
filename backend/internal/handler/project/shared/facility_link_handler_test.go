package shared

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	domainproject "github.com/besart951/go_infra_link/backend/internal/domain/project"
	"github.com/besart951/go_infra_link/backend/internal/handler/middleware"
	facilityservice "github.com/besart951/go_infra_link/backend/internal/service/facility"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestFacilityLinkHandlerStartCopyPersistsProjectCommand(t *testing.T) {
	gin.SetMode(gin.TestMode)
	manager := newDurableCopyJobManager(t)

	projectID := uuid.New()
	actorID := uuid.New()
	operationID := uuid.New()
	sourceID := uuid.New()
	received := make(chan domainproject.FacilityCopyJobPayload, 1)
	manager.RegisterTask(domainproject.TaskCopyProjectControlCabinet, func(_ context.Context, job facilityservice.CopyJob, _ func(facilityservice.FacilityJobProgress)) (facilityservice.FacilityJobTaskResult, error) {
		var payload domainproject.FacilityCopyJobPayload
		if err := json.Unmarshal(job.Payload, &payload); err != nil {
			return facilityservice.FacilityJobTaskResult{}, err
		}
		received <- payload
		return facilityservice.FacilityJobTaskResult{}, nil
	})

	handler := NewFacilityLinkHandler(nil, nil)
	handler.ConfigureCopyJobs(manager)
	c, recorder := copyRequestContext(projectID, operationID, actorID)

	handler.StartCopy(c, ProjectCopyCommand{
		ProjectID: projectID,
		SourceID:  sourceID,
		Kind:      facilityservice.CopyJobKindControlCabinet,
		Task:      domainproject.TaskCopyProjectControlCabinet,
	})
	if recorder.Code != http.StatusAccepted {
		t.Fatalf("StartCopy() status = %d, body = %s", recorder.Code, recorder.Body.String())
	}

	select {
	case payload := <-received:
		if payload.ProjectID != projectID || payload.SourceID != sourceID {
			t.Fatalf("unexpected persisted payload %#v", payload)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("persisted project copy task was not dispatched")
	}
}

func TestFacilityLinkHandlerRejectsCopyWithoutDurableStore(t *testing.T) {
	manager := facilityservice.NewCopyJobManager(nil)
	t.Cleanup(manager.Close)
	handler := NewFacilityLinkHandler(nil, nil)
	handler.ConfigureCopyJobs(manager)
	c, recorder := copyRequestContext(uuid.New(), uuid.New(), uuid.New())

	handler.StartCopy(c, ProjectCopyCommand{Task: domainproject.TaskCopyProjectControlCabinet})
	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected status 503, got %d", recorder.Code)
	}
}

func newDurableCopyJobManager(t *testing.T) *facilityservice.CopyJobManager {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "facility-jobs.db")), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	if err := facilityservice.MigrateCopyJobs(db); err != nil {
		t.Fatalf("migrate jobs: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("database handle: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
	manager := facilityservice.NewCopyJobManagerWithDB(nil, db)
	t.Cleanup(manager.Close)
	return manager
}

func copyRequestContext(projectID, operationID, actorID uuid.UUID) (*gin.Context, *httptest.ResponseRecorder) {
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	request := httptest.NewRequest(http.MethodPost, "/projects/"+projectID.String()+"/copy", nil)
	request.Header.Set("X-Copy-Operation-ID", operationID.String())
	c.Request = request
	c.Set(middleware.ContextUserIDKey, actorID)
	return c, recorder
}
