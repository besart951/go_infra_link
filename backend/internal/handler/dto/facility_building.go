package dto

import (
	"time"

	"github.com/google/uuid"
)

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

type BuildingBulkRequest struct {
	Ids []uuid.UUID `json:"ids" binding:"required,min=1,dive,required"`
}

type BuildingBulkResponse struct {
	Items []BuildingResponse `json:"items"`
}
