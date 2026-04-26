package handlerutil

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	"github.com/gin-gonic/gin"
)

func TestRespondDomainErrorUsesValidationError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)

	handled := RespondDomainError(
		context,
		domain.NewValidationError().Add("email", "is required"),
		PlainError(http.StatusInternalServerError, "internal", "internal error"),
	)

	if !handled {
		t.Fatal("expected validation error to be handled")
	}
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", recorder.Code)
	}
}

func TestRespondDomainErrorMatchesMappedError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)

	handled := RespondDomainError(
		context,
		fmtWrapped(domain.ErrNotFound),
		PlainError(http.StatusInternalServerError, "internal", "internal error"),
		MapError(domain.ErrNotFound, PlainError(http.StatusNotFound, "not_found", "missing")),
	)

	if !handled {
		t.Fatal("expected mapped error to be handled")
	}
	if recorder.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", recorder.Code)
	}
}

func TestRespondMappedDomainErrorDoesNotWriteFallback(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)

	handled := RespondMappedDomainError(
		context,
		errors.New("unmapped"),
		MapError(domain.ErrNotFound, PlainError(http.StatusNotFound, "not_found", "missing")),
	)

	if handled {
		t.Fatal("expected unmapped error not to be handled")
	}
	if recorder.Code != http.StatusOK {
		t.Fatalf("expected recorder to remain unwritten, got %d", recorder.Code)
	}
}

func TestRespondMappedDomainErrorWritesMappedError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)

	handled := RespondMappedDomainError(
		context,
		fmtWrapped(domain.ErrNotFound),
		MapError(domain.ErrNotFound, PlainError(http.StatusNotFound, "not_found", "missing")),
	)

	if !handled {
		t.Fatal("expected mapped error to be handled")
	}
	if recorder.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", recorder.Code)
	}
}

func TestRespondErrorSuppressesCanceledRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	ginContext, _ := gin.CreateTestContext(recorder)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	ctx, cancel := context.WithCancel(req.Context())
	cancel()
	ginContext.Request = req.WithContext(ctx)

	RespondError(ginContext, http.StatusInternalServerError, "fetch_failed", "failed")

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected no response to be written, got %d", recorder.Code)
	}
	if recorder.Body.Len() != 0 {
		t.Fatalf("expected empty body, got %q", recorder.Body.String())
	}
}

func fmtWrapped(err error) error {
	return errors.Join(err)
}
