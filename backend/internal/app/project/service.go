package project

import (
	"context"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	"github.com/google/uuid"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateProject(ctx context.Context, name string, ownerID string) (string, error) {
	project := &domain.Project{
		ID:      uuid.NewString(),
		Name:    name,
		OwnerID: ownerID,
	}

	if err := s.repo.Create(ctx, project); err != nil {
		return "", err
	}

	ownerMember := &domain.ProjectMember{
		ID:        uuid.NewString(),
		ProjectID: project.ID,
		UserID:    ownerID,
		Role:      "owner",
	}

	if err := s.repo.AddMember(ctx, ownerMember); err != nil {
		return "", err
	}

	return project.ID, nil
}

func (s *Service) GetProject(ctx context.Context, id string) (*domain.Project, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *Service) AddMember(ctx context.Context, projectID string, userID string, role string) error {
	member := &domain.ProjectMember{
		ID:        uuid.NewString(),
		ProjectID: projectID,
		UserID:    userID,
		Role:      role,
	}

	return s.repo.AddMember(ctx, member)
}
