package mapper

import (
	"github.com/besart951/go_infra_link/backend/internal/domain/team"
	"github.com/besart951/go_infra_link/backend/internal/handler/dto"
)

// ToTeamModel converts a CreateTeamRequest to a Team domain model
func ToTeamModel(req dto.CreateTeamRequest) *team.Team {
	return &team.Team{
		Name:        req.Name,
		Description: req.Description,
	}
}

// ApplyTeamUpdate applies UpdateTeamRequest fields to an existing Team
func ApplyTeamUpdate(target *team.Team, req dto.UpdateTeamRequest) {
	if req.Name != "" {
		target.Name = req.Name
	}
	if req.Description != nil {
		target.Description = req.Description
	}
}

// ToTeamResponse converts a Team domain model to a TeamResponse DTO
func ToTeamResponse(t *team.Team) dto.TeamResponse {
	return dto.TeamResponse{
		ID:          t.ID,
		Name:        t.Name,
		Description: t.Description,
		CreatedAt:   t.CreatedAt,
		UpdatedAt:   t.UpdatedAt,
	}
}

// ToTeamListResponse converts a list of Teams to TeamResponses
func ToTeamListResponse(teams []team.Team) []dto.TeamResponse {
	items := make([]dto.TeamResponse, len(teams))
	for i, t := range teams {
		items[i] = ToTeamResponse(&t)
	}
	return items
}

// ToTeamMemberResponse converts a TeamMember domain model to a TeamMemberResponse DTO
func ToTeamMemberResponse(m *team.TeamMember) dto.TeamMemberResponse {
	return dto.TeamMemberResponse{
		TeamID:   m.TeamID,
		UserID:   m.UserID,
		Role:     string(m.Role),
		JoinedAt: m.JoinedAt,
	}
}

// ToTeamMemberListResponse converts a list of TeamMembers to TeamMemberResponses
func ToTeamMemberListResponse(members []team.TeamMember) []dto.TeamMemberResponse {
	items := make([]dto.TeamMemberResponse, len(members))
	for i, m := range members {
		items[i] = ToTeamMemberResponse(&m)
	}
	return items
}
