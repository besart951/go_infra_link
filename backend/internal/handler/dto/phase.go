package dto

import (
	"time"

	"github.com/besart951/go_infra_link/backend/internal/domain/project"
	"github.com/besart951/go_infra_link/backend/internal/domain/user"
	"github.com/google/uuid"
)

// Phase DTOs

type CreatePhaseRequest struct {
	Name      string    `json:"name" binding:"required,min=1,max=255"`
	ProjectID uuid.UUID `json:"project_id" binding:"required"`
}

type UpdatePhaseRequest struct {
	Name string `json:"name" binding:"omitempty,min=1,max=255"`
}

type PhaseResponse struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	ProjectID uuid.UUID `json:"project_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type PhaseListResponse struct {
	Items      []PhaseResponse `json:"items"`
	Total      int64           `json:"total"`
	Page       int             `json:"page"`
	TotalPages int             `json:"total_pages"`
}

// PhasePermission DTOs

type CreatePhasePermissionRequest struct {
	PhaseID    uuid.UUID             `json:"phase_id" binding:"required"`
	Role       user.Role             `json:"role" binding:"required,oneof=user admin superadmin admin_planer planer admin_entrepreneur entrepreneur"`
	Permission project.PermissionType `json:"permission" binding:"required,oneof=edit suggest_changes view delete manage_users"`
}

type UpdatePhasePermissionRequest struct {
	Permission project.PermissionType `json:"permission" binding:"required,oneof=edit suggest_changes view delete manage_users"`
}

type PhasePermissionResponse struct {
	ID         uuid.UUID             `json:"id"`
	PhaseID    uuid.UUID             `json:"phase_id"`
	Role       user.Role             `json:"role"`
	Permission project.PermissionType `json:"permission"`
	CreatedAt  time.Time             `json:"created_at"`
	UpdatedAt  time.Time             `json:"updated_at"`
}

type PhasePermissionListResponse struct {
	Items      []PhasePermissionResponse `json:"items"`
	Total      int64                     `json:"total"`
	Page       int                       `json:"page"`
	TotalPages int                       `json:"total_pages"`
}
