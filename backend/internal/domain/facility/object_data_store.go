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
	GetPaginatedListWithFilters(ctx context.Context, params domain.PaginationParams, filters ObjectDataFilterParams) (*domain.PaginatedList[ObjectData], error)
}

type ObjectDataFilterParams struct {
	ProjectID    *uuid.UUID
	ApparatID    *uuid.UUID
	SystemPartID *uuid.UUID
}
