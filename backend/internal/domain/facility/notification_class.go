package facility

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
)

type NotificationClass struct {
	domain.Base
	EventCategory        string `gorm:"not null"`
	Nc                   int    `gorm:"not null"`
	ObjectDescription    string `gorm:"not null"`
	InternalDescription  string `gorm:"not null"`
	Meaning              string `gorm:"not null"`
	AckRequiredNotNormal bool   `gorm:"default:false"`
	AckRequiredError     bool   `gorm:"default:false"`
	AckRequiredNormal    bool   `gorm:"default:false"`
	NormNotNormal        int    `gorm:"default:0"`
	NormError            int    `gorm:"default:0"`
	NormNormal           int    `gorm:"default:0"`
}
