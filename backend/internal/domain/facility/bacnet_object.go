package facility

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
	"github.com/google/uuid"
)

type BacnetObject struct {
	domain.Base
	TextFix        string `gorm:"index:idx_field_device_textfix;not null"`
	Description    *string
	GMSVisible     bool `gorm:"default:false"`
	Optional       bool `gorm:"default:false"`
	TextIndividual *string

	SoftwareType   BacnetSoftwareType `gorm:"type:varchar(50);not null"`
	SoftwareNumber uint16             `gorm:"not null"`

	HardwareType     BacnetHardwareType `gorm:"type:varchar(50)"`
	HardwareQuantity uint8

	FieldDeviceID       *uuid.UUID         `gorm:"type:uuid;index;index:idx_field_device_textfix"`
	FieldDevice         *FieldDevice       `gorm:"foreignKey:FieldDeviceID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	SoftwareReferenceID *uuid.UUID         `gorm:"type:uuid;index"`
	SoftwareReference   *BacnetObject      `gorm:"foreignKey:SoftwareReferenceID"`
	StateTextID         *uuid.UUID         `gorm:"type:uuid;index"`
	StateText           *StateText         `gorm:"foreignKey:StateTextID"`
	NotificationClassID *uuid.UUID         `gorm:"type:uuid;index"`
	NotificationClass   *NotificationClass `gorm:"foreignKey:NotificationClassID"`
	AlarmTypeID         *uuid.UUID         `gorm:"type:uuid;index"`
	AlarmType           *AlarmType         `gorm:"foreignKey:AlarmTypeID"`
	AlarmDefinitionID   *uuid.UUID         `gorm:"-:all"`
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
	AlarmTypeID         *uuid.UUID
	AlarmDefinitionID   *uuid.UUID
}

// ApplyPatch applies the supported partial state transition to an existing
// BACnet object. Persistence, parent existence, and cross-object uniqueness
// remain application/service concerns.
func (b *BacnetObject) ApplyPatch(patch BacnetObjectPatch) error {
	if b == nil || b.ID == uuid.Nil {
		return domain.ErrInvalidArgument
	}
	if patch.ID != uuid.Nil && patch.ID != b.ID {
		return domain.ErrInvalidArgument
	}

	if patch.TextFix != nil {
		b.TextFix = *patch.TextFix
	}
	if patch.Description != nil {
		b.Description = cloneBacnetPointer(patch.Description)
	}
	if patch.GMSVisible != nil {
		b.GMSVisible = *patch.GMSVisible
	}
	if patch.Optional != nil {
		b.Optional = *patch.Optional
	}
	if patch.TextIndividual != nil {
		b.TextIndividual = cloneBacnetPointer(patch.TextIndividual)
	}
	if patch.SoftwareType != nil {
		b.SoftwareType = *patch.SoftwareType
	}
	if patch.SoftwareNumber != nil {
		b.SoftwareNumber = *patch.SoftwareNumber
	}
	if patch.HardwareType != nil {
		b.HardwareType = *patch.HardwareType
	}
	if patch.HardwareQuantity != nil {
		b.HardwareQuantity = *patch.HardwareQuantity
	}
	if patch.SoftwareReferenceID != nil {
		b.SoftwareReferenceID = cloneBacnetPointer(patch.SoftwareReferenceID)
	}
	if patch.StateTextID != nil {
		b.StateTextID = cloneBacnetPointer(patch.StateTextID)
	}
	if patch.NotificationClassID != nil {
		b.NotificationClassID = cloneBacnetPointer(patch.NotificationClassID)
	}
	if patch.AlarmTypeID != nil {
		b.AlarmTypeID = cloneBacnetPointer(patch.AlarmTypeID)
	}
	if patch.AlarmDefinitionID != nil {
		b.AlarmDefinitionID = cloneBacnetPointer(patch.AlarmDefinitionID)
	}

	return nil
}

// AssignToFieldDevice changes the direct facility owner. Project-link
// reconciliation and uniqueness validation are deliberately outside the
// entity because they require repositories.
func (b *BacnetObject) AssignToFieldDevice(fieldDeviceID uuid.UUID) error {
	if b == nil || b.ID == uuid.Nil || fieldDeviceID == uuid.Nil {
		return domain.ErrInvalidArgument
	}
	b.FieldDeviceID = &fieldDeviceID
	return nil
}

// DetachFromFieldDevice prepares a BACnet object for indirect ownership, such
// as an ObjectData template association managed by the application service.
func (b *BacnetObject) DetachFromFieldDevice() error {
	if b == nil || b.ID == uuid.Nil {
		return domain.ErrInvalidArgument
	}
	b.FieldDeviceID = nil
	return nil
}

func cloneBacnetPointer[T any](value *T) *T {
	if value == nil {
		return nil
	}
	clone := *value
	return &clone
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
	BacnetHardwareTypeEMPTY BacnetHardwareType = ""
	BacnetHardwareTypeDO    BacnetHardwareType = "do"
	BacnetHardwareTypeAO    BacnetHardwareType = "ao"
	BacnetHardwareTypeDI    BacnetHardwareType = "di"
	BacnetHardwareTypeAI    BacnetHardwareType = "ai"
)
