package realtime

import (
	"context"
	"testing"
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
