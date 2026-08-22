package exporting

import (
	"context"

	domainexport "github.com/besart951/go_infra_link/backend/internal/domain/exporting"
	domainuser "github.com/besart951/go_infra_link/backend/internal/domain/user"
	"github.com/google/uuid"
)

type globalExportPermissionChecker interface {
	HasPermission(ctx context.Context, role domainuser.Role, permission string) (bool, error)
}

type projectExportPermissionChecker interface {
	CanUseProjectPermissionForProject(ctx context.Context, requesterID, projectID uuid.UUID, requesterRole *domainuser.Role, permission string) (bool, error)
}

type DownloadPolicy struct {
	projects projectExportPermissionChecker
	global   globalExportPermissionChecker
}

func NewDownloadPolicy(projects projectExportPermissionChecker, global globalExportPermissionChecker) *DownloadPolicy {
	return &DownloadPolicy{projects: projects, global: global}
}

func (p *DownloadPolicy) CanDownload(ctx context.Context, authorization domainexport.DownloadAuthorization) (bool, error) {
	if authorization.Scope.Kind != domainexport.AccessScopeProject {
		return p.canDownloadGlobal(ctx, authorization.RequesterRole)
	}
	if p.projects == nil || len(authorization.Scope.ProjectIDs) == 0 {
		return false, nil
	}
	for _, projectID := range authorization.Scope.ProjectIDs {
		allowed, err := p.projects.CanUseProjectPermissionForProject(
			ctx,
			authorization.RequesterID,
			projectID,
			&authorization.RequesterRole,
			domainuser.PermissionProjectFieldDeviceRead,
		)
		if err != nil || !allowed {
			return allowed, err
		}
	}
	return true, nil
}

func (p *DownloadPolicy) canDownloadGlobal(ctx context.Context, role domainuser.Role) (bool, error) {
	if p.global == nil {
		return false, nil
	}
	return p.global.HasPermission(ctx, role, domainuser.PermissionFieldDeviceRead)
}
