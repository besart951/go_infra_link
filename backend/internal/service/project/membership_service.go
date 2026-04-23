package project

import (
	"context"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainProject "github.com/besart951/go_infra_link/backend/internal/domain/project"
	domainUser "github.com/besart951/go_infra_link/backend/internal/domain/user"
	"github.com/google/uuid"
)

type ProjectMembershipService struct {
	repo     domainProject.ProjectRepository
	userRepo domainUser.UserRepository
}

func (s *ProjectMembershipService) InviteUser(ctx context.Context, projectID, userID uuid.UUID) error {
	if _, err := domain.GetByID(ctx, s.repo, projectID); err != nil {
		return err
	}
	if _, err := domain.GetByID(ctx, s.userRepo, userID); err != nil {
		return err
	}
	return s.repo.AddUser(ctx, projectID, userID)
}

func (s *ProjectMembershipService) ListUsers(ctx context.Context, projectID uuid.UUID) ([]domainUser.User, error) {
	if _, err := domain.GetByID(ctx, s.repo, projectID); err != nil {
		return nil, err
	}
	return s.repo.ListUsers(ctx, projectID)
}

func (s *ProjectMembershipService) RemoveUser(ctx context.Context, projectID, userID uuid.UUID) error {
	if _, err := domain.GetByID(ctx, s.repo, projectID); err != nil {
		return err
	}
	if _, err := domain.GetByID(ctx, s.userRepo, userID); err != nil {
		return err
	}
	return s.repo.RemoveUser(ctx, projectID, userID)
}