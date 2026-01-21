package facility

import "github.com/google/uuid"

// ObjectDataBacnetObjectStore manages the object_data_bacnet_objects join table.
// It is used to attach/detach bacnet objects to object data templates.
type ObjectDataBacnetObjectStore interface {
	Add(objectDataID uuid.UUID, bacnetObjectID uuid.UUID) error
	Delete(objectDataID uuid.UUID, bacnetObjectID uuid.UUID) error
}
