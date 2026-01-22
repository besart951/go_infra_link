package facility

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
)

type SystemPart struct {
	domain.Base
	ShortName   string  `gorm:"not null"`
	Name        string  `gorm:"not null"`
	Description *string

	Apparats []*Apparat `gorm:"many2many:system_part_apparats;"`
}
