package auth

import (
	"time"

	"github.com/google/uuid"
)

// AuthStrategy defines the interface for different authentication strategies
// Following the Strategy Pattern and Open/Closed Principle
type AuthStrategy interface {
	// CreateToken creates an authentication token for the given user
	CreateToken(userID uuid.UUID, expiresAt time.Time) (string, error)

	// ValidateToken validates and parses an authentication token
	// Returns the user ID if valid, or an error if invalid/expired
	ValidateToken(token string) (uuid.UUID, error)

	// ParseToken validates and returns the full claims for backward compatibility
	// This allows strategies to return their native claim types
	ParseToken(token string) (interface{}, error)

	// Name returns the name of the authentication strategy
	Name() string
}
