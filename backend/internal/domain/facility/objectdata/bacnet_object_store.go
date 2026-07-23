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

	// AssignSoftwareReferenceIDs remaps references after a copied BACnet object
	// set has been created. The map key is the copied object ID and the value is
	// its copied reference target ID.
	AssignSoftwareReferenceIDs(ctx context.Context, assignments map[uuid.UUID]uuid.UUID) error

	GetByFieldDeviceIDs(ctx context.Context, ids []uuid.UUID) ([]*domainFacility.BacnetObject, error)
	DeleteByFieldDeviceIDs(ctx context.Context, ids []uuid.UUID) error
}

// ObjectDataBacnetObjectCreator is the specialized persistence capability for
// creating a BACnet object and its ObjectData owner link as one transaction
// phase. History decorators use this seam to resolve the owner after both rows
// exist instead of recording an unscoped create between the two writes.
type ObjectDataBacnetObjectCreator interface {
	CreateForObjectData(
		ctx context.Context,
		objectDataID uuid.UUID,
		entity *domainFacility.BacnetObject,
	) error
}
