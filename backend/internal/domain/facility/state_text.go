package facility

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
)

type StateText struct {
	domain.Base
	RefNumber   int
	StateText1  *string `gorm:"size:100"`
	StateText2  *string `gorm:"size:100"`
	StateText3  *string `gorm:"size:100"`
	StateText4  *string `gorm:"size:100"`
	StateText5  *string `gorm:"size:100"`
	StateText6  *string `gorm:"size:100"`
	StateText7  *string `gorm:"size:100"`
	StateText8  *string `gorm:"size:100"`
	StateText9  *string `gorm:"size:100"`
	StateText10 *string `gorm:"size:100"`
	StateText11 *string `gorm:"size:100"`
	StateText12 *string `gorm:"size:100"`
	StateText13 *string `gorm:"size:100"`
	StateText14 *string `gorm:"size:100"`
	StateText15 *string `gorm:"size:100"`
	StateText16 *string `gorm:"size:100"`
}
