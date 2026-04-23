package project

import (
	"context"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	"github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
)

type ProjectFieldDevice struct {
	domain.Base
	ProjectID     uuid.UUID
	Project       Project
	FieldDeviceID uuid.UUID
	FieldDevice   facility.FieldDevice
}

type ProjectFieldDeviceRepository interface {
	domain.Repository[ProjectFieldDevice]
	GetPaginatedListByProjectID(ctx context.Context, projectID uuid.UUID, params domain.PaginationParams) (*domain.PaginatedList[ProjectFieldDevice], error)
	GetByFieldDeviceIDs(ctx context.Context, fieldDeviceIDs []uuid.UUID) ([]*ProjectFieldDevice, error)
	DeleteByFieldDeviceIDs(ctx context.Context, fieldDeviceIDs []uuid.UUID) error
}
