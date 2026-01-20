//go:build legacydto
// +build legacydto

package dto

import (
	"time"

	"github.com/besart951/go_infra_link/backend/internal/domain/project"
	"github.com/google/uuid"
)

// Project DTOs

type CreateProjectRequest struct {
	Name        string     `json:"name" binding:"required,min=1,max=255"`
	Description string     `json:"description"`
	Status      string     `json:"status" binding:"omitempty,oneof=planned ongoing completed"`
	StartDate   *time.Time `json:"start_date"`
	PhaseID     *uuid.UUID `json:"phase_id"`
	CreatorID   uuid.UUID  `json:"creator_id" binding:"required"`
}

type UpdateProjectRequest struct {
	Name        string                `json:"name" binding:"omitempty,min=1,max=255"`
	Description string                `json:"description"`
	Status      project.ProjectStatus `json:"status" binding:"omitempty,oneof=planned ongoing completed"`
	StartDate   *time.Time            `json:"start_date"`
	PhaseID     *uuid.UUID            `json:"phase_id"`
}

type ProjectResponse struct {
	ID          uuid.UUID             `json:"id"`
	Name        string                `json:"name"`
	Description string                `json:"description"`
	Status      project.ProjectStatus `json:"status"`
	StartDate   *time.Time            `json:"start_date"`
	PhaseID     uuid.UUID             `json:"phase_id"`
	CreatorID   uuid.UUID             `json:"creator_id"`
	CreatedAt   time.Time             `json:"created_at"`
	UpdatedAt   time.Time             `json:"updated_at"`
}

type ProjectListResponse struct {
	Items      []ProjectResponse `json:"items"`
	Total      int64             `json:"total"`
	Page       int               `json:"page"`
	TotalPages int               `json:"total_pages"`
}

// User DTOs

type CreateUserRequest struct {
	FirstName   string     `json:"first_name" binding:"required,min=1,max=100"`
	LastName    string     `json:"last_name" binding:"required,min=1,max=100"`
	Email       string     `json:"email" binding:"required,email"`
	Password    string     `json:"password" binding:"required,min=8"`
	IsActive    bool       `json:"is_active"`
	CreatedByID *uuid.UUID `json:"created_by_id"`
}

type UpdateUserRequest struct {
	FirstName string `json:"first_name" binding:"omitempty,min=1,max=100"`
	LastName  string `json:"last_name" binding:"omitempty,min=1,max=100"`
	Email     string `json:"email" binding:"omitempty,email"`
	Password  string `json:"password" binding:"omitempty,min=8"`
	IsActive  *bool  `json:"is_active"`
}

type UserResponse struct {
	ID        uuid.UUID `json:"id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Email     string    `json:"email"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type UserListResponse struct {
	Items      []UserResponse `json:"items"`
	Total      int64          `json:"total"`
	Page       int            `json:"page"`
	TotalPages int            `json:"total_pages"`
}

// Pagination Query

type PaginationQuery struct {
	Page   int    `form:"page" binding:"omitempty,min=1"`
	Limit  int    `form:"limit" binding:"omitempty,min=1,max=100"`
	Search string `form:"search"`
}

// Error Response

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

// Facility DTOs - Building

type CreateBuildingRequest struct {
	IWSCode       string `json:"iws_code" binding:"required,len=4"`
	BuildingGroup int    `json:"building_group" binding:"required"`
}

type UpdateBuildingRequest struct {
	IWSCode       string `json:"iws_code" binding:"omitempty,len=4"`
	BuildingGroup int    `json:"building_group" binding:"omitempty"`
}

type BuildingResponse struct {
	ID            uuid.UUID `json:"id"`
	IWSCode       string    `json:"iws_code"`
	BuildingGroup int       `json:"building_group"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type BuildingListResponse struct {
	Items      []BuildingResponse `json:"items"`
	Total      int64              `json:"total"`
	Page       int                `json:"page"`
	TotalPages int                `json:"total_pages"`
}

// Facility DTOs - ControlCabinet

type CreateControlCabinetRequest struct {
	BuildingID       uuid.UUID  `json:"building_id" binding:"required"`
	ProjectID        *uuid.UUID `json:"project_id"`
	ControlCabinetNr *string    `json:"control_cabinet_nr" binding:"omitempty,max=11"`
}

type UpdateControlCabinetRequest struct {
	BuildingID       uuid.UUID  `json:"building_id"`
	ProjectID        *uuid.UUID `json:"project_id"`
	ControlCabinetNr *string    `json:"control_cabinet_nr" binding:"omitempty,max=11"`
}

type ControlCabinetResponse struct {
	ID               uuid.UUID  `json:"id"`
	BuildingID       uuid.UUID  `json:"building_id"`
	ProjectID        *uuid.UUID `json:"project_id"`
	ControlCabinetNr *string    `json:"control_cabinet_nr"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

type ControlCabinetListResponse struct {
	Items      []ControlCabinetResponse `json:"items"`
	Total      int64                    `json:"total"`
	Page       int                      `json:"page"`
	TotalPages int                      `json:"total_pages"`
}

// Facility DTOs - SPSController

type CreateSPSControllerRequest struct {
	ControlCabinetID  uuid.UUID  `json:"control_cabinet_id" binding:"required"`
	ProjectID         *uuid.UUID `json:"project_id"`
	GADevice          *string    `json:"ga_device" binding:"omitempty,max=10"`
	DeviceName        string     `json:"device_name" binding:"required,max=100"`
	DeviceDescription *string    `json:"device_description" binding:"omitempty,max=250"`
	DeviceLocation    *string    `json:"device_location" binding:"omitempty,max=250"`
	IPAddress         *string    `json:"ip_address" binding:"omitempty,max=50"`
	Subnet            *string    `json:"subnet" binding:"omitempty,max=50"`
	Gateway           *string    `json:"gateway" binding:"omitempty,max=50"`
	Vlan              *string    `json:"vlan" binding:"omitempty,max=50"`
}

type UpdateSPSControllerRequest struct {
	ControlCabinetID  uuid.UUID  `json:"control_cabinet_id"`
	ProjectID         *uuid.UUID `json:"project_id"`
	GADevice          *string    `json:"ga_device" binding:"omitempty,max=10"`
	DeviceName        string     `json:"device_name" binding:"omitempty,max=100"`
	DeviceDescription *string    `json:"device_description" binding:"omitempty,max=250"`
	DeviceLocation    *string    `json:"device_location" binding:"omitempty,max=250"`
	IPAddress         *string    `json:"ip_address" binding:"omitempty,max=50"`
	Subnet            *string    `json:"subnet" binding:"omitempty,max=50"`
	Gateway           *string    `json:"gateway" binding:"omitempty,max=50"`
	Vlan              *string    `json:"vlan" binding:"omitempty,max=50"`
}

type SPSControllerResponse struct {
	ID                uuid.UUID  `json:"id"`
	ControlCabinetID  uuid.UUID  `json:"control_cabinet_id"`
	ProjectID         *uuid.UUID `json:"project_id"`
	GADevice          *string    `json:"ga_device"`
	DeviceName        string     `json:"device_name"`
	DeviceDescription *string    `json:"device_description"`
	DeviceLocation    *string    `json:"device_location"`
	IPAddress         *string    `json:"ip_address"`
	Subnet            *string    `json:"subnet"`
	Gateway           *string    `json:"gateway"`
	Vlan              *string    `json:"vlan"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

type SPSControllerListResponse struct {
	Items      []SPSControllerResponse `json:"items"`
	Total      int64                   `json:"total"`
	Page       int                     `json:"page"`
	TotalPages int                     `json:"total_pages"`
}

// Facility DTOs - FieldDevice

type CreateFieldDeviceRequest struct {
	BMK                       *string    `json:"bmk" binding:"omitempty,max=10"`
	Description               *string    `json:"description" binding:"omitempty,max=250"`
	ApparatNr                 *int       `json:"apparat_nr"`
	SPSControllerSystemTypeID uuid.UUID  `json:"sps_controller_system_type_id" binding:"required"`
	SystemPartID              *uuid.UUID `json:"system_part_id"`
	SpecificationID           *uuid.UUID `json:"specification_id"`
	ProjectID                 *uuid.UUID `json:"project_id"`
	ApparatID                 uuid.UUID  `json:"apparat_id" binding:"required"`
}

type UpdateFieldDeviceRequest struct {
	BMK                       *string    `json:"bmk" binding:"omitempty,max=10"`
	Description               *string    `json:"description" binding:"omitempty,max=250"`
	ApparatNr                 *int       `json:"apparat_nr"`
	SPSControllerSystemTypeID uuid.UUID  `json:"sps_controller_system_type_id"`
	SystemPartID              *uuid.UUID `json:"system_part_id"`
	SpecificationID           *uuid.UUID `json:"specification_id"`
	ProjectID                 *uuid.UUID `json:"project_id"`
	ApparatID                 uuid.UUID  `json:"apparat_id"`
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

// Facility DTOs - Apparat

type CreateApparatRequest struct {
	ShortName   string  `json:"short_name" binding:"required,max=255"`
	Name        string  `json:"name" binding:"required,max=250"`
	Description *string `json:"description" binding:"omitempty,max=250"`
}

type UpdateApparatRequest struct {
	ShortName   string  `json:"short_name" binding:"omitempty,max=255"`
	Name        string  `json:"name" binding:"omitempty,max=250"`
	Description *string `json:"description" binding:"omitempty,max=250"`
}

type ApparatResponse struct {
	ID          uuid.UUID `json:"id"`
	ShortName   string    `json:"short_name"`
	Name        string    `json:"name"`
	Description *string   `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type ApparatListResponse struct {
	Items      []ApparatResponse `json:"items"`
	Total      int64             `json:"total"`
	Page       int               `json:"page"`
	TotalPages int               `json:"total_pages"`
}

// Facility DTOs - SystemType

type CreateSystemTypeRequest struct {
	NumberMin int    `json:"number_min" binding:"required"`
	NumberMax int    `json:"number_max" binding:"required"`
	Name      string `json:"name" binding:"required,max=150"`
}

type UpdateSystemTypeRequest struct {
	NumberMin int    `json:"number_min"`
	NumberMax int    `json:"number_max"`
	Name      string `json:"name" binding:"omitempty,max=150"`
}

type SystemTypeResponse struct {
	ID        uuid.UUID `json:"id"`
	NumberMin int       `json:"number_min"`
	NumberMax int       `json:"number_max"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type SystemTypeListResponse struct {
	Items      []SystemTypeResponse `json:"items"`
	Total      int64                `json:"total"`
	Page       int                  `json:"page"`
	TotalPages int                  `json:"total_pages"`
}

// Facility DTOs - SystemPart

type CreateSystemPartRequest struct {
	ShortName   string  `json:"short_name" binding:"required,max=10"`
	Name        string  `json:"name" binding:"required,max=250"`
	Description *string `json:"description" binding:"omitempty,max=250"`
}

type UpdateSystemPartRequest struct {
	ShortName   string  `json:"short_name" binding:"omitempty,max=10"`
	Name        string  `json:"name" binding:"omitempty,max=250"`
	Description *string `json:"description" binding:"omitempty,max=250"`
}

type SystemPartResponse struct {
	ID          uuid.UUID `json:"id"`
	ShortName   string    `json:"short_name"`
	Name        string    `json:"name"`
	Description *string   `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type SystemPartListResponse struct {
	Items      []SystemPartResponse `json:"items"`
	Total      int64                `json:"total"`
	Page       int                  `json:"page"`
	TotalPages int                  `json:"total_pages"`
}

// Facility DTOs - Specification

type CreateSpecificationRequest struct {
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

type SpecificationResponse struct {
	ID                                        uuid.UUID `json:"id"`
	SpecificationSupplier                     *string   `json:"specification_supplier"`
	SpecificationBrand                        *string   `json:"specification_brand"`
	SpecificationType                         *string   `json:"specification_type"`
	AdditionalInfoMotorValve                  *string   `json:"additional_info_motor_valve"`
	AdditionalInfoSize                        *int      `json:"additional_info_size"`
	AdditionalInformationInstallationLocation *string   `json:"additional_information_installation_location"`
	ElectricalConnectionPH                    *int      `json:"electrical_connection_ph"`
	ElectricalConnectionACDC                  *string   `json:"electrical_connection_acdc"`
	ElectricalConnectionAmperage              *float64  `json:"electrical_connection_amperage"`
	ElectricalConnectionPower                 *float64  `json:"electrical_connection_power"`
	ElectricalConnectionRotation              *int      `json:"electrical_connection_rotation"`
	CreatedAt                                 time.Time `json:"created_at"`
	UpdatedAt                                 time.Time `json:"updated_at"`
}

type SpecificationListResponse struct {
	Items      []SpecificationResponse `json:"items"`
	Total      int64                   `json:"total"`
	Page       int                     `json:"page"`
	TotalPages int                     `json:"total_pages"`
}
