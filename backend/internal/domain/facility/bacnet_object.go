package facility

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
	"github.com/google/uuid"
)

type BacnetObject struct {
	domain.Base
	TextFix        string  `gorm:"size:250;uniqueIndex:idx_Bacnet_fd_text"`
	Description    *string `gorm:"type:text"`
	GMSVisible     bool    `gorm:"default:false"`
	Optional       bool    `gorm:"default:false"`
	TextIndividual *string `gorm:"size:250"`

	SoftwareType   BacnetSoftwareType `gorm:"type:varchar(2);not null"`
	SoftwareNumber uint16

	HardwareType     BacnetHardwareType `gorm:"type:varchar(2);not null"`
	HardwareQuantity uint8              `gorm:"default:1"`

	FieldDeviceID       *uuid.UUID   `gorm:"uniqueIndex:idx_Bacnet_fd_text"`
	FieldDevice         *FieldDevice `gorm:"foreignKey:FieldDeviceID;constraint:OnDelete:CASCADE"`
	SoftwareReferenceID *uuid.UUID
	SoftwareReference   *BacnetObject `gorm:"foreignKey:SoftwareReferenceID;constraint:OnDelete:SET NULL"`
	StateTextID         *uuid.UUID
	StateText           *StateText `gorm:"foreignKey:StateTextID;constraint:OnDelete:SET NULL"`
	NotificationClassID *uuid.UUID
	NotificationClass   *NotificationClass `gorm:"foreignKey:NotificationClassID;constraint:OnDelete:SET NULL"`
	AlarmDefinitionID   *uuid.UUID
	AlarmDefinition     *AlarmDefinition `gorm:"foreignKey:AlarmDefinitionID;constraint:OnDelete:SET NULL"`
}

type BacnetSoftwareType string

const (
	BacnetSoftwareTypeAI BacnetSoftwareType = "ai"
	BacnetSoftwareTypeAO BacnetSoftwareType = "ao"
	BacnetSoftwareTypeAV BacnetSoftwareType = "av"
	BacnetSoftwareTypeBI BacnetSoftwareType = "bi"
	BacnetSoftwareTypeBO BacnetSoftwareType = "bo"
	BacnetSoftwareTypeBV BacnetSoftwareType = "bv"
	BacnetSoftwareTypeMI BacnetSoftwareType = "mi"
	BacnetSoftwareTypeMO BacnetSoftwareType = "mo"
	BacnetSoftwareTypeMV BacnetSoftwareType = "mv"
	BacnetSoftwareTypeCA BacnetSoftwareType = "ca"
	BacnetSoftwareTypeEE BacnetSoftwareType = "ee"
	BacnetSoftwareTypeLP BacnetSoftwareType = "lp"
	BacnetSoftwareTypeNC BacnetSoftwareType = "nc"
	BacnetSoftwareTypeSC BacnetSoftwareType = "sc"
	BacnetSoftwareTypeTL BacnetSoftwareType = "tl"
)

type BacnetHardwareType string

const (
	BacnetHardwareTypeDO BacnetHardwareType = "do"
	BacnetHardwareTypeAO BacnetHardwareType = "ao"
	BacnetHardwareTypeDI BacnetHardwareType = "di"
	BacnetHardwareTypeAI BacnetHardwareType = "ai"
)
