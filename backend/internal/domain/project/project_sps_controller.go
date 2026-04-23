package project

import (
	"context"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	"github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
)

type ProjectSPSController struct {
	domain.Base
	ProjectID       uuid.UUID
	Project         Project
	SPSControllerID uuid.UUID
	SPSController   facility.SPSController
}

type ProjectSPSControllerRepository interface {
	domain.Repository[ProjectSPSController]
	GetPaginatedListByProjectID(ctx context.Context, projectID uuid.UUID, params domain.PaginationParams) (*domain.PaginatedList[ProjectSPSController], error)
	GetBySPSControllerID(ctx context.Context, spsControllerID uuid.UUID) ([]*ProjectSPSController, error)
	GetBySPSControllerIDs(ctx context.Context, spsControllerIDs []uuid.UUID) ([]*ProjectSPSController, error)
	DeleteBySPSControllerIDs(ctx context.Context, spsControllerIDs []uuid.UUID) error
}
