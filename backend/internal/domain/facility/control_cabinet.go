package facility

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
	"github.com/google/uuid"
)

type ControlCabinet struct {
	domain.Base
	BuildingID       uuid.UUID
	Building         Building `gorm:"foreignKey:BuildingID;constraint:OnDelete:CASCADE"`
	ProjectID        *uuid.UUID
	Project          *domain.Project `gorm:"foreignKey:ProjectID;constraint:OnDelete:SET NULL"`
	ControlCabinetNr *string         `gorm:"size:11;index"`

	SPSControllers []SPSController `gorm:"foreignKey:ControlCabinetID"`
}
