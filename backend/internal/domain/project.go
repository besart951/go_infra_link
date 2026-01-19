package domain

import (
	"time"
	"github.com/google/uuid"
)

type ProjectStatus string

const (
	StatusPlanned   ProjectStatus = "planned"
	StatusOngoing   ProjectStatus = "ongoing"
	StatusCompleted ProjectStatus = "completed"
)

type Project struct {
	Base
	Name        string        `gorm:"size:255;not null"`
	Description string        `gorm:"type:text"`
	Status      ProjectStatus `gorm:"type:varchar(20);default:'planned'"`
	StartDate   *time.Time
	PhaseID     uuid.UUID     `json:"phase_id"`
	Phase       Phase         `gorm:"foreignKey:PhaseID"`
	CreatorID   uuid.UUID     `json:"creator_id"`
	Creator     User          `gorm:"foreignKey:CreatorID"`
	
	Users []*User `gorm:"many2many:project_users;" json:"users,omitempty"`
}

type Phase struct {
	Base
	Name string `gorm:"uniqueIndex"`
}

type ProjectRepository interface {
	GetByIds(ids []uuid.UUID) ([]*Project, error)
	Create(entity *Project) error
	Update(entity *Project) error
	DeleteByIds(ids []uuid.UUID) error
	GetPaginatedList(params PaginationParams) (*PaginatedList[Project], error)
}
