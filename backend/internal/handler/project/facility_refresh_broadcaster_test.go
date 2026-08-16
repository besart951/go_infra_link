package project

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	domainProject "github.com/besart951/go_infra_link/backend/internal/domain/project"
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
		Base:             domain.Base{ID: controlCabinetID},
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
		Base:             domain.Base{ID: spsControllerID},
		ControlCabinetID: uuid.New(),
		DeviceName:       deviceName,
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

func TestFacilityRefreshBroadcasterRecordsPreciseSystemTypeChange(t *testing.T) {
	ctx := context.Background()
	projectID, controllerID, systemTypeID := uuid.New(), uuid.New(), uuid.New()
	publisher := &fakeProjectRefreshPublisher{}
	changes := &preciseChangeRecorder{}
	broadcaster := NewFacilityRefreshBroadcaster(
		&fakeFacilityProjectLookup{spsControllerProjectIDs: []uuid.UUID{projectID}},
		publisher,
		changes,
	)

	broadcaster.BroadcastSPSControllerSystemTypeChange(ctx, nil, controllerID, systemTypeID, "updated", "document_name")

	if len(changes.calls) != 1 {
		t.Fatalf("recorded %d changes, want 1", len(changes.calls))
	}
	call := changes.calls[0]
	if call.eventType != "project.sps_controller_system_type.updated" || call.entityID != systemTypeID.String() {
		t.Fatalf("unexpected change call: %+v", call)
	}
	if got, want := call.changedFields, []string{"document_name"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("changed fields = %#v, want %#v", got, want)
	}
	if len(publisher.projectChanges) != 1 || publisher.projectChanges[0].AggregateID != systemTypeID {
		t.Fatalf("expected durable project event for system type, got %+v", publisher.projectChanges)
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
	calls                       []projectRefreshCall
	controlCabinetDeltas        []domainFacility.ControlCabinet
	controlCabinetDeltaProjects []uuid.UUID
	spsControllerDeltas         []domainFacility.SPSController
	spsControllerDeltaProjects  []uuid.UUID
	projectChanges              []ProjectChange
}

type projectRefreshCall struct {
	projectID uuid.UUID
	scope     string
	entityIDs []string
}

func (f *fakeProjectRefreshPublisher) BroadcastRefreshRequest(projectID uuid.UUID, _ *uuid.UUID, scope string, entityIDs []string) {
	f.calls = append(f.calls, projectRefreshCall{projectID: projectID, scope: scope, entityIDs: append([]string(nil), entityIDs...)})
}

func (f *fakeProjectRefreshPublisher) BroadcastControlCabinetDelta(projectID uuid.UUID, _ *uuid.UUID, controlCabinet domainFacility.ControlCabinet) {
	f.controlCabinetDeltaProjects = append(f.controlCabinetDeltaProjects, projectID)
	f.controlCabinetDeltas = append(f.controlCabinetDeltas, controlCabinet)
}

func (f *fakeProjectRefreshPublisher) BroadcastSPSControllerDelta(projectID uuid.UUID, _ *uuid.UUID, spsController domainFacility.SPSController) {
	f.spsControllerDeltaProjects = append(f.spsControllerDeltaProjects, projectID)
	f.spsControllerDeltas = append(f.spsControllerDeltas, spsController)
}

func (f *fakeProjectRefreshPublisher) BroadcastProjectChange(_ context.Context, change ProjectChange) error {
	f.projectChanges = append(f.projectChanges, change)
	return nil
}

type preciseChangeCall struct {
	eventType     string
	entityID      string
	changedFields []string
}

type preciseChangeRecorder struct {
	calls []preciseChangeCall
}

func (r *preciseChangeRecorder) ListAfter(context.Context, uuid.UUID, uint64, int) (*domainProject.ChangePage, error) {
	return nil, nil
}

func (r *preciseChangeRecorder) RecordEvent(context.Context, uuid.UUID, string, *uuid.UUID, ...string) error {
	return nil
}

func (r *preciseChangeRecorder) RecordEventsWithFields(_ context.Context, projectID uuid.UUID, eventType string, actorID *uuid.UUID, changedFields []string, entityIDs ...string) ([]domainProject.Change, error) {
	entityID := uuid.Nil
	if len(entityIDs) > 0 {
		entityID, _ = uuid.Parse(entityIDs[0])
	}
	r.calls = append(r.calls, preciseChangeCall{eventType: eventType, entityID: entityID.String(), changedFields: append([]string(nil), changedFields...)})
	return []domainProject.Change{{
		ProjectID: projectID, AggregateType: "sps_controller_system_type", AggregateID: &entityID, Action: domainProject.ChangeUpdated, ActorID: actorID, ChangedFields: append([]string(nil), changedFields...),
	}}, nil
}
