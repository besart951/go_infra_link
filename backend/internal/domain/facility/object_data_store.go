package facility

import "github.com/google/uuid"

// ObjectDataStore extends the base repository with helper methods
// needed to apply ObjectData templates.
type ObjectDataStore interface {
	ObjectDataRepository

	GetBacnetObjectIDs(objectDataID uuid.UUID) ([]uuid.UUID, error)
}
