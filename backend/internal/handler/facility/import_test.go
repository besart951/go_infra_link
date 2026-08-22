package facility

import (
	"bytes"
	"context"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	fielddeviceimport "github.com/besart951/go_infra_link/backend/internal/application/fielddeviceimport"
	"github.com/besart951/go_infra_link/backend/internal/handler/middleware"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type importServiceStub struct {
	result  fielddeviceimport.Result
	err     error
	ownerID uuid.UUID
}

func (s *importServiceStub) Import(_ context.Context, command fielddeviceimport.Command) (fielddeviceimport.Result, error) {
	s.ownerID = command.OwnerID
	return s.result, s.err
}

func TestImportFieldDevicesReturnsValidationReport(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ownerID := uuid.New()
	service := &importServiceStub{
		result: fielddeviceimport.Result{Issues: []fielddeviceimport.Issue{{Code: "missing_owner"}}},
		err:    fielddeviceimport.ErrInvalidWorkbook,
	}
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = multipartImportRequest(t)
	ctx.Set(middleware.ContextUserIDKey, ownerID)

	NewImportHandler(service).ImportFieldDevices(ctx)

	if recorder.Code != http.StatusUnprocessableEntity || service.ownerID != ownerID {
		t.Fatalf("status=%d owner=%s body=%s", recorder.Code, service.ownerID, recorder.Body.String())
	}
}

func TestImportFieldDevicesRejectsUnexpectedFailure(t *testing.T) {
	gin.SetMode(gin.TestMode)
	service := &importServiceStub{err: errors.New("database unavailable")}
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = multipartImportRequest(t)
	ctx.Set(middleware.ContextUserIDKey, uuid.New())

	NewImportHandler(service).ImportFieldDevices(ctx)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
	}
}

func multipartImportRequest(t *testing.T) *http.Request {
	t.Helper()
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("file", "facility.xlsx")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := part.Write([]byte("workbook")); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest(http.MethodPost, "/api/v1/facility/imports/field-devices", &body)
	request.Header.Set("Content-Type", writer.FormDataContentType())
	return request
}
