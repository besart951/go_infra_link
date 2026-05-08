package user

import "context"

type UserRoleRepository interface {
	ListByRoles(ctx context.Context, roles []Role) ([]User, error)
}
