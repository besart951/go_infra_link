package facility

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
)

type Apparat struct {
	domain.Base
	ShortName   string  `gorm:"uniqueIndex:idx_apparat_composite"`
	Name        string  `gorm:"size:250;uniqueIndex:idx_apparat_composite"`
	Description *string `gorm:"size:250"`

	SystemParts  []*SystemPart `gorm:"many2many:apparat_system_parts;"`
	FieldDevices []FieldDevice `gorm:"foreignKey:ApparatID"`
}
