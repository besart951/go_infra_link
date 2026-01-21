package facility

import "github.com/google/uuid"

// SpecificationStore extends the generic repository with access patterns
// needed for the 1:1 (nullable) FieldDevice <-> Specification relationship.
//
// A FieldDevice may have 0 or 1 Specification.
// A Specification belongs to exactly 1 FieldDevice (when present).
type SpecificationStore interface {
	SpecificationRepository

	GetByFieldDeviceIDs(fieldDeviceIDs []uuid.UUID) ([]*Specification, error)
	SoftDeleteByFieldDeviceIDs(fieldDeviceIDs []uuid.UUID) error
}
