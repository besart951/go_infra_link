package shared

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/besart951/go_infra_link/backend/internal/domain/project"
	domainUser "github.com/besart951/go_infra_link/backend/internal/domain/user"
	dto "github.com/besart951/go_infra_link/backend/internal/handler/dto/common"
	"github.com/besart951/go_infra_link/backend/internal/handler/middleware"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func TestEnsureProjectAnyPermissionForProjectRespondsWithDenialDetails(t *testing.T) {
	gin.SetMode(gin.TestMode)

	userID := uuid.New()
	projectID := uuid.New()
	role := domainUser.RoleAdminPlaner
	message := "Das Projekt befindet sich in der Phase \"SIA 51\". In dieser Phase hat Ihre Rolle \"Planner Administrator\" keine Berechtigung für \"project.fielddevice.update\"."

	access := &permissionDenialAccessFake{
		details: &project.PermissionDenialDetails{
			Reason:             project.PermissionDenialReasonPhaseBlocked,
			Permission:         domainUser.PermissionProjectFieldDeviceUpdate,
			RequesterRole:      role,
			RequesterRoleLabel: domainUser.RoleDisplayName(role),
			PhaseName:          "SIA 51",
			Message:            message,
		},
	}

	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodPatch, "/projects/"+projectID.String(), nil)
	c.Set(middleware.ContextUserIDKey, userID)
	c.Set(middleware.ContextUserRoleKey, role)

	if EnsureProjectAnyPermissionForProject(c, access, projectID, domainUser.PermissionProjectFieldDeviceUpdate) {
		t.Fatal("expected permission check to deny")
	}
	if recorder.Code != http.StatusForbidden {
		t.Fatalf("expected status 403, got %d", recorder.Code)
	}

	var response dto.ErrorResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("expected JSON error response, got %v", err)
	}
	if response.Code != "authorization_failed" {
		t.Fatalf("expected authorization_failed code, got %q", response.Code)
	}
	if response.Message != message {
		t.Fatalf("expected specific denial message, got %q", response.Message)
	}

	details, ok := response.Details.(map[string]any)
	if !ok {
		t.Fatalf("expected structured denial details, got %#v", response.Details)
	}
	if details["reason"] != string(project.PermissionDenialReasonPhaseBlocked) {
		t.Fatalf("expected phase blocked reason, got %#v", details["reason"])
	}
	if details["phase_name"] != "SIA 51" {
		t.Fatalf("expected phase name in details, got %#v", details["phase_name"])
	}
}

type permissionDenialAccessFake struct {
	details *project.PermissionDenialDetails
}

func (f *permissionDenialAccessFake) CanAccessProject(context.Context, uuid.UUID, uuid.UUID, *domainUser.Role) (bool, error) {
	return true, nil
}

func (f *permissionDenialAccessFake) CanUseProjectPermission(context.Context, uuid.UUID, *domainUser.Role, string) (bool, error) {
	return false, nil
}

func (f *permissionDenialAccessFake) CanUseProjectPermissionForProject(context.Context, uuid.UUID, uuid.UUID, *domainUser.Role, string) (bool, error) {
	return false, nil
}

func (f *permissionDenialAccessFake) ExplainProjectPermissionDenial(context.Context, uuid.UUID, *domainUser.Role, []string) (*project.PermissionDenialDetails, error) {
	return f.details, nil
}

func (f *permissionDenialAccessFake) ExplainProjectScopedPermissionDenial(context.Context, uuid.UUID, uuid.UUID, *domainUser.Role, []string) (*project.PermissionDenialDetails, error) {
	return f.details, nil
}
