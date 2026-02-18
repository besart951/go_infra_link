package facility

import "github.com/google/uuid"

// BacnetObjectStore extends the base repository with helper methods
// used for FieldDevice hydration and bulk operations.
type BacnetObjectStore interface {
	BacnetObjectRepository

	GetByFieldDeviceIDs(ids []uuid.UUID) ([]*BacnetObject, error)
	DeleteByFieldDeviceIDs(ids []uuid.UUID) error
}
