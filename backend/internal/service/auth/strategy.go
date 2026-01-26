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
	
	// Name returns the name of the authentication strategy
	Name() string
}
