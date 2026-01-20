package dto

import (
	"time"

	"github.com/google/uuid"
)

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
