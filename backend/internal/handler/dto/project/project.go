package project

import (
	"time"

	domainproject "github.com/besart951/go_infra_link/backend/internal/domain/project"
	"github.com/google/uuid"
)

// Project DTOs

type CreateProjectRequest struct {
	Name        string         `json:"name" binding:"required,min=1,max=255"`
	Description string         `json:"description"`
	Status      string         `json:"status" binding:"omitempty,oneof=planned ongoing completed"`
	StartDate   *SwissDateTime `json:"start_date"`
	PhaseID     uuid.UUID      `json:"phase_id" binding:"required"`
}

type UpdateProjectRequest struct {
	BaseVersion *uint64                      `json:"base_version" binding:"omitempty,min=1"`
	Name        *string                      `json:"name" binding:"omitempty,max=255"`
	Description *string                      `json:"description"`
	Status      *domainproject.ProjectStatus `json:"status" binding:"omitempty,oneof=planned ongoing completed"`
	StartDate   OptionalSwissDateTime        `json:"start_date" swaggertype:"string" format:"date-time" extensions:"x-nullable"`
	PhaseID     *uuid.UUID                   `json:"phase_id"`
}

type ProjectResponse struct {
	ID          uuid.UUID                   `json:"id"`
	Version     uint64                      `json:"version"`
	Name        string                      `json:"name"`
	Description string                      `json:"description"`
	Status      domainproject.ProjectStatus `json:"status"`
	StartDate   *time.Time                  `json:"start_date"`
	PhaseID     uuid.UUID                   `json:"phase_id"`
	Phase       *PhaseResponse              `json:"phase,omitempty"`
	CreatorID   uuid.UUID                   `json:"creator_id"`
	CreatedAt   time.Time                   `json:"created_at"`
	UpdatedAt   time.Time                   `json:"updated_at"`
}

type ProjectListResponse struct {
	Items      []ProjectResponse `json:"items"`
	Total      int64             `json:"total"`
	Page       int               `json:"page"`
	TotalPages int               `json:"total_pages"`
}

// ProjectCapabilitiesResponse contains the project-scoped permissions that are
// effective for the authenticated user in one concrete project.
type ProjectCapabilitiesResponse struct {
	Permissions []string `json:"permissions"`
}

type ListProjectsQuery struct {
	PaginationQuery
	Status  domainproject.ProjectStatus `form:"status" binding:"omitempty,oneof=planned ongoing completed"`
	PhaseID string                      `form:"phase_id" binding:"omitempty,uuid"`
}
