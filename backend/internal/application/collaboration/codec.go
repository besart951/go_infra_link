package collaboration

import (
	"context"
	"encoding/json"
	"fmt"

	domainCollaboration "github.com/besart951/go_infra_link/backend/internal/domain/collaboration"
)

// EncodedCommand is the stable payload stored in the collaboration outbox.
// The type discriminator is deliberately separate from JSON so command renames
// cannot be hidden by Go's permissive unmarshalling.
type EncodedCommand struct {
	Type    string
	Payload json.RawMessage
}

// NewOutboxEvent converts a sealed application command into the durable record
// a transaction-scoped infrastructure adapter writes before commit.
func NewOutboxEvent(command Command) (*domainCollaboration.OutboxEvent, error) {
	encoded, err := EncodeCommand(command)
	if err != nil {
		return nil, err
	}
	envelope, ok := commandEnvelope(command)
	if !ok || envelope.EventID == [16]byte{} || envelope.OperationID == [16]byte{} ||
		envelope.CorrelationID == [16]byte{} || envelope.ProjectID == [16]byte{} ||
		envelope.OccurredAt.IsZero() || envelope.SchemaVersion != SchemaVersionV2 {
		return nil, fmt.Errorf("collaboration command %s has incomplete envelope", encoded.Type)
	}
	occurredAt := envelope.OccurredAt.UTC()
	return &domainCollaboration.OutboxEvent{
		EventID:       envelope.EventID,
		EventType:     encoded.Type,
		SchemaVersion: envelope.SchemaVersion,
		OperationID:   envelope.OperationID,
		ProjectID:     envelope.ProjectID,
		Payload:       encoded.Payload,
		NextAttemptAt: occurredAt,
	}, nil
}

func CommandEnvelope(command Command) (Envelope, bool) {
	return commandEnvelope(command)
}

// EnqueueCommand writes through the transaction-scoped store installed by the
// transaction runner. The boolean is false for non-production test adapters
// that deliberately do not provide that infrastructure capability.
func EnqueueCommand(ctx context.Context, command Command) (bool, error) {
	store, ok := domainCollaboration.OutboxStoreFromContext(ctx)
	if !ok {
		return false, nil
	}
	event, err := NewOutboxEvent(command)
	if err != nil {
		return true, err
	}
	return true, store.Enqueue(ctx, event)
}

func OutboxConfigured(ctx context.Context) bool {
	_, ok := domainCollaboration.OutboxStoreFromContext(ctx)
	return ok
}

func EncodeCommand(command Command) (EncodedCommand, error) {
	commandType, ok := commandTypeOf(command)
	if !ok {
		return EncodedCommand{}, fmt.Errorf("unsupported collaboration command %T", command)
	}
	payload, err := json.Marshal(command)
	if err != nil {
		return EncodedCommand{}, fmt.Errorf("marshal collaboration command %s: %w", commandType, err)
	}
	return EncodedCommand{Type: commandType, Payload: payload}, nil
}

func DecodeCommand(encoded EncodedCommand) (Command, error) {
	if len(encoded.Payload) == 0 {
		return nil, fmt.Errorf("collaboration command %q has empty payload", encoded.Type)
	}
	var target any
	switch encoded.Type {
	case "facility_hierarchy_refresh_required":
		target = &FacilityHierarchyRefreshRequired{}
	case "control_cabinet_created":
		target = &ControlCabinetCreated{}
	case "control_cabinet_cloned":
		target = &ControlCabinetCloned{}
	case "control_cabinet_deleted":
		target = &ControlCabinetDeleted{}
	case "control_cabinet_updated":
		target = &ControlCabinetUpdated{}
	case "control_cabinet_moved":
		target = &ControlCabinetMoved{}
	case "field_device_updated":
		target = &FieldDeviceUpdated{}
	case "field_device_moved":
		target = &FieldDeviceMoved{}
	case "field_device_deleted":
		target = &FieldDeviceDeleted{}
	case "field_devices_created":
		target = &FieldDevicesCreated{}
	case "bacnet_object_created":
		target = &BacnetObjectCreated{}
	case "bacnet_object_updated":
		target = &BacnetObjectUpdated{}
	case "sps_controller_created":
		target = &SPSControllerCreated{}
	case "sps_controller_cloned":
		target = &SPSControllerCloned{}
	case "sps_controller_system_type_cloned":
		target = &SPSControllerSystemTypeCloned{}
	case "sps_controller_updated":
		target = &SPSControllerUpdated{}
	case "sps_controller_moved":
		target = &SPSControllerMoved{}
	case "sps_controller_deleted":
		target = &SPSControllerDeleted{}
	default:
		return nil, fmt.Errorf("unsupported collaboration command type %q", encoded.Type)
	}
	if err := json.Unmarshal(encoded.Payload, target); err != nil {
		return nil, fmt.Errorf("unmarshal collaboration command %s: %w", encoded.Type, err)
	}
	// Command is intentionally sealed. Return value types so Dispatcher retains
	// its existing value-type switch contract.
	switch target.(type) {
	case *FacilityHierarchyRefreshRequired:
		return *(target.(*FacilityHierarchyRefreshRequired)), nil
	case *ControlCabinetCreated:
		return *(target.(*ControlCabinetCreated)), nil
	case *ControlCabinetCloned:
		return *(target.(*ControlCabinetCloned)), nil
	case *ControlCabinetDeleted:
		return *(target.(*ControlCabinetDeleted)), nil
	case *ControlCabinetUpdated:
		return *(target.(*ControlCabinetUpdated)), nil
	case *ControlCabinetMoved:
		return *(target.(*ControlCabinetMoved)), nil
	case *FieldDeviceUpdated:
		return *(target.(*FieldDeviceUpdated)), nil
	case *FieldDeviceMoved:
		return *(target.(*FieldDeviceMoved)), nil
	case *FieldDeviceDeleted:
		return *(target.(*FieldDeviceDeleted)), nil
	case *FieldDevicesCreated:
		return *(target.(*FieldDevicesCreated)), nil
	case *BacnetObjectCreated:
		return *(target.(*BacnetObjectCreated)), nil
	case *BacnetObjectUpdated:
		return *(target.(*BacnetObjectUpdated)), nil
	case *SPSControllerCreated:
		return *(target.(*SPSControllerCreated)), nil
	case *SPSControllerCloned:
		return *(target.(*SPSControllerCloned)), nil
	case *SPSControllerSystemTypeCloned:
		return *(target.(*SPSControllerSystemTypeCloned)), nil
	case *SPSControllerUpdated:
		return *(target.(*SPSControllerUpdated)), nil
	case *SPSControllerMoved:
		return *(target.(*SPSControllerMoved)), nil
	case *SPSControllerDeleted:
		return *(target.(*SPSControllerDeleted)), nil
	default:
		panic("command type switch is exhaustive")
	}
}

func commandTypeOf(command Command) (string, bool) {
	switch command.(type) {
	case FacilityHierarchyRefreshRequired:
		return "facility_hierarchy_refresh_required", true
	case ControlCabinetCreated:
		return "control_cabinet_created", true
	case ControlCabinetCloned:
		return "control_cabinet_cloned", true
	case ControlCabinetDeleted:
		return "control_cabinet_deleted", true
	case ControlCabinetUpdated:
		return "control_cabinet_updated", true
	case ControlCabinetMoved:
		return "control_cabinet_moved", true
	case FieldDeviceUpdated:
		return "field_device_updated", true
	case FieldDeviceMoved:
		return "field_device_moved", true
	case FieldDeviceDeleted:
		return "field_device_deleted", true
	case FieldDevicesCreated:
		return "field_devices_created", true
	case BacnetObjectCreated:
		return "bacnet_object_created", true
	case BacnetObjectUpdated:
		return "bacnet_object_updated", true
	case SPSControllerCreated:
		return "sps_controller_created", true
	case SPSControllerCloned:
		return "sps_controller_cloned", true
	case SPSControllerSystemTypeCloned:
		return "sps_controller_system_type_cloned", true
	case SPSControllerUpdated:
		return "sps_controller_updated", true
	case SPSControllerMoved:
		return "sps_controller_moved", true
	case SPSControllerDeleted:
		return "sps_controller_deleted", true
	default:
		return "", false
	}
}

func commandEnvelope(command Command) (Envelope, bool) {
	switch value := command.(type) {
	case FacilityHierarchyRefreshRequired:
		return value.Envelope, true
	case ControlCabinetCreated:
		return value.Envelope, true
	case ControlCabinetCloned:
		return value.Envelope, true
	case ControlCabinetDeleted:
		return value.Envelope, true
	case ControlCabinetUpdated:
		return value.Envelope, true
	case ControlCabinetMoved:
		return value.Envelope, true
	case FieldDeviceUpdated:
		return value.Envelope, true
	case FieldDeviceMoved:
		return value.Envelope, true
	case FieldDeviceDeleted:
		return value.Envelope, true
	case FieldDevicesCreated:
		return value.Envelope, true
	case BacnetObjectCreated:
		return value.Envelope, true
	case BacnetObjectUpdated:
		return value.Envelope, true
	case SPSControllerCreated:
		return value.Envelope, true
	case SPSControllerCloned:
		return value.Envelope, true
	case SPSControllerSystemTypeCloned:
		return value.Envelope, true
	case SPSControllerUpdated:
		return value.Envelope, true
	case SPSControllerMoved:
		return value.Envelope, true
	case SPSControllerDeleted:
		return value.Envelope, true
	default:
		return Envelope{}, false
	}
}
