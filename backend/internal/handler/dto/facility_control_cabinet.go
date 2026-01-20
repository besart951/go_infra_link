package dto

import (
	"time"

	"github.com/google/uuid"
)

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
