package facility

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
	"github.com/google/uuid"
)

type FieldDevice struct {
	domain.Base
	BMK                       *string
	Description               *string
	ApparatNr                 *int
	SPSControllerSystemTypeID uuid.UUID               `gorm:"type:uuid;not null;index"`
	SPSControllerSystemType   SPSControllerSystemType `gorm:"foreignKey:SPSControllerSystemTypeID"`
	SystemPartID              *uuid.UUID              `gorm:"type:uuid;index"`
	SystemPart                *SystemPart             `gorm:"foreignKey:SystemPartID"`
	SpecificationID           *uuid.UUID              `gorm:"type:uuid;index"`
	Specification             *Specification          `gorm:"foreignKey:SpecificationID"`
	ApparatID                 uuid.UUID               `gorm:"type:uuid;not null;index"`
	Apparat                   Apparat                 `gorm:"foreignKey:ApparatID"`

	BacnetObjects []BacnetObject `gorm:"foreignKey:FieldDeviceID"`
}
