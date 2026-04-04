package facility

import (
	"context"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	"github.com/google/uuid"
)

// ObjectDataStore extends the base repository with helper methods
// needed to apply ObjectData templates.
type ObjectDataStore interface {
	ObjectDataRepository

	GetBacnetObjectIDs(ctx context.Context, objectDataID uuid.UUID) ([]uuid.UUID, error)
	ExistsByDescription(ctx context.Context, projectID *uuid.UUID, description string, excludeID *uuid.UUID) (bool, error)
	GetTemplates(ctx context.Context) ([]*ObjectData, error)
	GetTemplatesLite(ctx context.Context) ([]*ObjectData, error)
	GetForProject(ctx context.Context, projectID uuid.UUID) ([]*ObjectData, error)
	GetForProjectLite(ctx context.Context, projectID uuid.UUID) ([]*ObjectData, error)
	GetPaginatedListForProject(ctx context.Context, projectID uuid.UUID, params domain.PaginationParams) (*domain.PaginatedList[ObjectData], error)
	GetPaginatedListByApparatID(ctx context.Context, apparatID uuid.UUID, params domain.PaginationParams) (*domain.PaginatedList[ObjectData], error)
	GetPaginatedListBySystemPartID(ctx context.Context, systemPartID uuid.UUID, params domain.PaginationParams) (*domain.PaginatedList[ObjectData], error)
	GetPaginatedListByApparatAndSystemPartID(ctx context.Context, apparatID, systemPartID uuid.UUID, params domain.PaginationParams) (*domain.PaginatedList[ObjectData], error)
	GetPaginatedListForProjectByApparatID(ctx context.Context, projectID, apparatID uuid.UUID, params domain.PaginationParams) (*domain.PaginatedList[ObjectData], error)
	GetPaginatedListForProjectBySystemPartID(ctx context.Context, projectID, systemPartID uuid.UUID, params domain.PaginationParams) (*domain.PaginatedList[ObjectData], error)
	GetPaginatedListForProjectByApparatAndSystemPartID(ctx context.Context, projectID, apparatID, systemPartID uuid.UUID, params domain.PaginationParams) (*domain.PaginatedList[ObjectData], error)
}
