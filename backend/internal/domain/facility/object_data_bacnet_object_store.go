package facility

import (
	"context"

	"github.com/google/uuid"
)

// ObjectDataBacnetObjectStore manages the object_data_bacnet_objects join table.
// It is used to attach/detach bacnet objects to object data templates.
type ObjectDataBacnetObjectStore interface {
	Add(ctx context.Context, objectDataID uuid.UUID, bacnetObjectID uuid.UUID) error
	Delete(ctx context.Context, objectDataID uuid.UUID, bacnetObjectID uuid.UUID) error
	DeleteByObjectDataID(ctx context.Context, objectDataID uuid.UUID) error
}
