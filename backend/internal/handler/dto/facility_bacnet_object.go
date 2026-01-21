package dto

import (
	"time"

	"github.com/google/uuid"
)

// Facility DTOs - BacnetObject

type CreateBacnetObjectRequest struct {
	FieldDeviceID *uuid.UUID `json:"field_device_id"`
	ObjectDataID  *uuid.UUID `json:"object_data_id"`

	BacnetObjectInput
}

// UpdateBacnetObjectRequest is a full update (PUT-style) payload.
// If object_data_id is provided, the object will also be attached to that object data template.
type UpdateBacnetObjectRequest struct {
	FieldDeviceID *uuid.UUID `json:"field_device_id"`
	ObjectDataID  *uuid.UUID `json:"object_data_id"`

	BacnetObjectInput
}

// BacnetObjectInput is used for nested create/update under FieldDevice.
// Numbers are ints for easier JSON validation and are converted to domain types in the handler/service.
type BacnetObjectInput struct {
	TextFix        string  `json:"text_fix" binding:"required,max=250"`
	Description    *string `json:"description" binding:"omitempty"`
	GMSVisible     bool    `json:"gms_visible"`
	Optional       bool    `json:"optional"`
	TextIndividual *string `json:"text_individual" binding:"omitempty,max=250"`

	SoftwareType   string `json:"software_type" binding:"required"`
	SoftwareNumber int    `json:"software_number" binding:"required,min=0,max=65535"`

	HardwareType     string `json:"hardware_type" binding:"required"`
	HardwareQuantity int    `json:"hardware_quantity" binding:"required,min=1,max=255"`

	SoftwareReferenceID *uuid.UUID `json:"software_reference_id"`
	StateTextID         *uuid.UUID `json:"state_text_id"`
	NotificationClassID *uuid.UUID `json:"notification_class_id"`
	AlarmDefinitionID   *uuid.UUID `json:"alarm_definition_id"`
}

type BacnetObjectResponse struct {
	ID string `json:"id"`

	TextFix        string  `json:"text_fix"`
	Description    *string `json:"description"`
	GMSVisible     bool    `json:"gms_visible"`
	Optional       bool    `json:"optional"`
	TextIndividual *string `json:"text_individual"`

	SoftwareType   string `json:"software_type"`
	SoftwareNumber int    `json:"software_number"`

	HardwareType     string `json:"hardware_type"`
	HardwareQuantity int    `json:"hardware_quantity"`

	FieldDeviceID *uuid.UUID `json:"field_device_id"`

	SoftwareReferenceID *uuid.UUID `json:"software_reference_id"`
	StateTextID         *uuid.UUID `json:"state_text_id"`
	NotificationClassID *uuid.UUID `json:"notification_class_id"`
	AlarmDefinitionID   *uuid.UUID `json:"alarm_definition_id"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
