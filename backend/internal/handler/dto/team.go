package dto

import (
	"time"

	"github.com/google/uuid"
)

type CreateTeamRequest struct {
	Name        string  `json:"name" binding:"required,min=1,max=150"`
	Description *string `json:"description"`
}

type UpdateTeamRequest struct {
	Name        string  `json:"name" binding:"omitempty,min=1,max=150"`
	Description *string `json:"description"`
}

type TeamResponse struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description *string   `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type TeamListResponse struct {
	Items      []TeamResponse `json:"items"`
	Total      int64          `json:"total"`
	Page       int            `json:"page"`
	TotalPages int            `json:"total_pages"`
}

type AddTeamMemberRequest struct {
	UserID uuid.UUID `json:"user_id" binding:"required"`
	Role   string    `json:"role" binding:"required,oneof=member manager owner"`
}

type TeamMemberResponse struct {
	TeamID   uuid.UUID `json:"team_id"`
	UserID   uuid.UUID `json:"user_id"`
	Role     string    `json:"role"`
	JoinedAt time.Time `json:"joined_at"`
}

type TeamMemberListResponse struct {
	Items      []TeamMemberResponse `json:"items"`
	Total      int64                `json:"total"`
	Page       int                  `json:"page"`
	TotalPages int                  `json:"total_pages"`
}
