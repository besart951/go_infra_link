package facility

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
	"github.com/google/uuid"
)

type SPSController struct {
	domain.Base
	ControlCabinetID  uuid.UUID
	ControlCabinet    ControlCabinet
	GADevice          *string
	DeviceName        string
	DeviceDescription *string
	DeviceLocation    *string
	IPAddress         *string
	Subnet            *string
	Gateway           *string
	Vlan              *string

	SPSControllerSystemTypes []SPSControllerSystemType
}
