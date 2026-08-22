package controlcabinet

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

func TestUpdateProjectControlCabinetRelinksTheRequestedLink(t *testing.T) {
	gin.SetMode(gin.TestMode)
	projectID, linkID, controlCabinetID := uuid.New(), uuid.New(), uuid.New()
	service := &contractFacilityLinkService{}
	handler := NewHandler(contractAccessPolicy{}, service, nil)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodPut, "/projects/"+projectID.String()+"/control-cabinets/"+linkID.String(), strings.NewReader(`{"control_cabinet_id":"`+controlCabinetID.String()+`","base_version":1}`))
	ctx.Request.Header.Set("Content-Type", "application/json")
	ctx.Params = gin.Params{{Key: "id", Value: projectID.String()}, {Key: "linkId", Value: linkID.String()}}
	ctx.Set(middleware.ContextUserIDKey, uuid.New())

	handler.UpdateProjectControlCabinet(ctx)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", recorder.Code, recorder.Body.String())
	}
	if service.linkID != linkID || service.projectID != projectID || service.controlCabinetID != controlCabinetID {
		t.Fatalf("expected relink ids link=%s project=%s controlCabinet=%s, got link=%s project=%s controlCabinet=%s", linkID, projectID, controlCabinetID, service.linkID, service.projectID, service.controlCabinetID)
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
	linkID           uuid.UUID
	projectID        uuid.UUID
	controlCabinetID uuid.UUID
}

func (s *contractFacilityLinkService) CreateControlCabinet(context.Context, uuid.UUID, uuid.UUID) (*domainProject.ProjectControlCabinet, error) {
	return nil, nil
}

func (s *contractFacilityLinkService) CopyControlCabinet(context.Context, uuid.UUID, uuid.UUID) (*domainFacility.ControlCabinet, error) {
	return nil, nil
}

func (s *contractFacilityLinkService) UpdateControlCabinet(_ context.Context, command domainProject.FacilityLinkUpdate) (*domainProject.ProjectControlCabinet, error) {
	s.linkID, s.projectID, s.controlCabinetID = command.LinkID, command.ProjectID, command.TargetID
	return &domainProject.ProjectControlCabinet{ProjectID: command.ProjectID, ControlCabinetID: command.TargetID}, nil
}

func (s *contractFacilityLinkService) DeleteControlCabinet(context.Context, domainProject.FacilityLinkDelete) error {
	return nil
}

func (s *contractFacilityLinkService) ListControlCabinets(context.Context, uuid.UUID, int, int) (*domain.PaginatedList[domainProject.ProjectControlCabinet], error) {
	return nil, nil
}
