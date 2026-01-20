package facility

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
	"github.com/besart951/go_infra_link/backend/internal/domain/project"
	"github.com/google/uuid"
)

type FieldDevice struct {
	domain.Base
	BMK                       *string `gorm:"size:10;index"`
	Description               *string `gorm:"size:250"`
	ApparatNr                 *int    `gorm:"index"`
	SPSControllerSystemTypeID uuid.UUID
	SPSControllerSystemType   SPSControllerSystemType `gorm:"foreignKey:SPSControllerSystemTypeID;constraint:OnDelete:RESTRICT"`
	SystemPartID              *uuid.UUID
	SystemPart                *SystemPart `gorm:"foreignKey:SystemPartID;constraint:OnDelete:CASCADE"`
	SpecificationID           *uuid.UUID
	Specification             *Specification `gorm:"foreignKey:SpecificationID;constraint:OnDelete:CASCADE"`
	ProjectID                 *uuid.UUID
	Project                   *project.Project `gorm:"foreignKey:ProjectID;constraint:OnDelete:SET NULL"`
	ApparatID                 uuid.UUID
	Apparat                   Apparat `gorm:"foreignKey:ApparatID;constraint:OnDelete:RESTRICT"`

	BacnetObjects []BacnetObject `gorm:"foreignKey:FieldDeviceID"`
}
