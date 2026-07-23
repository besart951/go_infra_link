package realtime

import (
	"context"
	"reflect"
	"strings"
	"testing"
	"time"

	appcollaboration "github.com/besart951/go_infra_link/backend/internal/application/collaboration"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
)

type collaborationRefreshPublisherSpy struct {
	projectID    uuid.UUID
	actorID      *uuid.UUID
	scope        string
	entityIDs    []string
	cabinet      *domainFacility.ControlCabinet
	controller   *domainFacility.SPSController
	fieldDevices []map[string]any
}

func (s *collaborationRefreshPublisherSpy) BroadcastFieldDeviceDelta(
	projectID uuid.UUID,
	actorID *uuid.UUID,
	fieldDevices []map[string]any,
) {
	s.projectID = projectID
	s.actorID = actorID
	s.fieldDevices = cloneFieldDeviceDeltas(fieldDevices)
}

func (s *collaborationRefreshPublisherSpy) BroadcastSPSControllerDelta(
	projectID uuid.UUID,
	actorID *uuid.UUID,
	spsController domainFacility.SPSController,
) {
	s.projectID = projectID
	s.actorID = actorID
	clone := spsController
	clone.GADevice = clonePointer(spsController.GADevice)
	clone.DeviceDescription = clonePointer(spsController.DeviceDescription)
	clone.DeviceLocation = clonePointer(spsController.DeviceLocation)
	clone.IPAddress = clonePointer(spsController.IPAddress)
	clone.Subnet = clonePointer(spsController.Subnet)
	clone.Gateway = clonePointer(spsController.Gateway)
	clone.Vlan = clonePointer(spsController.Vlan)
	s.controller = &clone
}

func (s *collaborationRefreshPublisherSpy) BroadcastControlCabinetDelta(
	projectID uuid.UUID,
	actorID *uuid.UUID,
	controlCabinet domainFacility.ControlCabinet,
) {
	s.projectID = projectID
	s.actorID = actorID
	clone := controlCabinet
	if controlCabinet.ControlCabinetNr != nil {
		value := *controlCabinet.ControlCabinetNr
		clone.ControlCabinetNr = &value
	}
	s.cabinet = &clone
}

func (s *collaborationRefreshPublisherSpy) BroadcastRefreshRequest(
	projectID uuid.UUID,
	actorID *uuid.UUID,
	scope string,
	entityIDs []string,
) {
	s.projectID = projectID
	s.actorID = actorID
	s.scope = scope
	s.entityIDs = append([]string(nil), entityIDs...)
}

func TestCollaborationCommandAdapterMapsTypedRefreshToVersionOnePublisher(t *testing.T) {
	publisher := &collaborationRefreshPublisherSpy{}
	adapter := NewCollaborationCommandAdapter(publisher)
	projectID := uuid.New()
	actorID := uuid.New()
	entityOne := uuid.New()
	entityTwo := uuid.New()

	err := adapter.PublishFacilityHierarchyRefresh(context.Background(),
		appcollaboration.FacilityHierarchyRefreshRequired{
			Envelope: appcollaboration.Envelope{
				ProjectID: projectID,
				ActorID:   &actorID,
			},
			Scope:     appcollaboration.FacilityScopeFieldDevice,
			EntityIDs: []uuid.UUID{entityOne, entityTwo},
		},
	)
	if err != nil {
		t.Fatalf("PublishFacilityHierarchyRefresh returned error: %v", err)
	}

	if publisher.projectID != projectID {
		t.Fatalf("expected project %s, got %s", projectID, publisher.projectID)
	}
	if publisher.actorID == nil || *publisher.actorID != actorID {
		t.Fatalf("expected actor %s, got %v", actorID, publisher.actorID)
	}
	if publisher.scope != "field_device" {
		t.Fatalf("expected field_device scope, got %q", publisher.scope)
	}
	if want := []string{entityOne.String(), entityTwo.String()}; !reflect.DeepEqual(publisher.entityIDs, want) {
		t.Fatalf("entity IDs: got %v, want %v", publisher.entityIDs, want)
	}
}

func TestCollaborationCommandAdapterOmitsIDsForFullRefresh(t *testing.T) {
	publisher := &collaborationRefreshPublisherSpy{}
	adapter := NewCollaborationCommandAdapter(publisher)

	err := adapter.PublishFacilityHierarchyRefresh(context.Background(),
		appcollaboration.FacilityHierarchyRefreshRequired{
			Envelope:    appcollaboration.Envelope{ProjectID: uuid.New()},
			Scope:       appcollaboration.FacilityScopeProject,
			EntityIDs:   []uuid.UUID{uuid.New()},
			FullRefresh: true,
		},
	)
	if err != nil {
		t.Fatalf("PublishFacilityHierarchyRefresh returned error: %v", err)
	}
	if len(publisher.entityIDs) != 0 {
		t.Fatalf("expected no IDs for full refresh, got %v", publisher.entityIDs)
	}
	if publisher.scope != "project" {
		t.Fatalf("expected project scope, got %q", publisher.scope)
	}
}

func TestCollaborationCommandAdapterMapsControlCabinetUpdatedToVersionOneDelta(t *testing.T) {
	publisher := &collaborationRefreshPublisherSpy{}
	adapter := NewCollaborationCommandAdapter(publisher)
	projectID := uuid.New()
	actorID := uuid.New()
	cabinetID := uuid.New()
	buildingID := uuid.New()
	number := "AK01"
	createdAt := time.Date(2026, time.July, 20, 20, 0, 0, 0, time.UTC)
	updatedAt := createdAt.Add(time.Minute)

	err := adapter.PublishControlCabinetUpdated(
		context.Background(),
		appcollaboration.ControlCabinetUpdated{
			Envelope: appcollaboration.Envelope{
				ProjectID: projectID,
				ActorID:   &actorID,
			},
			ControlCabinet: appcollaboration.ControlCabinetState{
				ID:               cabinetID,
				BuildingID:       buildingID,
				ControlCabinetNr: &number,
				CreatedAt:        createdAt,
				UpdatedAt:        updatedAt,
			},
		},
	)
	if err != nil {
		t.Fatalf("PublishControlCabinetUpdated returned error: %v", err)
	}
	if publisher.projectID != projectID ||
		publisher.actorID == nil || *publisher.actorID != actorID ||
		publisher.cabinet == nil || publisher.cabinet.ID != cabinetID ||
		publisher.cabinet.BuildingID != buildingID ||
		publisher.cabinet.ControlCabinetNr == nil ||
		*publisher.cabinet.ControlCabinetNr != number ||
		!publisher.cabinet.CreatedAt.Equal(createdAt) ||
		!publisher.cabinet.UpdatedAt.Equal(updatedAt) {
		t.Fatalf("unexpected v1 cabinet delta: %+v", publisher)
	}
}

func TestCollaborationCommandAdapterMapsControlCabinetCreatedToVersionOneDelta(t *testing.T) {
	publisher := &collaborationRefreshPublisherSpy{}
	adapter := NewCollaborationCommandAdapter(publisher)
	projectID := uuid.New()
	actorID := uuid.New()
	cabinetID := uuid.New()
	buildingID := uuid.New()
	number := "AK01"

	err := adapter.PublishControlCabinetCreated(
		context.Background(),
		appcollaboration.ControlCabinetCreated{
			Envelope: appcollaboration.Envelope{
				ProjectID: projectID,
				ActorID:   &actorID,
			},
			ControlCabinet: appcollaboration.ControlCabinetState{
				ID:               cabinetID,
				BuildingID:       buildingID,
				ControlCabinetNr: &number,
			},
		},
	)
	if err != nil {
		t.Fatalf("PublishControlCabinetCreated returned error: %v", err)
	}
	if publisher.projectID != projectID ||
		publisher.actorID == nil || *publisher.actorID != actorID ||
		publisher.cabinet == nil || publisher.cabinet.ID != cabinetID ||
		publisher.cabinet.BuildingID != buildingID ||
		publisher.cabinet.ControlCabinetNr == nil ||
		*publisher.cabinet.ControlCabinetNr != number {
		t.Fatalf("unexpected v1 cabinet delta: %+v", publisher)
	}
}

func TestCollaborationCommandAdapterMapsControlCabinetClonedToVersionOneDelta(t *testing.T) {
	publisher := &collaborationRefreshPublisherSpy{}
	adapter := NewCollaborationCommandAdapter(publisher)
	projectID := uuid.New()
	actorID := uuid.New()
	sourceID := uuid.New()
	cabinetID := uuid.New()
	buildingID := uuid.New()
	number := "AK02"

	err := adapter.PublishControlCabinetCloned(
		context.Background(),
		appcollaboration.ControlCabinetCloned{
			Envelope: appcollaboration.Envelope{
				ProjectID: projectID,
				ActorID:   &actorID,
			},
			SourceControlCabinetID: sourceID,
			ControlCabinet: appcollaboration.ControlCabinetState{
				ID:               cabinetID,
				BuildingID:       buildingID,
				ControlCabinetNr: &number,
			},
		},
	)
	if err != nil {
		t.Fatalf("PublishControlCabinetCloned returned error: %v", err)
	}
	if publisher.projectID != projectID ||
		publisher.actorID == nil || *publisher.actorID != actorID ||
		publisher.cabinet == nil || publisher.cabinet.ID != cabinetID ||
		publisher.cabinet.BuildingID != buildingID ||
		publisher.cabinet.ControlCabinetNr == nil ||
		*publisher.cabinet.ControlCabinetNr != number {
		t.Fatalf("unexpected v1 cabinet delta: %+v", publisher)
	}
}

func TestCollaborationCommandAdapterMapsControlCabinetDeletedToTargetedVersionOneRefresh(t *testing.T) {
	publisher := &collaborationRefreshPublisherSpy{}
	adapter := NewCollaborationCommandAdapter(publisher)
	projectID := uuid.New()
	actorID := uuid.New()
	cabinetID := uuid.New()

	err := adapter.PublishControlCabinetDeleted(
		context.Background(),
		appcollaboration.ControlCabinetDeleted{
			Envelope: appcollaboration.Envelope{
				ProjectID: projectID,
				ActorID:   &actorID,
			},
			ControlCabinetID: cabinetID,
			BuildingID:       uuid.New(),
		},
	)
	if err != nil {
		t.Fatalf("PublishControlCabinetDeleted returned error: %v", err)
	}
	if publisher.projectID != projectID ||
		publisher.actorID == nil || *publisher.actorID != actorID ||
		publisher.scope != "control_cabinet" ||
		!reflect.DeepEqual(publisher.entityIDs, []string{cabinetID.String()}) {
		t.Fatalf("unexpected v1 cabinet refresh: %+v", publisher)
	}
}

func TestCollaborationCommandAdapterMapsControlCabinetMovedToVersionOneDelta(t *testing.T) {
	publisher := &collaborationRefreshPublisherSpy{}
	adapter := NewCollaborationCommandAdapter(publisher)
	cabinetID := uuid.New()

	err := adapter.PublishControlCabinetMoved(
		context.Background(),
		appcollaboration.ControlCabinetMoved{
			Envelope:       appcollaboration.Envelope{ProjectID: uuid.New()},
			ControlCabinet: appcollaboration.ControlCabinetState{ID: cabinetID},
			FromBuildingID: uuid.New(),
			ToBuildingID:   uuid.New(),
		},
	)
	if err != nil {
		t.Fatalf("PublishControlCabinetMoved returned error: %v", err)
	}
	if publisher.cabinet == nil || publisher.cabinet.ID != cabinetID {
		t.Fatalf("unexpected v1 cabinet delta: %+v", publisher.cabinet)
	}
}

func TestCollaborationCommandAdapterMapsFieldDeviceUpdatedToTargetedVersionOneRefresh(t *testing.T) {
	publisher := &collaborationRefreshPublisherSpy{}
	adapter := NewCollaborationCommandAdapter(publisher)
	projectID := uuid.New()
	actorID := uuid.New()
	fieldDeviceID := uuid.New()

	err := adapter.PublishFieldDeviceUpdated(
		context.Background(),
		appcollaboration.FieldDeviceUpdated{
			Envelope: appcollaboration.Envelope{
				ProjectID: projectID,
				ActorID:   &actorID,
			},
			FieldDeviceID: fieldDeviceID,
		},
	)
	if err != nil {
		t.Fatalf("PublishFieldDeviceUpdated returned error: %v", err)
	}

	if publisher.projectID != projectID {
		t.Fatalf("expected project %s, got %s", projectID, publisher.projectID)
	}
	if publisher.actorID == nil || *publisher.actorID != actorID {
		t.Fatalf("expected actor %s, got %v", actorID, publisher.actorID)
	}
	if publisher.scope != "field_device" {
		t.Fatalf("expected field_device scope, got %q", publisher.scope)
	}
	if want := []string{fieldDeviceID.String()}; !reflect.DeepEqual(publisher.entityIDs, want) {
		t.Fatalf("entity IDs: got %v, want %v", publisher.entityIDs, want)
	}
}

func TestCollaborationCommandAdapterMapsFieldDeviceMovedToTargetedVersionOneRefresh(t *testing.T) {
	publisher := &collaborationRefreshPublisherSpy{}
	adapter := NewCollaborationCommandAdapter(publisher)
	projectID := uuid.New()
	actorID := uuid.New()
	fieldDeviceID := uuid.New()

	err := adapter.PublishFieldDeviceMoved(
		context.Background(),
		appcollaboration.FieldDeviceMoved{
			Envelope: appcollaboration.Envelope{
				ProjectID: projectID,
				ActorID:   &actorID,
			},
			FieldDeviceID:                 fieldDeviceID,
			FromSPSControllerSystemTypeID: uuid.New(),
			ToSPSControllerSystemTypeID:   uuid.New(),
		},
	)
	if err != nil {
		t.Fatalf("PublishFieldDeviceMoved returned error: %v", err)
	}

	if publisher.projectID != projectID {
		t.Fatalf("expected project %s, got %s", projectID, publisher.projectID)
	}
	if publisher.actorID == nil || *publisher.actorID != actorID {
		t.Fatalf("expected actor %s, got %v", actorID, publisher.actorID)
	}
	if publisher.scope != "field_device" {
		t.Fatalf("expected field_device scope, got %q", publisher.scope)
	}
	if want := []string{fieldDeviceID.String()}; !reflect.DeepEqual(publisher.entityIDs, want) {
		t.Fatalf("entity IDs: got %v, want %v", publisher.entityIDs, want)
	}
}

func TestCollaborationCommandAdapterMapsFieldDeviceDeletedToTargetedVersionOneRefresh(t *testing.T) {
	publisher := &collaborationRefreshPublisherSpy{}
	adapter := NewCollaborationCommandAdapter(publisher)
	projectID := uuid.New()
	actorID := uuid.New()
	fieldDeviceID := uuid.New()

	err := adapter.PublishFieldDeviceDeleted(
		context.Background(),
		appcollaboration.FieldDeviceDeleted{
			Envelope: appcollaboration.Envelope{
				ProjectID: projectID,
				ActorID:   &actorID,
			},
			FieldDeviceID:             fieldDeviceID,
			SPSControllerSystemTypeID: uuid.New(),
		},
	)
	if err != nil {
		t.Fatalf("PublishFieldDeviceDeleted returned error: %v", err)
	}
	if publisher.projectID != projectID ||
		publisher.actorID == nil || *publisher.actorID != actorID ||
		publisher.scope != "field_device" {
		t.Fatalf("unexpected v1 refresh envelope: %+v", publisher)
	}
	if want := []string{fieldDeviceID.String()}; !reflect.DeepEqual(publisher.entityIDs, want) {
		t.Fatalf("entity IDs: got %v, want %v", publisher.entityIDs, want)
	}
}

func TestCollaborationCommandAdapterMapsBacnetObjectUpdatedToParentFieldDeviceRefresh(t *testing.T) {
	publisher := &collaborationRefreshPublisherSpy{}
	adapter := NewCollaborationCommandAdapter(publisher)
	projectID := uuid.New()
	actorID := uuid.New()
	fieldDeviceOne := uuid.New()
	fieldDeviceTwo := uuid.New()

	err := adapter.PublishBacnetObjectUpdated(
		context.Background(),
		appcollaboration.BacnetObjectUpdated{
			Envelope: appcollaboration.Envelope{
				ProjectID: projectID,
				ActorID:   &actorID,
			},
			BacnetObjectID: uuid.New(),
			FieldDeviceIDs: []uuid.UUID{fieldDeviceOne, uuid.Nil, fieldDeviceTwo},
		},
	)
	if err != nil {
		t.Fatalf("PublishBacnetObjectUpdated returned error: %v", err)
	}
	if publisher.projectID != projectID ||
		publisher.actorID == nil || *publisher.actorID != actorID ||
		publisher.scope != "field_device" {
		t.Fatalf("unexpected v1 refresh envelope: %+v", publisher)
	}
	if want := []string{fieldDeviceOne.String(), fieldDeviceTwo.String()}; !reflect.DeepEqual(publisher.entityIDs, want) {
		t.Fatalf("entity IDs: got %v, want %v", publisher.entityIDs, want)
	}
}

func TestCollaborationCommandAdapterMapsBacnetObjectCreatedToParentFieldDeviceRefresh(t *testing.T) {
	publisher := &collaborationRefreshPublisherSpy{}
	adapter := NewCollaborationCommandAdapter(publisher)
	projectID := uuid.New()
	actorID := uuid.New()
	fieldDeviceID := uuid.New()

	err := adapter.PublishBacnetObjectCreated(
		context.Background(),
		appcollaboration.BacnetObjectCreated{
			Envelope: appcollaboration.Envelope{
				ProjectID: projectID,
				ActorID:   &actorID,
			},
			BacnetObjectID: uuid.New(),
			FieldDeviceID:  fieldDeviceID,
		},
	)
	if err != nil {
		t.Fatalf("PublishBacnetObjectCreated returned error: %v", err)
	}
	if publisher.projectID != projectID ||
		publisher.actorID == nil || *publisher.actorID != actorID ||
		publisher.scope != "field_device" {
		t.Fatalf("unexpected v1 refresh envelope: %+v", publisher)
	}
	if want := []string{fieldDeviceID.String()}; !reflect.DeepEqual(publisher.entityIDs, want) {
		t.Fatalf("entity IDs: got %v, want %v", publisher.entityIDs, want)
	}
}

func TestCollaborationCommandAdapterMapsSPSControllerUpdatedToTargetedVersionOneRefresh(t *testing.T) {
	publisher := &collaborationRefreshPublisherSpy{}
	adapter := NewCollaborationCommandAdapter(publisher)
	projectID := uuid.New()
	actorID := uuid.New()
	spsControllerID := uuid.New()

	err := adapter.PublishSPSControllerUpdated(
		context.Background(),
		appcollaboration.SPSControllerUpdated{
			Envelope: appcollaboration.Envelope{
				ProjectID: projectID,
				ActorID:   &actorID,
			},
			SPSControllerID: spsControllerID,
		},
	)
	if err != nil {
		t.Fatalf("PublishSPSControllerUpdated returned error: %v", err)
	}

	if publisher.projectID != projectID ||
		publisher.actorID == nil || *publisher.actorID != actorID ||
		publisher.scope != "sps_controller" {
		t.Fatalf("unexpected v1 refresh envelope: %+v", publisher)
	}
	if want := []string{spsControllerID.String()}; !reflect.DeepEqual(publisher.entityIDs, want) {
		t.Fatalf("entity IDs: got %v, want %v", publisher.entityIDs, want)
	}
}

func TestCollaborationCommandAdapterMapsSPSControllerCreatedToVersionOneDelta(t *testing.T) {
	publisher := &collaborationRefreshPublisherSpy{}
	adapter := NewCollaborationCommandAdapter(publisher)
	projectID := uuid.New()
	actorID := uuid.New()
	controllerID := uuid.New()
	cabinetID := uuid.New()
	gaDevice := "AAA"
	description := "Automation controller"
	createdAt := time.Date(2026, time.July, 20, 23, 30, 0, 0, time.UTC)

	err := adapter.PublishSPSControllerCreated(
		context.Background(),
		appcollaboration.SPSControllerCreated{
			Envelope: appcollaboration.Envelope{
				ProjectID: projectID,
				ActorID:   &actorID,
			},
			SPSController: appcollaboration.SPSControllerState{
				ID:                controllerID,
				ControlCabinetID:  cabinetID,
				GADevice:          &gaDevice,
				DeviceName:        "BLD_AK01_AAA",
				DeviceDescription: &description,
				CreatedAt:         createdAt,
				UpdatedAt:         createdAt,
			},
		},
	)
	if err != nil {
		t.Fatalf("PublishSPSControllerCreated returned error: %v", err)
	}
	if publisher.projectID != projectID ||
		publisher.actorID == nil || *publisher.actorID != actorID ||
		publisher.controller == nil || publisher.controller.ID != controllerID ||
		publisher.controller.ControlCabinetID != cabinetID ||
		publisher.controller.GADevice == nil || *publisher.controller.GADevice != gaDevice ||
		publisher.controller.DeviceName != "BLD_AK01_AAA" ||
		publisher.controller.DeviceDescription == nil || *publisher.controller.DeviceDescription != description ||
		!publisher.controller.CreatedAt.Equal(createdAt) {
		t.Fatalf("unexpected v1 SPS delta: %+v", publisher)
	}
}

func TestCollaborationCommandAdapterMapsSPSControllerClonedToVersionOneDelta(t *testing.T) {
	publisher := &collaborationRefreshPublisherSpy{}
	adapter := NewCollaborationCommandAdapter(publisher)
	projectID := uuid.New()
	actorID := uuid.New()
	controllerID := uuid.New()
	cabinetID := uuid.New()
	gaDevice := "AAB"
	createdAt := time.Date(2026, time.July, 20, 23, 45, 0, 0, time.UTC)

	err := adapter.PublishSPSControllerCloned(
		context.Background(),
		appcollaboration.SPSControllerCloned{
			Envelope: appcollaboration.Envelope{
				ProjectID: projectID,
				ActorID:   &actorID,
			},
			SourceSPSControllerID: uuid.New(),
			SPSController: appcollaboration.SPSControllerState{
				ID:               controllerID,
				ControlCabinetID: cabinetID,
				GADevice:         &gaDevice,
				DeviceName:       "BLD_AK01_AAB",
				CreatedAt:        createdAt,
				UpdatedAt:        createdAt,
			},
		},
	)
	if err != nil {
		t.Fatalf("PublishSPSControllerCloned returned error: %v", err)
	}
	if publisher.projectID != projectID ||
		publisher.actorID == nil || *publisher.actorID != actorID ||
		publisher.controller == nil || publisher.controller.ID != controllerID ||
		publisher.controller.ControlCabinetID != cabinetID ||
		publisher.controller.GADevice == nil || *publisher.controller.GADevice != gaDevice ||
		publisher.controller.DeviceName != "BLD_AK01_AAB" ||
		!publisher.controller.CreatedAt.Equal(createdAt) {
		t.Fatalf("unexpected v1 cloned SPS delta: %+v", publisher)
	}
}

func TestCollaborationCommandAdapterMapsSPSControllerSystemTypeClonedToOwningSPSRefresh(t *testing.T) {
	publisher := &collaborationRefreshPublisherSpy{}
	adapter := NewCollaborationCommandAdapter(publisher)
	projectID := uuid.New()
	actorID := uuid.New()
	spsControllerID := uuid.New()

	err := adapter.PublishSPSControllerSystemTypeCloned(
		context.Background(),
		appcollaboration.SPSControllerSystemTypeCloned{
			Envelope: appcollaboration.Envelope{
				ProjectID: projectID,
				ActorID:   &actorID,
			},
			SourceSPSControllerSystemTypeID: uuid.New(),
			SPSControllerSystemTypeID:       uuid.New(),
			SPSControllerID:                 spsControllerID,
		},
	)
	if err != nil {
		t.Fatalf("PublishSPSControllerSystemTypeCloned returned error: %v", err)
	}
	if publisher.projectID != projectID ||
		publisher.actorID == nil || *publisher.actorID != actorID ||
		publisher.scope != "sps_controller" {
		t.Fatalf("unexpected v1 SPS system-type refresh envelope: %+v", publisher)
	}
	if want := []string{spsControllerID.String()}; !reflect.DeepEqual(publisher.entityIDs, want) {
		t.Fatalf("entity IDs: got %v, want %v", publisher.entityIDs, want)
	}
}

func TestCollaborationCommandAdapterMapsSPSControllerMovedToTargetedVersionOneRefresh(t *testing.T) {
	publisher := &collaborationRefreshPublisherSpy{}
	adapter := NewCollaborationCommandAdapter(publisher)
	spsControllerID := uuid.New()

	err := adapter.PublishSPSControllerMoved(
		context.Background(),
		appcollaboration.SPSControllerMoved{
			Envelope:             appcollaboration.Envelope{ProjectID: uuid.New()},
			SPSControllerID:      spsControllerID,
			FromControlCabinetID: uuid.New(),
			ToControlCabinetID:   uuid.New(),
		},
	)
	if err != nil {
		t.Fatalf("PublishSPSControllerMoved returned error: %v", err)
	}

	if publisher.scope != "sps_controller" {
		t.Fatalf("expected sps_controller scope, got %q", publisher.scope)
	}
	if want := []string{spsControllerID.String()}; !reflect.DeepEqual(publisher.entityIDs, want) {
		t.Fatalf("entity IDs: got %v, want %v", publisher.entityIDs, want)
	}
}

func TestCollaborationCommandAdapterMapsSPSControllerDeletedToTargetedVersionOneRefresh(t *testing.T) {
	publisher := &collaborationRefreshPublisherSpy{}
	adapter := NewCollaborationCommandAdapter(publisher)
	projectID := uuid.New()
	actorID := uuid.New()
	spsControllerID := uuid.New()

	err := adapter.PublishSPSControllerDeleted(
		context.Background(),
		appcollaboration.SPSControllerDeleted{
			Envelope: appcollaboration.Envelope{
				ProjectID: projectID,
				ActorID:   &actorID,
			},
			SPSControllerID:  spsControllerID,
			ControlCabinetID: uuid.New(),
		},
	)
	if err != nil {
		t.Fatalf("PublishSPSControllerDeleted returned error: %v", err)
	}

	if publisher.projectID != projectID ||
		publisher.actorID == nil || *publisher.actorID != actorID ||
		publisher.scope != "sps_controller" {
		t.Fatalf("unexpected v1 refresh envelope: %+v", publisher)
	}
	if want := []string{spsControllerID.String()}; !reflect.DeepEqual(publisher.entityIDs, want) {
		t.Fatalf("entity IDs: got %v, want %v", publisher.entityIDs, want)
	}
}

func TestCollaborationCommandAdapterMapsFieldDevicesCreatedToExactVersionOneDelta(t *testing.T) {
	publisher := &collaborationRefreshPublisherSpy{}
	adapter := NewCollaborationCommandAdapter(publisher)
	projectID := uuid.New()
	actorID := uuid.New()
	fieldDeviceID := uuid.New()
	parentID := uuid.New()
	systemPartID := uuid.New()
	specificationID := uuid.New()
	apparatID := uuid.New()
	bmk := "M01"
	description := "supply air sensor"
	textFix := "AI01"
	createdAt := time.Date(2026, time.July, 21, 9, 0, 0, 0, time.UTC)
	updatedAt := createdAt.Add(time.Minute)

	err := adapter.PublishFieldDevicesCreated(
		context.Background(),
		appcollaboration.FieldDevicesCreated{
			Envelope: appcollaboration.Envelope{
				ProjectID:  projectID,
				ActorID:    &actorID,
				OccurredAt: updatedAt,
			},
			FieldDevices: []appcollaboration.FieldDeviceState{{
				ID:                        fieldDeviceID,
				BMK:                       &bmk,
				Description:               &description,
				TextFix:                   &textFix,
				ApparatNumber:             8,
				SPSControllerSystemTypeID: parentID,
				SystemPartID:              systemPartID,
				SpecificationID:           &specificationID,
				ApparatID:                 apparatID,
				CreatedAt:                 createdAt,
				UpdatedAt:                 updatedAt,
			}},
		},
	)
	if err != nil {
		t.Fatalf("PublishFieldDevicesCreated returned error: %v", err)
	}
	if publisher.projectID != projectID || publisher.actorID == nil ||
		*publisher.actorID != actorID || publisher.scope != "" ||
		len(publisher.entityIDs) != 0 || len(publisher.fieldDevices) != 1 {
		t.Fatalf("unexpected v1 delta envelope: %+v", publisher)
	}
	payload := publisher.fieldDevices[0]
	if len(payload) != 11 || payload["id"] != fieldDeviceID ||
		payload["sps_controller_system_type_id"] != parentID ||
		payload["apparat_id"] != apparatID || payload["created_at"] != createdAt ||
		payload["updated_at"] != updatedAt {
		t.Fatalf("unexpected v1 delta payload: %+v", payload)
	}
	assertStringPointer := func(key, want string) {
		t.Helper()
		value, ok := payload[key].(*string)
		if !ok || value == nil || *value != want {
			t.Fatalf("%s: got %#v, want pointer to %q", key, payload[key], want)
		}
	}
	assertStringPointer("bmk", bmk)
	assertStringPointer("description", description)
	assertStringPointer("text_fix", textFix)
	if value, ok := payload["apparat_nr"].(*int); !ok || value == nil || *value != 8 {
		t.Fatalf("apparat_nr: %#v", payload["apparat_nr"])
	}
	if value, ok := payload["system_part_id"].(*uuid.UUID); !ok || value == nil || *value != systemPartID {
		t.Fatalf("system_part_id: %#v", payload["system_part_id"])
	}
	if value, ok := payload["specification_id"].(*uuid.UUID); !ok || value == nil || *value != specificationID {
		t.Fatalf("specification_id: %#v", payload["specification_id"])
	}
}

func TestCollaborationCommandAdapterFallsBackToFullRefreshForOversizedFieldDeviceDelta(t *testing.T) {
	for _, tc := range []struct {
		name   string
		states []appcollaboration.FieldDeviceState
	}{
		{
			name:   "item count",
			states: makeFieldDeviceStates(projectCollaborationMaxFieldDeviceDeltas+1, ""),
		},
		{
			name: "encoded bytes",
			states: makeFieldDeviceStates(
				projectCollaborationMaxFieldDeviceDeltas,
				strings.Repeat("x", 250),
			),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			publisher := &collaborationRefreshPublisherSpy{}
			adapter := NewCollaborationCommandAdapter(publisher)
			projectID := uuid.New()
			actorID := uuid.New()

			err := adapter.PublishFieldDevicesCreated(
				context.Background(),
				appcollaboration.FieldDevicesCreated{
					Envelope: appcollaboration.Envelope{
						ProjectID:  projectID,
						ActorID:    &actorID,
						OccurredAt: time.Now().UTC(),
					},
					FieldDevices: tc.states,
				},
			)
			if err != nil {
				t.Fatalf("PublishFieldDevicesCreated returned error: %v", err)
			}
			if len(publisher.fieldDevices) != 0 || publisher.projectID != projectID ||
				publisher.actorID == nil || *publisher.actorID != actorID ||
				publisher.scope != "field_device" || len(publisher.entityIDs) != 0 {
				t.Fatalf("expected project-scoped full refresh, got %+v", publisher)
			}
		})
	}
}

func makeFieldDeviceStates(count int, text string) []appcollaboration.FieldDeviceState {
	states := make([]appcollaboration.FieldDeviceState, count)
	for index := range states {
		value := text
		states[index] = appcollaboration.FieldDeviceState{
			ID:                        uuid.New(),
			Description:               &value,
			TextFix:                   &value,
			ApparatNumber:             index + 1,
			SPSControllerSystemTypeID: uuid.New(),
			SystemPartID:              uuid.New(),
			ApparatID:                 uuid.New(),
			CreatedAt:                 time.Now().UTC(),
			UpdatedAt:                 time.Now().UTC(),
		}
	}
	return states
}
