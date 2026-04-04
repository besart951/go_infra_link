package project

import (
	"context"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	"github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
)

type ProjectFieldDevice struct {
	domain.Base
	ProjectID     uuid.UUID            `gorm:"type:uuid;not null;uniqueIndex:idx_project_field_device_unique"`
	Project       Project              `gorm:"foreignKey:ProjectID;references:ID"`
	FieldDeviceID uuid.UUID            `gorm:"type:uuid;not null;uniqueIndex:idx_project_field_device_unique"`
	FieldDevice   facility.FieldDevice `gorm:"foreignKey:FieldDeviceID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

func (ProjectFieldDevice) TableName() string {
	return "project_field_devices"
}

type ProjectFieldDeviceRepository interface {
	domain.Repository[ProjectFieldDevice]
	GetPaginatedListByProjectID(ctx context.Context, projectID uuid.UUID, params domain.PaginationParams) (*domain.PaginatedList[ProjectFieldDevice], error)
}
