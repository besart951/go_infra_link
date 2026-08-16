package spscontroller

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

func TestUpdateProjectSPSControllerRelinksTheRequestedLink(t *testing.T) {
	gin.SetMode(gin.TestMode)
	projectID, linkID, controllerID := uuid.New(), uuid.New(), uuid.New()
	service := &contractFacilityLinkService{}
	handler := NewHandler(contractAccessPolicy{}, service, nil)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodPut, "/projects/"+projectID.String()+"/sps-controllers/"+linkID.String(), strings.NewReader(`{"sps_controller_id":"`+controllerID.String()+`"}`))
	ctx.Request.Header.Set("Content-Type", "application/json")
	ctx.Params = gin.Params{{Key: "id", Value: projectID.String()}, {Key: "linkId", Value: linkID.String()}}
	ctx.Set(middleware.ContextUserIDKey, uuid.New())

	handler.UpdateProjectSPSController(ctx)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", recorder.Code, recorder.Body.String())
	}
	if service.linkID != linkID || service.projectID != projectID || service.controllerID != controllerID {
		t.Fatalf("expected relink ids link=%s project=%s controller=%s, got link=%s project=%s controller=%s", linkID, projectID, controllerID, service.linkID, service.projectID, service.controllerID)
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
	linkID       uuid.UUID
	projectID    uuid.UUID
	controllerID uuid.UUID
}

func (s *contractFacilityLinkService) CreateSPSController(context.Context, uuid.UUID, uuid.UUID) (*domainProject.ProjectSPSController, error) {
	return nil, nil
}

func (s *contractFacilityLinkService) CopySPSController(context.Context, uuid.UUID, uuid.UUID) (*domainFacility.SPSController, error) {
	return nil, nil
}

func (s *contractFacilityLinkService) CopySPSControllerSystemType(context.Context, uuid.UUID, uuid.UUID) (*domainFacility.SPSControllerSystemType, error) {
	return nil, nil
}

func (s *contractFacilityLinkService) UpdateSPSController(_ context.Context, linkID, projectID, controllerID uuid.UUID) (*domainProject.ProjectSPSController, error) {
	s.linkID, s.projectID, s.controllerID = linkID, projectID, controllerID
	return &domainProject.ProjectSPSController{ProjectID: projectID, SPSControllerID: controllerID}, nil
}

func (s *contractFacilityLinkService) DeleteSPSController(context.Context, uuid.UUID, uuid.UUID) error {
	return nil
}

func (s *contractFacilityLinkService) ListSPSControllers(context.Context, uuid.UUID, int, int) (*domain.PaginatedList[domainProject.ProjectSPSController], error) {
	return nil, nil
}
