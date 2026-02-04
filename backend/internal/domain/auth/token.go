package auth

import (
	"time"

	"github.com/google/uuid"
)

// TokenValidator validates access tokens and extracts user identity.
type TokenValidator interface {
	ValidateToken(token string) (uuid.UUID, error)
}

// TokenCreator creates signed access tokens.
type TokenCreator interface {
	CreateAccessToken(userID uuid.UUID, expiresAt time.Time) (string, error)
}

// TokenService combines token validation and creation.
type TokenService interface {
	TokenValidator
	TokenCreator
}
