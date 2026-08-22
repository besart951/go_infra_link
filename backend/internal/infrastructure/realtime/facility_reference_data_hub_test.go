package realtime

import (
	"context"
	"testing"

	apprealtime "github.com/besart951/go_infra_link/backend/internal/application/realtime"
	"github.com/google/uuid"
)

func TestFacilityReferenceDataRoutesAcrossInstances(t *testing.T) {
	bus := NewInMemoryBus()
	defer bus.Close()

	hubA := NewFacilityReferenceDataHub(WithFacilityReferenceDataBus(bus, "node-a"))
	defer hubA.Close()
	hubB := NewFacilityReferenceDataHub(WithFacilityReferenceDataBus(bus, "node-b"))
	defer hubB.Close()

	client := &facilityReferenceDataClient{hub: hubB, socket: newTestSocket(4)}
	hubB.register(client)

	hubA.BroadcastFacilityReferenceDataChange(
		context.Background(),
		FacilityReferenceDataResourceSystemParts,
		FacilityReferenceDataResourceApparats,
	)

	message := receiveSocketMessageOfType(t, client.socket, facilityReferenceDataEventChanged)
	resources, ok := message["resources"].([]any)
	if !ok {
		t.Fatalf("resources = %T, want []any", message["resources"])
	}
	if len(resources) != 2 || resources[0] != FacilityReferenceDataResourceApparats || resources[1] != FacilityReferenceDataResourceSystemParts {
		t.Fatalf("resources = %#v, want apparats and system_parts", resources)
	}
}

func TestFacilityReferenceDataDoesNotDuplicateOwnBusEventToLocalClient(t *testing.T) {
	bus := NewInMemoryBus()
	defer bus.Close()

	hub := NewFacilityReferenceDataHub(WithFacilityReferenceDataBus(bus, "node-a"))
	defer hub.Close()

	client := &facilityReferenceDataClient{hub: hub, socket: newTestSocket(4)}
	hub.register(client)

	hub.BroadcastFacilityReferenceDataChange(context.Background(), FacilityReferenceDataResourceApparats)

	_ = receiveSocketMessageOfType(t, client.socket, facilityReferenceDataEventChanged)
	assertNoSocketMessageOfType(t, client.socket, facilityReferenceDataEventChanged)
}

func TestFacilityJobProgressIsOnlyDeliveredToTheOwningUser(t *testing.T) {
	bus := NewInMemoryBus()
	defer bus.Close()

	hubA := NewFacilityReferenceDataHub(WithFacilityReferenceDataBus(bus, "node-a"))
	defer hubA.Close()
	hubB := NewFacilityReferenceDataHub(WithFacilityReferenceDataBus(bus, "node-b"))
	defer hubB.Close()

	ownerID := uuid.New()
	ownerClient := &facilityReferenceDataClient{hub: hubB, userID: ownerID, socket: newTestSocket(4)}
	otherClient := &facilityReferenceDataClient{hub: hubB, userID: uuid.New(), socket: newTestSocket(4)}
	hubB.register(ownerClient)
	hubB.register(otherClient)

	jobID := uuid.New()
	hubA.BroadcastFacilityJobProgress(context.Background(), apprealtime.FacilityJobProgressEvent{
		JobID: jobID, OwnerID: ownerID, Kind: "control_cabinet", Status: "running",
		Progress: 50, Stage: "copying_controllers",
	})

	message := receiveSocketMessageOfType(t, ownerClient.socket, facilityJobProgressEvent)
	if message["job_id"] != jobID.String() || message["progress"] != float64(50) {
		t.Fatalf("job message = %#v, want job ID %s and progress 50", message, jobID)
	}
	assertNoSocketMessageOfType(t, otherClient.socket, facilityJobProgressEvent)
}

func TestFacilityChangeIsFilteredByReadableResource(t *testing.T) {
	hub := NewFacilityReferenceDataHub()
	defer hub.Close()

	allowedClient := &facilityReferenceDataClient{
		hub: hub, readableResources: map[string]struct{}{"apparats": {}}, socket: newTestSocket(4),
	}
	deniedClient := &facilityReferenceDataClient{
		hub: hub, readableResources: map[string]struct{}{"system_parts": {}}, socket: newTestSocket(4),
	}
	hub.register(allowedClient)
	hub.register(deniedClient)

	id := uuid.New()
	hub.BroadcastFacilityChange(context.Background(), "apparats", "updated", []uuid.UUID{id, id}, nil)

	message := receiveSocketMessageOfType(t, allowedClient.socket, facilityChangedEvent)
	if message["resource"] != "apparats" || message["action"] != "updated" {
		t.Fatalf("facility event = %#v", message)
	}
	ids, ok := message["ids"].([]any)
	if !ok || len(ids) != 1 || ids[0] != id.String() {
		t.Fatalf("IDs = %#v, want one %s", message["ids"], id)
	}
	assertNoSocketMessageOfType(t, deniedClient.socket, facilityChangedEvent)
}

func TestFacilityReferenceDataChangeFiltersIndividualResources(t *testing.T) {
	hub := NewFacilityReferenceDataHub()
	defer hub.Close()

	client := &facilityReferenceDataClient{
		hub: hub, readableResources: map[string]struct{}{"apparats": {}}, socket: newTestSocket(4),
	}
	hub.register(client)
	hub.BroadcastFacilityReferenceDataChange(context.Background(), FacilityReferenceDataResourceApparats, FacilityReferenceDataResourceSystemParts)

	message := receiveSocketMessageOfType(t, client.socket, facilityReferenceDataEventChanged)
	resources, ok := message["resources"].([]any)
	if !ok || len(resources) != 1 || resources[0] != FacilityReferenceDataResourceApparats {
		t.Fatalf("resources = %#v, want only apparats", message["resources"])
	}
}
