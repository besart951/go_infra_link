package facility

import (
	"context"

	"github.com/google/uuid"
)

// BacnetObjectStore extends the base repository with helper methods
// used for FieldDevice hydration and bulk operations.
type BacnetObjectStore interface {
	BacnetObjectRepository

	// BulkCreate creates multiple BACnet objects in batches.
	BulkCreate(ctx context.Context, entities []*BacnetObject, batchSize int) error

	GetByFieldDeviceIDs(ctx context.Context, ids []uuid.UUID) ([]*BacnetObject, error)
	DeleteByFieldDeviceIDs(ctx context.Context, ids []uuid.UUID) error
}
