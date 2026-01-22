package facility

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
)

type Building struct {
	domain.Base
	IWSCode       string `gorm:"index"`
	BuildingGroup int    `gorm:"not null"`

	ControlCabinets []ControlCabinet `gorm:"foreignKey:BuildingID"`
}
