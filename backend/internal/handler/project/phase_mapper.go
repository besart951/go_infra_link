package project

import (
	"github.com/besart951/go_infra_link/backend/internal/domain/project"
	dto "github.com/besart951/go_infra_link/backend/internal/handler/dto/project"
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
