package project

import (
	"github.com/besart951/go_infra_link/backend/internal/domain/user"
	"github.com/google/uuid"
)

type PermissionDenialReason string

const (
	PermissionDenialReasonMissingGeneral PermissionDenialReason = "missing_general_permission"
	PermissionDenialReasonPhaseBlocked   PermissionDenialReason = "phase_blocked"
	PermissionDenialReasonForbidden      PermissionDenialReason = "forbidden"
)

type PermissionRequiredRole struct {
	Role  user.Role `json:"role"`
	Label string    `json:"label"`
}

type PermissionDenialDetails struct {
	Reason             PermissionDenialReason   `json:"reason"`
	Permission         string                   `json:"permission,omitempty"`
	Permissions        []string                 `json:"permissions,omitempty"`
	ProjectID          *uuid.UUID               `json:"project_id,omitempty"`
	PhaseID            *uuid.UUID               `json:"phase_id,omitempty"`
	PhaseName          string                   `json:"phase_name,omitempty"`
	RequesterRole      user.Role                `json:"requester_role,omitempty"`
	RequesterRoleLabel string                   `json:"requester_role_label,omitempty"`
	MinimumRole        user.Role                `json:"minimum_role,omitempty"`
	MinimumRoleLabel   string                   `json:"minimum_role_label,omitempty"`
	RequiredRoles      []PermissionRequiredRole `json:"required_roles,omitempty"`
	Message            string                   `json:"message"`
}
