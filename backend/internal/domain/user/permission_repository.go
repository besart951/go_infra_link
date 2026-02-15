package user

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
	"github.com/google/uuid"
)

// PermissionRepository manages permission types.
type PermissionRepository interface {
	domain.Repository[Permission]
	GetByName(name string) (*Permission, error)
	ListAll() ([]Permission, error)
	ListByNames(names []string) ([]Permission, error)
}

// RolePermissionRepository manages role-to-permission assignments.
type RolePermissionRepository interface {
	Create(entity *RolePermission) error
	GetByIds(ids []uuid.UUID) ([]*RolePermission, error)
	Update(entity *RolePermission) error
	DeleteByIds(ids []uuid.UUID) error
	GetPaginatedList(params domain.PaginationParams) (*domain.PaginatedList[RolePermission], error)
	ListByRole(role Role) ([]RolePermission, error)
	ListByRoles(roles []Role) ([]RolePermission, error)
	ReplaceRolePermissions(role Role, permissions []string) error
	AddPermissionToRole(role Role, permission string) (*RolePermission, error)
	RemovePermissionFromRole(role Role, permission string) error
	DeleteByPermissionName(permission string) error
}
