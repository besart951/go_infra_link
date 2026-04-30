package project

import (
	"time"

	domainUser "github.com/besart951/go_infra_link/backend/internal/domain/user"
	"github.com/google/uuid"
)

type CreatePhasePermissionRequest struct {
	PhaseID     uuid.UUID       `json:"phase_id" binding:"required"`
	Role        domainUser.Role `json:"role" binding:"required,oneof=admin_fzag fzag admin_planer planer admin_entrepreneur entrepreneur"`
	Permissions []string        `json:"permissions"`
	Permission  *string         `json:"permission"`
}

type UpdatePhasePermissionRequest struct {
	PhaseID     *uuid.UUID       `json:"phase_id"`
	Role        *domainUser.Role `json:"role" binding:"omitempty,oneof=admin_fzag fzag admin_planer planer admin_entrepreneur entrepreneur"`
	Permissions *[]string        `json:"permissions"`
	Permission  *string          `json:"permission"`
}

type PhasePermissionResponse struct {
	ID          uuid.UUID       `json:"id"`
	PhaseID     uuid.UUID       `json:"phase_id"`
	Role        domainUser.Role `json:"role"`
	Permissions []string        `json:"permissions"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

type PhasePermissionListResponse struct {
	Items []PhasePermissionResponse `json:"items"`
}
