package project

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
	"github.com/google/uuid"
)

type ProjectFieldDevice struct {
	domain.Base
	ProjectID     uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_project_field_device_unique"`
	FieldDeviceID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_project_field_device_unique"`
}

func (ProjectFieldDevice) TableName() string {
	return "project_field_devices"
}

type ProjectFieldDeviceRepository = domain.Repository[ProjectFieldDevice]
