package user

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainUser "github.com/besart951/go_infra_link/backend/internal/domain/user"
	"github.com/google/uuid"
)

type UserService interface {
	CreateWithPassword(user *domainUser.User, password string) error
	UpdateWithPassword(user *domainUser.User, password *string) error
	GetByID(id uuid.UUID) (*domainUser.User, error)
	List(page, limit int, search, orderBy, order string) (*domain.PaginatedList[domainUser.User], error)
	DeleteByID(id uuid.UUID) error
}

type AdminService interface {
	DisableUser(userID uuid.UUID) error
	EnableUser(userID uuid.UUID) error
	SetUserRole(userID uuid.UUID, role domainUser.Role) error
}

type RoleQueryService interface {
	GetGlobalRole(userID uuid.UUID) (domainUser.Role, error)
	GetAllowedRoles(requesterRole domainUser.Role) []domainUser.Role
}

type PermissionService interface {
	ListPermissions() ([]domainUser.Permission, error)
	GetPermissionByID(id uuid.UUID) (*domainUser.Permission, error)
	CreatePermission(permission *domainUser.Permission) error
	UpdatePermission(permission *domainUser.Permission) error
	DeletePermission(id uuid.UUID) error
}

type RolePermissionService interface {
	ListRolesWithPermissions() ([]domainUser.RoleInfo, error)
	UpdateRolePermissions(role domainUser.Role, permissions []string) ([]string, error)
	AddRolePermission(role domainUser.Role, permission string) (*domainUser.RolePermission, error)
	RemoveRolePermission(role domainUser.Role, permission string) error
}
