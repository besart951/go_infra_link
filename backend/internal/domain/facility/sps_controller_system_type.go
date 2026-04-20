package facility

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
	"github.com/google/uuid"
)

type SPSControllerSystemType struct {
	domain.Base
	Number            *int
	DocumentName      *string
	SPSControllerID   uuid.UUID     `gorm:"type:uuid;not null;index"`
	SPSController     SPSController `gorm:"foreignKey:SPSControllerID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	SystemTypeID      uuid.UUID     `gorm:"type:uuid;not null;index"`
	SystemType        SystemType    `gorm:"foreignKey:SystemTypeID"`
	FieldDevicesCount int           `gorm:"->;column:field_devices_count;-:migration"`

	FieldDevices []FieldDevice `gorm:"foreignKey:SPSControllerSystemTypeID"`
}
