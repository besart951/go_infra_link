package notification

import (
	"time"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	"github.com/google/uuid"
)

type SystemNotification struct {
	domain.Base
	RecipientID  uuid.UUID         `gorm:"type:uuid;not null;index"`
	ActorID      *uuid.UUID        `gorm:"type:uuid"`
	EventKey     string            `gorm:"type:varchar(128);not null;index"`
	Title        string            `gorm:"not null"`
	Body         string            `gorm:"type:text"`
	ResourceType string            `gorm:"type:varchar(64)"`
	ResourceID   *uuid.UUID        `gorm:"type:uuid;index"`
	Metadata     map[string]string `gorm:"serializer:json;type:text"`
	ReadAt       *time.Time        `gorm:"index"`
}

func (n *SystemNotification) GetBase() *domain.Base {
	return &n.Base
}

func (n *SystemNotification) MarkRead(now time.Time) {
	if n.ReadAt == nil {
		n.ReadAt = &now
	}
	n.TouchForUpdate(now)
}
