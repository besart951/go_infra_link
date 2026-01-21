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
	SystemPartID              *uuid.UUID          `json:"system_part_id"`
	ProjectID                 *uuid.UUID          `json:"project_id"`
	ApparatID                 uuid.UUID           `json:"apparat_id" binding:"required"`
	ObjectDataID              *uuid.UUID          `json:"object_data_id"`
	BacnetObjects             []BacnetObjectInput `json:"bacnet_objects" binding:"omitempty,dive"`
}

type UpdateFieldDeviceRequest struct {
	BMK                       *string              `json:"bmk" binding:"omitempty,max=10"`
	Description               *string              `json:"description" binding:"omitempty,max=250"`
	ApparatNr                 *int                 `json:"apparat_nr" binding:"omitempty,min=1,max=99"`
	SPSControllerSystemTypeID uuid.UUID            `json:"sps_controller_system_type_id"`
	SystemPartID              *uuid.UUID           `json:"system_part_id"`
	ProjectID                 *uuid.UUID           `json:"project_id"`
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
	ProjectID                 *uuid.UUID `json:"project_id"`
	ApparatID                 uuid.UUID  `json:"apparat_id"`
	CreatedAt                 time.Time  `json:"created_at"`
	UpdatedAt                 time.Time  `json:"updated_at"`
}

type FieldDeviceListResponse struct {
	Items      []FieldDeviceResponse `json:"items"`
	Total      int64                 `json:"total"`
	Page       int                   `json:"page"`
	TotalPages int                   `json:"total_pages"`
}
