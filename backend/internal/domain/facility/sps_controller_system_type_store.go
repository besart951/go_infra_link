package facility

import (
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
	GetPaginatedListBySPSControllerID(spsControllerID uuid.UUID, params domain.PaginationParams) (*domain.PaginatedList[SPSControllerSystemType], error)
	ListBySPSControllerID(spsControllerID uuid.UUID) ([]*SPSControllerSystemType, error)
	GetIDsBySPSControllerIDs(ids []uuid.UUID) ([]uuid.UUID, error)
	DeleteBySPSControllerIDs(ids []uuid.UUID) error
}
