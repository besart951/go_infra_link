package auth

import (
	"time"

	domainAuth "github.com/besart951/go_infra_link/backend/internal/domain/auth"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// jwtAuthStrategy implements AuthStrategy using JWT tokens
// This is the concrete strategy for JWT-based authentication
type jwtAuthStrategy struct {
	secret []byte
	issuer string
}

// NewJWTAuthStrategy creates a new JWT authentication strategy
func NewJWTAuthStrategy(secret, issuer string) domainAuth.AuthStrategy {
	return &jwtAuthStrategy{
		secret: []byte(secret),
		issuer: issuer,
	}
}

// CreateToken creates a JWT access token for the given user
func (s *jwtAuthStrategy) CreateToken(userID uuid.UUID, expiresAt time.Time) (string, error) {
	claims := jwt.RegisteredClaims{
		Subject:   userID.String(),
		Issuer:    s.issuer,
		IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		ExpiresAt: jwt.NewNumericDate(expiresAt.UTC()),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secret)
}

// ValidateToken validates and parses a JWT token, returning the user ID
func (s *jwtAuthStrategy) ValidateToken(tokenString string) (uuid.UUID, error) {
	claims, err := s.parseAndValidateToken(tokenString)
	if err != nil {
		return uuid.Nil, err
	}

	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		// Wrap error with context about invalid UUID in token subject
		return uuid.Nil, jwt.ErrInvalidType
	}
	return userID, nil
}

// ParseToken validates and returns the full JWT claims
// This satisfies the ParseToken method added to AuthStrategy for better encapsulation
func (s *jwtAuthStrategy) ParseToken(tokenString string) (interface{}, error) {
	return s.parseAndValidateToken(tokenString)
}

// parseAndValidateToken is a helper that parses and validates a JWT token
// This eliminates duplication between ValidateToken and ParseToken methods
func (s *jwtAuthStrategy) parseAndValidateToken(tokenString string) (*jwt.RegisteredClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(t *jwt.Token) (interface{}, error) {
		return s.secret, nil
	})
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*jwt.RegisteredClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, jwt.ErrTokenInvalidClaims
}

// Name returns the name of this authentication strategy
func (s *jwtAuthStrategy) Name() string {
	return "JWT"
}
