package object

import (
	"context"

	"github.com/besart951/go_infra_link/backend/internal/domain"
)

type Repository interface {
	Create(ctx context.Context, o *domain.Object) error
	FindByID(ctx context.Context, id string) (*domain.Object, error)
	GrantAccess(ctx context.Context, p *domain.ObjectPermission) error
}
