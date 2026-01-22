package project

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
	"github.com/google/uuid"
)

type ProjectSPSController struct {
	domain.Base
	ProjectID       uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_project_sps_controller_unique"`
	SPSControllerID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_project_sps_controller_unique"`
}

func (ProjectSPSController) TableName() string {
	return "project_sps_controllers"
}

type ProjectSPSControllerRepository = domain.Repository[ProjectSPSController]
