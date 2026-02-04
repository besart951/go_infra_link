package auth

import (
	"time"

	domainAuth "github.com/besart951/go_infra_link/backend/internal/domain/auth"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type jwtService struct {
	strategy domainAuth.AuthStrategy
	issuer   string
}

// NewJWTService creates a new JWT service using the strategy pattern.
// Returns domainAuth.TokenService to satisfy the domain port.
func NewJWTService(secret, issuer string) domainAuth.TokenService {
	return &jwtService{
		strategy: NewJWTAuthStrategy(secret, issuer),
		issuer:   issuer,
	}
}

// CreateAccessToken creates a JWT access token using the strategy.
func (s *jwtService) CreateAccessToken(userID uuid.UUID, expiresAt time.Time) (string, error) {
	return s.strategy.CreateToken(userID, expiresAt)
}

// ValidateToken validates a token and returns the user ID.
func (s *jwtService) ValidateToken(token string) (uuid.UUID, error) {
	return s.strategy.ValidateToken(token)
}

// ParseAccessToken parses and validates a JWT token.
func (s *jwtService) ParseAccessToken(tokenString string) (*jwt.RegisteredClaims, error) {
	claims, err := s.strategy.ParseToken(tokenString)
	if err != nil {
		return nil, err
	}

	if jwtClaims, ok := claims.(*jwt.RegisteredClaims); ok {
		return jwtClaims, nil
	}

	return nil, jwt.ErrTokenInvalidClaims
}

// GetStrategy returns the underlying authentication strategy.
func (s *jwtService) GetStrategy() domainAuth.AuthStrategy {
	return s.strategy
}
