package auth

import (
	"time"

	domainUser "github.com/besart951/go_infra_link/backend/internal/domain/user"
)

// LoginResult represents the outcome of a successful authentication.
type LoginResult struct {
	User               *domainUser.User
	AccessToken        string
	AccessTokenExpiry  time.Time
	RefreshToken       string
	RefreshTokenExpiry time.Time
	CSRFFriendlyToken  string
}
