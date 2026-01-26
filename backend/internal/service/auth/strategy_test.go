package auth_test

import (
	"testing"
	"time"

	"github.com/besart951/go_infra_link/backend/internal/service/auth"
	"github.com/google/uuid"
)

// TestJWTAuthStrategy verifies the JWT authentication strategy works correctly
func TestJWTAuthStrategy(t *testing.T) {
	secret := "test-secret-key-for-testing"
	issuer := "test-issuer"
	
	strategy := auth.NewJWTAuthStrategy(secret, issuer)
	
	if strategy.Name() != "JWT" {
		t.Errorf("Expected strategy name to be 'JWT', got '%s'", strategy.Name())
	}
	
	userID := uuid.New()
	expiresAt := time.Now().Add(1 * time.Hour)
	
	// Test token creation
	token, err := strategy.CreateToken(userID, expiresAt)
	if err != nil {
		t.Fatalf("Failed to create token: %v", err)
	}
	
	if token == "" {
		t.Fatal("Expected non-empty token")
	}
	
	// Test token validation
	validatedUserID, err := strategy.ValidateToken(token)
	if err != nil {
		t.Fatalf("Failed to validate token: %v", err)
	}
	
	if validatedUserID != userID {
		t.Errorf("Expected userID %s, got %s", userID, validatedUserID)
	}
}

// TestJWTAuthStrategy_InvalidToken verifies that invalid tokens are rejected
func TestJWTAuthStrategy_InvalidToken(t *testing.T) {
	secret := "test-secret-key"
	issuer := "test-issuer"
	
	strategy := auth.NewJWTAuthStrategy(secret, issuer)
	
	invalidToken := "invalid.token.here"
	
	_, err := strategy.ValidateToken(invalidToken)
	if err == nil {
		t.Error("Expected error for invalid token, got nil")
	}
}

// TestJWTAuthStrategy_ExpiredToken verifies that expired tokens are rejected
func TestJWTAuthStrategy_ExpiredToken(t *testing.T) {
	secret := "test-secret-key"
	issuer := "test-issuer"
	
	strategy := auth.NewJWTAuthStrategy(secret, issuer)
	
	userID := uuid.New()
	expiresAt := time.Now().Add(-1 * time.Hour) // Already expired
	
	token, err := strategy.CreateToken(userID, expiresAt)
	if err != nil {
		t.Fatalf("Failed to create token: %v", err)
	}
	
	_, err = strategy.ValidateToken(token)
	if err == nil {
		t.Error("Expected error for expired token, got nil")
	}
}

// TestJWTService_UsesStrategy verifies that JWTService uses the strategy pattern
func TestJWTService_UsesStrategy(t *testing.T) {
	secret := "test-secret-key"
	issuer := "test-issuer"
	
	jwtService := auth.NewJWTService(secret, issuer)
	
	strategy := jwtService.GetStrategy()
	if strategy == nil {
		t.Fatal("Expected strategy to be set")
	}
	
	if strategy.Name() != "JWT" {
		t.Errorf("Expected strategy name to be 'JWT', got '%s'", strategy.Name())
	}
	
	// Verify token operations work through the service
	userID := uuid.New()
	expiresAt := time.Now().Add(1 * time.Hour)
	
	token, err := jwtService.CreateAccessToken(userID, expiresAt)
	if err != nil {
		t.Fatalf("Failed to create access token: %v", err)
	}
	
	claims, err := jwtService.ParseAccessToken(token)
	if err != nil {
		t.Fatalf("Failed to parse access token: %v", err)
	}
	
	if claims.Subject != userID.String() {
		t.Errorf("Expected subject %s, got %s", userID.String(), claims.Subject)
	}
}
