package history

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	appcontrolcabinet "github.com/besart951/go_infra_link/backend/internal/application/facility/controlcabinet"
	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainHistory "github.com/besart951/go_infra_link/backend/internal/domain/history"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type projectControlCabinetRestorerStub struct {
	command appcontrolcabinet.RestoreForProjectCommand
	result  *domainHistory.RestoreResult
	err     error
	calls   int
}

func (s *projectControlCabinetRestorerStub) RestoreForProject(
	_ context.Context,
	command appcontrolcabinet.RestoreForProjectCommand,
) (*domainHistory.RestoreResult, error) {
	s.calls++
	s.command = command
	return s.result, s.err
}

func TestRestoreProjectControlCabinetDelegatesTypedCommandAndPreservesResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)
	projectID := uuid.New()
	clientProjectID := uuid.New()
	controlCabinetID := uuid.New()
	eventID := uuid.New()
	asOf := time.Date(2026, time.July, 22, 9, 0, 0, 0, time.UTC)
	result := &domainHistory.RestoreResult{
		RestoredCount: 3,
		DeletedCount:  2,
		SkippedCount:  1,
		BatchID:       uuid.New(),
	}
	restorer := &projectControlCabinetRestorerStub{result: result}
	handler := NewHandler(historyServiceStub{}, restorer, nil)
	router := gin.New()
	router.POST(
		"/projects/:id/history/control-cabinets/:controlCabinetId/restore",
		handler.RestoreProjectControlCabinet,
	)
	body, err := json.Marshal(domainHistory.RestoreControlCabinetRequest{
		AsOf:      &asOf,
		EventID:   &eventID,
		ProjectID: &clientProjectID,
	})
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}
	req := httptest.NewRequest(
		http.MethodPost,
		"/projects/"+projectID.String()+"/history/control-cabinets/"+
			controlCabinetID.String()+"/restore",
		bytes.NewReader(body),
	)
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("status: got %d body=%s", res.Code, res.Body.String())
	}
	if restorer.calls != 1 || restorer.command.ProjectID != projectID ||
		restorer.command.ControlCabinetID != controlCabinetID ||
		restorer.command.AsOf == nil || !restorer.command.AsOf.Equal(asOf) ||
		restorer.command.EventID == nil || *restorer.command.EventID != eventID {
		t.Fatalf("typed command: %+v", restorer.command)
	}
	var got domainHistory.RestoreResult
	if err := json.Unmarshal(res.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if got != *result {
		t.Fatalf("response: got %+v, want %+v", got, *result)
	}
}

func TestRestoreProjectControlCabinetMapsScopeDenialAndDoesNotUseLegacyRestore(t *testing.T) {
	gin.SetMode(gin.TestMode)
	restorer := &projectControlCabinetRestorerStub{err: appcontrolcabinet.ErrProjectRestoreAccessDenied}
	handler := NewHandler(historyServiceStub{}, restorer, nil)
	router := gin.New()
	router.POST(
		"/projects/:id/history/control-cabinets/:controlCabinetId/restore",
		handler.RestoreProjectControlCabinet,
	)
	req := httptest.NewRequest(
		http.MethodPost,
		"/projects/"+uuid.NewString()+"/history/control-cabinets/"+
			uuid.NewString()+"/restore",
		nil,
	)
	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)

	if res.Code != http.StatusForbidden {
		t.Fatalf("status: got %d body=%s", res.Code, res.Body.String())
	}
	if restorer.calls != 1 {
		t.Fatalf("application restore calls: got %d, want 1", restorer.calls)
	}
}

func TestRestoreProjectControlCabinetRequiresConfiguredApplicationHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewHandler(historyServiceStub{}, nil, nil)
	router := gin.New()
	router.POST(
		"/projects/:id/history/control-cabinets/:controlCabinetId/restore",
		handler.RestoreProjectControlCabinet,
	)
	req := httptest.NewRequest(
		http.MethodPost,
		"/projects/"+uuid.NewString()+"/history/control-cabinets/"+
			uuid.NewString()+"/restore",
		nil,
	)
	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)

	if res.Code != http.StatusInternalServerError {
		t.Fatalf("status: got %d body=%s", res.Code, res.Body.String())
	}
}

func TestRestoreProjectControlCabinetMapsNotFoundAndInvalidArgument(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, test := range []struct {
		name       string
		err        error
		wantStatus int
	}{
		{name: "scope not found", err: domain.ErrNotFound, wantStatus: http.StatusNotFound},
		{name: "invalid command", err: domain.ErrInvalidArgument, wantStatus: http.StatusBadRequest},
	} {
		t.Run(test.name, func(t *testing.T) {
			restorer := &projectControlCabinetRestorerStub{err: test.err}
			handler := NewHandler(historyServiceStub{}, restorer, nil)
			router := gin.New()
			router.POST(
				"/projects/:id/history/control-cabinets/:controlCabinetId/restore",
				handler.RestoreProjectControlCabinet,
			)
			req := httptest.NewRequest(
				http.MethodPost,
				"/projects/"+uuid.NewString()+"/history/control-cabinets/"+
					uuid.NewString()+"/restore",
				nil,
			)
			res := httptest.NewRecorder()

			router.ServeHTTP(res, req)

			if res.Code != test.wantStatus {
				t.Fatalf("status: got %d body=%s", res.Code, res.Body.String())
			}
		})
	}
}
