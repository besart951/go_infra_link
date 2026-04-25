package project

import (
	"context"

	"github.com/google/uuid"
)

type facilityProjectLookup interface {
	ListProjectIDsByControlCabinetID(ctx context.Context, controlCabinetID uuid.UUID) ([]uuid.UUID, error)
	ListProjectIDsBySPSControllerID(ctx context.Context, spsControllerID uuid.UUID) ([]uuid.UUID, error)
}

type projectRefreshPublisher interface {
	BroadcastRefreshRequest(projectID uuid.UUID, actorID *uuid.UUID, scope string, deviceIDs []string)
}

type FacilityRefreshBroadcaster struct {
	lookup    facilityProjectLookup
	publisher projectRefreshPublisher
}

func NewFacilityRefreshBroadcaster(lookup facilityProjectLookup, publisher projectRefreshPublisher) *FacilityRefreshBroadcaster {
	return &FacilityRefreshBroadcaster{
		lookup:    lookup,
		publisher: publisher,
	}
}

func (b *FacilityRefreshBroadcaster) BroadcastRefreshForControlCabinet(ctx context.Context, controlCabinetID uuid.UUID, scope string) {
	if b == nil || b.lookup == nil || b.publisher == nil {
		return
	}

	projectIDs, err := b.lookup.ListProjectIDsByControlCabinetID(ctx, controlCabinetID)
	if err != nil {
		return
	}

	for _, projectID := range projectIDs {
		b.publisher.BroadcastRefreshRequest(projectID, nil, scope, nil)
	}
}

func (b *FacilityRefreshBroadcaster) BroadcastRefreshForSPSController(ctx context.Context, spsControllerID uuid.UUID, scope string) {
	if b == nil || b.lookup == nil || b.publisher == nil {
		return
	}

	projectIDs, err := b.lookup.ListProjectIDsBySPSControllerID(ctx, spsControllerID)
	if err != nil {
		return
	}

	for _, projectID := range projectIDs {
		b.publisher.BroadcastRefreshRequest(projectID, nil, scope, nil)
	}
}
