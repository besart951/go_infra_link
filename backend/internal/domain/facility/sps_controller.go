package facility

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
	"github.com/google/uuid"
)

type SPSController struct {
	domain.Base
	ControlCabinetID  uuid.UUID      `gorm:"type:uuid;not null;index;uniqueIndex:idx_cabinet_devicename"`
	ControlCabinet    ControlCabinet `gorm:"foreignKey:ControlCabinetID"`
	GADevice          *string
	DeviceName        string `gorm:"not null;uniqueIndex:idx_cabinet_devicename"`
	DeviceDescription *string
	DeviceLocation    *string
	IPAddress         *string `gorm:"uniqueIndex:idx_vlan_ip"`
	Subnet            *string
	Gateway           *string
	Vlan              *string `gorm:"uniqueIndex:idx_vlan_ip"`

	SPSControllerSystemTypes []SPSControllerSystemType `gorm:"foreignKey:SPSControllerID"`
}
