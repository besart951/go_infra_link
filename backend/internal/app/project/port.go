package project

import (
	"context"

	"github.com/besart951/go_infra_link/backend/internal/domain"
)

type Repository interface {
	Create(ctx context.Context, p *domain.Project) error
	FindByID(ctx context.Context, id string) (*domain.Project, error)
	AddMember(ctx context.Context, m *domain.ProjectMember) error
}
