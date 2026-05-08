package objectdata

import (
	"context"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
)

// ObjectDataStore extends the base repository with helper methods
// needed to apply ObjectData templates.
type ObjectDataStore interface {
	domainFacility.ObjectDataRepository

	GetBacnetObjectIDs(ctx context.Context, objectDataID uuid.UUID) ([]uuid.UUID, error)
	ExistsByDescription(ctx context.Context, projectID *uuid.UUID, description string, excludeID *uuid.UUID) (bool, error)
	GetTemplates(ctx context.Context) ([]*domainFacility.ObjectData, error)
	GetTemplatesLite(ctx context.Context) ([]*domainFacility.ObjectData, error)
	GetForProject(ctx context.Context, projectID uuid.UUID) ([]*domainFacility.ObjectData, error)
	GetForProjectLite(ctx context.Context, projectID uuid.UUID) ([]*domainFacility.ObjectData, error)
	GetPaginatedListWithFilters(ctx context.Context, params domain.PaginationParams, filters ObjectDataFilterParams) (*domain.PaginatedList[domainFacility.ObjectData], error)
}

type ObjectDataFilterParams struct {
	ProjectID    *uuid.UUID
	ApparatID    *uuid.UUID
	SystemPartID *uuid.UUID
}
