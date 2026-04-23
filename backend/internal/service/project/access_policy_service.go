package project

import (
	"context"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainProject "github.com/besart951/go_infra_link/backend/internal/domain/project"
	domainUser "github.com/besart951/go_infra_link/backend/internal/domain/user"
	"github.com/google/uuid"
)

type ProjectAccessPolicyService struct {
	repo     domainProject.ProjectRepository
	userRepo domainUser.UserRepository
}

func (s *ProjectAccessPolicyService) CanAccessProject(ctx context.Context, requesterID, projectID uuid.UUID) (bool, error) {
	project, err := domain.GetByID(ctx, s.repo, projectID)
	if err != nil {
		return false, err
	}

	if project.CreatorID == requesterID {
		return true, nil
	}

	if s.userRepo != nil {
		users, err := s.userRepo.GetByIds(ctx, []uuid.UUID{requesterID})
		if err != nil {
			return false, err
		}
		if len(users) > 0 && domainUser.IsAdmin(users[0].Role) {
			return true, nil
		}
	}

	return s.repo.HasUser(ctx, projectID, requesterID)
}
