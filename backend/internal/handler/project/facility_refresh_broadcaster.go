package project

import (
	"context"

	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
)

type facilityProjectLookup interface {
	ListProjectIDsByControlCabinetID(ctx context.Context, controlCabinetID uuid.UUID) ([]uuid.UUID, error)
	ListProjectIDsBySPSControllerID(ctx context.Context, spsControllerID uuid.UUID) ([]uuid.UUID, error)
}

type projectRefreshPublisher interface {
	BroadcastRefreshRequest(projectID uuid.UUID, actorID *uuid.UUID, scope string, entityIDs []string)
	BroadcastControlCabinetDelta(projectID uuid.UUID, actorID *uuid.UUID, controlCabinet domainFacility.ControlCabinet)
	BroadcastSPSControllerDelta(projectID uuid.UUID, actorID *uuid.UUID, spsController domainFacility.SPSController)
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

func (b *FacilityRefreshBroadcaster) BroadcastRefreshForControlCabinet(ctx context.Context, actorID *uuid.UUID, controlCabinetID uuid.UUID, scope string) {
	if b == nil || b.lookup == nil || b.publisher == nil {
		return
	}

	projectIDs, err := b.lookup.ListProjectIDsByControlCabinetID(ctx, controlCabinetID)
	if err != nil {
		return
	}

	for _, projectID := range projectIDs {
		b.publisher.BroadcastRefreshRequest(projectID, actorID, scope, []string{controlCabinetID.String()})
	}
}

func (b *FacilityRefreshBroadcaster) BroadcastRefreshForSPSController(ctx context.Context, actorID *uuid.UUID, spsControllerID uuid.UUID, scope string) {
	if b == nil || b.lookup == nil || b.publisher == nil {
		return
	}

	projectIDs, err := b.lookup.ListProjectIDsBySPSControllerID(ctx, spsControllerID)
	if err != nil {
		return
	}

	for _, projectID := range projectIDs {
		b.publisher.BroadcastRefreshRequest(projectID, actorID, scope, []string{spsControllerID.String()})
	}
}

func (b *FacilityRefreshBroadcaster) BroadcastControlCabinetDelta(ctx context.Context, actorID *uuid.UUID, controlCabinet domainFacility.ControlCabinet) {
	if b == nil || b.lookup == nil || b.publisher == nil {
		return
	}

	projectIDs, err := b.lookup.ListProjectIDsByControlCabinetID(ctx, controlCabinet.ID)
	if err != nil {
		return
	}

	for _, projectID := range projectIDs {
		b.publisher.BroadcastControlCabinetDelta(projectID, actorID, controlCabinet)
	}

}

func (b *FacilityRefreshBroadcaster) BroadcastSPSControllerDelta(ctx context.Context, actorID *uuid.UUID, spsController domainFacility.SPSController) {
	if b == nil || b.lookup == nil || b.publisher == nil {
		return
	}

	projectIDs, err := b.lookup.ListProjectIDsBySPSControllerID(ctx, spsController.ID)
	if err != nil {
		return
	}

	for _, projectID := range projectIDs {
		b.publisher.BroadcastSPSControllerDelta(projectID, actorID, spsController)
	}
}
