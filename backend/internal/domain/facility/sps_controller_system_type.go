package facility

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
	"github.com/google/uuid"
)

type SPSControllerSystemType struct {
	domain.Base
	Number          *int
	DocumentName    *string `gorm:"size:250"`
	SPSControllerID uuid.UUID
	SPSController   SPSController `gorm:"foreignKey:SPSControllerID;constraint:OnDelete:CASCADE"`
	SystemTypeID    uuid.UUID
	SystemType      SystemType `gorm:"foreignKey:SystemTypeID;constraint:OnDelete:RESTRICT"`

	FieldDevices []FieldDevice `gorm:"foreignKey:SPSControllerSystemTypeID"`
}
