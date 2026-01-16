package object

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

func (s *Service) CreateObject(ctx context.Context, projectID string, name string) (string, error) {
	obj := &domain.Object{
		ID:        uuid.NewString(),
		ProjectID: projectID,
		Name:      name,
	}

	if err := s.repo.Create(ctx, obj); err != nil {
		return "", err
	}

	return obj.ID, nil
}

func (s *Service) GetObject(ctx context.Context, id string) (*domain.Object, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *Service) GrantAccess(ctx context.Context, objectID string, userID string, role string) error {
	perm := &domain.ObjectPermission{
		ID:       uuid.NewString(),
		ObjectID: objectID,
		UserID:   userID,
		Role:     role,
	}

	return s.repo.GrantAccess(ctx, perm)
}
