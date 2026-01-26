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
func (s *jwtService) ParseAccessToken(tokenString string) (*jwt.RegisteredClaims, error) {
	// Use the strategy to validate the token
	_, err := s.strategy.ValidateToken(tokenString)
	if err != nil {
		return nil, err
	}

	// Parse again to get the full claims for backward compatibility
	// This maintains the existing interface while using the strategy internally
	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(t *jwt.Token) (interface{}, error) {
		// Access the secret through the strategy (we know it's JWT strategy)
		if jwtStrat, ok := s.strategy.(*jwtAuthStrategy); ok {
			return jwtStrat.secret, nil
		}
		return nil, jwt.ErrTokenInvalidClaims
	})
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*jwt.RegisteredClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, jwt.ErrTokenInvalidClaims
}

// GetStrategy returns the underlying authentication strategy
// This allows for future extensibility and testing
func (s *jwtService) GetStrategy() AuthStrategy {
	return s.strategy
}
