package bacnetobject

import (
	"context"
	"testing"
	"time"

	appcollaboration "github.com/besart951/go_infra_link/backend/internal/application/collaboration"
	domainCollaboration "github.com/besart951/go_infra_link/backend/internal/domain/collaboration"
	domainObjectData "github.com/besart951/go_infra_link/backend/internal/domain/facility/objectdata"
	domainProject "github.com/besart951/go_infra_link/backend/internal/domain/project"
	"github.com/google/uuid"
)

type collaborationOutboxResolverStub struct {
	links  []*domainProject.ProjectFieldDevice
	owners []domainObjectData.BacnetObjectOwner
}

func (s *collaborationOutboxResolverStub) GetByFieldDeviceIDs(
	context.Context,
	[]uuid.UUID,
) ([]*domainProject.ProjectFieldDevice, error) {
	return s.links, nil
}

func (s *collaborationOutboxResolverStub) GetByBacnetObjectIDs(
	context.Context,
	[]uuid.UUID,
) ([]domainObjectData.BacnetObjectOwner, error) {
	return s.owners, nil
}

type collaborationOutboxStoreStub struct {
	events []*domainCollaboration.OutboxEvent
}

func (s *collaborationOutboxStoreStub) Enqueue(
	_ context.Context,
	event *domainCollaboration.OutboxEvent,
) error {
	s.events = append(s.events, event)
	return nil
}

func (*collaborationOutboxStoreStub) ClaimDue(
	context.Context,
	time.Time,
	int,
) ([]domainCollaboration.OutboxEvent, error) {
	return nil, nil
}

func (*collaborationOutboxStoreStub) WasProcessed(
	context.Context,
	string,
	uuid.UUID,
) (bool, error) {
	return false, nil
}

func (*collaborationOutboxStoreStub) MarkDelivered(
	context.Context,
	string,
	domainCollaboration.OutboxEvent,
	time.Time,
) error {
	return nil
}

func (*collaborationOutboxStoreStub) MarkFailed(
	context.Context,
	domainCollaboration.OutboxEvent,
	string,
	time.Time,
	time.Time,
) error {
	return nil
}

func TestTransactionalBACnetOutboxUsesNarrowScopeAndMixedOwnerFallback(t *testing.T) {
	bacnetObjectID := bacnetTestUUID(201)
	fieldDeviceID := bacnetTestUUID(202)
	objectDataID := bacnetTestUUID(203)
	directProjectID := bacnetTestUUID(211)
	objectDataProjectID := bacnetTestUUID(212)
	mixedProjectID := bacnetTestUUID(213)
	operationID := bacnetTestUUID(220)
	eventIDs := []uuid.UUID{
		bacnetTestUUID(221),
		bacnetTestUUID(222),
		bacnetTestUUID(223),
	}
	resolver := &collaborationOutboxResolverStub{
		links: []*domainProject.ProjectFieldDevice{
			{ProjectID: directProjectID, FieldDeviceID: fieldDeviceID},
			{ProjectID: mixedProjectID, FieldDeviceID: fieldDeviceID},
		},
		owners: []domainObjectData.BacnetObjectOwner{
			{BacnetObjectID: bacnetObjectID, ObjectDataID: objectDataID, ProjectID: &objectDataProjectID},
			{BacnetObjectID: bacnetObjectID, ObjectDataID: objectDataID, ProjectID: &mixedProjectID},
		},
	}
	store := &collaborationOutboxStoreStub{}
	ctx := domainCollaboration.WithOutboxStore(context.Background(), store)
	nextID := func() uuid.UUID {
		id := eventIDs[0]
		eventIDs = eventIDs[1:]
		return id
	}

	projectIDs, err := enqueueTransactionalMutation(
		ctx,
		resolver,
		bacnetObjectID,
		0,
		[]uuid.UUID{fieldDeviceID},
		operationID,
		nil,
		time.Date(2026, time.July, 23, 16, 30, 0, 0, time.UTC),
		nextID,
	)
	if err != nil {
		t.Fatalf("enqueue mutation: %v", err)
	}
	wantProjects := []uuid.UUID{directProjectID, objectDataProjectID, mixedProjectID}
	if len(projectIDs) != len(wantProjects) {
		t.Fatalf("projects: got %v, want %v", projectIDs, wantProjects)
	}
	for i := range wantProjects {
		if projectIDs[i] != wantProjects[i] {
			t.Fatalf("projects: got %v, want %v", projectIDs, wantProjects)
		}
	}
	if len(store.events) != 3 {
		t.Fatalf("outbox events: got %d, want 3", len(store.events))
	}

	decoded := make(map[uuid.UUID]appcollaboration.Command, len(store.events))
	for _, event := range store.events {
		command, err := appcollaboration.DecodeCommand(appcollaboration.EncodedCommand{
			Type: event.EventType, Payload: event.Payload,
		})
		if err != nil {
			t.Fatalf("decode event: %v", err)
		}
		envelope, _ := appcollaboration.CommandEnvelope(command)
		decoded[envelope.ProjectID] = command
	}
	if _, ok := decoded[directProjectID].(appcollaboration.BacnetObjectUpdated); !ok {
		t.Fatalf("direct project command: %T", decoded[directProjectID])
	}
	objectDataCommand, ok := decoded[objectDataProjectID].(appcollaboration.FacilityHierarchyRefreshRequired)
	if !ok || objectDataCommand.Scope != appcollaboration.FacilityScopeObjectData ||
		len(objectDataCommand.EntityIDs) != 1 || objectDataCommand.EntityIDs[0] != objectDataID {
		t.Fatalf("ObjectData project command: %#v", decoded[objectDataProjectID])
	}
	mixedCommand, ok := decoded[mixedProjectID].(appcollaboration.FacilityHierarchyRefreshRequired)
	if !ok || mixedCommand.Scope != appcollaboration.FacilityScopeProject || !mixedCommand.FullRefresh {
		t.Fatalf("mixed-owner project command: %#v", decoded[mixedProjectID])
	}
}
