package dto

import (
	"time"

	"github.com/besart951/go_infra_link/backend/internal/domain/project"
	"github.com/google/uuid"
)

type DashboardProjectResponse struct {
	ID        uuid.UUID             `json:"id"`
	Name      string                `json:"name"`
	Status    project.ProjectStatus `json:"status"`
	Phase     string                `json:"phase"`
	UpdatedAt time.Time             `json:"updated_at"`
}

type DashboardUserPresenceResponse struct {
	ID          uuid.UUID  `json:"id"`
	FirstName   string     `json:"first_name"`
	LastName    string     `json:"last_name"`
	Email       string     `json:"email"`
	LastLoginAt *time.Time `json:"last_login_at,omitempty"`
	IsOnline    bool       `json:"is_online"`
}

type DashboardTeamMemberResponse struct {
	UserID      uuid.UUID  `json:"user_id"`
	FirstName   string     `json:"first_name"`
	LastName    string     `json:"last_name"`
	Email       string     `json:"email"`
	Role        string     `json:"role"`
	LastLoginAt *time.Time `json:"last_login_at,omitempty"`
	IsOnline    bool       `json:"is_online"`
}

type DashboardTeamSummaryResponse struct {
	ID       uuid.UUID `json:"id"`
	Name     string    `json:"name"`
	Role     string    `json:"role"`
	JoinedAt time.Time `json:"joined_at"`
}

type DashboardTeamResponse struct {
	ID      uuid.UUID                     `json:"id"`
	Name    string                        `json:"name"`
	Role    string                        `json:"role"`
	Members []DashboardTeamMemberResponse `json:"members"`
}

type DashboardResponse struct {
	LastProject *DashboardProjectResponse       `json:"last_project,omitempty"`
	PrimaryTeam *DashboardTeamResponse          `json:"primary_team,omitempty"`
	Teams       []DashboardTeamSummaryResponse  `json:"teams"`
	OnlineUsers []DashboardUserPresenceResponse `json:"online_users"`
}
