package project

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
	"github.com/google/uuid"
)

type ProjectSPSController struct {
	domain.Base
	ProjectID       uuid.UUID `gorm:"type:uuid;not null;index:idx_project_sps_controller"`
	SPSControllerID uuid.UUID `gorm:"type:uuid;not null;index:idx_project_sps_controller"`
}

func (ProjectSPSController) TableName() string {
	return "project_sps_controllers"
}

type ProjectSPSControllerRepository = domain.Repository[ProjectSPSController]
