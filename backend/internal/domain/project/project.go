package project

import (
	"time"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	"github.com/besart951/go_infra_link/backend/internal/domain/user"
	"github.com/google/uuid"
)

type ProjectStatus string

const (
	StatusPlanned   ProjectStatus = "planned"
	StatusOngoing   ProjectStatus = "ongoing"
	StatusCompleted ProjectStatus = "completed"
)

type Project struct {
	domain.Base
	Name        string
	Description string
	Status      ProjectStatus
	StartDate   *time.Time
	PhaseID     uuid.UUID `json:"phase_id"`
	Phase       Phase
	CreatorID   uuid.UUID `json:"creator_id"`
	Creator     user.User

	Users []*user.User `json:"users,omitempty"`
}

type Phase struct {
	domain.Base
	Name string
}

type ProjectRepository = domain.Repository[Project]
