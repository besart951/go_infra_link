package facility

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
	"github.com/google/uuid"
)

type Apparat struct {
	domain.Base
	ShortName   string `gorm:"uniqueIndex;idx_name_deleted;not null"`
	Name        string `gorm:"uniqueIndex;idx_name_deleted;not null"`
	Description *string

	SystemParts  []*SystemPart `gorm:"many2many:system_part_apparats;"`
	FieldDevices []FieldDevice `gorm:"foreignKey:ApparatID"`
}

type ApparatFilterParams struct {
	ObjectDataID *uuid.UUID
	SystemPartID *uuid.UUID
}
