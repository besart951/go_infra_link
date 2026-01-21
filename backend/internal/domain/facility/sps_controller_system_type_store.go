package facility

import "github.com/google/uuid"

// SPSControllerSystemTypeStore extends the basic CRUD repository with helper operations
// needed by the service layer to keep associations consistent.
//
// Read-Check rule: the service layer uses this to replace associations without
// relying on database FK errors.
type SPSControllerSystemTypeStore interface {
	SPSControllerSystemTypeRepository
	SoftDeleteBySPSControllerIDs(ids []uuid.UUID) error
}
