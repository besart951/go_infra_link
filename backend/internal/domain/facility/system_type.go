package facility

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
)

type SystemType struct {
	domain.Base
	NumberMin int    `gorm:"not null"`
	NumberMax int    `gorm:"not null"`
	Name      string `gorm:"not null"`
}
