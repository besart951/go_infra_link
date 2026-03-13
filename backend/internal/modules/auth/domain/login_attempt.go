package auth

import (
	"time"

	"github.com/google/uuid"
)

type LoginAttempt struct {
	ID            uuid.UUID  `gorm:"type:uuid;primaryKey"`
	CreatedAt     time.Time  `gorm:"autoCreateTime;index"`
	UserID        *uuid.UUID `gorm:"type:uuid;index"`
	Email         *string    `gorm:"index"`
	IP            *string
	UserAgent     *string
	Success       bool `gorm:"index"`
	FailureReason *string
}
