package project

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
	"github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
)

type ProjectSPSController struct {
	domain.Base
	ProjectID       uuid.UUID              `gorm:"type:uuid;not null;uniqueIndex:idx_project_sps_controller_unique"`
	Project         Project                `gorm:"foreignKey:ProjectID;references:ID"`
	SPSControllerID uuid.UUID              `gorm:"type:uuid;not null;uniqueIndex:idx_project_sps_controller_unique"`
	SPSController   facility.SPSController `gorm:"foreignKey:SPSControllerID;references:ID"`
}

func (ProjectSPSController) TableName() string {
	return "project_sps_controllers"
}

type ProjectSPSControllerRepository interface {
	domain.Repository[ProjectSPSController]
	GetPaginatedListByProjectID(projectID uuid.UUID, params domain.PaginationParams) (*domain.PaginatedList[ProjectSPSController], error)
}
