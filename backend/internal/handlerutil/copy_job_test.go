package handlerutil

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func TestParseCopyOperationID(t *testing.T) {
	validID := uuid.New()
	tests := []struct {
		name    string
		header  string
		wantErr bool
		wantID  uuid.UUID
	}{
		{name: "missing header creates an ID"},
		{name: "valid client operation ID", header: validID.String(), wantID: validID},
		{name: "invalid operation ID", header: "not-a-uuid", wantErr: true},
		{name: "nil operation ID", header: uuid.Nil.String(), wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
			ctx.Request = httptest.NewRequest("POST", "/", nil)
			if tt.header != "" {
				ctx.Request.Header.Set(CopyOperationIDHeader, tt.header)
			}

			gotID, err := ParseCopyOperationID(ctx)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ParseCopyOperationID() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if tt.wantID != uuid.Nil && gotID != tt.wantID {
				t.Fatalf("ParseCopyOperationID() = %s, want %s", gotID, tt.wantID)
			}
			if gotID == uuid.Nil {
				t.Fatal("ParseCopyOperationID() returned nil UUID")
			}
		})
	}
}
