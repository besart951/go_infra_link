package auth

import (
	"time"

	"github.com/google/uuid"
)

type LoginAttempt struct {
	ID            uuid.UUID
	CreatedAt     time.Time
	UserID        *uuid.UUID
	Email         *string
	IP            *string
	UserAgent     *string
	Success       bool
	FailureReason *string
}
