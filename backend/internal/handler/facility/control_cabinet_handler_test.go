package facility

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	appcontrolcabinet "github.com/besart951/go_infra_link/backend/internal/application/facility/controlcabinet"
	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type controlCabinetUpdaterStub struct {
	command appcontrolcabinet.UpdateCommand
	result  *domainFacility.ControlCabinet
	err     error
}

type controlCabinetCreatorStub struct {
	command appcontrolcabinet.CreateCommand
	result  *domainFacility.ControlCabinet
	err     error
	calls   int
}

type controlCabinetClonerStub struct {
	command appcontrolcabinet.CloneCommand
	result  *domainFacility.ControlCabinet
	err     error
	calls   int
}

type controlCabinetDeleterStub struct {
	command appcontrolcabinet.DeleteCommand
	err     error
	calls   int
}

func (s *controlCabinetDeleterStub) Delete(
	_ context.Context,
	command appcontrolcabinet.DeleteCommand,
) error {
	s.calls++
	s.command = command
	return s.err
}

func (s *controlCabinetClonerStub) Clone(
	_ context.Context,
	command appcontrolcabinet.CloneCommand,
) (*domainFacility.ControlCabinet, error) {
	s.calls++
	s.command = command
	return s.result, s.err
}

func (s *controlCabinetCreatorStub) Create(
	_ context.Context,
	command appcontrolcabinet.CreateCommand,
) (*domainFacility.ControlCabinet, error) {
	s.calls++
	s.command = command
	return s.result, s.err
}

func TestCreateControlCabinetMapsRequestToApplicationCommandWithoutDirectBroadcast(t *testing.T) {
	gin.SetMode(gin.TestMode)
	buildingID := uuid.New()
	cabinetID := uuid.New()
	number := "AK01"
	creator := &controlCabinetCreatorStub{result: &domainFacility.ControlCabinet{
		Base:             domain.Base{ID: cabinetID},
		BuildingID:       buildingID,
		ControlCabinetNr: &number,
	}}
	handler := NewControlCabinetHandler(nil, creator, nil, nil, nil)
	router := gin.New()
	router.POST("/facility/control-cabinets", handler.CreateControlCabinet)
	request := httptest.NewRequest(
		http.MethodPost,
		"/facility/control-cabinets",
		strings.NewReader(`{"building_id":"`+buildingID.String()+`","control_cabinet_nr":"AK01"}`),
	)
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("response status: got %d, body=%s", recorder.Code, recorder.Body.String())
	}
	if creator.calls != 1 || creator.command.BuildingID != buildingID ||
		creator.command.ControlCabinetNr == nil || *creator.command.ControlCabinetNr != number {
		t.Fatalf("unexpected create command: %+v", creator.command)
	}
	var response map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response["id"] != cabinetID.String() || response["control_cabinet_nr"] != number {
		t.Fatalf("unexpected response: %+v", response)
	}
}

func TestCopyControlCabinetDelegatesToApplicationCommandWithoutDirectBroadcast(t *testing.T) {
	gin.SetMode(gin.TestMode)
	sourceID := uuid.New()
	copyID := uuid.New()
	buildingID := uuid.New()
	number := "AK02"
	cloner := &controlCabinetClonerStub{result: &domainFacility.ControlCabinet{
		Base:             domain.Base{ID: copyID},
		BuildingID:       buildingID,
		ControlCabinetNr: &number,
	}}
	handler := NewControlCabinetHandler(nil, nil, cloner, nil, nil)
	router := gin.New()
	router.POST("/facility/control-cabinets/:id/copy", handler.CopyControlCabinet)
	request := httptest.NewRequest(
		http.MethodPost,
		"/facility/control-cabinets/"+sourceID.String()+"/copy",
		nil,
	)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("response status: got %d, body=%s", recorder.Code, recorder.Body.String())
	}
	if cloner.calls != 1 || cloner.command.SourceControlCabinetID != sourceID {
		t.Fatalf("unexpected clone command: calls=%d command=%+v", cloner.calls, cloner.command)
	}
	var response map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response["id"] != copyID.String() || response["control_cabinet_nr"] != number {
		t.Fatalf("unexpected response: %+v", response)
	}
}

func TestCopyControlCabinetPreservesNotFoundMapping(t *testing.T) {
	gin.SetMode(gin.TestMode)
	sourceID := uuid.New()
	cloner := &controlCabinetClonerStub{err: domain.ErrNotFound}
	handler := NewControlCabinetHandler(nil, nil, cloner, nil, nil)
	router := gin.New()
	router.POST("/facility/control-cabinets/:id/copy", handler.CopyControlCabinet)
	request := httptest.NewRequest(
		http.MethodPost,
		"/facility/control-cabinets/"+sourceID.String()+"/copy",
		nil,
	)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("response status: got %d, want %d; body=%s",
			recorder.Code,
			http.StatusNotFound,
			recorder.Body.String(),
		)
	}
}

func (s *controlCabinetUpdaterStub) Update(
	_ context.Context,
	command appcontrolcabinet.UpdateCommand,
) (*domainFacility.ControlCabinet, error) {
	s.command = command
	return s.result, s.err
}

func TestUpdateControlCabinetMapsCompatibilityRequestToApplicationCommand(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cabinetID := uuid.New()
	buildingID := uuid.New()
	number := "AK02"
	updater := &controlCabinetUpdaterStub{result: &domainFacility.ControlCabinet{
		Base:             domain.Base{ID: cabinetID},
		BuildingID:       buildingID,
		ControlCabinetNr: &number,
	}}
	handler := NewControlCabinetHandler(nil, nil, nil, updater, nil)
	router := gin.New()
	router.PUT("/facility/control-cabinets/:id", handler.UpdateControlCabinet)
	request := httptest.NewRequest(
		http.MethodPut,
		"/facility/control-cabinets/"+cabinetID.String(),
		strings.NewReader(`{"expected_version":3,"building_id":"`+buildingID.String()+`","control_cabinet_nr":"AK02"}`),
	)
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("response status: got %d, body=%s", recorder.Code, recorder.Body.String())
	}
	command := updater.command
	if command.ControlCabinetID != cabinetID ||
		command.BuildingID == nil || *command.BuildingID != buildingID ||
		command.ControlCabinetNr == nil || *command.ControlCabinetNr != number {
		t.Fatalf("unexpected update command: %+v", command)
	}

	var response map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response["id"] != cabinetID.String() || response["control_cabinet_nr"] != number {
		t.Fatalf("unexpected response: %+v", response)
	}
}

func TestUpdateControlCabinetMapsApplicationLoadNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cabinetID := uuid.New()
	handler := NewControlCabinetHandler(nil, nil, nil, &controlCabinetUpdaterStub{
		err: &appcontrolcabinet.LoadError{Err: domain.ErrNotFound},
	}, nil)
	router := gin.New()
	router.PUT("/facility/control-cabinets/:id", handler.UpdateControlCabinet)
	request := httptest.NewRequest(
		http.MethodPut,
		"/facility/control-cabinets/"+cabinetID.String(),
		strings.NewReader(`{"expected_version":3}`),
	)
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("response status: got %d, want %d; body=%s", recorder.Code, http.StatusNotFound, recorder.Body.String())
	}
}

func TestDeleteControlCabinetDelegatesToTypedApplicationCommand(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cabinetID := uuid.New()
	deleter := &controlCabinetDeleterStub{}
	handler := NewControlCabinetHandler(nil, nil, nil, nil, deleter)
	router := gin.New()
	router.DELETE("/facility/control-cabinets/:id", handler.DeleteControlCabinet)
	request := httptest.NewRequest(
		http.MethodDelete,
		"/facility/control-cabinets/"+cabinetID.String(),
		nil,
	)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("response status: got %d, body=%s", recorder.Code, recorder.Body.String())
	}
	if deleter.calls != 1 || deleter.command.ControlCabinetID != cabinetID {
		t.Fatalf("delete command: calls=%d command=%+v", deleter.calls, deleter.command)
	}
}

func TestDeleteControlCabinetPreservesNotFoundMapping(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cabinetID := uuid.New()
	deleter := &controlCabinetDeleterStub{err: domain.ErrNotFound}
	handler := NewControlCabinetHandler(nil, nil, nil, nil, deleter)
	router := gin.New()
	router.DELETE("/facility/control-cabinets/:id", handler.DeleteControlCabinet)
	request := httptest.NewRequest(
		http.MethodDelete,
		"/facility/control-cabinets/"+cabinetID.String(),
		nil,
	)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("response status: got %d, want %d; body=%s",
			recorder.Code,
			http.StatusNotFound,
			recorder.Body.String(),
		)
	}
}
