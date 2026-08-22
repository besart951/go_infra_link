package fielddevice

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	domainProject "github.com/besart951/go_infra_link/backend/internal/domain/project"
	domainUser "github.com/besart951/go_infra_link/backend/internal/domain/user"
	"github.com/besart951/go_infra_link/backend/internal/handler/middleware"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func TestProjectFieldDeviceLinkHandlersCreateAndRelinkTheRequestedLink(t *testing.T) {
	gin.SetMode(gin.TestMode)
	projectID, linkID, initialDeviceID, replacementDeviceID := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	service := &contractFacilityLinkService{}
	handler := NewHandler(contractAccessPolicy{}, service, nil)

	createRecorder := httptest.NewRecorder()
	createContext, _ := gin.CreateTestContext(createRecorder)
	createContext.Request = httptest.NewRequest(http.MethodPost, "/projects/"+projectID.String()+"/field-devices", strings.NewReader(`{"field_device_id":"`+initialDeviceID.String()+`"}`))
	createContext.Request.Header.Set("Content-Type", "application/json")
	createContext.Params = gin.Params{{Key: "id", Value: projectID.String()}}
	createContext.Set(middleware.ContextUserIDKey, uuid.New())
	handler.CreateProjectFieldDevice(createContext)

	if createRecorder.Code != http.StatusCreated {
		t.Fatalf("expected create status 201, got %d: %s", createRecorder.Code, createRecorder.Body.String())
	}
	if service.projectID != projectID || service.fieldDeviceID != initialDeviceID {
		t.Fatalf("expected create for project=%s fieldDevice=%s, got project=%s fieldDevice=%s", projectID, initialDeviceID, service.projectID, service.fieldDeviceID)
	}

	updateRecorder := httptest.NewRecorder()
	updateContext, _ := gin.CreateTestContext(updateRecorder)
	updateContext.Request = httptest.NewRequest(http.MethodPut, "/projects/"+projectID.String()+"/field-devices/"+linkID.String(), strings.NewReader(`{"field_device_id":"`+replacementDeviceID.String()+`","base_version":1}`))
	updateContext.Request.Header.Set("Content-Type", "application/json")
	updateContext.Params = gin.Params{{Key: "id", Value: projectID.String()}, {Key: "linkId", Value: linkID.String()}}
	updateContext.Set(middleware.ContextUserIDKey, uuid.New())
	handler.UpdateProjectFieldDevice(updateContext)

	if updateRecorder.Code != http.StatusOK {
		t.Fatalf("expected update status 200, got %d: %s", updateRecorder.Code, updateRecorder.Body.String())
	}
	if service.linkID != linkID || service.projectID != projectID || service.fieldDeviceID != replacementDeviceID {
		t.Fatalf("expected relink ids link=%s project=%s fieldDevice=%s, got link=%s project=%s fieldDevice=%s", linkID, projectID, replacementDeviceID, service.linkID, service.projectID, service.fieldDeviceID)
	}
}

type contractAccessPolicy struct{}

func (contractAccessPolicy) CanAccessProject(context.Context, uuid.UUID, uuid.UUID, *domainUser.Role) (bool, error) {
	return true, nil
}

func (contractAccessPolicy) CanUseProjectPermission(context.Context, uuid.UUID, *domainUser.Role, string) (bool, error) {
	return true, nil
}

func (contractAccessPolicy) CanUseProjectPermissionForProject(context.Context, uuid.UUID, uuid.UUID, *domainUser.Role, string) (bool, error) {
	return true, nil
}

type contractFacilityLinkService struct {
	linkID        uuid.UUID
	projectID     uuid.UUID
	fieldDeviceID uuid.UUID
}

func (s *contractFacilityLinkService) CreateFieldDevice(_ context.Context, projectID, fieldDeviceID uuid.UUID) (*domainProject.ProjectFieldDevice, error) {
	s.projectID, s.fieldDeviceID = projectID, fieldDeviceID
	return &domainProject.ProjectFieldDevice{ProjectID: projectID, FieldDeviceID: fieldDeviceID}, nil
}

func (s *contractFacilityLinkService) UpdateFieldDevice(_ context.Context, command domainProject.FacilityLinkUpdate) (*domainProject.ProjectFieldDevice, error) {
	s.linkID, s.projectID, s.fieldDeviceID = command.LinkID, command.ProjectID, command.TargetID
	return &domainProject.ProjectFieldDevice{ProjectID: command.ProjectID, FieldDeviceID: command.TargetID}, nil
}

func (s *contractFacilityLinkService) DeleteFieldDevice(context.Context, domainProject.FacilityLinkDelete) error {
	return nil
}

func (s *contractFacilityLinkService) MultiCreateFieldDevices(context.Context, uuid.UUID, []uuid.UUID) ([]uuid.UUID, []string) {
	return nil, nil
}

func (s *contractFacilityLinkService) MultiCreateAndAssignFieldDevices(context.Context, uuid.UUID, []domainFacility.FieldDeviceCreateItem) (*domainFacility.FieldDeviceMultiCreateResult, error) {
	return nil, nil
}

func (s *contractFacilityLinkService) ListFieldDevices(context.Context, uuid.UUID, int, int) (*domain.PaginatedList[domainProject.ProjectFieldDevice], error) {
	return nil, nil
}
