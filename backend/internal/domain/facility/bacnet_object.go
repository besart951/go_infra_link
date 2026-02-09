package facility

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
	"github.com/google/uuid"
)

type BacnetObject struct {
	domain.Base
	TextFix        string `gorm:"uniqueIndex:idx_field_device_textfix;not null"`
	Description    *string
	GMSVisible     bool `gorm:"default:false"`
	Optional       bool `gorm:"default:false"`
	TextIndividual *string

	SoftwareType   BacnetSoftwareType `gorm:"type:varchar(50);not null"`
	SoftwareNumber uint16             `gorm:"not null"`

	HardwareType     BacnetHardwareType `gorm:"type:varchar(50)"`
	HardwareQuantity uint8

	FieldDeviceID       *uuid.UUID         `gorm:"type:uuid;index;uniqueIndex:idx_field_device_textfix"`
	FieldDevice         *FieldDevice       `gorm:"foreignKey:FieldDeviceID"`
	SoftwareReferenceID *uuid.UUID         `gorm:"type:uuid;index"`
	SoftwareReference   *BacnetObject      `gorm:"foreignKey:SoftwareReferenceID"`
	StateTextID         *uuid.UUID         `gorm:"type:uuid;index"`
	StateText           *StateText         `gorm:"foreignKey:StateTextID"`
	NotificationClassID *uuid.UUID         `gorm:"type:uuid;index"`
	NotificationClass   *NotificationClass `gorm:"foreignKey:NotificationClassID"`
	AlarmDefinitionID   *uuid.UUID         `gorm:"type:uuid;index"`
	AlarmDefinition     *AlarmDefinition   `gorm:"foreignKey:AlarmDefinitionID"`
}

// BacnetObjectPatch represents a partial update for a bacnet object.
// Only non-nil fields are applied.
type BacnetObjectPatch struct {
	ID                  uuid.UUID
	TextFix             *string
	Description         *string
	GMSVisible          *bool
	Optional            *bool
	TextIndividual      *string
	SoftwareType        *BacnetSoftwareType
	SoftwareNumber      *uint16
	HardwareType        *BacnetHardwareType
	HardwareQuantity    *uint8
	SoftwareReferenceID *uuid.UUID
	StateTextID         *uuid.UUID
	NotificationClassID *uuid.UUID
	AlarmDefinitionID   *uuid.UUID
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
