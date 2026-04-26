package project

import (
	"context"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainProject "github.com/besart951/go_infra_link/backend/internal/domain/project"
	domainUser "github.com/besart951/go_infra_link/backend/internal/domain/user"
	"github.com/google/uuid"
)

// ProjectWorkflowService centralizes project write workflows behind one seam.
type ProjectWorkflowService struct {
	lifecycle  workflowLifecycle
	membership workflowMembership
}

type workflowLifecycle interface {
	Create(ctx context.Context, project *domainProject.Project) error
}

type workflowMembership interface {
	InviteUser(ctx context.Context, projectID, userID uuid.UUID) error
	RemoveUser(ctx context.Context, projectID, userID uuid.UUID) error
	ListUsers(ctx context.Context, projectID uuid.UUID) ([]domainUser.User, error)
}

func newProjectWorkflowService(lifecycle workflowLifecycle, membership workflowMembership) *ProjectWorkflowService {
	return &ProjectWorkflowService{lifecycle: lifecycle, membership: membership}
}

func (s *ProjectWorkflowService) CreateProject(ctx context.Context, project *domainProject.Project) error {
	if s == nil || s.lifecycle == nil || project == nil {
		return domain.ErrInvalidArgument
	}
	return s.lifecycle.Create(ctx, project)
}

func (s *ProjectWorkflowService) InviteUser(ctx context.Context, projectID, userID uuid.UUID) error {
	if s == nil || s.membership == nil {
		return domain.ErrInvalidArgument
	}
	return s.membership.InviteUser(ctx, projectID, userID)
}

func (s *ProjectWorkflowService) RemoveUser(ctx context.Context, projectID, userID uuid.UUID) error {
	if s == nil || s.membership == nil {
		return domain.ErrInvalidArgument
	}
	return s.membership.RemoveUser(ctx, projectID, userID)
}

func (s *ProjectWorkflowService) ListUsers(ctx context.Context, projectID uuid.UUID) ([]domainUser.User, error) {
	if s == nil || s.membership == nil {
		return nil, domain.ErrInvalidArgument
	}
	return s.membership.ListUsers(ctx, projectID)
}
