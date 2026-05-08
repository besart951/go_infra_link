package handlerutil

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	dto "github.com/besart951/go_infra_link/backend/internal/handler/dto/common"
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

func TestRespondMappedDomainErrorUsesDefaultDomainMapping(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)

	handled := RespondMappedDomainError(context, fmtWrapped(domain.ErrConflict))

	if !handled {
		t.Fatal("expected default conflict mapping to be handled")
	}
	if recorder.Code != http.StatusConflict {
		t.Fatalf("expected status 409, got %d", recorder.Code)
	}

	var response dto.ErrorResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("expected response json to decode, got %v", err)
	}
	if response.Code != "conflict" || response.Error != "conflict" || response.LocalizedKey != "errors.conflict" {
		t.Fatalf("expected unified conflict response, got %+v", response)
	}
}

func TestRespondMappedDomainErrorLetsCustomMappingOverrideDefault(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)

	handled := RespondMappedDomainError(
		context,
		fmtWrapped(domain.ErrNotFound),
		MapError(domain.ErrNotFound, PlainError(http.StatusBadRequest, "invalid_reference", "invalid reference")),
	)

	if !handled {
		t.Fatal("expected custom mapping to be handled")
	}
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", recorder.Code)
	}
	var response dto.ErrorResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("expected response json to decode, got %v", err)
	}
	if response.Code != "invalid_reference" {
		t.Fatalf("expected custom code, got %+v", response)
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
