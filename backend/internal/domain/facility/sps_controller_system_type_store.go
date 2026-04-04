package facility

import (
	"context"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	"github.com/google/uuid"
)

// SPSControllerSystemTypeStore extends the basic CRUD repository with helper operations
// needed by the service layer to keep associations consistent.
//
// Read-Check rule: the service layer uses this to replace associations without
// relying on database FK errors.
type SPSControllerSystemTypeStore interface {
	SPSControllerSystemTypeRepository
	GetPaginatedListBySPSControllerID(ctx context.Context, spsControllerID uuid.UUID, params domain.PaginationParams) (*domain.PaginatedList[SPSControllerSystemType], error)
	GetPaginatedListByProjectID(ctx context.Context, projectID uuid.UUID, params domain.PaginationParams) (*domain.PaginatedList[SPSControllerSystemType], error)
	ListBySPSControllerID(ctx context.Context, spsControllerID uuid.UUID) ([]*SPSControllerSystemType, error)
	GetIDsBySPSControllerIDs(ctx context.Context, ids []uuid.UUID) ([]uuid.UUID, error)
	DeleteBySPSControllerIDs(ctx context.Context, ids []uuid.UUID) error
}
