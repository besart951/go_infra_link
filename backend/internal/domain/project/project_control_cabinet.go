package project

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
	"github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
)

type ProjectControlCabinet struct {
	domain.Base
	ProjectID        uuid.UUID               `gorm:"type:uuid;not null;uniqueIndex:idx_project_control_cabinet_unique"`
	Project          Project                 `gorm:"foreignKey:ProjectID;references:ID"`
	ControlCabinetID uuid.UUID               `gorm:"type:uuid;not null;uniqueIndex:idx_project_control_cabinet_unique"`
	ControlCabinet   facility.ControlCabinet `gorm:"foreignKey:ControlCabinetID;references:ID"`
}

func (ProjectControlCabinet) TableName() string {
	return "project_control_cabinets"
}

type ProjectControlCabinetRepository interface {
	domain.Repository[ProjectControlCabinet]
	GetPaginatedListByProjectID(projectID uuid.UUID, params domain.PaginationParams) (*domain.PaginatedList[ProjectControlCabinet], error)
}
