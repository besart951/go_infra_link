package dto

import (
	"time"

	"github.com/google/uuid"
)

type AdminPasswordResetResponse struct {
	ResetToken string    `json:"reset_token"`
	ExpiresAt  time.Time `json:"expires_at"`
}

type AdminLockUserRequest struct {
	Until time.Time `json:"until" binding:"required"`
}

type AdminSetUserRoleRequest struct {
	Role string `json:"role" binding:"required,oneof=user admin superadmin"`
}

type LoginAttemptResponse struct {
	ID            uuid.UUID  `json:"id"`
	CreatedAt     time.Time  `json:"created_at"`
	UserID        *uuid.UUID `json:"user_id,omitempty"`
	Email         *string    `json:"email,omitempty"`
	IP            *string    `json:"ip,omitempty"`
	UserAgent     *string    `json:"user_agent,omitempty"`
	Success       bool       `json:"success"`
	FailureReason *string    `json:"failure_reason,omitempty"`
}

type LoginAttemptListResponse struct {
	Items      []LoginAttemptResponse `json:"items"`
	Total      int64                  `json:"total"`
	Page       int                    `json:"page"`
	TotalPages int                    `json:"total_pages"`
}
