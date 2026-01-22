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
	Name        string        `gorm:"not null"`
	Description string
	Status      ProjectStatus `gorm:"type:varchar(50);not null"`
	StartDate   *time.Time
	PhaseID     uuid.UUID `json:"phase_id" gorm:"type:uuid;not null"`
	Phase       Phase     `gorm:"foreignKey:PhaseID"`
	CreatorID   uuid.UUID `json:"creator_id" gorm:"type:uuid;not null"`
	Creator     user.User `gorm:"foreignKey:CreatorID"`

	Users []*user.User `json:"users,omitempty" gorm:"many2many:project_users;"`
}

type Phase struct {
	domain.Base
	Name string `gorm:"not null"`
}

type ProjectRepository = domain.Repository[Project]
