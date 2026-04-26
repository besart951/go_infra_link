package project

import (
	"context"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainProject "github.com/besart951/go_infra_link/backend/internal/domain/project"
	domainUser "github.com/besart951/go_infra_link/backend/internal/domain/user"
	"github.com/google/uuid"
)

type workflowFromServices struct {
	lifecycle  ProjectLifecycleService
	membership ProjectMembershipService
}

func newWorkflowFromServices(lifecycle ProjectLifecycleService, membership ProjectMembershipService) ProjectWorkflowService {
	return &workflowFromServices{lifecycle: lifecycle, membership: membership}
}

func (w *workflowFromServices) CreateProject(ctx context.Context, project *domainProject.Project) error {
	if w == nil || w.lifecycle == nil || project == nil {
		return domain.ErrInvalidArgument
	}
	return w.lifecycle.Create(ctx, project)
}

func (w *workflowFromServices) InviteUser(ctx context.Context, projectID, userID uuid.UUID) error {
	if w == nil || w.membership == nil {
		return domain.ErrInvalidArgument
	}
	return w.membership.InviteUser(ctx, projectID, userID)
}

func (w *workflowFromServices) ListUsers(ctx context.Context, projectID uuid.UUID) ([]domainUser.User, error) {
	if w == nil || w.membership == nil {
		return nil, domain.ErrInvalidArgument
	}
	return w.membership.ListUsers(ctx, projectID)
}

func (w *workflowFromServices) RemoveUser(ctx context.Context, projectID, userID uuid.UUID) error {
	if w == nil || w.membership == nil {
		return domain.ErrInvalidArgument
	}
	return w.membership.RemoveUser(ctx, projectID, userID)
}
