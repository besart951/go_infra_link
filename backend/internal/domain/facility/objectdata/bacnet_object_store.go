package objectdata

import (
	"context"

	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
)

// BacnetObjectStore extends the base repository with helper methods
// used for FieldDevice hydration and bulk operations.
type BacnetObjectStore interface {
	domainFacility.BacnetObjectRepository

	// BulkCreate creates multiple BACnet objects in batches.
	BulkCreate(ctx context.Context, entities []*domainFacility.BacnetObject, batchSize int) error

	GetByFieldDeviceIDs(ctx context.Context, ids []uuid.UUID) ([]*domainFacility.BacnetObject, error)
	DeleteByFieldDeviceIDs(ctx context.Context, ids []uuid.UUID) error
}
