package project

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

// stubPhaseService is a minimal in-test stub that implements PhaseService.
type stubPhaseService struct {
	deleteByIDErr error
}

func (s *stubPhaseService) Create(_ context.Context, _ *domainProject.Phase) error { return nil }
func (s *stubPhaseService) GetByID(_ context.Context, _ uuid.UUID) (*domainProject.Phase, error) {
	return nil, domain.ErrNotFound
}
func (s *stubPhaseService) List(_ context.Context, _, _ int, _ string) (*domain.PaginatedList[domainProject.Phase], error) {
	return nil, nil
}
func (s *stubPhaseService) Update(_ context.Context, _ *domainProject.Phase) error { return nil }
func (s *stubPhaseService) DeleteByID(_ context.Context, _ uuid.UUID) error {
	return s.deleteByIDErr
}

// setupDeletePhaseRouter wires up a gin router for DELETE /phases/:id backed by the given service.
func setupDeletePhaseRouter(svc PhaseService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewPhaseHandler(svc)
	r.DELETE("/phases/:id", h.DeletePhase)
	return r
}

// TestDeletePhaseServiceErrorReturnsSingleErrorResponse is the regression test for the
// double-response bug: when DeleteByID fails the handler must respond with exactly one
// 500 response and must NOT also write a 204.
func TestDeletePhaseServiceErrorReturnsSingleErrorResponse(t *testing.T) {
	svc := &stubPhaseService{deleteByIDErr: errors.New("db error")}
	router := setupDeletePhaseRouter(svc)

	req := httptest.NewRequest(http.MethodDelete, "/phases/"+uuid.New().String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.Code)
	}
}

// TestDeletePhaseSuccessReturns204 verifies the happy path still works.
func TestDeletePhaseSuccessReturns204(t *testing.T) {
	svc := &stubPhaseService{deleteByIDErr: nil}
	router := setupDeletePhaseRouter(svc)

	req := httptest.NewRequest(http.MethodDelete, "/phases/"+uuid.New().String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("expected status 204, got %d", w.Code)
	}
}
