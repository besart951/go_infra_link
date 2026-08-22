package handlerutil

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func TestParseIdempotencyKey(t *testing.T) {
	gin.SetMode(gin.TestMode)
	expected := uuid.New()
	context, _ := gin.CreateTestContext(httptest.NewRecorder())
	context.Request = httptest.NewRequest("POST", "/", nil)
	context.Request.Header.Set(IdempotencyKeyHeader, expected.String())

	actual, err := ParseIdempotencyKey(context)
	if err != nil || actual != expected {
		t.Fatalf("idempotency key = %s, %v", actual, err)
	}
}

func TestParseIdempotencyKeyRejectsInvalidValue(t *testing.T) {
	context, _ := gin.CreateTestContext(httptest.NewRecorder())
	context.Request = httptest.NewRequest("POST", "/", nil)
	context.Request.Header.Set(IdempotencyKeyHeader, "not-a-uuid")
	if _, err := ParseIdempotencyKey(context); err == nil {
		t.Fatal("expected invalid idempotency key to fail")
	}
}
