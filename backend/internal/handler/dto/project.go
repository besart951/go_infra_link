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
	PhaseID     uuid.UUID  `json:"phase_id" binding:"required"`
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
