package project

import (
	"time"

	"github.com/besart951/go_infra_link/backend/internal/domain/project"
	dto "github.com/besart951/go_infra_link/backend/internal/handler/dto/project"
)

// ToProjectModel converts a CreateProjectRequest to a Project domain model
func ToProjectModel(req dto.CreateProjectRequest) *project.Project {
	proj := &project.Project{
		Name:        req.Name,
		Description: req.Description,
		Status:      project.ProjectStatus(req.Status),
		StartDate:   toTimePtr(req.StartDate),
		PhaseID:     req.PhaseID,
	}

	return proj
}

// ApplyProjectUpdate applies UpdateProjectRequest fields to an existing Project
func ApplyProjectUpdate(target *project.Project, req dto.UpdateProjectRequest) {
	if req.Name != nil {
		target.Name = *req.Name
	}
	if req.Description != nil {
		target.Description = *req.Description
	}
	if req.Status != nil {
		target.Status = *req.Status
	}
	if req.StartDate.Set {
		target.StartDate = toTimePtr(req.StartDate.Value)
	}
	if req.PhaseID != nil {
		target.PhaseID = *req.PhaseID
	}
}

func toTimePtr(value *dto.SwissDateTime) *time.Time {
	if value == nil {
		return nil
	}
	parsed := value.Time
	return &parsed
}

// ToProjectResponse converts a Project domain model to a ProjectResponse DTO
func ToProjectResponse(p *project.Project) dto.ProjectResponse {
	return dto.ProjectResponse{
		ID:          p.ID,
		Name:        p.Name,
		Description: p.Description,
		Status:      p.Status,
		StartDate:   p.StartDate,
		PhaseID:     p.PhaseID,
		CreatorID:   p.CreatorID,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}
}

// ToProjectListResponse converts a list of Projects to ProjectResponses
func ToProjectListResponse(projects []project.Project) []dto.ProjectResponse {
	items := make([]dto.ProjectResponse, len(projects))
	for i, p := range projects {
		items[i] = ToProjectResponse(&p)
	}
	return items
}
