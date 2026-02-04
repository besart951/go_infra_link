package dto

import (
	"time"

	"github.com/google/uuid"
)

// Facility DTOs - FieldDevice

type CreateFieldDeviceRequest struct {
	BMK                       *string             `json:"bmk" binding:"omitempty,max=10"`
	Description               *string             `json:"description" binding:"omitempty,max=250"`
	ApparatNr                 *int                `json:"apparat_nr" binding:"required,min=1,max=99"`
	SPSControllerSystemTypeID uuid.UUID           `json:"sps_controller_system_type_id" binding:"required"`
	SystemPartID              uuid.UUID           `json:"system_part_id" binding:"required"`
	ApparatID                 uuid.UUID           `json:"apparat_id" binding:"required"`
	ObjectDataID              *uuid.UUID          `json:"object_data_id"`
	BacnetObjects             []BacnetObjectInput `json:"bacnet_objects" binding:"omitempty,dive"`
}

type UpdateFieldDeviceRequest struct {
	BMK                       *string              `json:"bmk" binding:"omitempty,max=10"`
	Description               *string              `json:"description" binding:"omitempty,max=250"`
	ApparatNr                 *int                 `json:"apparat_nr" binding:"omitempty,min=1,max=99"`
	SPSControllerSystemTypeID uuid.UUID            `json:"sps_controller_system_type_id"`
	SystemPartID              uuid.UUID            `json:"system_part_id" binding:"required"`
	ApparatID                 uuid.UUID            `json:"apparat_id"`
	ObjectDataID              *uuid.UUID           `json:"object_data_id"`
	BacnetObjects             *[]BacnetObjectInput `json:"bacnet_objects" binding:"omitempty,dive"`
}

type FieldDeviceResponse struct {
	ID                        uuid.UUID  `json:"id"`
	BMK                       *string    `json:"bmk"`
	Description               *string    `json:"description"`
	ApparatNr                 *int       `json:"apparat_nr"`
	SPSControllerSystemTypeID uuid.UUID  `json:"sps_controller_system_type_id"`
	SystemPartID              *uuid.UUID `json:"system_part_id"`
	SpecificationID           *uuid.UUID `json:"specification_id"`
	ApparatID                 uuid.UUID  `json:"apparat_id"`
	CreatedAt                 time.Time  `json:"created_at"`
	UpdatedAt                 time.Time  `json:"updated_at"`

	// Embedded related entities for display
	SPSControllerSystemType *SPSControllerSystemTypeResponse `json:"sps_controller_system_type,omitempty"`
	Apparat                 *ApparatResponse                 `json:"apparat,omitempty"`
	SystemPart              *SystemPartResponse              `json:"system_part,omitempty"`
	Specification           *SpecificationResponse           `json:"specification,omitempty"`
	BacnetObjects           []BacnetObjectResponse           `json:"bacnet_objects,omitempty"`
}

type FieldDeviceListResponse struct {
	Items      []FieldDeviceResponse `json:"items"`
	Total      int64                 `json:"total"`
	Page       int                   `json:"page"`
	TotalPages int                   `json:"total_pages"`
}

type AvailableApparatNumbersResponse struct {
	Available []int `json:"available"`
}

// FieldDeviceOptionsResponse contains all metadata needed for creating/editing field devices
// This implements the "Single-Fetch Metadata Strategy" to avoid multiple API calls
type FieldDeviceOptionsResponse struct {
	Apparats            []ApparatResponse    `json:"apparats"`
	SystemParts         []SystemPartResponse `json:"system_parts"`
	ObjectDatas         []ObjectDataResponse `json:"object_datas"`
	ApparatToSystemPart map[string][]string  `json:"apparat_to_system_part"` // apparat_id -> [system_part_ids]
	ObjectDataToApparat map[string][]string  `json:"object_data_to_apparat"` // object_data_id -> [apparat_ids]
}

// MultiCreateFieldDeviceRequest represents a request to create multiple field devices
type MultiCreateFieldDeviceRequest struct {
	FieldDevices []CreateFieldDeviceRequest `json:"field_devices" binding:"required,min=1,dive"`
}

// FieldDeviceCreateResultResponse represents the result of creating a single field device
type FieldDeviceCreateResultResponse struct {
	Index       int                  `json:"index"`        // Index in the original request array
	Success     bool                 `json:"success"`      // Whether the creation succeeded
	FieldDevice *FieldDeviceResponse `json:"field_device"` // The created field device (null if failed)
	Error       string               `json:"error"`        // Error message if failed (empty if succeeded)
	ErrorField  string               `json:"error_field"`  // Specific field that caused the error (if applicable)
}

// MultiCreateFieldDeviceResponse represents the response from a multi-create operation
type MultiCreateFieldDeviceResponse struct {
	Results       []FieldDeviceCreateResultResponse `json:"results"`
	TotalRequests int                               `json:"total_requests"`
	SuccessCount  int                               `json:"success_count"`
	FailureCount  int                               `json:"failure_count"`
}

// BulkUpdateFieldDeviceItem represents a single field device update in a bulk operation
type BulkUpdateFieldDeviceItem struct {
	ID          uuid.UUID `json:"id" binding:"required"`
	BMK         *string   `json:"bmk" binding:"omitempty,max=10"`
	Description *string   `json:"description" binding:"omitempty,max=250"`
	ApparatNr   *int      `json:"apparat_nr" binding:"omitempty,min=1,max=99"`
	ApparatID   *uuid.UUID `json:"apparat_id"`
	SystemPartID *uuid.UUID `json:"system_part_id"`
}

// BulkUpdateFieldDeviceRequest represents a request to update multiple field devices
type BulkUpdateFieldDeviceRequest struct {
	Updates []BulkUpdateFieldDeviceItem `json:"updates" binding:"required,min=1,dive"`
}

// BulkOperationResultItem represents the result of a single item in a bulk operation
type BulkOperationResultItem struct {
	ID      uuid.UUID `json:"id"`
	Success bool      `json:"success"`
	Error   string    `json:"error,omitempty"`
}

// BulkUpdateFieldDeviceResponse represents the response from a bulk update operation
type BulkUpdateFieldDeviceResponse struct {
	Results      []BulkOperationResultItem `json:"results"`
	TotalCount   int                       `json:"total_count"`
	SuccessCount int                       `json:"success_count"`
	FailureCount int                       `json:"failure_count"`
}

// BulkDeleteFieldDeviceRequest represents a request to delete multiple field devices
type BulkDeleteFieldDeviceRequest struct {
	IDs []uuid.UUID `json:"ids" binding:"required,min=1"`
}

// BulkDeleteFieldDeviceResponse represents the response from a bulk delete operation
type BulkDeleteFieldDeviceResponse struct {
	Results      []BulkOperationResultItem `json:"results"`
	TotalCount   int                       `json:"total_count"`
	SuccessCount int                       `json:"success_count"`
	FailureCount int                       `json:"failure_count"`
}

