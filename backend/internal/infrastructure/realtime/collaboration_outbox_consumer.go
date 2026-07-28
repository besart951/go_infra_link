package realtime

import (
	"context"
	"fmt"

	appcollaboration "github.com/besart951/go_infra_link/backend/internal/application/collaboration"
	domainCollaboration "github.com/besart951/go_infra_link/backend/internal/domain/collaboration"
	"github.com/google/uuid"
)

const collaborationOutboxConsumerID = "project-websocket-v2"

type CollaborationOutboxConsumer struct {
	hub *ProjectCollaborationHub
}

func NewCollaborationOutboxConsumer(hub *ProjectCollaborationHub) *CollaborationOutboxConsumer {
	return &CollaborationOutboxConsumer{hub: hub}
}

func (*CollaborationOutboxConsumer) ConsumerID() string {
	return collaborationOutboxConsumerID
}

func (c *CollaborationOutboxConsumer) Deliver(
	_ context.Context,
	event domainCollaboration.OutboxEvent,
) error {
	if c == nil || c.hub == nil {
		return fmt.Errorf("project collaboration hub is not configured")
	}
	command, err := appcollaboration.DecodeCommand(appcollaboration.EncodedCommand{
		Type:    event.EventType,
		Payload: event.Payload,
	})
	if err != nil {
		return err
	}
	envelope, ok := appcollaboration.CommandEnvelope(command)
	if !ok || envelope.EventID != event.EventID ||
		envelope.OperationID != event.OperationID ||
		envelope.ProjectID != event.ProjectID {
		return fmt.Errorf("collaboration outbox envelope does not match persisted event")
	}
	scope, entityIDs, err := v2CommandRefresh(command)
	if err != nil {
		return err
	}
	actorID := ""
	if envelope.ActorID != nil {
		actorID = envelope.ActorID.String()
	}
	return c.hub.BroadcastCommittedEvent(ProjectCollaborationCommittedEventMessage{
		SchemaVersion:   event.SchemaVersion,
		EventID:         event.EventID,
		OperationID:     event.OperationID,
		CorrelationID:   envelope.CorrelationID,
		ProjectID:       event.ProjectID,
		ActorID:         actorID,
		Sequence:        event.Sequence,
		EventType:       event.EventType,
		Scope:           string(scope),
		EntityIDs:       uuidStrings(entityIDs),
		EntityRevisions: cloneEntityRevisions(envelope.EntityRevisions),
		OccurredAt:      envelope.OccurredAt,
	})
}

func cloneEntityRevisions(source map[string]uint64) map[string]uint64 {
	if len(source) == 0 {
		return nil
	}
	cloned := make(map[string]uint64, len(source))
	for id, revision := range source {
		if id != "" && revision != 0 {
			cloned[id] = revision
		}
	}
	if len(cloned) == 0 {
		return nil
	}
	return cloned
}

func v2CommandRefresh(command appcollaboration.Command) (
	appcollaboration.FacilityScope,
	[]uuid.UUID,
	error,
) {
	switch value := command.(type) {
	case appcollaboration.FacilityHierarchyRefreshRequired:
		return value.Scope, value.EntityIDs, nil
	case appcollaboration.FieldDeviceUpdated:
		return appcollaboration.FacilityScopeFieldDevice, []uuid.UUID{value.FieldDeviceID}, nil
	case appcollaboration.FieldDeviceMoved:
		return appcollaboration.FacilityScopeFieldDevice, []uuid.UUID{value.FieldDeviceID}, nil
	case appcollaboration.FieldDeviceDeleted:
		return appcollaboration.FacilityScopeFieldDevice, []uuid.UUID{value.FieldDeviceID}, nil
	case appcollaboration.FieldDevicesCreated:
		ids := make([]uuid.UUID, 0, len(value.FieldDevices))
		for _, item := range value.FieldDevices {
			ids = append(ids, item.ID)
		}
		return appcollaboration.FacilityScopeFieldDevice, ids, nil
	case appcollaboration.BacnetObjectCreated:
		return appcollaboration.FacilityScopeFieldDevice, []uuid.UUID{value.FieldDeviceID}, nil
	case appcollaboration.BacnetObjectUpdated:
		return appcollaboration.FacilityScopeFieldDevice, value.FieldDeviceIDs, nil
	case appcollaboration.SPSControllerCreated:
		return appcollaboration.FacilityScopeSPSController, []uuid.UUID{value.SPSController.ID}, nil
	case appcollaboration.SPSControllerCloned:
		return appcollaboration.FacilityScopeSPSController, []uuid.UUID{value.SPSController.ID}, nil
	case appcollaboration.SPSControllerSystemTypeCloned:
		return appcollaboration.FacilityScopeSPSController, []uuid.UUID{value.SPSControllerID}, nil
	case appcollaboration.SPSControllerUpdated:
		return appcollaboration.FacilityScopeSPSController, []uuid.UUID{value.SPSControllerID}, nil
	case appcollaboration.SPSControllerMoved:
		return appcollaboration.FacilityScopeSPSController, []uuid.UUID{value.SPSControllerID}, nil
	case appcollaboration.SPSControllerDeleted:
		return appcollaboration.FacilityScopeSPSController, []uuid.UUID{value.SPSControllerID}, nil
	case appcollaboration.ControlCabinetCreated:
		return appcollaboration.FacilityScopeControlCabinet, []uuid.UUID{value.ControlCabinet.ID}, nil
	case appcollaboration.ControlCabinetCloned:
		return appcollaboration.FacilityScopeControlCabinet, []uuid.UUID{value.ControlCabinet.ID}, nil
	case appcollaboration.ControlCabinetUpdated:
		return appcollaboration.FacilityScopeControlCabinet, []uuid.UUID{value.ControlCabinet.ID}, nil
	case appcollaboration.ControlCabinetMoved:
		return appcollaboration.FacilityScopeControlCabinet, []uuid.UUID{value.ControlCabinet.ID}, nil
	case appcollaboration.ControlCabinetDeleted:
		return appcollaboration.FacilityScopeControlCabinet, []uuid.UUID{value.ControlCabinetID}, nil
	default:
		return "", nil, fmt.Errorf("unsupported version-2 collaboration command %T", command)
	}
}

func uuidStrings(ids []uuid.UUID) []string {
	result := make([]string, 0, len(ids))
	for _, id := range ids {
		if id != uuid.Nil {
			result = append(result, id.String())
		}
	}
	return result
}
