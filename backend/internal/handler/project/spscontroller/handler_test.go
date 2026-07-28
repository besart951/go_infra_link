package spscontroller

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

	appspscontroller "github.com/besart951/go_infra_link/backend/internal/application/facility/spscontroller"
	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	domainProject "github.com/besart951/go_infra_link/backend/internal/domain/project"
	domainUser "github.com/besart951/go_infra_link/backend/internal/domain/user"
	projectdto "github.com/besart951/go_infra_link/backend/internal/handler/dto/project"
	"github.com/besart951/go_infra_link/backend/internal/handler/middleware"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type projectCloneAccessStub struct {
	hasAccess           bool
	hasPermission       bool
	accessCalls         int
	permissionCalls     int
	requesterID         uuid.UUID
	projectID           uuid.UUID
	requestedPermission string
}

func (s *projectCloneAccessStub) CanAccessProject(
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

func (s *projectCloneAccessStub) CanUseProjectPermission(
	context.Context,
	uuid.UUID,
	*domainUser.Role,
	string,
) (bool, error) {
	return s.hasPermission, nil
}

func (s *projectCloneAccessStub) CanUseProjectPermissionForProject(
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

type projectSPSControllerClonerStub struct {
	command appspscontroller.CloneForProjectCommand
	result  *domainFacility.SPSController
	err     error
	calls   int
}

type projectSPSControllerSystemTypeClonerStub struct {
	command appspscontroller.CloneSystemTypeForProjectCommand
	result  *domainFacility.SPSControllerSystemType
	err     error
	calls   int
}

type projectSPSControllerAssignerStub struct {
	command appspscontroller.AssignToProjectCommand
	result  *domainProject.ProjectSPSController
	err     error
	calls   int
}

type projectSPSControllerReassignerStub struct {
	command appspscontroller.ReassignProjectLinkCommand
	result  *domainProject.ProjectSPSController
	err     error
	calls   int
}

func (s *projectSPSControllerAssignerStub) AssignToProject(
	_ context.Context,
	command appspscontroller.AssignToProjectCommand,
) (*domainProject.ProjectSPSController, error) {
	s.calls++
	s.command = command
	return s.result, s.err
}

func (s *projectSPSControllerReassignerStub) ReassignProjectLink(
	_ context.Context,
	command appspscontroller.ReassignProjectLinkCommand,
) (*domainProject.ProjectSPSController, error) {
	s.calls++
	s.command = command
	return s.result, s.err
}

type projectSPSControllerLinkStub struct {
	FacilityLinkService
	createCalls int
	updateCalls int
}

func (s *projectSPSControllerLinkStub) UpdateSPSController(
	context.Context,
	uuid.UUID,
	uuid.UUID,
	uuid.UUID,
) (*domainProject.ProjectSPSController, error) {
	s.updateCalls++
	return nil, errors.New("legacy UpdateSPSController must not be called")
}

func (s *projectSPSControllerLinkStub) CreateSPSController(
	context.Context,
	uuid.UUID,
	uuid.UUID,
) (*domainProject.ProjectSPSController, error) {
	s.createCalls++
	return nil, errors.New("legacy CreateSPSController must not be called")
}

func (s *projectSPSControllerSystemTypeClonerStub) CloneSystemTypeForProject(
	_ context.Context,
	command appspscontroller.CloneSystemTypeForProjectCommand,
) (*domainFacility.SPSControllerSystemType, error) {
	s.calls++
	s.command = command
	return s.result, s.err
}

func (s *projectSPSControllerClonerStub) CloneForProject(
	_ context.Context,
	command appspscontroller.CloneForProjectCommand,
) (*domainFacility.SPSController, error) {
	s.calls++
	s.command = command
	return s.result, s.err
}

func TestCreateProjectSPSControllerAuthorizesThenUsesTypedAssignmentAndNotificationOnly(t *testing.T) {
	gin.SetMode(gin.TestMode)
	requesterID := uuid.New()
	projectID := uuid.New()
	spsControllerID := uuid.New()
	linkID := uuid.New()
	createdAt := time.Date(2026, time.July, 22, 2, 0, 0, 0, time.UTC)
	access := &projectCloneAccessStub{hasAccess: true, hasPermission: true}
	linkService := &projectSPSControllerLinkStub{}
	assigned := &domainProject.ProjectSPSController{
		ProjectID:       projectID,
		SPSControllerID: spsControllerID,
	}
	assigned.ID = linkID
	assigned.CreatedAt = createdAt
	assigned.UpdatedAt = createdAt
	assigner := &projectSPSControllerAssignerStub{result: assigned}
	combinedNotifyCalls := 0
	var notifiedProjectID uuid.UUID
	var notifiedEvent string
	var notifiedIDs []string
	handler := NewHandler(
		access,
		linkService,
		&projectSPSControllerClonerStub{},
		&projectSPSControllerSystemTypeClonerStub{},
		assigner,
		nil,
		func(*gin.Context, uuid.UUID, string, ...string) { combinedNotifyCalls++ },
		func(_ *gin.Context, projectID uuid.UUID, event string, ids ...string) {
			notifiedProjectID = projectID
			notifiedEvent = event
			notifiedIDs = append([]string(nil), ids...)
		},
	)
	body, err := json.Marshal(projectdto.CreateProjectSPSControllerRequest{
		SPSControllerID: spsControllerID,
	})
	if err != nil {
		t.Fatalf("encode request: %v", err)
	}
	router := projectSPSControllerCreateTestRouter(requesterID, handler)
	request := httptest.NewRequest(
		http.MethodPost,
		"/projects/"+projectID.String()+"/sps-controllers",
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
		access.requestedPermission != domainUser.PermissionProjectSPSControllerCreate {
		t.Fatalf("access checks: %+v", access)
	}
	if assigner.calls != 1 || assigner.command.ProjectID != projectID ||
		assigner.command.SPSControllerID != spsControllerID || linkService.createCalls != 0 {
		t.Fatalf("typed assignment routing: assigner=%+v link=%+v", assigner, linkService)
	}
	if combinedNotifyCalls != 0 || notifiedProjectID != projectID ||
		notifiedEvent != "project.sps_controller.created" ||
		!reflect.DeepEqual(notifiedIDs, []string{spsControllerID.String()}) {
		t.Fatalf("notification split: combined=%d project=%s event=%q ids=%v",
			combinedNotifyCalls,
			notifiedProjectID,
			notifiedEvent,
			notifiedIDs,
		)
	}
	var response projectdto.ProjectSPSControllerResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response.ID != linkID || response.ProjectID != projectID ||
		response.SPSControllerID != spsControllerID ||
		!response.CreatedAt.Equal(createdAt) {
		t.Fatalf("response contract changed: %+v", response)
	}
	for _, internalKey := range []string{"operation_id", "batch_id", "changes"} {
		if strings.Contains(recorder.Body.String(), `"`+internalKey+`"`) {
			t.Fatalf("internal mutation metadata leaked: %s", recorder.Body.String())
		}
	}
}

func TestCreateProjectSPSControllerRejectsUnauthorizedProjectBeforeAssignment(t *testing.T) {
	gin.SetMode(gin.TestMode)
	requesterID := uuid.New()
	projectID := uuid.New()
	spsControllerID := uuid.New()
	access := &projectCloneAccessStub{hasAccess: false, hasPermission: true}
	assigner := &projectSPSControllerAssignerStub{}
	notifyCalls := 0
	handler := NewHandler(
		access,
		&projectSPSControllerLinkStub{},
		&projectSPSControllerClonerStub{},
		&projectSPSControllerSystemTypeClonerStub{},
		assigner,
		nil,
		nil,
		func(*gin.Context, uuid.UUID, string, ...string) { notifyCalls++ },
	)
	body, err := json.Marshal(projectdto.CreateProjectSPSControllerRequest{
		SPSControllerID: spsControllerID,
	})
	if err != nil {
		t.Fatalf("encode request: %v", err)
	}
	router := projectSPSControllerCreateTestRouter(requesterID, handler)
	request := httptest.NewRequest(
		http.MethodPost,
		"/projects/"+projectID.String()+"/sps-controllers",
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

func projectSPSControllerCreateTestRouter(
	requesterID uuid.UUID,
	handler *Handler,
) *gin.Engine {
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set(middleware.ContextUserIDKey, requesterID)
		c.Next()
	})
	router.POST(
		"/projects/:id/sps-controllers",
		handler.CreateProjectSPSController,
	)
	return router
}

func TestUpdateProjectSPSControllerAuthorizesThenUsesTypedReassignmentAndNotificationOnly(t *testing.T) {
	gin.SetMode(gin.TestMode)
	requesterID := uuid.New()
	projectID := uuid.New()
	linkID := uuid.New()
	spsControllerID := uuid.New()
	createdAt := time.Date(2026, time.July, 22, 3, 0, 0, 0, time.UTC)
	updatedAt := createdAt.Add(time.Minute)
	access := &projectCloneAccessStub{hasAccess: true, hasPermission: true}
	linkService := &projectSPSControllerLinkStub{}
	updated := &domainProject.ProjectSPSController{
		Base: domain.Base{
			ID:        linkID,
			CreatedAt: createdAt,
			UpdatedAt: updatedAt,
		},
		ProjectID:       projectID,
		SPSControllerID: spsControllerID,
	}
	reassigner := &projectSPSControllerReassignerStub{result: updated}
	combinedNotifyCalls := 0
	var notifiedProjectID uuid.UUID
	var notifiedEvent string
	var notifiedIDs []string
	handler := NewHandler(
		access,
		linkService,
		nil,
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
	body, err := json.Marshal(projectdto.UpdateProjectSPSControllerRequest{
		SPSControllerID: spsControllerID,
		ExpectedVersion: 1,
	})
	if err != nil {
		t.Fatalf("encode request: %v", err)
	}
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set(middleware.ContextUserIDKey, requesterID)
		c.Next()
	})
	router.PUT(
		"/projects/:id/sps-controllers/:linkId",
		handler.UpdateProjectSPSController,
	)
	request := httptest.NewRequest(
		http.MethodPut,
		"/projects/"+projectID.String()+"/sps-controllers/"+linkID.String(),
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
		access.requestedPermission != domainUser.PermissionProjectSPSControllerUpdate {
		t.Fatalf("access checks: %+v", access)
	}
	if reassigner.calls != 1 || reassigner.command.ProjectID != projectID ||
		reassigner.command.LinkID != linkID ||
		reassigner.command.SPSControllerID != spsControllerID || linkService.updateCalls != 0 {
		t.Fatalf("typed reassignment routing: reassigner=%+v link=%+v", reassigner, linkService)
	}
	if combinedNotifyCalls != 0 || notifiedProjectID != projectID ||
		notifiedEvent != "project.sps_controller.updated" ||
		!reflect.DeepEqual(notifiedIDs, []string{spsControllerID.String()}) {
		t.Fatalf("notification split: combined=%d project=%s event=%q ids=%v",
			combinedNotifyCalls,
			notifiedProjectID,
			notifiedEvent,
			notifiedIDs,
		)
	}
	var response projectdto.ProjectSPSControllerResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response.ID != linkID || response.ProjectID != projectID ||
		response.SPSControllerID != spsControllerID ||
		!response.CreatedAt.Equal(createdAt) || !response.UpdatedAt.Equal(updatedAt) {
		t.Fatalf("response contract changed: %+v", response)
	}
	for _, internalKey := range []string{"operation_id", "batch_id", "changes"} {
		if strings.Contains(recorder.Body.String(), `"`+internalKey+`"`) {
			t.Fatalf("internal mutation metadata leaked: %s", recorder.Body.String())
		}
	}
}

func TestUpdateProjectSPSControllerRejectsUnauthorizedProjectBeforeReassignment(t *testing.T) {
	gin.SetMode(gin.TestMode)
	requesterID := uuid.New()
	projectID := uuid.New()
	linkID := uuid.New()
	spsControllerID := uuid.New()
	access := &projectCloneAccessStub{hasAccess: false, hasPermission: true}
	reassigner := &projectSPSControllerReassignerStub{}
	notifyCalls := 0
	handler := NewHandler(
		access,
		&projectSPSControllerLinkStub{},
		nil,
		nil,
		nil,
		reassigner,
		nil,
		func(*gin.Context, uuid.UUID, string, ...string) { notifyCalls++ },
	)
	body, err := json.Marshal(projectdto.UpdateProjectSPSControllerRequest{
		SPSControllerID: spsControllerID,
		ExpectedVersion: 1,
	})
	if err != nil {
		t.Fatalf("encode request: %v", err)
	}
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set(middleware.ContextUserIDKey, requesterID)
		c.Next()
	})
	router.PUT(
		"/projects/:id/sps-controllers/:linkId",
		handler.UpdateProjectSPSController,
	)
	request := httptest.NewRequest(
		http.MethodPut,
		"/projects/"+projectID.String()+"/sps-controllers/"+linkID.String(),
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

func TestCopyProjectSPSControllerAuthorizesScopeThenUsesTypedApplicationCommand(t *testing.T) {
	gin.SetMode(gin.TestMode)
	requesterID := uuid.New()
	projectID := uuid.New()
	sourceID := uuid.New()
	copyID := uuid.New()
	cabinetID := uuid.New()
	gaDevice := "AAD"
	access := &projectCloneAccessStub{hasAccess: true, hasPermission: true}
	cloner := &projectSPSControllerClonerStub{result: &domainFacility.SPSController{
		Base:             domain.Base{ID: copyID},
		ControlCabinetID: cabinetID,
		GADevice:         &gaDevice,
		DeviceName:       "BLD_AK01_AAD",
	}}
	notifyCalls := 0
	handler := NewHandler(
		access,
		nil,
		cloner,
		nil,
		nil,
		nil,
		func(*gin.Context, uuid.UUID, string, ...string) { notifyCalls++ },
		nil,
	)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set(middleware.ContextUserIDKey, requesterID)
		c.Next()
	})
	router.POST(
		"/projects/:id/sps-controllers/:spsControllerId/copy",
		handler.CopyProjectSPSController,
	)
	request := httptest.NewRequest(
		http.MethodPost,
		"/projects/"+projectID.String()+"/sps-controllers/"+sourceID.String()+"/copy",
		nil,
	)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("response status: got %d, body=%s", recorder.Code, recorder.Body.String())
	}
	if access.accessCalls != 1 || access.permissionCalls != 1 ||
		access.requesterID != requesterID || access.projectID != projectID ||
		access.requestedPermission != domainUser.PermissionProjectSPSControllerCreate {
		t.Fatalf("access checks: %+v", access)
	}
	if cloner.calls != 1 || cloner.command.ProjectID != projectID ||
		cloner.command.SourceSPSControllerID != sourceID {
		t.Fatalf("clone command: calls=%d command=%+v", cloner.calls, cloner.command)
	}
	if notifyCalls != 0 {
		t.Fatalf("project handler bypassed application dispatch with %d notifications", notifyCalls)
	}
	var response map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response["id"] != copyID.String() || response["device_name"] != "BLD_AK01_AAD" {
		t.Fatalf("unexpected response: %+v", response)
	}
}

func TestCopyProjectSPSControllerDoesNotTrustUnauthorizedProjectID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	requesterID := uuid.New()
	projectID := uuid.New()
	access := &projectCloneAccessStub{hasAccess: false, hasPermission: true}
	cloner := &projectSPSControllerClonerStub{}
	handler := NewHandler(access, nil, cloner, nil, nil, nil, nil, nil)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set(middleware.ContextUserIDKey, requesterID)
		c.Next()
	})
	router.POST(
		"/projects/:id/sps-controllers/:spsControllerId/copy",
		handler.CopyProjectSPSController,
	)
	request := httptest.NewRequest(
		http.MethodPost,
		"/projects/"+projectID.String()+"/sps-controllers/"+uuid.NewString()+"/copy",
		nil,
	)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusForbidden {
		t.Fatalf("response status: got %d, body=%s", recorder.Code, recorder.Body.String())
	}
	if cloner.calls != 0 || access.permissionCalls != 0 {
		t.Fatalf("unauthorized request advanced: cloner=%d permissions=%d", cloner.calls, access.permissionCalls)
	}
}

func TestCopyProjectSPSControllerPreservesNotFoundMapping(t *testing.T) {
	gin.SetMode(gin.TestMode)
	requesterID := uuid.New()
	projectID := uuid.New()
	access := &projectCloneAccessStub{hasAccess: true, hasPermission: true}
	cloner := &projectSPSControllerClonerStub{err: domain.ErrNotFound}
	handler := NewHandler(access, nil, cloner, nil, nil, nil, nil, nil)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set(middleware.ContextUserIDKey, requesterID)
		c.Next()
	})
	router.POST(
		"/projects/:id/sps-controllers/:spsControllerId/copy",
		handler.CopyProjectSPSController,
	)
	request := httptest.NewRequest(
		http.MethodPost,
		"/projects/"+projectID.String()+"/sps-controllers/"+uuid.NewString()+"/copy",
		nil,
	)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("response status: got %d, body=%s", recorder.Code, recorder.Body.String())
	}
	if cloner.calls != 1 {
		t.Fatalf("application cloner calls: got %d, want 1", cloner.calls)
	}
}

func TestCopyProjectSPSControllerSystemTypeUsesTypedCommandAndNotificationOnlySeam(t *testing.T) {
	gin.SetMode(gin.TestMode)
	requesterID := uuid.New()
	projectID := uuid.New()
	sourceID := uuid.New()
	copyID := uuid.New()
	spsControllerID := uuid.New()
	systemTypeID := uuid.New()
	number := 3
	access := &projectCloneAccessStub{hasAccess: true, hasPermission: true}
	cloner := &projectSPSControllerSystemTypeClonerStub{
		result: &domainFacility.SPSControllerSystemType{
			Base:            domain.Base{ID: copyID},
			Number:          &number,
			SPSControllerID: spsControllerID,
			SystemTypeID:    systemTypeID,
		},
	}
	refreshNotifierCalls := 0
	eventNotifierCalls := 0
	var notifiedProjectID uuid.UUID
	var eventType string
	var entityIDs []string
	handler := NewHandler(
		access,
		nil,
		nil,
		cloner,
		nil,
		nil,
		func(*gin.Context, uuid.UUID, string, ...string) { refreshNotifierCalls++ },
		func(_ *gin.Context, projectID uuid.UUID, event string, ids ...string) {
			eventNotifierCalls++
			notifiedProjectID = projectID
			eventType = event
			entityIDs = append([]string(nil), ids...)
		},
	)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set(middleware.ContextUserIDKey, requesterID)
		c.Next()
	})
	router.POST(
		"/projects/:id/sps-controller-system-types/:systemTypeId/copy",
		handler.CopyProjectSPSControllerSystemType,
	)
	request := httptest.NewRequest(
		http.MethodPost,
		"/projects/"+projectID.String()+"/sps-controller-system-types/"+sourceID.String()+"/copy",
		nil,
	)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("response status: got %d, body=%s", recorder.Code, recorder.Body.String())
	}
	if access.accessCalls != 1 || access.permissionCalls != 1 ||
		access.requestedPermission != domainUser.PermissionProjectSPSControllerSystemTypeCreate {
		t.Fatalf("access checks: %+v", access)
	}
	if cloner.calls != 1 || cloner.command.ProjectID != projectID ||
		cloner.command.SourceSPSControllerSystemTypeID != sourceID {
		t.Fatalf("system-type clone command: calls=%d command=%+v", cloner.calls, cloner.command)
	}
	if refreshNotifierCalls != 0 {
		t.Fatalf("handler emitted %d legacy refresh notifications", refreshNotifierCalls)
	}
	if eventNotifierCalls != 1 || notifiedProjectID != projectID ||
		eventType != "project.sps_controller_system_type.copied" ||
		len(entityIDs) != 1 || entityIDs[0] != spsControllerID.String() {
		t.Fatalf("notification event: calls=%d project=%s type=%q ids=%v",
			eventNotifierCalls,
			notifiedProjectID,
			eventType,
			entityIDs,
		)
	}
	var response map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response["id"] != copyID.String() || response["sps_controller_id"] != spsControllerID.String() ||
		response["system_type_id"] != systemTypeID.String() || response["number"] != float64(number) {
		t.Fatalf("unexpected response: %+v", response)
	}
}

func TestCopyProjectSPSControllerSystemTypeDoesNotTrustUnauthorizedProjectID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	requesterID := uuid.New()
	projectID := uuid.New()
	access := &projectCloneAccessStub{hasAccess: false, hasPermission: true}
	cloner := &projectSPSControllerSystemTypeClonerStub{}
	eventCalls := 0
	handler := NewHandler(
		access,
		nil,
		nil,
		cloner,
		nil,
		nil,
		nil,
		func(*gin.Context, uuid.UUID, string, ...string) { eventCalls++ },
	)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set(middleware.ContextUserIDKey, requesterID)
		c.Next()
	})
	router.POST(
		"/projects/:id/sps-controller-system-types/:systemTypeId/copy",
		handler.CopyProjectSPSControllerSystemType,
	)
	request := httptest.NewRequest(
		http.MethodPost,
		"/projects/"+projectID.String()+"/sps-controller-system-types/"+uuid.NewString()+"/copy",
		nil,
	)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusForbidden {
		t.Fatalf("response status: got %d, body=%s", recorder.Code, recorder.Body.String())
	}
	if cloner.calls != 0 || eventCalls != 0 || access.permissionCalls != 0 {
		t.Fatalf("unauthorized request advanced: cloner=%d events=%d permissions=%d",
			cloner.calls,
			eventCalls,
			access.permissionCalls,
		)
	}
}

func TestCopyProjectSPSControllerSystemTypePreservesNotFoundMapping(t *testing.T) {
	gin.SetMode(gin.TestMode)
	requesterID := uuid.New()
	projectID := uuid.New()
	access := &projectCloneAccessStub{hasAccess: true, hasPermission: true}
	cloner := &projectSPSControllerSystemTypeClonerStub{err: domain.ErrNotFound}
	eventCalls := 0
	handler := NewHandler(
		access,
		nil,
		nil,
		cloner,
		nil,
		nil,
		nil,
		func(*gin.Context, uuid.UUID, string, ...string) { eventCalls++ },
	)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set(middleware.ContextUserIDKey, requesterID)
		c.Next()
	})
	router.POST(
		"/projects/:id/sps-controller-system-types/:systemTypeId/copy",
		handler.CopyProjectSPSControllerSystemType,
	)
	request := httptest.NewRequest(
		http.MethodPost,
		"/projects/"+projectID.String()+"/sps-controller-system-types/"+uuid.NewString()+"/copy",
		nil,
	)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("response status: got %d, body=%s", recorder.Code, recorder.Body.String())
	}
	if cloner.calls != 1 || eventCalls != 0 {
		t.Fatalf("failed command side effects: cloner=%d events=%d", cloner.calls, eventCalls)
	}
}
