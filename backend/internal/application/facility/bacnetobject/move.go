package bacnetobject

import (
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
)

// MoveCommand captures a change to the direct FieldDevice owner. An
// ObjectData association remains an application/service operation because it
// is represented by a join table rather than a BacnetObject field.
type MoveCommand struct {
	BacnetObjectID    uuid.UUID
	FromFieldDeviceID *uuid.UUID
	ToFieldDeviceID   *uuid.UUID
}

func newMoveCommand(
	current *domainFacility.BacnetObject,
	command UpdateCommand,
) *MoveCommand {
	fieldDeviceSet := command.FieldDeviceSet || command.FieldDeviceID != nil
	if current == nil || (!fieldDeviceSet && command.ObjectDataID == nil) {
		return nil
	}

	to := clonePointer(current.FieldDeviceID)
	if fieldDeviceSet {
		to = clonePointer(command.FieldDeviceID)
	} else if command.ObjectDataID != nil {
		to = nil
	}
	if equalPointers(current.FieldDeviceID, to) {
		return nil
	}

	return &MoveCommand{
		BacnetObjectID:    current.ID,
		FromFieldDeviceID: clonePointer(current.FieldDeviceID),
		ToFieldDeviceID:   to,
	}
}

func (c MoveCommand) applyTo(object *domainFacility.BacnetObject) error {
	if c.ToFieldDeviceID == nil {
		return object.DetachFromFieldDevice()
	}
	return object.AssignToFieldDevice(*c.ToFieldDeviceID)
}

func (c UpdateCommand) applyAssignmentTo(object *domainFacility.BacnetObject) error {
	fieldDeviceSet := c.FieldDeviceSet || c.FieldDeviceID != nil
	if fieldDeviceSet && c.FieldDeviceID == nil {
		return object.DetachFromFieldDevice()
	}
	if fieldDeviceSet && c.FieldDeviceID != nil {
		return object.AssignToFieldDevice(*c.FieldDeviceID)
	}
	if c.ObjectDataID != nil {
		return object.DetachFromFieldDevice()
	}
	return nil
}
