package auth

import (
	"context"
	"net/http"

	domainAuth "github.com/besart951/go_infra_link/backend/internal/domain/auth"
	domainUser "github.com/besart951/go_infra_link/backend/internal/domain/user"
	"github.com/google/uuid"
)

type CookieSettings struct {
	Domain   string
	Secure   bool
	SameSite http.SameSite
}

type AuthService interface {
	Login(ctx context.Context, email, password string, userAgent, ip *string) (*domainAuth.LoginResult, error)
	Refresh(ctx context.Context, refreshToken string, userAgent, ip *string) (*domainAuth.LoginResult, error)
	Logout(ctx context.Context, refreshToken string) error
}

type UserService interface {
	GetByID(ctx context.Context, id uuid.UUID) (*domainUser.User, error)
}

type PermissionQueryService interface {
	GetRolePermissions(ctx context.Context, role domainUser.Role) ([]string, error)
	CanAccessUserDirectory(role domainUser.Role) bool
}
