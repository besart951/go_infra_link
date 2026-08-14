package changes

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	domainProject "github.com/besart951/go_infra_link/backend/internal/domain/project"
	domainUser "github.com/besart951/go_infra_link/backend/internal/domain/user"
	dto "github.com/besart951/go_infra_link/backend/internal/handler/dto/project"
	"github.com/besart951/go_infra_link/backend/internal/handler/middleware"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type serviceFake struct {
	projectID uuid.UUID
	after     uint64
	limit     int
	page      *domainProject.ChangePage
}

func (f *serviceFake) ListAfter(_ context.Context, projectID uuid.UUID, after uint64, limit int) (*domainProject.ChangePage, error) {
	f.projectID, f.after, f.limit = projectID, after, limit
	return f.page, nil
}

type accessFake struct{}

func (accessFake) CanAccessProject(context.Context, uuid.UUID, uuid.UUID, *domainUser.Role) (bool, error) {
	return true, nil
}
func (accessFake) CanUseProjectPermission(context.Context, uuid.UUID, *domainUser.Role, string) (bool, error) {
	return true, nil
}
func (accessFake) CanUseProjectPermissionForProject(context.Context, uuid.UUID, uuid.UUID, *domainUser.Role, string) (bool, error) {
	return true, nil
}

func TestListReturnsV2ProjectChangeContract(t *testing.T) {
	gin.SetMode(gin.TestMode)
	projectID, eventID, aggregateID := uuid.New(), uuid.New(), uuid.New()
	service := &serviceFake{page: &domainProject.ChangePage{
		ProjectID: projectID, CurrentRevision: 9, HasMore: true,
		Events: []domainProject.Change{{
			EventID: eventID, ProjectID: projectID, Revision: 8,
			AggregateType: "field_device", AggregateID: &aggregateID,
			Action: domainProject.ChangeUpdated, ChangedFields: []string{"description"},
			ParentRefs: map[string]uuid.UUID{"project_id": projectID}, OccurredAt: time.Now().UTC(),
		}},
	}}

	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Params = gin.Params{{Key: "id", Value: projectID.String()}}
	c.Request = httptest.NewRequest(http.MethodGet, "/projects/"+projectID.String()+"/changes?after_revision=7&limit=1", nil)
	c.Set(middleware.ContextUserIDKey, uuid.New())

	NewHandler(accessFake{}, service).List(c)
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", recorder.Code, recorder.Body.String())
	}
	var response dto.ProjectChangesResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if service.after != 7 || service.limit != 1 || response.CurrentRevision != 9 || len(response.Events) != 1 {
		t.Fatalf("unexpected response/service call: response=%+v after=%d limit=%d", response, service.after, service.limit)
	}
	if response.Events[0].EventID != eventID || response.Events[0].AggregateType != "field_device" {
		t.Fatalf("unexpected event: %+v", response.Events[0])
	}
}
