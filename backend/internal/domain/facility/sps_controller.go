package facility

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
	"github.com/google/uuid"
)

type SPSController struct {
	domain.Base
	ControlCabinetID  uuid.UUID
	ControlCabinet    ControlCabinet `gorm:"foreignKey:ControlCabinetID;constraint:OnDelete:CASCADE"`
	ProjectID         *uuid.UUID
	Project           *domain.Project `gorm:"foreignKey:ProjectID;constraint:OnDelete:SET NULL"`
	GADevice          *string         `gorm:"size:10"`
	DeviceName        string          `gorm:"size:100"`
	DeviceDescription *string         `gorm:"size:250"`
	DeviceLocation    *string         `gorm:"size:250"`
	IPAddress         *string         `gorm:"size:50;index"`
	Subnet            *string         `gorm:"size:50"`
	Gateway           *string         `gorm:"size:50"`
	Vlan              *string         `gorm:"size:50;index"`

	SPSControllerSystemTypes []SPSControllerSystemType `gorm:"foreignKey:SPSControllerID"`
}
