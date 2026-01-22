package facility

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
	"github.com/besart951/go_infra_link/backend/internal/domain/project"
	"github.com/google/uuid"
)

type SPSController struct {
	domain.Base
	ControlCabinetID  uuid.UUID        `gorm:"type:uuid;not null;index"`
	ControlCabinet    ControlCabinet   `gorm:"foreignKey:ControlCabinetID"`
	ProjectID         *uuid.UUID       `gorm:"type:uuid;index"`
	Project           *project.Project `gorm:"foreignKey:ProjectID"`
	GADevice          *string
	DeviceName        string `gorm:"not null"`
	DeviceDescription *string
	DeviceLocation    *string
	IPAddress         *string
	Subnet            *string
	Gateway           *string
	Vlan              *string

	SPSControllerSystemTypes []SPSControllerSystemType `gorm:"foreignKey:SPSControllerID"`
}
