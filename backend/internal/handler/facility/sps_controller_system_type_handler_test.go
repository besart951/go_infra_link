package facility

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	appspscontroller "github.com/besart951/go_infra_link/backend/internal/application/facility/spscontroller"
	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type spsControllerSystemTypeClonerStub struct {
	command appspscontroller.CloneSystemTypeCommand
	result  *domainFacility.SPSControllerSystemType
	err     error
	calls   int
}

type spsControllerSystemTypeDeleterStub struct {
	command appspscontroller.DeleteSystemTypeCommand
	err     error
	calls   int
}

func (s *spsControllerSystemTypeDeleterStub) DeleteSystemType(
	_ context.Context,
	command appspscontroller.DeleteSystemTypeCommand,
) error {
	s.calls++
	s.command = command
	return s.err
}

func (s *spsControllerSystemTypeClonerStub) CloneSystemType(
	_ context.Context,
	command appspscontroller.CloneSystemTypeCommand,
) (*domainFacility.SPSControllerSystemType, error) {
	s.calls++
	s.command = command
	return s.result, s.err
}

func TestCopySPSControllerSystemTypeUsesTypedApplicationCommand(t *testing.T) {
	gin.SetMode(gin.TestMode)
	sourceID := uuid.New()
	copyID := uuid.New()
	spsControllerID := uuid.New()
	systemTypeID := uuid.New()
	number := 4
	documentName := "HLK-04.pdf"
	cloner := &spsControllerSystemTypeClonerStub{
		result: &domainFacility.SPSControllerSystemType{
			Base:              domain.Base{ID: copyID},
			Number:            &number,
			DocumentName:      &documentName,
			SPSControllerID:   spsControllerID,
			SystemTypeID:      systemTypeID,
			FieldDevicesCount: 7,
		},
	}
	handler := NewSPSControllerSystemTypeHandler(nil, cloner, nil)
	router := gin.New()
	router.POST(
		"/facility/sps-controller-system-types/:id/copy",
		handler.CopySPSControllerSystemType,
	)
	request := httptest.NewRequest(
		http.MethodPost,
		"/facility/sps-controller-system-types/"+sourceID.String()+"/copy",
		nil,
	)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("response status: got %d, body=%s", recorder.Code, recorder.Body.String())
	}
	if cloner.calls != 1 || cloner.command.SourceSPSControllerSystemTypeID != sourceID {
		t.Fatalf("clone command: calls=%d command=%+v", cloner.calls, cloner.command)
	}
	var response map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response["id"] != copyID.String() ||
		response["sps_controller_id"] != spsControllerID.String() ||
		response["system_type_id"] != systemTypeID.String() ||
		response["number"] != float64(number) ||
		response["document_name"] != documentName ||
		response["field_devices_count"] != float64(7) {
		t.Fatalf("unexpected response: %+v", response)
	}
}

func TestCopySPSControllerSystemTypePreservesNotFoundMapping(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cloner := &spsControllerSystemTypeClonerStub{err: domain.ErrNotFound}
	handler := NewSPSControllerSystemTypeHandler(nil, cloner, nil)
	router := gin.New()
	router.POST(
		"/facility/sps-controller-system-types/:id/copy",
		handler.CopySPSControllerSystemType,
	)
	request := httptest.NewRequest(
		http.MethodPost,
		"/facility/sps-controller-system-types/"+uuid.NewString()+"/copy",
		nil,
	)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("response status: got %d, body=%s", recorder.Code, recorder.Body.String())
	}
	if cloner.calls != 1 {
		t.Fatalf("cloner calls: got %d, want 1", cloner.calls)
	}
}

func TestDeleteSPSControllerSystemTypeUsesTypedApplicationCommand(t *testing.T) {
	gin.SetMode(gin.TestMode)
	systemTypeID := uuid.New()
	deleter := &spsControllerSystemTypeDeleterStub{}
	handler := NewSPSControllerSystemTypeHandler(nil, nil, deleter)
	router := gin.New()
	router.DELETE(
		"/facility/sps-controller-system-types/:id",
		handler.DeleteSPSControllerSystemType,
	)
	request := httptest.NewRequest(
		http.MethodDelete,
		"/facility/sps-controller-system-types/"+systemTypeID.String(),
		nil,
	)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("response status: got %d, body=%s", recorder.Code, recorder.Body.String())
	}
	if deleter.calls != 1 || deleter.command.SPSControllerSystemTypeID != systemTypeID {
		t.Fatalf("delete command: calls=%d command=%+v", deleter.calls, deleter.command)
	}
}

func TestDeleteSPSControllerSystemTypePreservesNotFoundMapping(t *testing.T) {
	gin.SetMode(gin.TestMode)
	deleter := &spsControllerSystemTypeDeleterStub{err: domain.ErrNotFound}
	handler := NewSPSControllerSystemTypeHandler(nil, nil, deleter)
	router := gin.New()
	router.DELETE(
		"/facility/sps-controller-system-types/:id",
		handler.DeleteSPSControllerSystemType,
	)
	request := httptest.NewRequest(
		http.MethodDelete,
		"/facility/sps-controller-system-types/"+uuid.NewString(),
		nil,
	)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("response status: got %d, body=%s", recorder.Code, recorder.Body.String())
	}
	if deleter.calls != 1 {
		t.Fatalf("deleter calls: got %d, want 1", deleter.calls)
	}
}
