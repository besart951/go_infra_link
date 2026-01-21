package facility

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
	"github.com/google/uuid"
)

type BacnetObject struct {
	domain.Base
	TextFix        string
	Description    *string
	GMSVisible     bool
	Optional       bool
	TextIndividual *string

	SoftwareType   BacnetSoftwareType
	SoftwareNumber uint16

	HardwareType     BacnetHardwareType
	HardwareQuantity uint8

	FieldDeviceID       *uuid.UUID
	FieldDevice         *FieldDevice
	SoftwareReferenceID *uuid.UUID
	SoftwareReference   *BacnetObject
	StateTextID         *uuid.UUID
	StateText           *StateText
	NotificationClassID *uuid.UUID
	NotificationClass   *NotificationClass
	AlarmDefinitionID   *uuid.UUID
	AlarmDefinition     *AlarmDefinition
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
