package exporting

import (
	"context"
	"encoding/json"
	"testing"

	domainexport "github.com/besart951/go_infra_link/backend/internal/domain/exporting"
	domainuser "github.com/besart951/go_infra_link/backend/internal/domain/user"
	"github.com/google/uuid"
)

type globalPermissionStub struct {
	allowed    bool
	permission string
}

func (s *globalPermissionStub) HasPermission(_ context.Context, _ domainuser.Role, permission string) (bool, error) {
	s.permission = permission
	return s.allowed, nil
}

type projectPermissionStub struct {
	allowed    map[uuid.UUID]bool
	checked    []uuid.UUID
	userID     uuid.UUID
	permission string
}

func (s *projectPermissionStub) CanUseProjectPermissionForProject(_ context.Context, requesterID, projectID uuid.UUID, _ *domainuser.Role, permission string) (bool, error) {
	s.checked = append(s.checked, projectID)
	s.userID = requesterID
	s.permission = permission
	return s.allowed[projectID], nil
}

func TestDownloadPolicyRequiresGlobalFieldDeviceReadForGlobalExport(t *testing.T) {
	global := &globalPermissionStub{allowed: true}
	policy := NewDownloadPolicy(nil, global)

	allowed, err := policy.CanDownload(t.Context(), domainexport.DownloadAuthorization{
		RequesterID: uuid.New(), Scope: domainexport.Scope{Kind: domainexport.AccessScopeGlobal},
	})
	if err != nil || !allowed {
		t.Fatalf("CanDownload() = %v, %v", allowed, err)
	}
	if global.permission != domainuser.PermissionFieldDeviceRead {
		t.Fatalf("permission = %q", global.permission)
	}
}

func TestDownloadPolicyRechecksEveryProjectInScope(t *testing.T) {
	requesterID := uuid.New()
	firstProjectID := uuid.New()
	secondProjectID := uuid.New()
	projects := &projectPermissionStub{allowed: map[uuid.UUID]bool{firstProjectID: true, secondProjectID: false}}
	policy := NewDownloadPolicy(projects, nil)

	allowed, err := policy.CanDownload(t.Context(), domainexport.DownloadAuthorization{
		RequesterID: requesterID,
		Scope: domainexport.Scope{
			Kind:       domainexport.AccessScopeProject,
			ProjectIDs: []uuid.UUID{firstProjectID, secondProjectID},
		},
	})
	if err != nil || allowed {
		t.Fatalf("CanDownload() = %v, %v", allowed, err)
	}
	if len(projects.checked) != 2 || projects.userID != requesterID {
		t.Fatalf("unexpected project checks %#v", projects.checked)
	}
	if projects.permission != domainuser.PermissionProjectFieldDeviceRead {
		t.Fatalf("permission = %q", projects.permission)
	}
}

func TestDownloadPolicyRejectsEmptyProjectScope(t *testing.T) {
	policy := NewDownloadPolicy(&projectPermissionStub{}, nil)
	allowed, err := policy.CanDownload(t.Context(), domainexport.DownloadAuthorization{
		RequesterID: uuid.New(), Scope: domainexport.Scope{Kind: domainexport.AccessScopeProject},
	})
	if err != nil || allowed {
		t.Fatalf("CanDownload() = %v, %v", allowed, err)
	}
}

func TestExportScopeFromPayloadDefaultsLegacyJobsToGlobal(t *testing.T) {
	projectID := uuid.New()
	payload, err := json.Marshal(domainexport.Request{ProjectIDs: []uuid.UUID{projectID}})
	if err != nil {
		t.Fatal(err)
	}

	scope, err := exportScopeFromPayload(payload)
	if err != nil {
		t.Fatal(err)
	}
	if scope.Kind != domainexport.AccessScopeGlobal || len(scope.ProjectIDs) != 1 || scope.ProjectIDs[0] != projectID {
		t.Fatalf("unexpected scope %#v", scope)
	}
}
