package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// JWTService interface for JWT token operations
// Now uses AuthStrategy pattern internally for better extensibility
type JWTService interface {
	CreateAccessToken(userID uuid.UUID, expiresAt time.Time) (string, error)
	ParseAccessToken(tokenString string) (*jwt.RegisteredClaims, error)
	GetStrategy() AuthStrategy
}

type jwtService struct {
	strategy AuthStrategy
	issuer   string
}

// NewJWTService creates a new JWT service using the strategy pattern
func NewJWTService(secret, issuer string) JWTService {
	return &jwtService{
		strategy: NewJWTAuthStrategy(secret, issuer),
		issuer:   issuer,
	}
}

// CreateAccessToken creates a JWT access token using the strategy
func (s *jwtService) CreateAccessToken(userID uuid.UUID, expiresAt time.Time) (string, error) {
	return s.strategy.CreateToken(userID, expiresAt)
}

// ParseAccessToken parses and validates a JWT token
// Now uses the strategy's ParseToken method for proper encapsulation
func (s *jwtService) ParseAccessToken(tokenString string) (*jwt.RegisteredClaims, error) {
	// Use the strategy's ParseToken method which handles validation and parsing
	claims, err := s.strategy.ParseToken(tokenString)
	if err != nil {
		return nil, err
	}

	// Type assert to jwt.RegisteredClaims (safe since we know it's JWT strategy)
	if jwtClaims, ok := claims.(*jwt.RegisteredClaims); ok {
		return jwtClaims, nil
	}

	return nil, jwt.ErrTokenInvalidClaims
}

// GetStrategy returns the underlying authentication strategy
// This allows for future extensibility and testing
func (s *jwtService) GetStrategy() AuthStrategy {
	return s.strategy
}
