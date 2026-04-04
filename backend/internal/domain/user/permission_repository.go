package user

import (
	"context"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	"github.com/google/uuid"
)

// PermissionRepository manages permission types.
type PermissionRepository interface {
	domain.Repository[Permission]
	GetByName(ctx context.Context, name string) (*Permission, error)
	ListAll(ctx context.Context) ([]Permission, error)
	ListByNames(ctx context.Context, names []string) ([]Permission, error)
}

// RolePermissionRepository manages role-to-permission assignments.
type RolePermissionRepository interface {
	Create(ctx context.Context, entity *RolePermission) error
	GetByIds(ctx context.Context, ids []uuid.UUID) ([]*RolePermission, error)
	Update(ctx context.Context, entity *RolePermission) error
	DeleteByIds(ctx context.Context, ids []uuid.UUID) error
	GetPaginatedList(ctx context.Context, params domain.PaginationParams) (*domain.PaginatedList[RolePermission], error)
	ListByRole(ctx context.Context, role Role) ([]RolePermission, error)
	ListByRoles(ctx context.Context, roles []Role) ([]RolePermission, error)
	ReplaceRolePermissions(ctx context.Context, role Role, permissions []string) error
	AddPermissionToRole(ctx context.Context, role Role, permission string) (*RolePermission, error)
	RemovePermissionFromRole(ctx context.Context, role Role, permission string) error
	DeleteByPermissionName(ctx context.Context, permission string) error
}
