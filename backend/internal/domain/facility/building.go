package facility

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
)

type Building struct {
	domain.Base
	IWSCode       string `gorm:"index;uniqueIndex:idx_building_iws_group"`
	BuildingGroup int    `gorm:"not null;uniqueIndex:idx_building_iws_group"`

	ControlCabinets []ControlCabinet `gorm:"foreignKey:BuildingID"`
}
