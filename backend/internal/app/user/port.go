package user

import (
	"context"

	"github.com/besart951/go_infra_link/backend/internal/domain"
)

type Repository interface {
	FindByID(ctx context.Context, id string) (*domain.User, error)
	Create(ctx context.Context, u *domain.User) error
}
