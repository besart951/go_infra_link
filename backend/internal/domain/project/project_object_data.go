package project

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
	"github.com/google/uuid"
)

type ProjectObjectData struct {
	domain.Base
	ProjectID     uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_project_object_data_unique"`
	ObjectDataID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_project_object_data_unique"`
}

func (ProjectObjectData) TableName() string {
	return "project_object_data"
}

type ProjectObjectDataRepository = domain.Repository[ProjectObjectData]
