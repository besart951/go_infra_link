package project

import (
	"context"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	"github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
)

type ProjectControlCabinet struct {
	domain.Base
	ProjectID        uuid.UUID
	Project          Project
	ControlCabinetID uuid.UUID
	ControlCabinet   facility.ControlCabinet
}

type ProjectControlCabinetRepository interface {
	domain.Repository[ProjectControlCabinet]
	GetPaginatedListByProjectID(ctx context.Context, projectID uuid.UUID, params domain.PaginationParams) (*domain.PaginatedList[ProjectControlCabinet], error)
	GetByControlCabinetID(ctx context.Context, controlCabinetID uuid.UUID) ([]*ProjectControlCabinet, error)
	GetByControlCabinetIDs(ctx context.Context, controlCabinetIDs []uuid.UUID) ([]*ProjectControlCabinet, error)
	DeleteByControlCabinetIDs(ctx context.Context, controlCabinetIDs []uuid.UUID) error
}
