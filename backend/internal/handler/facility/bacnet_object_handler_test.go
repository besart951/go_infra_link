package facility

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	appbacnetobject "github.com/besart951/go_infra_link/backend/internal/application/facility/bacnetobject"
	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type bacnetObjectUpdaterStub struct {
	command appbacnetobject.UpdateCommand
	result  *domainFacility.BacnetObject
	err     error
}

type bacnetObjectCreatorStub struct {
	fieldDeviceCommand appbacnetobject.CreateForFieldDeviceCommand
	objectDataCommand  appbacnetobject.CreateForObjectDataCommand
	result             *domainFacility.BacnetObject
	err                error
	fieldDeviceCalls   int
	objectDataCalls    int
}

func (s *bacnetObjectCreatorStub) CreateForFieldDevice(
	_ context.Context,
	command appbacnetobject.CreateForFieldDeviceCommand,
) (*domainFacility.BacnetObject, error) {
	s.fieldDeviceCalls++
	s.fieldDeviceCommand = command
	return s.result, s.err
}

func (s *bacnetObjectCreatorStub) CreateForObjectData(
	_ context.Context,
	command appbacnetobject.CreateForObjectDataCommand,
) (*domainFacility.BacnetObject, error) {
	s.objectDataCalls++
	s.objectDataCommand = command
	return s.result, s.err
}

func bacnetHandlerTestUUID(value byte) uuid.UUID {
	var id uuid.UUID
	id[15] = value
	return id
}

func TestCreateBacnetObjectRoutesFieldDeviceParentThroughApplicationCommand(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fieldDeviceID := bacnetHandlerTestUUID(1)
	objectID := bacnetHandlerTestUUID(2)
	alarmDefinitionID := bacnetHandlerTestUUID(3)
	description := "Room temperature"
	creator := &bacnetObjectCreatorStub{result: &domainFacility.BacnetObject{
		Base:              domain.Base{ID: objectID},
		TextFix:           "AI",
		Description:       &description,
		SoftwareType:      domainFacility.BacnetSoftwareTypeAI,
		SoftwareNumber:    7,
		FieldDeviceID:     &fieldDeviceID,
		AlarmDefinitionID: &alarmDefinitionID,
	}}
	handler := NewBacnetObjectHandler(creator, nil)
	router := gin.New()
	router.POST("/facility/bacnet-objects", handler.CreateBacnetObject)
	request := httptest.NewRequest(
		http.MethodPost,
		"/facility/bacnet-objects",
		strings.NewReader(`{
			"field_device_id":"`+fieldDeviceID.String()+`",
			"text_fix":"AI",
			"description":"Room temperature",
			"gms_visible":true,
			"text_individual":"",
			"software_type":"ai",
			"software_number":7,
			"hardware_type":"ai",
			"hardware_quantity":1,
			"alarm_definition_id":"`+alarmDefinitionID.String()+`"
		}`),
	)
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("response status: got %d, body=%s", recorder.Code, recorder.Body.String())
	}
	if creator.fieldDeviceCalls != 1 || creator.objectDataCalls != 0 {
		t.Fatalf("create calls: field device=%d object data=%d", creator.fieldDeviceCalls, creator.objectDataCalls)
	}
	command := creator.fieldDeviceCommand
	if command.FieldDeviceID != fieldDeviceID || command.Input.TextFix != "AI" ||
		command.Input.Description == nil || *command.Input.Description != description ||
		!command.Input.GMSVisible || command.Input.TextIndividual != nil ||
		command.Input.SoftwareType != domainFacility.BacnetSoftwareTypeAI ||
		command.Input.SoftwareNumber != 7 ||
		command.Input.HardwareType != domainFacility.BacnetHardwareTypeAI ||
		command.Input.HardwareQuantity != 1 ||
		command.Input.AlarmDefinitionID == nil || *command.Input.AlarmDefinitionID != alarmDefinitionID {
		t.Fatalf("unexpected create command: %+v", command)
	}
	var response map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response["id"] != objectID.String() || response["text_fix"] != "AI" {
		t.Fatalf("unexpected response: %+v", response)
	}
}

func TestCreateBacnetObjectRoutesObjectDataParentThroughApplicationCommand(t *testing.T) {
	gin.SetMode(gin.TestMode)
	objectDataID := bacnetHandlerTestUUID(4)
	objectID := bacnetHandlerTestUUID(5)
	creator := &bacnetObjectCreatorStub{result: &domainFacility.BacnetObject{
		Base:           domain.Base{ID: objectID},
		TextFix:        "AI",
		SoftwareType:   domainFacility.BacnetSoftwareTypeAI,
		SoftwareNumber: 1,
	}}
	handler := NewBacnetObjectHandler(creator, nil)
	router := gin.New()
	router.POST("/facility/bacnet-objects", handler.CreateBacnetObject)
	request := httptest.NewRequest(
		http.MethodPost,
		"/facility/bacnet-objects",
		strings.NewReader(`{
			"object_data_id":"`+objectDataID.String()+`",
			"text_fix":"AI",
			"software_type":"ai",
			"software_number":1
		}`),
	)
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("response status: got %d, body=%s", recorder.Code, recorder.Body.String())
	}
	if creator.fieldDeviceCalls != 0 || creator.objectDataCalls != 1 ||
		creator.objectDataCommand.ObjectDataID != objectDataID ||
		creator.objectDataCommand.Input.TextFix != "AI" ||
		creator.objectDataCommand.Input.SoftwareType != domainFacility.BacnetSoftwareTypeAI ||
		creator.objectDataCommand.Input.SoftwareNumber != 1 {
		t.Fatalf("unexpected ObjectData command: %+v", creator.objectDataCommand)
	}
}

func TestCreateBacnetObjectRejectsMissingOrAmbiguousParentBeforeApplicationCall(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fieldDeviceID := bacnetHandlerTestUUID(6)
	objectDataID := bacnetHandlerTestUUID(7)
	for _, test := range []struct {
		name string
		body string
	}{
		{
			name: "missing",
			body: `{"text_fix":"AI","software_type":"ai","software_number":1}`,
		},
		{
			name: "both",
			body: `{"field_device_id":"` + fieldDeviceID.String() +
				`","object_data_id":"` + objectDataID.String() +
				`","text_fix":"AI","software_type":"ai","software_number":1}`,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			creator := &bacnetObjectCreatorStub{}
			handler := NewBacnetObjectHandler(creator, nil)
			router := gin.New()
			router.POST("/facility/bacnet-objects", handler.CreateBacnetObject)
			request := httptest.NewRequest(
				http.MethodPost,
				"/facility/bacnet-objects",
				strings.NewReader(test.body),
			)
			request.Header.Set("Content-Type", "application/json")
			recorder := httptest.NewRecorder()

			router.ServeHTTP(recorder, request)

			if recorder.Code != http.StatusBadRequest {
				t.Fatalf("response status: got %d, body=%s", recorder.Code, recorder.Body.String())
			}
			if creator.fieldDeviceCalls != 0 || creator.objectDataCalls != 0 {
				t.Fatalf("application called for invalid parent: %+v", creator)
			}
		})
	}
}

func (s *bacnetObjectUpdaterStub) Update(
	_ context.Context,
	command appbacnetobject.UpdateCommand,
) (*domainFacility.BacnetObject, error) {
	s.command = command
	return s.result, s.err
}

func TestUpdateBacnetObjectMapsCompatibilityRequestToApplicationCommand(t *testing.T) {
	gin.SetMode(gin.TestMode)
	objectID := uuid.New()
	fieldDeviceID := uuid.New()
	softwareReferenceID := uuid.New()
	updater := &bacnetObjectUpdaterStub{result: &domainFacility.BacnetObject{
		Base:                domain.Base{ID: objectID},
		TextFix:             "NEW",
		SoftwareType:        domainFacility.BacnetSoftwareTypeAO,
		SoftwareNumber:      7,
		FieldDeviceID:       &fieldDeviceID,
		SoftwareReferenceID: &softwareReferenceID,
	}}
	handler := NewBacnetObjectHandler(nil, updater)
	router := gin.New()
	router.PUT("/facility/bacnet-objects/:id", handler.UpdateBacnetObject)
	request := httptest.NewRequest(
		http.MethodPut,
		"/facility/bacnet-objects/"+objectID.String(),
		strings.NewReader(`{
			"field_device_id":"`+fieldDeviceID.String()+`",
			"text_fix":"NEW",
			"gms_visible":true,
			"software_type":"ao",
			"software_number":7,
			"hardware_quantity":2,
			"software_reference_id":"`+softwareReferenceID.String()+`"
		}`),
	)
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("response status: got %d, body=%s", recorder.Code, recorder.Body.String())
	}
	command := updater.command
	if command.BacnetObjectID != objectID ||
		command.FieldDeviceID == nil || *command.FieldDeviceID != fieldDeviceID ||
		command.ObjectDataID != nil || command.Patch.ID != objectID ||
		command.Patch.TextFix == nil || *command.Patch.TextFix != "NEW" ||
		command.Patch.GMSVisible == nil || !*command.Patch.GMSVisible ||
		command.Patch.SoftwareType == nil || *command.Patch.SoftwareType != domainFacility.BacnetSoftwareTypeAO ||
		command.Patch.SoftwareNumber == nil || *command.Patch.SoftwareNumber != 7 ||
		command.Patch.HardwareQuantity == nil || *command.Patch.HardwareQuantity != 2 ||
		command.Patch.SoftwareReferenceID == nil || *command.Patch.SoftwareReferenceID != softwareReferenceID {
		t.Fatalf("unexpected update command: %+v", command)
	}

	var response map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response["id"] != objectID.String() || response["text_fix"] != "NEW" {
		t.Fatalf("unexpected response: %+v", response)
	}
}

func TestUpdateBacnetObjectPreservesExcelSoftwareReferenceOnlyPayload(t *testing.T) {
	gin.SetMode(gin.TestMode)
	objectID := uuid.New()
	softwareReferenceID := uuid.New()
	updater := &bacnetObjectUpdaterStub{result: &domainFacility.BacnetObject{
		Base:         domain.Base{ID: objectID},
		TextFix:      "AI",
		SoftwareType: domainFacility.BacnetSoftwareTypeAI,
	}}
	handler := NewBacnetObjectHandler(nil, updater)
	router := gin.New()
	router.PUT("/facility/bacnet-objects/:id", handler.UpdateBacnetObject)
	request := httptest.NewRequest(
		http.MethodPut,
		"/facility/bacnet-objects/"+objectID.String(),
		strings.NewReader(`{"software_reference_id":"`+softwareReferenceID.String()+`"}`),
	)
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("response status: got %d, body=%s", recorder.Code, recorder.Body.String())
	}
	patch := updater.command.Patch
	if patch.SoftwareReferenceID == nil || *patch.SoftwareReferenceID != softwareReferenceID ||
		patch.TextFix != nil || updater.command.FieldDeviceID != nil || updater.command.ObjectDataID != nil {
		t.Fatalf("unexpected Excel compatibility command: %+v", updater.command)
	}
}

func TestUpdateBacnetObjectMapsApplicationLoadNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	objectID := uuid.New()
	handler := NewBacnetObjectHandler(nil, &bacnetObjectUpdaterStub{
		err: &appbacnetobject.LoadError{Err: domain.ErrNotFound},
	})
	router := gin.New()
	router.PUT("/facility/bacnet-objects/:id", handler.UpdateBacnetObject)
	request := httptest.NewRequest(
		http.MethodPut,
		"/facility/bacnet-objects/"+objectID.String(),
		strings.NewReader(`{}`),
	)
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("response status: got %d, want %d; body=%s", recorder.Code, http.StatusNotFound, recorder.Body.String())
	}
}
