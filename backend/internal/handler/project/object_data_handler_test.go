package project

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/besart951/go_infra_link/backend/internal/handler/middleware"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func TestListProjectObjectDataReturnsAfterInvalidApparatID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	projectID := uuid.New()
	userID := uuid.New()
	accessService := &fakeProjectAccessPolicyService{hasAccess: true}
	handler := NewProjectHandler(nil, accessService, nil, nil)

	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	tracker := &statusTrackingWriter{ResponseWriter: context.Writer}
	context.Writer = tracker
	context.Request = httptest.NewRequest(http.MethodGet, "/projects/"+projectID.String()+"/object-data?apparat_id=not-a-uuid", nil)
	context.Params = gin.Params{{Key: "id", Value: projectID.String()}}
	context.Set(middleware.ContextUserIDKey, userID)

	handler.ListProjectObjectData(context)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d body=%s", recorder.Code, recorder.Body.String())
	}
	if accessService.calls != 1 {
		t.Fatalf("expected access check to run once, got %d call(s)", accessService.calls)
	}
	if accessService.lastRequesterID != userID || accessService.lastProjectID != projectID {
		t.Fatalf("expected access check for requester=%s project=%s, got requester=%s project=%s", userID, projectID, accessService.lastRequesterID, accessService.lastProjectID)
	}
	if len(tracker.statusWrites) != 1 {
		t.Fatalf("expected exactly one status write, got %v", tracker.statusWrites)
	}
	if tracker.statusWrites[0] != http.StatusBadRequest {
		t.Fatalf("expected only status write to be 400, got %v", tracker.statusWrites)
	}
}

type fakeProjectAccessPolicyService struct {
	hasAccess       bool
	err             error
	calls           int
	lastRequesterID uuid.UUID
	lastProjectID   uuid.UUID
}

func (s *fakeProjectAccessPolicyService) CanAccessProject(_ context.Context, requesterID, projectID uuid.UUID) (bool, error) {
	s.calls++
	s.lastRequesterID = requesterID
	s.lastProjectID = projectID
	if s.err != nil {
		return false, s.err
	}
	return s.hasAccess, nil
}
