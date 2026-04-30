package project

import (
	"context"
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
	Phase       *Phase
	CreatorID   uuid.UUID `json:"creator_id"`
	Creator     user.User

	Users []*user.User `json:"users,omitempty"`
}

type Phase struct {
	domain.Base
	Name string
}

type PhasePermission struct {
	domain.Base
	PhaseID     uuid.UUID `json:"phase_id" gorm:"type:uuid;not null;uniqueIndex:idx_phase_permission_phase_role"`
	Phase       *Phase
	Role        user.Role `json:"role" gorm:"type:varchar(50);not null;uniqueIndex:idx_phase_permission_phase_role"`
	Permissions []string  `json:"permissions" gorm:"serializer:json;type:text;not null"`
}

type ProjectRepository interface {
	domain.Repository[Project]
	GetPaginatedListForUser(ctx context.Context, params domain.PaginationParams, userID uuid.UUID) (*domain.PaginatedList[Project], error)
	GetPaginatedListWithStatus(ctx context.Context, params domain.PaginationParams, status *ProjectStatus) (*domain.PaginatedList[Project], error)
	GetPaginatedListForUserWithStatus(ctx context.Context, params domain.PaginationParams, userID uuid.UUID, status *ProjectStatus) (*domain.PaginatedList[Project], error)
	HasUser(ctx context.Context, projectID, userID uuid.UUID) (bool, error)
	AddUser(ctx context.Context, projectID, userID uuid.UUID) error
	RemoveUser(ctx context.Context, projectID, userID uuid.UUID) error
	ListUsers(ctx context.Context, projectID uuid.UUID) ([]user.User, error)
}

type PhaseRepository = domain.Repository[Phase]

type PhasePermissionRepository interface {
	domain.Repository[PhasePermission]
	GetByPhaseAndRole(ctx context.Context, phaseID uuid.UUID, role user.Role) (*PhasePermission, error)
	List(ctx context.Context, phaseID *uuid.UUID) ([]PhasePermission, error)
}
