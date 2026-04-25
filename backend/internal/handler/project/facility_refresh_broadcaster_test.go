package project

import (
	"context"
	"errors"
	"testing"

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

	broadcaster.BroadcastRefreshForControlCabinet(ctx, controlCabinetID, "control_cabinet")

	if len(publisher.calls) != 2 {
		t.Fatalf("expected two refresh broadcasts, got %+v", publisher.calls)
	}
	if publisher.calls[0].projectID != projectOneID || publisher.calls[1].projectID != projectTwoID {
		t.Fatalf("expected refreshes for linked projects, got %+v", publisher.calls)
	}
	if publisher.calls[0].scope != "control_cabinet" || publisher.calls[1].scope != "control_cabinet" {
		t.Fatalf("expected control cabinet scope, got %+v", publisher.calls)
	}
}

func TestFacilityRefreshBroadcasterSkipsPublishWhenLookupFails(t *testing.T) {
	publisher := &fakeProjectRefreshPublisher{}
	broadcaster := NewFacilityRefreshBroadcaster(
		&fakeFacilityProjectLookup{err: errors.New("lookup failed")},
		publisher,
	)

	broadcaster.BroadcastRefreshForSPSController(context.Background(), uuid.New(), "sps_controller")

	if len(publisher.calls) != 0 {
		t.Fatalf("expected no refresh broadcasts after lookup error, got %+v", publisher.calls)
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
	calls []projectRefreshCall
}

type projectRefreshCall struct {
	projectID uuid.UUID
	scope     string
}

func (f *fakeProjectRefreshPublisher) BroadcastRefreshRequest(projectID uuid.UUID, _ *uuid.UUID, scope string, _ []string) {
	f.calls = append(f.calls, projectRefreshCall{projectID: projectID, scope: scope})
}
