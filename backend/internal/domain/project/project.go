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
	Name        string `gorm:"not null"`
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
	Name      string    `gorm:"not null"`
	ProjectID uuid.UUID `json:"project_id" gorm:"type:uuid;not null"`
	Project   *Project  `gorm:"foreignKey:ProjectID"`
}

type PermissionType string

const (
	PermissionEdit            PermissionType = "edit"
	PermissionSuggestChanges  PermissionType = "suggest_changes"
	PermissionView            PermissionType = "view"
	PermissionDelete          PermissionType = "delete"
	PermissionManageUsers     PermissionType = "manage_users"
)

// PhasePermission defines what a specific role can do in a specific phase
type PhasePermission struct {
	domain.Base
	PhaseID    uuid.UUID      `json:"phase_id" gorm:"type:uuid;not null;index:idx_phase_role,unique"`
	Phase      *Phase         `gorm:"foreignKey:PhaseID"`
	Role       user.Role      `json:"role" gorm:"type:varchar(50);not null;index:idx_phase_role,unique"`
	Permission PermissionType `json:"permission" gorm:"type:varchar(50);not null"`
}

type ProjectRepository interface {
	domain.Repository[Project]
	AddUser(projectID, userID uuid.UUID) error
	RemoveUser(projectID, userID uuid.UUID) error
	ListUsers(projectID uuid.UUID) ([]user.User, error)
}

type PhaseRepository = domain.Repository[Phase]

type PhasePermissionRepository interface {
	domain.Repository[PhasePermission]
	GetByPhaseAndRole(phaseID uuid.UUID, role user.Role) (*PhasePermission, error)
	ListByPhase(phaseID uuid.UUID) ([]PhasePermission, error)
	DeleteByPhaseAndRole(phaseID uuid.UUID, role user.Role) error
}
