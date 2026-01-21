package auth

import "errors"

var (
	ErrInvalidCredentials = errors.New("invalid_credentials")
	ErrAccountDisabled    = errors.New("account_disabled")
	ErrAccountLocked      = errors.New("account_locked")
	ErrInvalidToken       = errors.New("invalid_token")
	ErrTokenExpired       = errors.New("token_expired")
	ErrTokenRevoked       = errors.New("token_revoked")

	ErrPasswordResetTokenInvalid = errors.New("password_reset_token_invalid")
	ErrPasswordResetTokenExpired = errors.New("password_reset_token_expired")
	ErrPasswordResetTokenUsed    = errors.New("password_reset_token_used")
)
