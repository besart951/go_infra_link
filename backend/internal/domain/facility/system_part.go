package facility

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
)

type SystemPart struct {
	domain.Base
	ShortName   string  `gorm:"size:10;unique"`
	Name        string  `gorm:"size:250;unique"`
	Description *string `gorm:"size:250"`

	Apparats []*Apparat `gorm:"many2many:apparat_system_parts;"`
}
