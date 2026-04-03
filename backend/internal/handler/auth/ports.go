package auth

import (
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
	Login(email, password string, userAgent, ip *string) (*domainAuth.LoginResult, error)
	Refresh(refreshToken string, userAgent, ip *string) (*domainAuth.LoginResult, error)
	Logout(refreshToken string) error
}

type UserService interface {
	GetByID(id uuid.UUID) (*domainUser.User, error)
}

type PermissionQueryService interface {
	GetRolePermissions(role domainUser.Role) ([]string, error)
}
