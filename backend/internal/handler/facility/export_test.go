package facility

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	domainexport "github.com/besart951/go_infra_link/backend/internal/domain/exporting"
	domainuser "github.com/besart951/go_infra_link/backend/internal/domain/user"
	"github.com/besart951/go_infra_link/backend/internal/handler/middleware"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type exportServiceStub struct {
	job domainexport.Job
}

func (s exportServiceStub) Create(context.Context, uuid.UUID, uuid.UUID, domainexport.Request) (domainexport.Job, error) {
	return s.job, nil
}

func (s exportServiceStub) Get(context.Context, uuid.UUID, uuid.UUID) (domainexport.Job, error) {
	return s.job, nil
}

type exportAuthorizerStub struct {
	allowed       bool
	authorization domainexport.DownloadAuthorization
}

func (s *exportAuthorizerStub) CanDownload(_ context.Context, authorization domainexport.DownloadAuthorization) (bool, error) {
	s.authorization = authorization
	return s.allowed, nil
}

func TestDownloadExportRechecksPersistedScope(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ownerID := uuid.New()
	jobID := uuid.New()
	projectID := uuid.New()
	authorizer := &exportAuthorizerStub{}
	handler := NewExportHandler(exportServiceStub{job: domainexport.Job{
		ID: jobID, Status: domainexport.StatusCompleted,
		Scope: domainexport.Scope{Kind: domainexport.AccessScopeProject, ProjectIDs: []uuid.UUID{projectID}},
	}}, authorizer)

	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodGet, "/facility/jobs/"+jobID.String()+"/download", nil)
	c.Params = gin.Params{{Key: "id", Value: jobID.String()}}
	c.Set(middleware.ContextUserIDKey, ownerID)
	c.Set(middleware.ContextUserRoleKey, domainuser.Role("project_member"))

	handler.DownloadExport(c)

	if recorder.Code != http.StatusForbidden {
		t.Fatalf("status = %d, body = %s", recorder.Code, recorder.Body.String())
	}
	if authorizer.authorization.RequesterID != ownerID || len(authorizer.authorization.Scope.ProjectIDs) != 1 || authorizer.authorization.Scope.ProjectIDs[0] != projectID {
		t.Fatalf("unexpected authorization %#v", authorizer.authorization)
	}
}
