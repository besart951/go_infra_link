package facility

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
	"github.com/google/uuid"
)

type ControlCabinet struct {
	domain.Base
	BuildingID       uuid.UUID `gorm:"type:uuid;not null;index;uniqueIndex:idx_cabinet_building_nr"`
	Building         Building  `gorm:"foreignKey:BuildingID"`
	ControlCabinetNr *string   `gorm:"uniqueIndex:idx_cabinet_building_nr"`

	SPSControllers []SPSController `gorm:"foreignKey:ControlCabinetID"`
}
