package user

import (
	"context"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainUser "github.com/besart951/go_infra_link/backend/internal/domain/user"
	userdirectory "github.com/besart951/go_infra_link/backend/internal/service/userdirectory"
	userregistration "github.com/besart951/go_infra_link/backend/internal/service/userregistration"
	"github.com/google/uuid"
)

type UserService interface {
	CreateWithPassword(ctx context.Context, user *domainUser.User, password string) error
	CreateWithPasswordForActor(ctx context.Context, actorID uuid.UUID, user *domainUser.User, password string) error
	UpdateProfileForActor(ctx context.Context, actorID uuid.UUID, user *domainUser.User) error
	UpdatePassword(ctx context.Context, userID uuid.UUID, currentPassword, newPassword string) (*domainUser.User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domainUser.User, error)
	List(ctx context.Context, page, limit int, search, orderBy, order string, includeDeleted bool) (*domain.PaginatedList[domainUser.User], error)
	DeleteByID(ctx context.Context, id uuid.UUID) error
	DeleteByIDForActor(ctx context.Context, actorID, userID uuid.UUID) error
	RestoreByIDForActor(ctx context.Context, actorID, userID uuid.UUID) error
}

type AdminService interface {
	DisableUser(ctx context.Context, actorID, userID uuid.UUID) error
	EnableUser(ctx context.Context, actorID, userID uuid.UUID) error
	SetUserRole(ctx context.Context, actorID, userID uuid.UUID, role domainUser.Role) error
}

type RoleQueryService interface {
	GetGlobalRole(ctx context.Context, userID uuid.UUID) (domainUser.Role, error)
	GetAllowedRoles(ctx context.Context, requesterRole domainUser.Role) ([]domainUser.Role, error)
}

type UserDirectoryService interface {
	List(ctx context.Context, requesterID uuid.UUID, page, limit int, search, teamID, role, orderBy, order string, includeDeleted bool) (*userdirectory.ListResult, error)
}

type UserRegistrationService interface {
	CreateInvitation(ctx context.Context, input userregistration.InviteInput) (*domainUser.User, *userregistration.Process, error)
	GetProcess(ctx context.Context, actorID, userID uuid.UUID) (*userregistration.Process, error)
	ListProcessesForUsers(ctx context.Context, users []domainUser.User) (map[uuid.UUID]*userregistration.Process, error)
	ResendInvitation(ctx context.Context, actorID, userID uuid.UUID) (*userregistration.Process, error)
}

type PermissionService interface {
	ListPermissions(ctx context.Context) ([]domainUser.Permission, error)
	GetPermissionByID(ctx context.Context, id uuid.UUID) (*domainUser.Permission, error)
	CreatePermission(ctx context.Context, permission *domainUser.Permission) error
	UpdatePermission(ctx context.Context, permission *domainUser.Permission) error
	DeletePermission(ctx context.Context, id uuid.UUID) error
}

type RolePermissionService interface {
	ListRolesWithPermissions(ctx context.Context) ([]domainUser.RoleInfo, error)
	GetAllowedRoles(ctx context.Context, role domainUser.Role) ([]domainUser.Role, error)
	UpdateRolePermissions(ctx context.Context, role domainUser.Role, permissions []string) ([]string, error)
	AddRolePermission(ctx context.Context, role domainUser.Role, permission string) (*domainUser.RolePermission, error)
	RemoveRolePermission(ctx context.Context, role domainUser.Role, permission string) error
}
