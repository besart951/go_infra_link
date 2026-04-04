package facility

import (
	"context"

	"github.com/google/uuid"
)

// SpecificationStore extends the generic repository with access patterns
// needed for the 1:1 (nullable) FieldDevice <-> Specification relationship.
//
// A FieldDevice may have 0 or 1 Specification.
// A Specification belongs to exactly 1 FieldDevice (when present).
type SpecificationStore interface {
	SpecificationRepository

	// BulkCreate creates multiple specifications in batches.
	BulkCreate(ctx context.Context, entities []*Specification, batchSize int) error

	GetByFieldDeviceIDs(ctx context.Context, fieldDeviceIDs []uuid.UUID) ([]*Specification, error)
	DeleteByFieldDeviceIDs(ctx context.Context, fieldDeviceIDs []uuid.UUID) error
}
