package handlerutil

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	dto "github.com/besart951/go_infra_link/backend/internal/handler/dto/common"
	"github.com/gin-gonic/gin"
)

func TestRespondErrorWritesUnifiedShapeWithCompatibilityAliases(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	context.Request.Header.Set("X-Request-ID", "req-123")

	RespondErrorWithLocalizedKey(context, http.StatusConflict, "conflict", "Already exists", "errors.conflict")

	var response dto.ErrorResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("expected error response to decode, got %v", err)
	}
	if recorder.Code != http.StatusConflict {
		t.Fatalf("expected status 409, got %d", recorder.Code)
	}
	if response.Error != "conflict" || response.Code != "conflict" {
		t.Fatalf("expected code and compatibility error to match, got %+v", response)
	}
	if response.LocalizedKey != "errors.conflict" || response.RequestID != "req-123" {
		t.Fatalf("expected localized key and request id, got %+v", response)
	}
}

func TestBindJSONReportsNestedFieldPaths(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"data":[{"items":[{}]}]}`))
	context.Request.Header.Set("Content-Type", "application/json")

	var request nestedValidationRequest
	if BindJSON(context, &request) {
		t.Fatal("expected binding to fail")
	}

	var response dto.ErrorResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("expected validation response to decode, got %v", err)
	}
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", recorder.Code)
	}
	if response.Fields["data[0].items[0].name"] != "is required" {
		t.Fatalf("expected nested compatibility field path, got %+v", response.Fields)
	}
	if len(response.FieldErrors) != 1 || response.FieldErrors[0].Path != "data[0].items[0].name" {
		t.Fatalf("expected nested field_errors path, got %+v", response.FieldErrors)
	}
}

func TestRespondErrorDoesNotWriteTwice(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = httptest.NewRequest(http.MethodGet, "/", nil)

	RespondError(context, http.StatusBadRequest, "bad", "bad request")
	RespondError(context, http.StatusInternalServerError, "internal", "internal error")

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected first status to win, got %d", recorder.Code)
	}
	var response dto.ErrorResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("expected error response to decode, got %v", err)
	}
	if response.Code != "bad" {
		t.Fatalf("expected first response to remain, got %+v", response)
	}
}

type nestedValidationRequest struct {
	Data []nestedValidationEntry `json:"data" binding:"required,dive"`
}

type nestedValidationEntry struct {
	Items []nestedValidationItem `json:"items" binding:"required,dive"`
}

type nestedValidationItem struct {
	Name string `json:"name" binding:"required"`
}
