package phase

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainProject "github.com/besart951/go_infra_link/backend/internal/domain/project"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func TestDeletePhaseReturnsAfterErrorResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)

	phaseID := uuid.New()
	handler := NewHandler(&fakePhaseService{deleteErr: errors.New("delete failed")})

	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	tracker := &statusTrackingWriter{ResponseWriter: context.Writer}
	context.Writer = tracker
	context.Request = httptest.NewRequest(http.MethodDelete, "/phases/"+phaseID.String(), nil)
	context.Params = gin.Params{{Key: "id", Value: phaseID.String()}}

	handler.DeletePhase(context)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", recorder.Code)
	}
	if len(tracker.statusWrites) != 1 {
		t.Fatalf("expected exactly one status write, got %v", tracker.statusWrites)
	}
	if tracker.statusWrites[0] != http.StatusInternalServerError {
		t.Fatalf("expected only status write to be 500, got %v", tracker.statusWrites)
	}
}

type statusTrackingWriter struct {
	gin.ResponseWriter
	statusWrites []int
}

func (w *statusTrackingWriter) WriteHeader(code int) {
	w.statusWrites = append(w.statusWrites, code)
	w.ResponseWriter.WriteHeader(code)
}

type fakePhaseService struct {
	deleteErr error
}

func (s *fakePhaseService) Create(context.Context, *domainProject.Phase) error {
	return nil
}

func (s *fakePhaseService) GetByID(context.Context, uuid.UUID) (*domainProject.Phase, error) {
	return nil, nil
}

func (s *fakePhaseService) List(context.Context, int, int, string) (*domain.PaginatedList[domainProject.Phase], error) {
	return nil, nil
}

func (s *fakePhaseService) Update(context.Context, *domainProject.Phase) error {
	return nil
}

func (s *fakePhaseService) DeleteByID(context.Context, uuid.UUID) error {
	return s.deleteErr
}
