package dto

import "github.com/google/uuid"

// Facility DTOs - Validation

type ValidateBuildingRequest struct {
	ID            *uuid.UUID `json:"id"`
	IWSCode       string     `json:"iws_code"`
	BuildingGroup int        `json:"building_group"`
}

type ValidateControlCabinetRequest struct {
	ID               *uuid.UUID `json:"id"`
	BuildingID       uuid.UUID  `json:"building_id"`
	ControlCabinetNr *string    `json:"control_cabinet_nr"`
}

type ValidateSPSControllerRequest struct {
	ID               *uuid.UUID `json:"id"`
	ControlCabinetID uuid.UUID  `json:"control_cabinet_id"`
	GADevice         *string    `json:"ga_device"`
	DeviceName       string     `json:"device_name"`
	IPAddress        *string    `json:"ip_address"`
	Subnet           *string    `json:"subnet"`
	Gateway          *string    `json:"gateway"`
	Vlan             *string    `json:"vlan"`
}
