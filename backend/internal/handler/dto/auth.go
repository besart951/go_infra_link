package dto

import (
	"time"

	"github.com/google/uuid"
)

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

type AuthUserResponse struct {
	ID        uuid.UUID `json:"id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Email     string    `json:"email"`
	IsActive  bool      `json:"is_active"`
}

type AuthResponse struct {
	User                  AuthUserResponse `json:"user"`
	AccessTokenExpiresAt  time.Time        `json:"access_token_expires_at"`
	RefreshTokenExpiresAt time.Time        `json:"refresh_token_expires_at"`
	CsrfToken             string           `json:"csrf_token"`
}

type PasswordResetConfirmRequest struct {
	Token       string `json:"token" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}
