package project

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
	"github.com/google/uuid"
)

type ProjectControlCabinet struct {
	domain.Base
	ProjectID        uuid.UUID `gorm:"type:uuid;not null;index:idx_project_control_cabinet"`
	ControlCabinetID uuid.UUID `gorm:"type:uuid;not null;index:idx_project_control_cabinet"`
}

func (ProjectControlCabinet) TableName() string {
	return "project_control_cabinets"
}

type ProjectControlCabinetRepository = domain.Repository[ProjectControlCabinet]
