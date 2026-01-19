package userservice

import (
	"context"

	"github.com/besart951/go_infra_link/backend/internal/core/domain/user"
	"github.com/google/uuid"
)

type UserService struct {
	repo user.UserRepository
}

func NewUserService(repo user.UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) Create(ctx context.Context, u *user.User) error {
	// Here you would normally hash the password before saving
	return s.repo.Create(ctx, u)
}

func (s *UserService) GetByID(ctx context.Context, id uuid.UUID) (*user.User, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *UserService) GetByUsername(ctx context.Context, username string) (*user.User, error) {
	return s.repo.GetByUsername(ctx, username)
}

func (s *UserService) GetAll(ctx context.Context) ([]user.User, error) {
	return s.repo.GetAll(ctx)
}

func (s *UserService) Update(ctx context.Context, u *user.User) error {
	return s.repo.Update(ctx, u)
}

func (s *UserService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}
