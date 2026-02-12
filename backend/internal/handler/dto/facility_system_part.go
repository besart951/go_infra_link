package dto

import (
	"time"

	"github.com/google/uuid"
)

// Facility DTOs - SystemPart

type CreateSystemPartRequest struct {
	ShortName   string  `json:"short_name" binding:"required,min=3,max=3"`
	Name        string  `json:"name" binding:"required,max=250"`
	Description *string `json:"description" binding:"omitempty,max=250"`
}

type UpdateSystemPartRequest struct {
	ShortName   string  `json:"short_name" binding:"omitempty,min=3,max=3"`
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
