package project

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
	"github.com/google/uuid"
)

type ProjectControlCabinet struct {
	domain.Base
	ProjectID        uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_project_control_cabinet_unique"`
	ControlCabinetID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_project_control_cabinet_unique"`
}

func (ProjectControlCabinet) TableName() string {
	return "project_control_cabinets"
}

type ProjectControlCabinetRepository = domain.Repository[ProjectControlCabinet]
