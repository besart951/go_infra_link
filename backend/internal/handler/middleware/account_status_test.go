package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	domainUser "github.com/besart951/go_infra_link/backend/internal/domain/user"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type stubUserStatusService struct {
	user *domainUser.User
	err  error
}

func (s stubUserStatusService) GetByID(ctx context.Context, id uuid.UUID) (*domainUser.User, error) {
	return s.user, s.err
}

func TestAccountStatusGuardSuppressesCanceledRequestError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	ginContext, _ := gin.CreateTestContext(recorder)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	ctx, cancel := context.WithCancel(req.Context())
	cancel()
	ginContext.Request = req.WithContext(ctx)
	ginContext.Set(ContextUserIDKey, uuid.New())

	guard := AccountStatusGuard(stubUserStatusService{err: context.Canceled})
	guard(ginContext)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected no response to be written, got %d", recorder.Code)
	}
	if recorder.Body.Len() != 0 {
		t.Fatalf("expected empty body, got %q", recorder.Body.String())
	}
	if !ginContext.IsAborted() {
		t.Fatal("expected request to be aborted")
	}
}

func TestAccountStatusGuardReturnsInternalServerErrorForNonCanceledError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	ginContext, _ := gin.CreateTestContext(recorder)
	ginContext.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	ginContext.Set(ContextUserIDKey, uuid.New())

	guard := AccountStatusGuard(stubUserStatusService{err: errors.New("db unavailable")})
	guard(ginContext)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", recorder.Code)
	}
	if !ginContext.IsAborted() {
		t.Fatal("expected request to be aborted")
	}
}
