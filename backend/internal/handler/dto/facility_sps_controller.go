package dto

import (
	"time"

	"github.com/google/uuid"
)

// Facility DTOs - SPSController

type CreateSPSControllerRequest struct {
	ControlCabinetID  uuid.UUID                      `json:"control_cabinet_id" binding:"required"`
	GADevice          *string                        `json:"ga_device" binding:"required,len=3"`
	DeviceName        string                         `json:"device_name" binding:"required,max=100"`
	DeviceDescription *string                        `json:"device_description" binding:"omitempty,max=250"`
	DeviceLocation    *string                        `json:"device_location" binding:"omitempty,max=250"`
	IPAddress         *string                        `json:"ip_address" binding:"omitempty,max=50"`
	Subnet            *string                        `json:"subnet" binding:"omitempty,max=50"`
	Gateway           *string                        `json:"gateway" binding:"omitempty,max=50"`
	Vlan              *string                        `json:"vlan" binding:"omitempty,max=50"`
	SystemTypes       []SPSControllerSystemTypeInput `json:"system_types" binding:"omitempty,dive"`
}

type UpdateSPSControllerRequest struct {
	ControlCabinetID  uuid.UUID                       `json:"control_cabinet_id"`
	GADevice          *string                         `json:"ga_device" binding:"omitempty,len=3"`
	DeviceName        string                          `json:"device_name" binding:"omitempty,max=100"`
	DeviceDescription *string                         `json:"device_description" binding:"omitempty,max=250"`
	DeviceLocation    *string                         `json:"device_location" binding:"omitempty,max=250"`
	IPAddress         *string                         `json:"ip_address" binding:"omitempty,max=50"`
	Subnet            *string                         `json:"subnet" binding:"omitempty,max=50"`
	Gateway           *string                         `json:"gateway" binding:"omitempty,max=50"`
	Vlan              *string                         `json:"vlan" binding:"omitempty,max=50"`
	SystemTypes       *[]SPSControllerSystemTypeInput `json:"system_types" binding:"omitempty,dive"`
}

type SPSControllerSystemTypeInput struct {
	SystemTypeID uuid.UUID `json:"system_type_id" binding:"required"`
	Number       *int      `json:"number"`
	DocumentName *string   `json:"document_name" binding:"omitempty,max=250"`
}

type SPSControllerResponse struct {
	ID                uuid.UUID `json:"id"`
	ControlCabinetID  uuid.UUID `json:"control_cabinet_id"`
	GADevice          *string   `json:"ga_device"`
	DeviceName        string    `json:"device_name"`
	DeviceDescription *string   `json:"device_description"`
	DeviceLocation    *string   `json:"device_location"`
	IPAddress         *string   `json:"ip_address"`
	Subnet            *string   `json:"subnet"`
	Gateway           *string   `json:"gateway"`
	Vlan              *string   `json:"vlan"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

type SPSControllerListResponse struct {
	Items      []SPSControllerResponse `json:"items"`
	Total      int64                   `json:"total"`
	Page       int                     `json:"page"`
	TotalPages int                     `json:"total_pages"`
}

type NextAvailableGADeviceResponse struct {
	GADevice string `json:"ga_device"`
}
