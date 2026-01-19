package facility

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
)

type NotificationClass struct {
	domain.Base
	EventCategory        string `gorm:"size:50"`
	Nc                   int
	ObjectDescription    string `gorm:"size:250"`
	InternalDescription  string `gorm:"size:250"`
	Meaning              string `gorm:"size:250"`
	AckRequiredNotNormal bool
	AckRequiredError     bool
	AckRequiredNormal    bool
	NormNotNormal        int
	NormError            int
	NormNormal           int
}
