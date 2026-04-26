package project

import (
	"context"
	"errors"
	"testing"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
)

func TestFacilityRefreshBroadcasterBroadcastsControlCabinetProjects(t *testing.T) {
	ctx := context.Background()
	controlCabinetID := uuid.New()
	projectOneID := uuid.New()
	projectTwoID := uuid.New()
	publisher := &fakeProjectRefreshPublisher{}
	broadcaster := NewFacilityRefreshBroadcaster(
		&fakeFacilityProjectLookup{controlCabinetProjectIDs: []uuid.UUID{projectOneID, projectTwoID}},
		publisher,
	)

	broadcaster.BroadcastRefreshForControlCabinet(ctx, nil, controlCabinetID, "control_cabinet")

	if len(publisher.calls) != 2 {
		t.Fatalf("expected two refresh broadcasts, got %+v", publisher.calls)
	}
	if publisher.calls[0].projectID != projectOneID || publisher.calls[1].projectID != projectTwoID {
		t.Fatalf("expected refreshes for linked projects, got %+v", publisher.calls)
	}
	if publisher.calls[0].scope != "control_cabinet" || publisher.calls[1].scope != "control_cabinet" {
		t.Fatalf("expected control cabinet scope, got %+v", publisher.calls)
	}
	if len(publisher.calls[0].entityIDs) != 1 || publisher.calls[0].entityIDs[0] != controlCabinetID.String() {
		t.Fatalf("expected control cabinet entity id, got %+v", publisher.calls)
	}
}

func TestFacilityRefreshBroadcasterSkipsPublishWhenLookupFails(t *testing.T) {
	publisher := &fakeProjectRefreshPublisher{}
	broadcaster := NewFacilityRefreshBroadcaster(
		&fakeFacilityProjectLookup{err: errors.New("lookup failed")},
		publisher,
	)

	broadcaster.BroadcastRefreshForSPSController(context.Background(), nil, uuid.New(), "sps_controller")

	if len(publisher.calls) != 0 {
		t.Fatalf("expected no refresh broadcasts after lookup error, got %+v", publisher.calls)
	}
}

func TestFacilityRefreshBroadcasterBroadcastsControlCabinetDelta(t *testing.T) {
	ctx := context.Background()
	controlCabinetID := uuid.New()
	projectID := uuid.New()
	publisher := &fakeProjectRefreshPublisher{}
	broadcaster := NewFacilityRefreshBroadcaster(
		&fakeFacilityProjectLookup{controlCabinetProjectIDs: []uuid.UUID{projectID}},
		publisher,
	)

	controlCabinetNr := "CC-1"
	broadcaster.BroadcastControlCabinetDelta(ctx, nil, domainFacility.ControlCabinet{
		Base: domain.Base{ID: controlCabinetID},
		BuildingID:       uuid.New(),
		ControlCabinetNr: &controlCabinetNr,
	})

	if len(publisher.controlCabinetDeltas) != 1 {
		t.Fatalf("expected control cabinet delta broadcast, got %+v", publisher.controlCabinetDeltas)
	}
	if publisher.controlCabinetDeltas[0].ID != controlCabinetID {
		t.Fatalf("expected control cabinet delta id, got %+v", publisher.controlCabinetDeltas)
	}
}

func TestFacilityRefreshBroadcasterBroadcastsSPSControllerDelta(t *testing.T) {
	ctx := context.Background()
	spsControllerID := uuid.New()
	projectID := uuid.New()
	publisher := &fakeProjectRefreshPublisher{}
	broadcaster := NewFacilityRefreshBroadcaster(
		&fakeFacilityProjectLookup{spsControllerProjectIDs: []uuid.UUID{projectID}},
		publisher,
	)

	deviceName := "SPS 1"
	broadcaster.BroadcastSPSControllerDelta(ctx, nil, domainFacility.SPSController{
		Base:              domain.Base{ID: spsControllerID},
		ControlCabinetID:  uuid.New(),
		DeviceName:        deviceName,
	})

	if len(publisher.spsControllerDeltas) != 1 {
		t.Fatalf("expected sps controller delta broadcast, got %+v", publisher.spsControllerDeltas)
	}
	if publisher.spsControllerDeltas[0].DeviceName != deviceName {
		t.Fatalf("expected sps controller delta payload, got %+v", publisher.spsControllerDeltas)
	}
	if publisher.spsControllerDeltaProjects[0] != projectID {
		t.Fatalf("expected sps delta for linked project, got %+v", publisher.spsControllerDeltaProjects)
	}
	if publisher.spsControllerDeltas[0].ID != spsControllerID {
		t.Fatalf("expected sps controller delta id, got %+v", publisher.spsControllerDeltas)
	}
}

type fakeFacilityProjectLookup struct {
	controlCabinetProjectIDs []uuid.UUID
	spsControllerProjectIDs  []uuid.UUID
	err                      error
}

func (f *fakeFacilityProjectLookup) ListProjectIDsByControlCabinetID(context.Context, uuid.UUID) ([]uuid.UUID, error) {
	if f.err != nil {
		return nil, f.err
	}
	return append([]uuid.UUID(nil), f.controlCabinetProjectIDs...), nil
}

func (f *fakeFacilityProjectLookup) ListProjectIDsBySPSControllerID(context.Context, uuid.UUID) ([]uuid.UUID, error) {
	if f.err != nil {
		return nil, f.err
	}
	return append([]uuid.UUID(nil), f.spsControllerProjectIDs...), nil
}

type fakeProjectRefreshPublisher struct {
	calls                     []projectRefreshCall
	controlCabinetDeltas      []projectCollaborationControlCabinet
	controlCabinetDeltaProjects []uuid.UUID
	spsControllerDeltas       []projectCollaborationSPSController
	spsControllerDeltaProjects []uuid.UUID
}

type projectRefreshCall struct {
	projectID uuid.UUID
	scope     string
	entityIDs []string
}

func (f *fakeProjectRefreshPublisher) BroadcastRefreshRequest(projectID uuid.UUID, _ *uuid.UUID, scope string, entityIDs []string) {
	f.calls = append(f.calls, projectRefreshCall{projectID: projectID, scope: scope, entityIDs: append([]string(nil), entityIDs...)})
}

func (f *fakeProjectRefreshPublisher) BroadcastControlCabinetDelta(projectID uuid.UUID, _ *uuid.UUID, controlCabinet projectCollaborationControlCabinet) {
	f.controlCabinetDeltaProjects = append(f.controlCabinetDeltaProjects, projectID)
	f.controlCabinetDeltas = append(f.controlCabinetDeltas, controlCabinet)
}

func (f *fakeProjectRefreshPublisher) BroadcastSPSControllerDelta(projectID uuid.UUID, _ *uuid.UUID, spsController projectCollaborationSPSController) {
	f.spsControllerDeltaProjects = append(f.spsControllerDeltaProjects, projectID)
	f.spsControllerDeltas = append(f.spsControllerDeltas, spsController)
}
