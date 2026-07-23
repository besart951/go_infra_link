package fielddevice

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	appfielddevice "github.com/besart951/go_infra_link/backend/internal/application/facility/fielddevice"
	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	domainProject "github.com/besart951/go_infra_link/backend/internal/domain/project"
	domainUser "github.com/besart951/go_infra_link/backend/internal/domain/user"
	facilitydto "github.com/besart951/go_infra_link/backend/internal/handler/dto/facility"
	projectdto "github.com/besart951/go_infra_link/backend/internal/handler/dto/project"
	"github.com/besart951/go_infra_link/backend/internal/handler/middleware"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type projectFieldDeviceAccessStub struct {
	hasAccess           bool
	hasPermission       bool
	accessCalls         int
	permissionCalls     int
	requesterID         uuid.UUID
	projectID           uuid.UUID
	requestedPermission string
}

func (s *projectFieldDeviceAccessStub) CanAccessProject(
	_ context.Context,
	requesterID uuid.UUID,
	projectID uuid.UUID,
	_ *domainUser.Role,
) (bool, error) {
	s.accessCalls++
	s.requesterID = requesterID
	s.projectID = projectID
	return s.hasAccess, nil
}

func (s *projectFieldDeviceAccessStub) CanUseProjectPermission(
	context.Context,
	uuid.UUID,
	*domainUser.Role,
	string,
) (bool, error) {
	return s.hasPermission, nil
}

func (s *projectFieldDeviceAccessStub) CanUseProjectPermissionForProject(
	_ context.Context,
	_ uuid.UUID,
	projectID uuid.UUID,
	_ *domainUser.Role,
	permission string,
) (bool, error) {
	s.permissionCalls++
	s.projectID = projectID
	s.requestedPermission = permission
	return s.hasPermission, nil
}

type projectFieldDeviceMultiCreatorStub struct {
	command appfielddevice.MultiCreateForProjectCommand
	result  *domainFacility.FieldDeviceMultiCreateResult
	err     error
	calls   int
}

type projectFieldDeviceAssignerStub struct {
	command appfielddevice.AssignToProjectCommand
	result  *domainProject.ProjectFieldDevice
	err     error
	calls   int
}

type projectFieldDeviceBulkAssignerStub struct {
	command appfielddevice.BulkAssignToProjectCommand
	result  appfielddevice.BulkAssignToProjectResult
	calls   int
}

type projectFieldDeviceReassignerStub struct {
	command appfielddevice.ReassignProjectLinkCommand
	result  *domainProject.ProjectFieldDevice
	err     error
	calls   int
}

func (s *projectFieldDeviceReassignerStub) ReassignProjectLink(
	_ context.Context,
	command appfielddevice.ReassignProjectLinkCommand,
) (*domainProject.ProjectFieldDevice, error) {
	s.calls++
	s.command = command
	return s.result, s.err
}

func (s *projectFieldDeviceBulkAssignerStub) BulkAssignToProject(
	_ context.Context,
	command appfielddevice.BulkAssignToProjectCommand,
) appfielddevice.BulkAssignToProjectResult {
	s.calls++
	s.command = command
	return s.result
}

func (s *projectFieldDeviceAssignerStub) AssignToProject(
	_ context.Context,
	command appfielddevice.AssignToProjectCommand,
) (*domainProject.ProjectFieldDevice, error) {
	s.calls++
	s.command = command
	return s.result, s.err
}

func (s *projectFieldDeviceMultiCreatorStub) MultiCreateForProject(
	_ context.Context,
	command appfielddevice.MultiCreateForProjectCommand,
) (*domainFacility.FieldDeviceMultiCreateResult, error) {
	s.calls++
	s.command = command
	return s.result, s.err
}

type projectFieldDeviceLinkStub struct {
	FacilityLinkService
	successIDs  []uuid.UUID
	errors      []string
	projectID   uuid.UUID
	inputIDs    []uuid.UUID
	calls       int
	createCalls int
	updateCalls int
}

func (s *projectFieldDeviceLinkStub) CreateFieldDevice(
	context.Context,
	uuid.UUID,
	uuid.UUID,
) (*domainProject.ProjectFieldDevice, error) {
	s.createCalls++
	return nil, errors.New("legacy CreateFieldDevice must not be called")
}

func (s *projectFieldDeviceLinkStub) MultiCreateFieldDevices(
	_ context.Context,
	projectID uuid.UUID,
	fieldDeviceIDs []uuid.UUID,
) ([]uuid.UUID, []string) {
	s.calls++
	s.projectID = projectID
	s.inputIDs = append([]uuid.UUID(nil), fieldDeviceIDs...)
	return s.successIDs, s.errors
}

func (s *projectFieldDeviceLinkStub) UpdateFieldDevice(
	context.Context,
	uuid.UUID,
	uuid.UUID,
	uuid.UUID,
) (*domainProject.ProjectFieldDevice, error) {
	s.updateCalls++
	return nil, errors.New("legacy UpdateFieldDevice must not be called")
}

func TestCreateProjectFieldDeviceAuthorizesThenUsesTypedAssignmentAndNotificationOnly(t *testing.T) {
	gin.SetMode(gin.TestMode)
	requesterID := uuid.New()
	projectID := uuid.New()
	fieldDeviceID := uuid.New()
	linkID := uuid.New()
	createdAt := time.Date(2026, time.July, 21, 19, 0, 0, 0, time.UTC)
	access := &projectFieldDeviceAccessStub{hasAccess: true, hasPermission: true}
	link := &projectFieldDeviceLinkStub{}
	assigned := &domainProject.ProjectFieldDevice{
		ProjectID:     projectID,
		FieldDeviceID: fieldDeviceID,
	}
	assigned.ID = linkID
	assigned.CreatedAt = createdAt
	assigned.UpdatedAt = createdAt
	assigner := &projectFieldDeviceAssignerStub{result: assigned}
	combinedNotifyCalls := 0
	var notifiedProjectID uuid.UUID
	var notifiedEvent string
	var notifiedIDs []string
	handler := NewHandler(
		access,
		link,
		&projectFieldDeviceMultiCreatorStub{},
		assigner,
		nil,
		nil,
		func(*gin.Context, uuid.UUID, string, ...string) { combinedNotifyCalls++ },
		func(_ *gin.Context, projectID uuid.UUID, event string, ids ...string) {
			notifiedProjectID = projectID
			notifiedEvent = event
			notifiedIDs = append([]string(nil), ids...)
		},
	)
	body, err := json.Marshal(projectdto.CreateProjectFieldDeviceRequest{
		FieldDeviceID: fieldDeviceID,
	})
	if err != nil {
		t.Fatalf("encode request: %v", err)
	}
	router := projectFieldDeviceTestRouter(requesterID, handler)
	request := httptest.NewRequest(
		http.MethodPost,
		"/projects/"+projectID.String()+"/field-devices",
		bytes.NewReader(body),
	)
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("response status: got %d body=%s", recorder.Code, recorder.Body.String())
	}
	if access.accessCalls != 1 || access.permissionCalls != 1 ||
		access.requesterID != requesterID || access.projectID != projectID ||
		access.requestedPermission != domainUser.PermissionProjectFieldDeviceCreate {
		t.Fatalf("access checks: %+v", access)
	}
	if assigner.calls != 1 || assigner.command.ProjectID != projectID ||
		assigner.command.FieldDeviceID != fieldDeviceID || link.createCalls != 0 {
		t.Fatalf("typed assignment routing: assigner=%+v link=%+v", assigner, link)
	}
	if combinedNotifyCalls != 0 || notifiedProjectID != projectID ||
		notifiedEvent != "project.field_device.created" ||
		!reflect.DeepEqual(notifiedIDs, []string{fieldDeviceID.String()}) {
		t.Fatalf("notification split: combined=%d project=%s event=%q ids=%v",
			combinedNotifyCalls,
			notifiedProjectID,
			notifiedEvent,
			notifiedIDs,
		)
	}
	var response projectdto.ProjectFieldDeviceResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response.ID != linkID || response.ProjectID != projectID ||
		response.FieldDeviceID != fieldDeviceID || !response.CreatedAt.Equal(createdAt) {
		t.Fatalf("response contract changed: %+v", response)
	}
	for _, internalKey := range []string{"operation_id", "batch_id", "changes"} {
		if strings.Contains(recorder.Body.String(), `"`+internalKey+`"`) {
			t.Fatalf("internal mutation metadata leaked: %s", recorder.Body.String())
		}
	}
}

func TestCreateProjectFieldDeviceRejectsUnauthorizedProjectBeforeAssignment(t *testing.T) {
	gin.SetMode(gin.TestMode)
	requesterID := uuid.New()
	projectID := uuid.New()
	fieldDeviceID := uuid.New()
	access := &projectFieldDeviceAccessStub{hasAccess: false, hasPermission: true}
	assigner := &projectFieldDeviceAssignerStub{}
	notifyCalls := 0
	handler := NewHandler(
		access,
		&projectFieldDeviceLinkStub{},
		&projectFieldDeviceMultiCreatorStub{},
		assigner,
		nil,
		nil,
		nil,
		func(*gin.Context, uuid.UUID, string, ...string) { notifyCalls++ },
	)
	body, err := json.Marshal(projectdto.CreateProjectFieldDeviceRequest{
		FieldDeviceID: fieldDeviceID,
	})
	if err != nil {
		t.Fatalf("encode request: %v", err)
	}
	router := projectFieldDeviceTestRouter(requesterID, handler)
	request := httptest.NewRequest(
		http.MethodPost,
		"/projects/"+projectID.String()+"/field-devices",
		bytes.NewReader(body),
	)
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusForbidden || assigner.calls != 0 ||
		access.permissionCalls != 0 || notifyCalls != 0 {
		t.Fatalf("unauthorized assignment advanced: status=%d assigner=%d permissions=%d notify=%d body=%s",
			recorder.Code,
			assigner.calls,
			access.permissionCalls,
			notifyCalls,
			recorder.Body.String(),
		)
	}
}

func TestMultiCreateProjectFieldDevicesAuthorizesThenUsesTypedCreateAndAssignCommand(t *testing.T) {
	gin.SetMode(gin.TestMode)
	requesterID := uuid.New()
	projectID := uuid.New()
	firstParentID := uuid.New()
	secondParentID := uuid.New()
	firstSystemPartID := uuid.New()
	secondSystemPartID := uuid.New()
	firstApparatID := uuid.New()
	secondApparatID := uuid.New()
	objectDataID := uuid.New()
	firstID := uuid.New()
	createdAt := time.Date(2026, time.July, 21, 10, 0, 0, 0, time.UTC)
	bmk := "M01"
	description := "supply air sensor"
	firstNumber := 4
	secondNumber := 5
	emptyText := ""
	access := &projectFieldDeviceAccessStub{hasAccess: true, hasPermission: true}
	creator := &projectFieldDeviceMultiCreatorStub{
		result: &domainFacility.FieldDeviceMultiCreateResult{
			Results: []domainFacility.FieldDeviceCreateResult{
				{
					Index:   0,
					Success: true,
					FieldDevice: &domainFacility.FieldDevice{
						Base:                      domain.Base{ID: firstID, CreatedAt: createdAt, UpdatedAt: createdAt},
						BMK:                       &bmk,
						Description:               &description,
						ApparatNr:                 firstNumber,
						SPSControllerSystemTypeID: firstParentID,
						SystemPartID:              firstSystemPartID,
						ApparatID:                 firstApparatID,
					},
				},
				{
					Index:      1,
					Error:      "BACnet validation failed",
					ErrorField: "bacnet_objects",
				},
			},
			TotalRequests: 2,
			SuccessCount:  1,
			FailureCount:  1,
		},
	}
	link := &projectFieldDeviceLinkStub{}
	notifyCalls := 0
	handler := NewHandler(
		access,
		link,
		creator,
		nil,
		nil,
		nil,
		func(*gin.Context, uuid.UUID, string, ...string) { notifyCalls++ },
		nil,
	)
	requestBody, err := json.Marshal(projectdto.MultiCreateProjectFieldDeviceRequest{
		FieldDevices: []facilitydto.CreateFieldDeviceRequest{
			{
				BMK:                       &bmk,
				Description:               &description,
				ApparatNr:                 &firstNumber,
				SPSControllerSystemTypeID: firstParentID,
				SystemPartID:              firstSystemPartID,
				ApparatID:                 firstApparatID,
				ObjectDataID:              &objectDataID,
			},
			{
				ApparatNr:                 &secondNumber,
				SPSControllerSystemTypeID: secondParentID,
				SystemPartID:              secondSystemPartID,
				ApparatID:                 secondApparatID,
				BacnetObjects: []facilitydto.BacnetObjectInput{{
					TextFix:        "AI05",
					TextIndividual: &emptyText,
					SoftwareType:   "AI",
					SoftwareNumber: 5,
				}},
			},
		},
	})
	if err != nil {
		t.Fatalf("encode request: %v", err)
	}
	router := projectFieldDeviceTestRouter(requesterID, handler)
	request := httptest.NewRequest(
		http.MethodPost,
		"/projects/"+projectID.String()+"/field-devices/multi-create",
		bytes.NewReader(requestBody),
	)
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("response status: got %d, body=%s", recorder.Code, recorder.Body.String())
	}
	if access.accessCalls != 1 || access.permissionCalls != 1 ||
		access.requesterID != requesterID || access.projectID != projectID ||
		access.requestedPermission != domainUser.PermissionProjectFieldDeviceCreate {
		t.Fatalf("access checks: %+v", access)
	}
	if creator.calls != 1 || creator.command.ProjectID != projectID ||
		len(creator.command.Items) != 2 || link.calls != 0 {
		t.Fatalf("command routing: creator=%+v link=%+v", creator, link)
	}
	firstItem := creator.command.Items[0]
	if firstItem.FieldDevice == nil || firstItem.FieldDevice.BMK == nil ||
		*firstItem.FieldDevice.BMK != bmk || firstItem.FieldDevice.ApparatNr != firstNumber ||
		firstItem.FieldDevice.SPSControllerSystemTypeID != firstParentID ||
		firstItem.FieldDevice.SystemPartID != firstSystemPartID ||
		firstItem.FieldDevice.ApparatID != firstApparatID || firstItem.ObjectDataID == nil ||
		*firstItem.ObjectDataID != objectDataID {
		t.Fatalf("first command item: %+v", firstItem)
	}
	secondItem := creator.command.Items[1]
	if secondItem.FieldDevice == nil || secondItem.FieldDevice.ApparatNr != secondNumber ||
		secondItem.FieldDevice.SPSControllerSystemTypeID != secondParentID ||
		len(secondItem.BacnetObjects) != 1 ||
		secondItem.BacnetObjects[0].SoftwareType != domainFacility.BacnetSoftwareType("AI") ||
		secondItem.BacnetObjects[0].TextIndividual != nil {
		t.Fatalf("second command item: %+v", secondItem)
	}
	if notifyCalls != 0 {
		t.Fatalf("handler bypassed application dispatch with %d direct notifications", notifyCalls)
	}
	var response facilitydto.MultiCreateFieldDeviceResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response.TotalRequests != 2 || response.SuccessCount != 1 || response.FailureCount != 1 ||
		len(response.Results) != 2 || !response.Results[0].Success ||
		response.Results[0].FieldDevice == nil || response.Results[0].FieldDevice.ID != firstID ||
		response.Results[1].Success || response.Results[1].Error != "BACnet validation failed" ||
		response.Results[1].ErrorField != "bacnet_objects" {
		t.Fatalf("response contract changed: %+v", response)
	}
	var responseFields map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &responseFields); err != nil {
		t.Fatalf("decode response fields: %v", err)
	}
	if _, ok := responseFields["operation_id"]; ok {
		t.Fatalf("internal mutation metadata leaked into v1 response: %+v", responseFields)
	}
}

func TestMultiCreateProjectFieldDevicesRejectsUnauthorizedProjectBeforeApplicationCommand(t *testing.T) {
	gin.SetMode(gin.TestMode)
	requesterID := uuid.New()
	projectID := uuid.New()
	access := &projectFieldDeviceAccessStub{hasAccess: false, hasPermission: true}
	creator := &projectFieldDeviceMultiCreatorStub{}
	handler := NewHandler(access, &projectFieldDeviceLinkStub{}, creator, nil, nil, nil, nil, nil)
	router := projectFieldDeviceTestRouter(requesterID, handler)
	request := httptest.NewRequest(
		http.MethodPost,
		"/projects/"+projectID.String()+"/field-devices/multi-create",
		bytes.NewReader([]byte(`{"field_devices":[{"apparat_nr":1}]}`)),
	)
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusForbidden {
		t.Fatalf("response status: got %d, body=%s", recorder.Code, recorder.Body.String())
	}
	if creator.calls != 0 || access.permissionCalls != 0 {
		t.Fatalf("unauthorized request advanced: creator=%d permissions=%d",
			creator.calls,
			access.permissionCalls,
		)
	}
}

func TestMultiCreateProjectFieldDevicesPreservesApplicationErrorMapping(t *testing.T) {
	gin.SetMode(gin.TestMode)
	requesterID := uuid.New()
	projectID := uuid.New()
	access := &projectFieldDeviceAccessStub{hasAccess: true, hasPermission: true}
	creator := &projectFieldDeviceMultiCreatorStub{err: domain.ErrNotFound}
	handler := NewHandler(access, &projectFieldDeviceLinkStub{}, creator, nil, nil, nil, nil, nil)
	router := projectFieldDeviceTestRouter(requesterID, handler)
	request := httptest.NewRequest(
		http.MethodPost,
		"/projects/"+projectID.String()+"/field-devices/multi-create",
		bytes.NewReader(validProjectFieldDeviceMultiCreateBody(t)),
	)
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNotFound || creator.calls != 1 {
		t.Fatalf("error mapping: status=%d calls=%d body=%s",
			recorder.Code,
			creator.calls,
			recorder.Body.String(),
		)
	}
}

func TestMultiCreateProjectFieldDevicesUsesTypedBulkAssignmentAndNotificationOnly(t *testing.T) {
	gin.SetMode(gin.TestMode)
	requesterID := uuid.New()
	projectID := uuid.New()
	firstID := uuid.New()
	secondID := uuid.New()
	access := &projectFieldDeviceAccessStub{hasAccess: true, hasPermission: true}
	creator := &projectFieldDeviceMultiCreatorStub{}
	link := &projectFieldDeviceLinkStub{}
	bulkAssigner := &projectFieldDeviceBulkAssignerStub{
		result: appfielddevice.BulkAssignToProjectResult{
			SuccessFieldDeviceIDs: []uuid.UUID{firstID},
			AssociationErrors:     []string{"second association failed"},
		},
	}
	combinedNotifyCalls := 0
	var notifiedProjectID uuid.UUID
	var notifiedEvent string
	var notifiedIDs []string
	handler := NewHandler(
		access,
		link,
		creator,
		nil,
		bulkAssigner,
		nil,
		func(*gin.Context, uuid.UUID, string, ...string) { combinedNotifyCalls++ },
		func(_ *gin.Context, projectID uuid.UUID, event string, ids ...string) {
			notifiedProjectID = projectID
			notifiedEvent = event
			notifiedIDs = append([]string(nil), ids...)
		},
	)
	body, err := json.Marshal(projectdto.MultiCreateProjectFieldDeviceRequest{
		FieldDeviceIDs: []uuid.UUID{firstID, secondID},
	})
	if err != nil {
		t.Fatalf("encode request: %v", err)
	}
	router := projectFieldDeviceTestRouter(requesterID, handler)
	request := httptest.NewRequest(
		http.MethodPost,
		"/projects/"+projectID.String()+"/field-devices/multi-create",
		bytes.NewReader(body),
	)
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK || link.calls != 0 || creator.calls != 0 ||
		bulkAssigner.calls != 1 || bulkAssigner.command.ProjectID != projectID ||
		!reflect.DeepEqual(bulkAssigner.command.FieldDeviceIDs, []uuid.UUID{firstID, secondID}) {
		t.Fatalf("existing-ID branch changed: status=%d link=%+v creator=%+v assigner=%+v body=%s",
			recorder.Code,
			link,
			creator,
			bulkAssigner,
			recorder.Body.String(),
		)
	}
	if combinedNotifyCalls != 0 || notifiedProjectID != projectID ||
		notifiedEvent != "project.field_device.multi_created" ||
		!reflect.DeepEqual(notifiedIDs, []string{firstID.String()}) {
		t.Fatalf("existing-ID notification changed: combined=%d project=%s event=%q ids=%v",
			combinedNotifyCalls,
			notifiedProjectID,
			notifiedEvent,
			notifiedIDs,
		)
	}
	var response projectdto.MultiCreateProjectFieldDeviceResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !reflect.DeepEqual(response.SuccessFieldDeviceIDs, []uuid.UUID{firstID}) ||
		!reflect.DeepEqual(response.AssociationErrors, []string{"second association failed"}) {
		t.Fatalf("existing-ID response changed: %+v", response)
	}
}

func TestUpdateProjectFieldDeviceAuthorizesThenUsesTypedReassignmentAndNotificationOnly(t *testing.T) {
	gin.SetMode(gin.TestMode)
	requesterID := uuid.New()
	projectID := uuid.New()
	linkID := uuid.New()
	fieldDeviceID := uuid.New()
	createdAt := time.Date(2026, time.July, 21, 22, 0, 0, 0, time.UTC)
	updatedAt := createdAt.Add(time.Hour)
	access := &projectFieldDeviceAccessStub{hasAccess: true, hasPermission: true}
	linkService := &projectFieldDeviceLinkStub{}
	updated := &domainProject.ProjectFieldDevice{
		ProjectID:     projectID,
		FieldDeviceID: fieldDeviceID,
	}
	updated.ID = linkID
	updated.CreatedAt = createdAt
	updated.UpdatedAt = updatedAt
	reassigner := &projectFieldDeviceReassignerStub{result: updated}
	combinedNotifyCalls := 0
	var notifiedProjectID uuid.UUID
	var notifiedEvent string
	var notifiedIDs []string
	handler := NewHandler(
		access,
		linkService,
		&projectFieldDeviceMultiCreatorStub{},
		nil,
		nil,
		reassigner,
		func(*gin.Context, uuid.UUID, string, ...string) { combinedNotifyCalls++ },
		func(_ *gin.Context, projectID uuid.UUID, event string, ids ...string) {
			notifiedProjectID = projectID
			notifiedEvent = event
			notifiedIDs = append([]string(nil), ids...)
		},
	)
	body, err := json.Marshal(projectdto.UpdateProjectFieldDeviceRequest{
		FieldDeviceID: fieldDeviceID,
	})
	if err != nil {
		t.Fatalf("encode request: %v", err)
	}
	router := projectFieldDeviceTestRouter(requesterID, handler)
	request := httptest.NewRequest(
		http.MethodPut,
		"/projects/"+projectID.String()+"/field-devices/"+linkID.String(),
		bytes.NewReader(body),
	)
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("response status: got %d body=%s", recorder.Code, recorder.Body.String())
	}
	if access.accessCalls != 1 || access.permissionCalls != 1 ||
		access.requesterID != requesterID || access.projectID != projectID ||
		access.requestedPermission != domainUser.PermissionProjectFieldDeviceUpdate {
		t.Fatalf("access checks: %+v", access)
	}
	if reassigner.calls != 1 || reassigner.command.ProjectID != projectID ||
		reassigner.command.LinkID != linkID ||
		reassigner.command.FieldDeviceID != fieldDeviceID || linkService.updateCalls != 0 {
		t.Fatalf("typed reassignment routing: reassigner=%+v link=%+v", reassigner, linkService)
	}
	if combinedNotifyCalls != 0 || notifiedProjectID != projectID ||
		notifiedEvent != "project.field_device.updated" ||
		!reflect.DeepEqual(notifiedIDs, []string{fieldDeviceID.String()}) {
		t.Fatalf("notification split: combined=%d project=%s event=%q ids=%v",
			combinedNotifyCalls,
			notifiedProjectID,
			notifiedEvent,
			notifiedIDs,
		)
	}
	var response projectdto.ProjectFieldDeviceResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response.ID != linkID || response.ProjectID != projectID ||
		response.FieldDeviceID != fieldDeviceID ||
		!response.CreatedAt.Equal(createdAt) || !response.UpdatedAt.Equal(updatedAt) {
		t.Fatalf("response contract changed: %+v", response)
	}
	for _, internalKey := range []string{"operation_id", "batch_id", "changes"} {
		if strings.Contains(recorder.Body.String(), `"`+internalKey+`"`) {
			t.Fatalf("internal mutation metadata leaked: %s", recorder.Body.String())
		}
	}
}

func TestUpdateProjectFieldDeviceRejectsUnauthorizedProjectBeforeReassignment(t *testing.T) {
	gin.SetMode(gin.TestMode)
	requesterID := uuid.New()
	projectID := uuid.New()
	linkID := uuid.New()
	fieldDeviceID := uuid.New()
	access := &projectFieldDeviceAccessStub{hasAccess: false, hasPermission: true}
	reassigner := &projectFieldDeviceReassignerStub{}
	notifyCalls := 0
	handler := NewHandler(
		access,
		&projectFieldDeviceLinkStub{},
		&projectFieldDeviceMultiCreatorStub{},
		nil,
		nil,
		reassigner,
		nil,
		func(*gin.Context, uuid.UUID, string, ...string) { notifyCalls++ },
	)
	body, err := json.Marshal(projectdto.UpdateProjectFieldDeviceRequest{
		FieldDeviceID: fieldDeviceID,
	})
	if err != nil {
		t.Fatalf("encode request: %v", err)
	}
	router := projectFieldDeviceTestRouter(requesterID, handler)
	request := httptest.NewRequest(
		http.MethodPut,
		"/projects/"+projectID.String()+"/field-devices/"+linkID.String(),
		bytes.NewReader(body),
	)
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusForbidden || reassigner.calls != 0 ||
		access.permissionCalls != 0 || notifyCalls != 0 {
		t.Fatalf("unauthorized reassignment advanced: status=%d reassigner=%d permissions=%d notify=%d body=%s",
			recorder.Code,
			reassigner.calls,
			access.permissionCalls,
			notifyCalls,
			recorder.Body.String(),
		)
	}
}

func projectFieldDeviceTestRouter(requesterID uuid.UUID, handler *Handler) *gin.Engine {
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set(middleware.ContextUserIDKey, requesterID)
		c.Next()
	})
	router.POST(
		"/projects/:id/field-devices/multi-create",
		handler.MultiCreateProjectFieldDevices,
	)
	router.POST(
		"/projects/:id/field-devices",
		handler.CreateProjectFieldDevice,
	)
	router.PUT(
		"/projects/:id/field-devices/:linkId",
		handler.UpdateProjectFieldDevice,
	)
	return router
}

func validProjectFieldDeviceMultiCreateBody(t *testing.T) []byte {
	t.Helper()
	number := 1
	body, err := json.Marshal(projectdto.MultiCreateProjectFieldDeviceRequest{
		FieldDevices: []facilitydto.CreateFieldDeviceRequest{{
			ApparatNr:                 &number,
			SPSControllerSystemTypeID: uuid.New(),
			SystemPartID:              uuid.New(),
			ApparatID:                 uuid.New(),
		}},
	})
	if err != nil {
		t.Fatalf("encode request: %v", err)
	}
	return body
}
