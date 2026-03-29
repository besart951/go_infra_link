package dto

import (
	"time"

	"github.com/google/uuid"
)

// Phase DTOs

type CreatePhaseRequest struct {
	Name string `json:"name" binding:"required,min=1,max=255"`
}

type UpdatePhaseRequest struct {
	Name string `json:"name" binding:"omitempty,min=1,max=255"`
}

type PhaseResponse struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type PhaseListResponse struct {
	Items      []PhaseResponse `json:"items"`
	Total      int64           `json:"total"`
	Page       int             `json:"page"`
	TotalPages int             `json:"total_pages"`
}
