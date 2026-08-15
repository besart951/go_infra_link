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

func TestFacilityCopyJobProgressIsOnlyDeliveredToTheOwningUser(t *testing.T) {
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
	hubA.BroadcastCopyJobProgress(context.Background(), apprealtime.CopyJobProgressEvent{
		JobID: jobID, OwnerID: ownerID, Kind: "control_cabinet", Status: "running",
		Progress: 50, Stage: "copying_controllers",
	})

	message := receiveSocketMessageOfType(t, ownerClient.socket, facilityCopyJobProgressEvent)
	if message["job_id"] != jobID.String() || message["progress"] != float64(50) {
		t.Fatalf("copy job message = %#v, want job ID %s and progress 50", message, jobID)
	}
	assertNoSocketMessageOfType(t, otherClient.socket, facilityCopyJobProgressEvent)
}
