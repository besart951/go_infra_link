package objectdata

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	appobjectdata "github.com/besart951/go_infra_link/backend/internal/application/facility/objectdata"
	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	domainUser "github.com/besart951/go_infra_link/backend/internal/domain/user"
	projectdto "github.com/besart951/go_infra_link/backend/internal/handler/dto/project"
	"github.com/besart951/go_infra_link/backend/internal/handler/middleware"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func TestListProjectObjectDataReturnsAfterInvalidApparatID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	projectID := uuid.New()
	userID := uuid.New()
	accessService := &fakeProjectAccessPolicyService{hasAccess: true}
	handler := NewHandler(accessService, nil, nil, nil, nil)

	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	tracker := &statusTrackingWriter{ResponseWriter: context.Writer}
	context.Writer = tracker
	context.Request = httptest.NewRequest(http.MethodGet, "/projects/"+projectID.String()+"/object-data?apparat_id=not-a-uuid", nil)
	context.Params = gin.Params{{Key: "id", Value: projectID.String()}}
	context.Set(middleware.ContextUserIDKey, userID)

	handler.ListProjectObjectData(context)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d body=%s", recorder.Code, recorder.Body.String())
	}
	if accessService.calls != 1 {
		t.Fatalf("expected access check to run once, got %d call(s)", accessService.calls)
	}
	if accessService.lastRequesterID != userID || accessService.lastProjectID != projectID {
		t.Fatalf("expected access check for requester=%s project=%s, got requester=%s project=%s", userID, projectID, accessService.lastRequesterID, accessService.lastProjectID)
	}
	if len(tracker.statusWrites) != 1 {
		t.Fatalf("expected exactly one status write, got %v", tracker.statusWrites)
	}
	if tracker.statusWrites[0] != http.StatusBadRequest {
		t.Fatalf("expected only status write to be 400, got %v", tracker.statusWrites)
	}
}

type projectObjectDataAssociationStub struct {
	attachCommand     appobjectdata.AttachToProjectCommand
	deactivateCommand appobjectdata.DeactivateForProjectCommand
	attachResult      *domainFacility.ObjectData
	deactivateResult  *domainFacility.ObjectData
	attachErr         error
	deactivateErr     error
	attachCalls       int
	deactivateCalls   int
}

func (s *projectObjectDataAssociationStub) AttachToProject(
	_ context.Context,
	command appobjectdata.AttachToProjectCommand,
) (*domainFacility.ObjectData, error) {
	s.attachCalls++
	s.attachCommand = command
	return s.attachResult, s.attachErr
}

func (s *projectObjectDataAssociationStub) DeactivateForProject(
	_ context.Context,
	command appobjectdata.DeactivateForProjectCommand,
) (*domainFacility.ObjectData, error) {
	s.deactivateCalls++
	s.deactivateCommand = command
	return s.deactivateResult, s.deactivateErr
}

func TestAddProjectObjectDataAuthorizesThenUsesTypedCommandAndNotificationOnly(t *testing.T) {
	gin.SetMode(gin.TestMode)
	requesterID := uuid.New()
	projectID := uuid.New()
	objectDataID := uuid.New()
	occurredAt := time.Date(2026, time.July, 22, 10, 0, 0, 0, time.UTC)
	access := &fakeProjectAccessPolicyService{hasAccess: true}
	association := &projectObjectDataAssociationStub{
		attachResult: &domainFacility.ObjectData{
			Base:        domain.Base{ID: objectDataID, CreatedAt: occurredAt, UpdatedAt: occurredAt},
			Description: "AHU",
			Version:     "1",
			IsActive:    true,
			ProjectID:   &projectID,
		},
	}
	var notifiedProjectID uuid.UUID
	var notifiedEvent string
	var notifiedIDs []string
	handler := NewHandler(
		access,
		nil,
		association,
		association,
		func(_ *gin.Context, gotProjectID uuid.UUID, event string, ids ...string) {
			notifiedProjectID = gotProjectID
			notifiedEvent = event
			notifiedIDs = append([]string(nil), ids...)
		},
	)
	body, err := json.Marshal(projectdto.CreateProjectObjectDataRequest{
		ObjectDataID: objectDataID,
	})
	if err != nil {
		t.Fatalf("encode request: %v", err)
	}
	router := projectObjectDataTestRouter(requesterID, handler)
	request := httptest.NewRequest(
		http.MethodPost,
		"/projects/"+projectID.String()+"/object-data",
		bytes.NewReader(body),
	)
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("response status: got %d body=%s", recorder.Code, recorder.Body.String())
	}
	if access.calls != 1 || access.permissionCalls != 1 ||
		access.lastRequesterID != requesterID || access.lastProjectID != projectID ||
		access.lastPermission != domainUser.PermissionProjectUpdate {
		t.Fatalf("access checks: %+v", access)
	}
	if association.attachCalls != 1 || association.deactivateCalls != 0 ||
		association.attachCommand.ProjectID != projectID ||
		association.attachCommand.ObjectDataID != objectDataID {
		t.Fatalf("typed command routing: %+v", association)
	}
	if notifiedProjectID != projectID || notifiedEvent != "project.object_data.created" ||
		len(notifiedIDs) != 0 {
		t.Fatalf(
			"notification-only callback: project=%s event=%q ids=%v",
			notifiedProjectID,
			notifiedEvent,
			notifiedIDs,
		)
	}
	var response projectdto.ObjectDataResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response.ID != objectDataID || response.ProjectID == nil ||
		*response.ProjectID != projectID || !response.IsActive ||
		response.Description != "AHU" || response.Version != "1" {
		t.Fatalf("response contract changed: %+v", response)
	}
	for _, internalKey := range []string{"operation_id", "batch_id", "changes"} {
		if strings.Contains(recorder.Body.String(), `"`+internalKey+`"`) {
			t.Fatalf("internal mutation metadata leaked: %s", recorder.Body.String())
		}
	}
}

func TestRemoveProjectObjectDataUsesDeactivateCommandAndPreservesResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)
	requesterID := uuid.New()
	projectID := uuid.New()
	objectDataID := uuid.New()
	access := &fakeProjectAccessPolicyService{hasAccess: true}
	association := &projectObjectDataAssociationStub{
		deactivateResult: &domainFacility.ObjectData{
			Base:        domain.Base{ID: objectDataID},
			Description: "AHU",
			Version:     "1",
			ProjectID:   &projectID,
		},
	}
	var notifiedEvent string
	handler := NewHandler(
		access,
		nil,
		association,
		association,
		func(_ *gin.Context, gotProjectID uuid.UUID, event string, ids ...string) {
			if gotProjectID != projectID || len(ids) != 0 {
				t.Fatalf("unexpected notification scope: project=%s ids=%v", gotProjectID, ids)
			}
			notifiedEvent = event
		},
	)
	router := projectObjectDataTestRouter(requesterID, handler)
	request := httptest.NewRequest(
		http.MethodDelete,
		"/projects/"+projectID.String()+"/object-data/"+objectDataID.String(),
		nil,
	)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("response status: got %d body=%s", recorder.Code, recorder.Body.String())
	}
	if association.deactivateCalls != 1 || association.attachCalls != 0 ||
		association.deactivateCommand.ProjectID != projectID ||
		association.deactivateCommand.ObjectDataID != objectDataID {
		t.Fatalf("typed command routing: %+v", association)
	}
	if notifiedEvent != "project.object_data.deleted" {
		t.Fatalf("notification event: got %q", notifiedEvent)
	}
	var response projectdto.ObjectDataResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response.ID != objectDataID || response.ProjectID == nil ||
		*response.ProjectID != projectID || response.IsActive {
		t.Fatalf("response contract changed: %+v", response)
	}
}

func TestAddProjectObjectDataRejectsUnauthorizedProjectBeforeCommand(t *testing.T) {
	gin.SetMode(gin.TestMode)
	requesterID := uuid.New()
	projectID := uuid.New()
	objectDataID := uuid.New()
	access := &fakeProjectAccessPolicyService{hasAccess: false}
	association := &projectObjectDataAssociationStub{}
	notifyCalls := 0
	handler := NewHandler(
		access,
		nil,
		association,
		association,
		func(*gin.Context, uuid.UUID, string, ...string) { notifyCalls++ },
	)
	body, err := json.Marshal(projectdto.CreateProjectObjectDataRequest{
		ObjectDataID: objectDataID,
	})
	if err != nil {
		t.Fatalf("encode request: %v", err)
	}
	router := projectObjectDataTestRouter(requesterID, handler)
	request := httptest.NewRequest(
		http.MethodPost,
		"/projects/"+projectID.String()+"/object-data",
		bytes.NewReader(body),
	)
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusForbidden || association.attachCalls != 0 ||
		access.permissionCalls != 0 || notifyCalls != 0 {
		t.Fatalf(
			"unauthorized command escaped gate: status=%d association=%+v access=%+v notify=%d",
			recorder.Code,
			association,
			access,
			notifyCalls,
		)
	}
}

func projectObjectDataTestRouter(
	requesterID uuid.UUID,
	handler *Handler,
) *gin.Engine {
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set(middleware.ContextUserIDKey, requesterID)
		c.Next()
	})
	router.POST("/projects/:id/object-data", handler.AddProjectObjectData)
	router.DELETE(
		"/projects/:id/object-data/:objectDataId",
		handler.RemoveProjectObjectData,
	)
	return router
}

type fakeProjectAccessPolicyService struct {
	hasAccess       bool
	err             error
	calls           int
	permissionCalls int
	lastRequesterID uuid.UUID
	lastProjectID   uuid.UUID
	lastPermission  string
}

func (s *fakeProjectAccessPolicyService) CanAccessProject(_ context.Context, requesterID, projectID uuid.UUID, _ *domainUser.Role) (bool, error) {
	s.calls++
	s.lastRequesterID = requesterID
	s.lastProjectID = projectID
	if s.err != nil {
		return false, s.err
	}
	return s.hasAccess, nil
}

func (s *fakeProjectAccessPolicyService) CanUseProjectPermission(_ context.Context, _ uuid.UUID, _ *domainUser.Role, _ string) (bool, error) {
	return true, nil
}

func (s *fakeProjectAccessPolicyService) CanUseProjectPermissionForProject(
	_ context.Context,
	_ uuid.UUID,
	projectID uuid.UUID,
	_ *domainUser.Role,
	permission string,
) (bool, error) {
	s.permissionCalls++
	s.lastProjectID = projectID
	s.lastPermission = permission
	return true, nil
}

type statusTrackingWriter struct {
	gin.ResponseWriter
	statusWrites []int
}

func (w *statusTrackingWriter) WriteHeader(code int) {
	w.statusWrites = append(w.statusWrites, code)
	w.ResponseWriter.WriteHeader(code)
}
