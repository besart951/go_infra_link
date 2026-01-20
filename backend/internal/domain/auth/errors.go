package auth

import "errors"

var (
	ErrInvalidCredentials = errors.New("invalid_credentials")
	ErrInvalidToken       = errors.New("invalid_token")
	ErrTokenExpired       = errors.New("token_expired")
	ErrTokenRevoked       = errors.New("token_revoked")
)
