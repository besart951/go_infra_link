package facility

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
	"github.com/besart951/go_infra_link/backend/internal/domain/project"
	"github.com/google/uuid"
)

type SPSController struct {
	domain.Base
	ControlCabinetID  uuid.UUID
	ControlCabinet    ControlCabinet
	ProjectID         *uuid.UUID
	Project           *project.Project
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
