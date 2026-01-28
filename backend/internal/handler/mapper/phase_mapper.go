package mapper

import (
	"github.com/besart951/go_infra_link/backend/internal/domain/project"
	"github.com/besart951/go_infra_link/backend/internal/handler/dto"
)

// Phase mappers

func ToPhaseModel(req dto.CreatePhaseRequest) *project.Phase {
	return &project.Phase{
		Name: req.Name,
	}
}

func ToPhaseResponse(phase *project.Phase) dto.PhaseResponse {
	return dto.PhaseResponse{
		ID:        phase.ID,
		Name:      phase.Name,
		CreatedAt: phase.CreatedAt,
		UpdatedAt: phase.UpdatedAt,
	}
}

func ToPhaseListResponse(phases []project.Phase) []dto.PhaseResponse {
	result := make([]dto.PhaseResponse, len(phases))
	for i, phase := range phases {
		result[i] = dto.PhaseResponse{
			ID:        phase.ID,
			Name:      phase.Name,
			CreatedAt: phase.CreatedAt,
			UpdatedAt: phase.UpdatedAt,
		}
	}
	return result
}

func ApplyPhaseUpdate(phase *project.Phase, req dto.UpdatePhaseRequest) {
	if req.Name != "" {
		phase.Name = req.Name
	}
}

// PhasePermission mappers

func ToPhasePermissionModel(req dto.CreatePhasePermissionRequest) *project.PhasePermission {
	return &project.PhasePermission{
		PhaseID:    req.PhaseID,
		Role:       req.Role,
		Permission: req.Permission,
	}
}

func ToPhasePermissionResponse(perm *project.PhasePermission) dto.PhasePermissionResponse {
	return dto.PhasePermissionResponse{
		ID:         perm.ID,
		PhaseID:    perm.PhaseID,
		Role:       perm.Role,
		Permission: perm.Permission,
		CreatedAt:  perm.CreatedAt,
		UpdatedAt:  perm.UpdatedAt,
	}
}

func ToPhasePermissionListResponse(perms []project.PhasePermission) []dto.PhasePermissionResponse {
	result := make([]dto.PhasePermissionResponse, len(perms))
	for i, perm := range perms {
		result[i] = dto.PhasePermissionResponse{
			ID:         perm.ID,
			PhaseID:    perm.PhaseID,
			Role:       perm.Role,
			Permission: perm.Permission,
			CreatedAt:  perm.CreatedAt,
			UpdatedAt:  perm.UpdatedAt,
		}
	}
	return result
}

func ApplyPhasePermissionUpdate(perm *project.PhasePermission, req dto.UpdatePhasePermissionRequest) {
	perm.Permission = req.Permission
}
