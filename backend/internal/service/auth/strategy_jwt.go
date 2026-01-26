package auth

import (
	"time"

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
func NewJWTAuthStrategy(secret, issuer string) AuthStrategy {
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
	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(t *jwt.Token) (interface{}, error) {
		return s.secret, nil
	})
	if err != nil {
		return uuid.Nil, err
	}

	if claims, ok := token.Claims.(*jwt.RegisteredClaims); ok && token.Valid {
		userID, err := uuid.Parse(claims.Subject)
		if err != nil {
			return uuid.Nil, jwt.ErrTokenInvalidClaims
		}
		return userID, nil
	}

	return uuid.Nil, jwt.ErrTokenInvalidClaims
}

// Name returns the name of this authentication strategy
func (s *jwtAuthStrategy) Name() string {
	return "JWT"
}
