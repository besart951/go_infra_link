package dto

import (
	"time"

	"github.com/google/uuid"
)

// Facility DTOs - Specification

type CreateSpecificationRequest struct {
	FieldDeviceID                             uuid.UUID `json:"field_device_id" binding:"required"`
	SpecificationSupplier                     *string   `json:"specification_supplier" binding:"omitempty,max=250"`
	SpecificationBrand                        *string   `json:"specification_brand" binding:"omitempty,max=250"`
	SpecificationType                         *string   `json:"specification_type" binding:"omitempty,max=250"`
	AdditionalInfoMotorValve                  *string   `json:"additional_info_motor_valve" binding:"omitempty,max=250"`
	AdditionalInfoSize                        *int      `json:"additional_info_size"`
	AdditionalInformationInstallationLocation *string   `json:"additional_information_installation_location" binding:"omitempty,max=250"`
	ElectricalConnectionPH                    *int      `json:"electrical_connection_ph"`
	ElectricalConnectionACDC                  *string   `json:"electrical_connection_acdc" binding:"omitempty,len=2"`
	ElectricalConnectionAmperage              *float64  `json:"electrical_connection_amperage"`
	ElectricalConnectionPower                 *float64  `json:"electrical_connection_power"`
	ElectricalConnectionRotation              *int      `json:"electrical_connection_rotation"`
}

type UpdateSpecificationRequest struct {
	SpecificationSupplier                     *string  `json:"specification_supplier" binding:"omitempty,max=250"`
	SpecificationBrand                        *string  `json:"specification_brand" binding:"omitempty,max=250"`
	SpecificationType                         *string  `json:"specification_type" binding:"omitempty,max=250"`
	AdditionalInfoMotorValve                  *string  `json:"additional_info_motor_valve" binding:"omitempty,max=250"`
	AdditionalInfoSize                        *int     `json:"additional_info_size"`
	AdditionalInformationInstallationLocation *string  `json:"additional_information_installation_location" binding:"omitempty,max=250"`
	ElectricalConnectionPH                    *int     `json:"electrical_connection_ph"`
	ElectricalConnectionACDC                  *string  `json:"electrical_connection_acdc" binding:"omitempty,len=2"`
	ElectricalConnectionAmperage              *float64 `json:"electrical_connection_amperage"`
	ElectricalConnectionPower                 *float64 `json:"electrical_connection_power"`
	ElectricalConnectionRotation              *int     `json:"electrical_connection_rotation"`
}

// FieldDevice-specific endpoints use the same payload shape as UpdateSpecificationRequest,
// but the FieldDevice ID comes from the URL path.
type CreateFieldDeviceSpecificationRequest = UpdateSpecificationRequest
type UpdateFieldDeviceSpecificationRequest = UpdateSpecificationRequest

type SpecificationResponse struct {
	ID                                        uuid.UUID  `json:"id"`
	FieldDeviceID                             *uuid.UUID `json:"field_device_id"`
	SpecificationSupplier                     *string    `json:"specification_supplier"`
	SpecificationBrand                        *string    `json:"specification_brand"`
	SpecificationType                         *string    `json:"specification_type"`
	AdditionalInfoMotorValve                  *string    `json:"additional_info_motor_valve"`
	AdditionalInfoSize                        *int       `json:"additional_info_size"`
	AdditionalInformationInstallationLocation *string    `json:"additional_information_installation_location"`
	ElectricalConnectionPH                    *int       `json:"electrical_connection_ph"`
	ElectricalConnectionACDC                  *string    `json:"electrical_connection_acdc"`
	ElectricalConnectionAmperage              *float64   `json:"electrical_connection_amperage"`
	ElectricalConnectionPower                 *float64   `json:"electrical_connection_power"`
	ElectricalConnectionRotation              *int       `json:"electrical_connection_rotation"`
	CreatedAt                                 time.Time  `json:"created_at"`
	UpdatedAt                                 time.Time  `json:"updated_at"`
}

type SpecificationListResponse struct {
	Items      []SpecificationResponse `json:"items"`
	Total      int64                   `json:"total"`
	Page       int                     `json:"page"`
	TotalPages int                     `json:"total_pages"`
}
