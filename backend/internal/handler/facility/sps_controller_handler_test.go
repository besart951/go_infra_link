package facility

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	appspscontroller "github.com/besart951/go_infra_link/backend/internal/application/facility/spscontroller"
	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type spsControllerUpdaterStub struct {
	command appspscontroller.UpdateCommand
	result  *domainFacility.SPSController
	err     error
}

type spsControllerCreatorStub struct {
	command appspscontroller.CreateCommand
	result  *domainFacility.SPSController
	err     error
	calls   int
}

type spsControllerDeleterStub struct {
	command appspscontroller.DeleteCommand
	err     error
	calls   int
}

type spsControllerClonerStub struct {
	command appspscontroller.CloneCommand
	result  *domainFacility.SPSController
	err     error
	calls   int
}

func (s *spsControllerClonerStub) Clone(
	_ context.Context,
	command appspscontroller.CloneCommand,
) (*domainFacility.SPSController, error) {
	s.calls++
	s.command = command
	return s.result, s.err
}

func (s *spsControllerDeleterStub) Delete(
	_ context.Context,
	command appspscontroller.DeleteCommand,
) error {
	s.calls++
	s.command = command
	return s.err
}

func (s *spsControllerCreatorStub) Create(
	_ context.Context,
	command appspscontroller.CreateCommand,
) (*domainFacility.SPSController, error) {
	s.calls++
	s.command = command
	return s.result, s.err
}

func TestCreateSPSControllerMapsRequestToApplicationCommandWithoutDirectBroadcast(t *testing.T) {
	gin.SetMode(gin.TestMode)
	controllerID := uuid.New()
	cabinetID := uuid.New()
	systemTypeID := uuid.New()
	gaDevice := "AAA"
	description := "Automation controller"
	creator := &spsControllerCreatorStub{result: &domainFacility.SPSController{
		Base:              domain.Base{ID: controllerID},
		ControlCabinetID:  cabinetID,
		GADevice:          &gaDevice,
		DeviceName:        "BLD_AK01_AAA",
		DeviceDescription: &description,
	}}
	handler := NewSPSControllerHandler(nil, creator, nil, nil, nil)
	router := gin.New()
	router.POST("/facility/sps-controllers", handler.CreateSPSController)
	request := httptest.NewRequest(
		http.MethodPost,
		"/facility/sps-controllers",
		strings.NewReader(`{
			"control_cabinet_id":"`+cabinetID.String()+`",
			"ga_device":"AAA",
			"device_name":"client value",
			"device_description":"Automation controller",
			"system_types":[{"system_type_id":"`+systemTypeID.String()+`","number":17}]
		}`),
	)
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("response status: got %d, body=%s", recorder.Code, recorder.Body.String())
	}
	command := creator.command
	if creator.calls != 1 || command.ControlCabinetID != cabinetID ||
		command.GADevice == nil || *command.GADevice != gaDevice ||
		command.DeviceName != "client value" ||
		command.DeviceDescription == nil || *command.DeviceDescription != description ||
		len(command.SystemTypes) != 1 || command.SystemTypes[0].SystemTypeID != systemTypeID ||
		command.SystemTypes[0].Number == nil || *command.SystemTypes[0].Number != 17 {
		t.Fatalf("unexpected create command: %+v", command)
	}
	var response map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response["id"] != controllerID.String() || response["device_name"] != "BLD_AK01_AAA" {
		t.Fatalf("unexpected response: %+v", response)
	}
}

func (s *spsControllerUpdaterStub) Update(
	_ context.Context,
	command appspscontroller.UpdateCommand,
) (*domainFacility.SPSController, error) {
	s.command = command
	return s.result, s.err
}

func TestUpdateSPSControllerMapsCompatibilityRequestToApplicationCommand(t *testing.T) {
	gin.SetMode(gin.TestMode)
	controllerID := uuid.New()
	cabinetID := uuid.New()
	systemTypeID := uuid.New()
	gaDevice := "AAA"
	updater := &spsControllerUpdaterStub{result: &domainFacility.SPSController{
		Base:             domain.Base{ID: controllerID},
		ControlCabinetID: cabinetID,
		GADevice:         &gaDevice,
		DeviceName:       "BLD_CC_AAA",
	}}
	handler := NewSPSControllerHandler(nil, nil, nil, updater, nil)
	router := gin.New()
	router.PUT("/facility/sps-controllers/:id", handler.UpdateSPSController)
	body := `{
		"expected_version":3,
		"control_cabinet_id":"` + cabinetID.String() + `",
		"ga_device":"AAA",
		"device_name":"ignored-by-generated-name-rule",
		"device_description":"Description",
		"system_types":[{"system_type_id":"` + systemTypeID.String() + `","number":17}]
	}`
	request := httptest.NewRequest(
		http.MethodPut,
		"/facility/sps-controllers/"+controllerID.String(),
		strings.NewReader(body),
	)
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("response status: got %d, body=%s", recorder.Code, recorder.Body.String())
	}
	command := updater.command
	if command.SPSControllerID != controllerID ||
		command.ControlCabinetID == nil || *command.ControlCabinetID != cabinetID ||
		command.GADevice == nil || *command.GADevice != gaDevice ||
		command.DeviceDescription == nil || *command.DeviceDescription != "Description" {
		t.Fatalf("unexpected update command: %+v", command)
	}
	if command.SystemTypes == nil || len(*command.SystemTypes) != 1 ||
		(*command.SystemTypes)[0].SystemTypeID != systemTypeID ||
		(*command.SystemTypes)[0].Number == nil || *(*command.SystemTypes)[0].Number != 17 {
		t.Fatalf("unexpected system type command: %+v", command.SystemTypes)
	}

	var response map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response["id"] != controllerID.String() || response["device_name"] != "BLD_CC_AAA" {
		t.Fatalf("unexpected response: %+v", response)
	}
}

func TestUpdateSPSControllerMapsApplicationLoadNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	controllerID := uuid.New()
	handler := NewSPSControllerHandler(nil, nil, nil, &spsControllerUpdaterStub{
		err: &appspscontroller.LoadError{Err: domain.ErrNotFound},
	}, nil)
	router := gin.New()
	router.PUT("/facility/sps-controllers/:id", handler.UpdateSPSController)
	request := httptest.NewRequest(
		http.MethodPut,
		"/facility/sps-controllers/"+controllerID.String(),
		strings.NewReader(`{"expected_version":3}`),
	)
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("response status: got %d, want %d; body=%s", recorder.Code, http.StatusNotFound, recorder.Body.String())
	}
}

func TestDeleteSPSControllerUsesApplicationCommandWithoutDirectBroadcast(t *testing.T) {
	gin.SetMode(gin.TestMode)
	controllerID := uuid.New()
	deleter := &spsControllerDeleterStub{}
	handler := NewSPSControllerHandler(nil, nil, nil, nil, deleter)
	router := gin.New()
	router.DELETE("/facility/sps-controllers/:id", handler.DeleteSPSController)
	request := httptest.NewRequest(
		http.MethodDelete,
		"/facility/sps-controllers/"+controllerID.String(),
		nil,
	)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("response status: got %d, body=%s", recorder.Code, recorder.Body.String())
	}
	if deleter.calls != 1 || deleter.command.SPSControllerID != controllerID {
		t.Fatalf("unexpected delete command: calls=%d command=%+v", deleter.calls, deleter.command)
	}
}

func TestCopySPSControllerUsesTypedCloneApplicationCommand(t *testing.T) {
	gin.SetMode(gin.TestMode)
	sourceID := uuid.New()
	copyID := uuid.New()
	cabinetID := uuid.New()
	gaDevice := "AAB"
	cloner := &spsControllerClonerStub{result: &domainFacility.SPSController{
		Base:             domain.Base{ID: copyID},
		ControlCabinetID: cabinetID,
		GADevice:         &gaDevice,
		DeviceName:       "BLD_AK01_AAB",
	}}
	handler := NewSPSControllerHandler(nil, nil, cloner, nil, nil)
	router := gin.New()
	router.POST("/facility/sps-controllers/:id/copy", handler.CopySPSController)
	request := httptest.NewRequest(
		http.MethodPost,
		"/facility/sps-controllers/"+sourceID.String()+"/copy",
		nil,
	)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("response status: got %d, body=%s", recorder.Code, recorder.Body.String())
	}
	if cloner.calls != 1 || cloner.command.SourceSPSControllerID != sourceID {
		t.Fatalf("unexpected clone command: calls=%d command=%+v", cloner.calls, cloner.command)
	}
	var response map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response["id"] != copyID.String() || response["device_name"] != "BLD_AK01_AAB" {
		t.Fatalf("unexpected response: %+v", response)
	}
}

func TestParseSPSControllerGAExcludeIDAcceptsLegacyFrontendAlias(t *testing.T) {
	given := uuid.New()
	ctx, recorder := newSPSControllerQueryContext(
		t,
		"?sps_controller_id="+given.String(),
	)

	got, ok := parseSPSControllerGAExcludeID(ctx)

	if !ok || got == nil || *got != given {
		t.Fatalf("legacy exclude ID: got %v (ok=%t), want %s", got, ok, given)
	}
	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected response status %d", recorder.Code)
	}
}

func TestParseSPSControllerGAExcludeIDPrefersCanonicalParameter(t *testing.T) {
	canonical := uuid.New()
	legacy := uuid.New()
	ctx, _ := newSPSControllerQueryContext(
		t,
		"?exclude_id="+canonical.String()+"&sps_controller_id="+legacy.String(),
	)

	got, ok := parseSPSControllerGAExcludeID(ctx)

	if !ok || got == nil || *got != canonical {
		t.Fatalf("canonical exclude ID: got %v (ok=%t), want %s", got, ok, canonical)
	}
}

func TestParseSPSControllerGAExcludeIDRejectsInvalidLegacyAlias(t *testing.T) {
	ctx, recorder := newSPSControllerQueryContext(t, "?sps_controller_id=invalid")

	got, ok := parseSPSControllerGAExcludeID(ctx)

	if ok || got != nil {
		t.Fatalf("invalid legacy exclude ID: got %v (ok=%t)", got, ok)
	}
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("response status: got %d, want %d", recorder.Code, http.StatusBadRequest)
	}
}

func newSPSControllerQueryContext(
	t *testing.T,
	rawQuery string,
) (*gin.Context, *httptest.ResponseRecorder) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/next-ga-device"+rawQuery, nil)
	return ctx, recorder
}
