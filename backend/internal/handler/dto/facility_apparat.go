package dto

import (
	"time"

	"github.com/google/uuid"
)

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
