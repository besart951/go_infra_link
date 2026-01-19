package facility

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
)

type Building struct {
	domain.Base
	IWSCode       string `gorm:"size:4;index"`
	BuildingGroup int

	ControlCabinets []ControlCabinet `gorm:"foreignKey:BuildingID"`
}
