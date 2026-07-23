package objectdata

import (
	"context"

	"github.com/google/uuid"
)

// BacnetObjectOwner is the lightweight ObjectData association projection used
// for project-scope resolution. One BACnet object may have multiple rows because
// the current schema permits multiple ObjectData links.
type BacnetObjectOwner struct {
	BacnetObjectID uuid.UUID
	ObjectDataID   uuid.UUID
	ProjectID      *uuid.UUID
}

// BacnetObjectOwnerReader resolves ObjectData ownership in batches without
// loading template children or complete facility hierarchies.
type BacnetObjectOwnerReader interface {
	GetByBacnetObjectIDs(
		ctx context.Context,
		bacnetObjectIDs []uuid.UUID,
	) ([]BacnetObjectOwner, error)
}
