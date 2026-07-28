package controlcabinet

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

	appcontrolcabinet "github.com/besart951/go_infra_link/backend/internal/application/facility/controlcabinet"
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

type projectControlCabinetClonerStub struct {
	command appcontrolcabinet.CloneForProjectCommand
	result  *domainFacility.ControlCabinet
	err     error
	calls   int
}

type projectControlCabinetAssignerStub struct {
	command appcontrolcabinet.AssignToProjectCommand
	result  *domainProject.ProjectControlCabinet
	err     error
	calls   int
}

type projectControlCabinetReassignerStub struct {
	command appcontrolcabinet.ReassignProjectLinkCommand
	result  *domainProject.ProjectControlCabinet
	err     error
	calls   int
}

func (s *projectControlCabinetAssignerStub) AssignToProject(
	_ context.Context,
	command appcontrolcabinet.AssignToProjectCommand,
) (*domainProject.ProjectControlCabinet, error) {
	s.calls++
	s.command = command
	return s.result, s.err
}

func (s *projectControlCabinetReassignerStub) ReassignProjectLink(
	_ context.Context,
	command appcontrolcabinet.ReassignProjectLinkCommand,
) (*domainProject.ProjectControlCabinet, error) {
	s.calls++
	s.command = command
	return s.result, s.err
}

type projectControlCabinetLinkStub struct {
	FacilityLinkService
	createCalls int
	updateCalls int
}

func (s *projectControlCabinetLinkStub) UpdateControlCabinet(
	context.Context,
	uuid.UUID,
	uuid.UUID,
	uuid.UUID,
) (*domainProject.ProjectControlCabinet, error) {
	s.updateCalls++
	return nil, errors.New("legacy UpdateControlCabinet must not be called")
}

func (s *projectControlCabinetLinkStub) CreateControlCabinet(
	context.Context,
	uuid.UUID,
	uuid.UUID,
) (*domainProject.ProjectControlCabinet, error) {
	s.createCalls++
	return nil, errors.New("legacy CreateControlCabinet must not be called")
}

func (s *projectControlCabinetClonerStub) CloneForProject(
	_ context.Context,
	command appcontrolcabinet.CloneForProjectCommand,
) (*domainFacility.ControlCabinet, error) {
	s.calls++
	s.command = command
	return s.result, s.err
}

func TestCreateProjectControlCabinetAuthorizesThenUsesTypedAssignmentAndNotificationOnly(t *testing.T) {
	gin.SetMode(gin.TestMode)
	requesterID := uuid.New()
	projectID := uuid.New()
	controlCabinetID := uuid.New()
	linkID := uuid.New()
	createdAt := time.Date(2026, time.July, 22, 0, 0, 0, 0, time.UTC)
	access := &projectCloneAccessStub{hasAccess: true, hasPermission: true}
	linkService := &projectControlCabinetLinkStub{}
	assigned := &domainProject.ProjectControlCabinet{
		ProjectID:        projectID,
		ControlCabinetID: controlCabinetID,
	}
	assigned.ID = linkID
	assigned.CreatedAt = createdAt
	assigned.UpdatedAt = createdAt
	assigner := &projectControlCabinetAssignerStub{result: assigned}
	combinedNotifyCalls := 0
	var notifiedProjectID uuid.UUID
	var notifiedEvent string
	var notifiedIDs []string
	handler := NewHandler(
		access,
		linkService,
		&projectControlCabinetClonerStub{},
		assigner,
		nil,
		func(*gin.Context, uuid.UUID, string, ...string) { combinedNotifyCalls++ },
		func(_ *gin.Context, projectID uuid.UUID, event string, ids ...string) {
			notifiedProjectID = projectID
			notifiedEvent = event
			notifiedIDs = append([]string(nil), ids...)
		},
	)
	body, err := json.Marshal(projectdto.CreateProjectControlCabinetRequest{
		ControlCabinetID: controlCabinetID,
	})
	if err != nil {
		t.Fatalf("encode request: %v", err)
	}
	router := projectControlCabinetCreateTestRouter(requesterID, handler)
	request := httptest.NewRequest(
		http.MethodPost,
		"/projects/"+projectID.String()+"/control-cabinets",
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
		access.requestedPermission != domainUser.PermissionProjectControlCabinetCreate {
		t.Fatalf("access checks: %+v", access)
	}
	if assigner.calls != 1 || assigner.command.ProjectID != projectID ||
		assigner.command.ControlCabinetID != controlCabinetID || linkService.createCalls != 0 {
		t.Fatalf("typed assignment routing: assigner=%+v link=%+v", assigner, linkService)
	}
	if combinedNotifyCalls != 0 || notifiedProjectID != projectID ||
		notifiedEvent != "project.control_cabinet.created" ||
		!reflect.DeepEqual(notifiedIDs, []string{controlCabinetID.String()}) {
		t.Fatalf("notification split: combined=%d project=%s event=%q ids=%v",
			combinedNotifyCalls,
			notifiedProjectID,
			notifiedEvent,
			notifiedIDs,
		)
	}
	var response projectdto.ProjectControlCabinetResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response.ID != linkID || response.ProjectID != projectID ||
		response.ControlCabinetID != controlCabinetID ||
		!response.CreatedAt.Equal(createdAt) {
		t.Fatalf("response contract changed: %+v", response)
	}
	for _, internalKey := range []string{"operation_id", "batch_id", "changes"} {
		if strings.Contains(recorder.Body.String(), `"`+internalKey+`"`) {
			t.Fatalf("internal mutation metadata leaked: %s", recorder.Body.String())
		}
	}
}

func TestCreateProjectControlCabinetRejectsUnauthorizedProjectBeforeAssignment(t *testing.T) {
	gin.SetMode(gin.TestMode)
	requesterID := uuid.New()
	projectID := uuid.New()
	controlCabinetID := uuid.New()
	access := &projectCloneAccessStub{hasAccess: false, hasPermission: true}
	assigner := &projectControlCabinetAssignerStub{}
	notifyCalls := 0
	handler := NewHandler(
		access,
		&projectControlCabinetLinkStub{},
		&projectControlCabinetClonerStub{},
		assigner,
		nil,
		nil,
		func(*gin.Context, uuid.UUID, string, ...string) { notifyCalls++ },
	)
	body, err := json.Marshal(projectdto.CreateProjectControlCabinetRequest{
		ControlCabinetID: controlCabinetID,
	})
	if err != nil {
		t.Fatalf("encode request: %v", err)
	}
	router := projectControlCabinetCreateTestRouter(requesterID, handler)
	request := httptest.NewRequest(
		http.MethodPost,
		"/projects/"+projectID.String()+"/control-cabinets",
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

func projectControlCabinetCreateTestRouter(
	requesterID uuid.UUID,
	handler *Handler,
) *gin.Engine {
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set(middleware.ContextUserIDKey, requesterID)
		c.Next()
	})
	router.POST(
		"/projects/:id/control-cabinets",
		handler.CreateProjectControlCabinet,
	)
	return router
}

func TestUpdateProjectControlCabinetAuthorizesThenUsesTypedReassignmentAndNotificationOnly(t *testing.T) {
	gin.SetMode(gin.TestMode)
	requesterID := uuid.New()
	projectID := uuid.New()
	linkID := uuid.New()
	controlCabinetID := uuid.New()
	createdAt := time.Date(2026, time.July, 22, 5, 0, 0, 0, time.UTC)
	updatedAt := createdAt.Add(time.Minute)
	access := &projectCloneAccessStub{hasAccess: true, hasPermission: true}
	linkService := &projectControlCabinetLinkStub{}
	updated := &domainProject.ProjectControlCabinet{
		Base: domain.Base{
			ID:        linkID,
			CreatedAt: createdAt,
			UpdatedAt: updatedAt,
		},
		ProjectID:        projectID,
		ControlCabinetID: controlCabinetID,
	}
	reassigner := &projectControlCabinetReassignerStub{result: updated}
	combinedNotifyCalls := 0
	var notifiedProjectID uuid.UUID
	var notifiedEvent string
	var notifiedIDs []string
	handler := NewHandler(
		access,
		linkService,
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
	body, err := json.Marshal(projectdto.UpdateProjectControlCabinetRequest{
		ControlCabinetID: controlCabinetID,
		ExpectedVersion:  1,
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
		"/projects/:id/control-cabinets/:linkId",
		handler.UpdateProjectControlCabinet,
	)
	request := httptest.NewRequest(
		http.MethodPut,
		"/projects/"+projectID.String()+"/control-cabinets/"+linkID.String(),
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
		access.requestedPermission != domainUser.PermissionProjectControlCabinetUpdate {
		t.Fatalf("access checks: %+v", access)
	}
	if reassigner.calls != 1 || reassigner.command.ProjectID != projectID ||
		reassigner.command.LinkID != linkID ||
		reassigner.command.ControlCabinetID != controlCabinetID || linkService.updateCalls != 0 {
		t.Fatalf("typed reassignment routing: reassigner=%+v link=%+v", reassigner, linkService)
	}
	if combinedNotifyCalls != 0 || notifiedProjectID != projectID ||
		notifiedEvent != "project.control_cabinet.updated" ||
		!reflect.DeepEqual(notifiedIDs, []string{controlCabinetID.String()}) {
		t.Fatalf("notification split: combined=%d project=%s event=%q ids=%v",
			combinedNotifyCalls,
			notifiedProjectID,
			notifiedEvent,
			notifiedIDs,
		)
	}
	var response projectdto.ProjectControlCabinetResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response.ID != linkID || response.ProjectID != projectID ||
		response.ControlCabinetID != controlCabinetID ||
		!response.CreatedAt.Equal(createdAt) || !response.UpdatedAt.Equal(updatedAt) {
		t.Fatalf("response contract changed: %+v", response)
	}
	for _, internalKey := range []string{"operation_id", "batch_id", "changes"} {
		if strings.Contains(recorder.Body.String(), `"`+internalKey+`"`) {
			t.Fatalf("internal mutation metadata leaked: %s", recorder.Body.String())
		}
	}
}

func TestUpdateProjectControlCabinetRejectsUnauthorizedProjectBeforeReassignment(t *testing.T) {
	gin.SetMode(gin.TestMode)
	requesterID := uuid.New()
	projectID := uuid.New()
	linkID := uuid.New()
	controlCabinetID := uuid.New()
	access := &projectCloneAccessStub{hasAccess: false, hasPermission: true}
	reassigner := &projectControlCabinetReassignerStub{}
	notifyCalls := 0
	handler := NewHandler(
		access,
		&projectControlCabinetLinkStub{},
		nil,
		nil,
		reassigner,
		nil,
		func(*gin.Context, uuid.UUID, string, ...string) { notifyCalls++ },
	)
	body, err := json.Marshal(projectdto.UpdateProjectControlCabinetRequest{
		ControlCabinetID: controlCabinetID,
		ExpectedVersion:  1,
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
		"/projects/:id/control-cabinets/:linkId",
		handler.UpdateProjectControlCabinet,
	)
	request := httptest.NewRequest(
		http.MethodPut,
		"/projects/"+projectID.String()+"/control-cabinets/"+linkID.String(),
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

func TestCopyProjectControlCabinetAuthorizesScopeThenUsesTypedApplicationCommand(t *testing.T) {
	gin.SetMode(gin.TestMode)
	requesterID := uuid.New()
	projectID := uuid.New()
	sourceID := uuid.New()
	copyID := uuid.New()
	buildingID := uuid.New()
	number := "AK04"
	access := &projectCloneAccessStub{hasAccess: true, hasPermission: true}
	cloner := &projectControlCabinetClonerStub{result: &domainFacility.ControlCabinet{
		Base:             domain.Base{ID: copyID},
		BuildingID:       buildingID,
		ControlCabinetNr: &number,
	}}
	notifyCalls := 0
	handler := NewHandler(
		access,
		nil,
		cloner,
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
		"/projects/:id/control-cabinets/:controlCabinetId/copy",
		handler.CopyProjectControlCabinet,
	)
	request := httptest.NewRequest(
		http.MethodPost,
		"/projects/"+projectID.String()+"/control-cabinets/"+sourceID.String()+"/copy",
		nil,
	)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("response status: got %d, body=%s", recorder.Code, recorder.Body.String())
	}
	if access.accessCalls != 1 || access.permissionCalls != 1 ||
		access.requesterID != requesterID || access.projectID != projectID ||
		access.requestedPermission != domainUser.PermissionProjectControlCabinetCreate {
		t.Fatalf("access checks: %+v", access)
	}
	if cloner.calls != 1 || cloner.command.ProjectID != projectID ||
		cloner.command.SourceControlCabinetID != sourceID {
		t.Fatalf("clone command: calls=%d command=%+v", cloner.calls, cloner.command)
	}
	if notifyCalls != 0 {
		t.Fatalf("project handler bypassed application dispatch with %d notifications", notifyCalls)
	}
	var response map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response["id"] != copyID.String() || response["control_cabinet_nr"] != number {
		t.Fatalf("unexpected response: %+v", response)
	}
}

func TestCopyProjectControlCabinetDoesNotTrustUnauthorizedProjectID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	requesterID := uuid.New()
	projectID := uuid.New()
	access := &projectCloneAccessStub{hasAccess: false, hasPermission: true}
	cloner := &projectControlCabinetClonerStub{}
	handler := NewHandler(access, nil, cloner, nil, nil, nil, nil)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set(middleware.ContextUserIDKey, requesterID)
		c.Next()
	})
	router.POST(
		"/projects/:id/control-cabinets/:controlCabinetId/copy",
		handler.CopyProjectControlCabinet,
	)
	request := httptest.NewRequest(
		http.MethodPost,
		"/projects/"+projectID.String()+"/control-cabinets/"+uuid.NewString()+"/copy",
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

func TestCopyProjectControlCabinetPreservesNotFoundMapping(t *testing.T) {
	gin.SetMode(gin.TestMode)
	requesterID := uuid.New()
	projectID := uuid.New()
	access := &projectCloneAccessStub{hasAccess: true, hasPermission: true}
	cloner := &projectControlCabinetClonerStub{err: domain.ErrNotFound}
	handler := NewHandler(access, nil, cloner, nil, nil, nil, nil)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set(middleware.ContextUserIDKey, requesterID)
		c.Next()
	})
	router.POST(
		"/projects/:id/control-cabinets/:controlCabinetId/copy",
		handler.CopyProjectControlCabinet,
	)
	request := httptest.NewRequest(
		http.MethodPost,
		"/projects/"+projectID.String()+"/control-cabinets/"+uuid.NewString()+"/copy",
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
