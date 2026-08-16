package project

import (
	"context"

	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	domainProject "github.com/besart951/go_infra_link/backend/internal/domain/project"
	"github.com/google/uuid"
)

type facilityProjectLookup interface {
	ListProjectIDsByControlCabinetID(ctx context.Context, controlCabinetID uuid.UUID) ([]uuid.UUID, error)
	ListProjectIDsBySPSControllerID(ctx context.Context, spsControllerID uuid.UUID) ([]uuid.UUID, error)
}

type fieldDeviceProjectLookup interface {
	ListProjectIDsByFieldDeviceID(ctx context.Context, fieldDeviceID uuid.UUID) ([]uuid.UUID, error)
}

type projectRefreshPublisher interface {
	BroadcastRefreshRequest(projectID uuid.UUID, actorID *uuid.UUID, scope string, entityIDs []string)
	BroadcastControlCabinetDelta(projectID uuid.UUID, actorID *uuid.UUID, controlCabinet domainFacility.ControlCabinet)
	BroadcastSPSControllerDelta(projectID uuid.UUID, actorID *uuid.UUID, spsController domainFacility.SPSController)
}

type durableProjectChangePublisher interface {
	BroadcastProjectChange(ctx context.Context, change ProjectChange) error
}

type FacilityRefreshBroadcaster struct {
	lookup    facilityProjectLookup
	publisher projectRefreshPublisher
	changes   ProjectChangeService
}

func NewFacilityRefreshBroadcaster(lookup facilityProjectLookup, publisher projectRefreshPublisher, changes ...ProjectChangeService) *FacilityRefreshBroadcaster {
	b := &FacilityRefreshBroadcaster{
		lookup:    lookup,
		publisher: publisher,
	}
	if len(changes) > 0 {
		b.changes = changes[0]
	}
	return b
}

// ProjectIDsForControlCabinet captures the collaboration audience before a
// destructive mutation removes the project link used by the normal lookup.
func (b *FacilityRefreshBroadcaster) ProjectIDsForControlCabinet(ctx context.Context, id uuid.UUID) ([]uuid.UUID, error) {
	if b == nil || b.lookup == nil {
		return nil, nil
	}
	return b.lookup.ListProjectIDsByControlCabinetID(ctx, id)
}

// ProjectIDsForSPSController captures the collaboration audience before a
// destructive mutation removes the project link used by the normal lookup.
func (b *FacilityRefreshBroadcaster) ProjectIDsForSPSController(ctx context.Context, id uuid.UUID) ([]uuid.UUID, error) {
	if b == nil || b.lookup == nil {
		return nil, nil
	}
	return b.lookup.ListProjectIDsBySPSControllerID(ctx, id)
}

// ProjectIDsForFieldDevice captures the collaboration audience before a
// destructive mutation removes the project link used by the normal lookup.
func (b *FacilityRefreshBroadcaster) ProjectIDsForFieldDevice(ctx context.Context, id uuid.UUID) ([]uuid.UUID, error) {
	if b == nil || b.lookup == nil {
		return nil, nil
	}
	lookup, ok := any(b.lookup).(fieldDeviceProjectLookup)
	if !ok {
		return nil, nil
	}
	return lookup.ListProjectIDsByFieldDeviceID(ctx, id)
}

// BroadcastChangeForProjects publishes to a previously captured audience.
// It is intentionally separate from lookup so successful deletes cannot lose
// their final invalidation event.
func (b *FacilityRefreshBroadcaster) BroadcastChangeForProjects(ctx context.Context, actorID *uuid.UUID, projectIDs []uuid.UUID, eventType string, entityID uuid.UUID) {
	for _, projectID := range projectIDs {
		b.recordAndPublish(ctx, projectID, actorID, eventType, entityID)
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
	if b.changes == nil {
		for _, projectID := range projectIDs {
			b.publisher.BroadcastRefreshRequest(projectID, actorID, scope, []string{controlCabinetID.String()})
		}
		return
	}

	for _, projectID := range projectIDs {
		b.recordAndPublish(ctx, projectID, actorID, "project.control_cabinet.updated", controlCabinetID)
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
	if b.changes == nil {
		for _, projectID := range projectIDs {
			b.publisher.BroadcastRefreshRequest(projectID, actorID, scope, []string{spsControllerID.String()})
		}
		return
	}

	for _, projectID := range projectIDs {
		b.recordAndPublish(ctx, projectID, actorID, "project.sps_controller.updated", spsControllerID)
	}
}

func (b *FacilityRefreshBroadcaster) BroadcastControlCabinetDelta(ctx context.Context, actorID *uuid.UUID, controlCabinet domainFacility.ControlCabinet, changedFields ...string) {
	if b == nil || b.lookup == nil || b.publisher == nil {
		return
	}

	projectIDs, err := b.lookup.ListProjectIDsByControlCabinetID(ctx, controlCabinet.ID)
	if err != nil {
		return
	}
	if b.changes == nil {
		for _, projectID := range projectIDs {
			b.publisher.BroadcastControlCabinetDelta(projectID, actorID, controlCabinet)
		}
		return
	}

	for _, projectID := range projectIDs {
		b.recordAndPublishWithFields(ctx, projectID, actorID, "project.control_cabinet.updated", controlCabinet.ID, changedFields)
	}

}

func (b *FacilityRefreshBroadcaster) BroadcastSPSControllerDelta(ctx context.Context, actorID *uuid.UUID, spsController domainFacility.SPSController, changedFields ...string) {
	if b == nil || b.lookup == nil || b.publisher == nil {
		return
	}

	projectIDs, err := b.lookup.ListProjectIDsBySPSControllerID(ctx, spsController.ID)
	if err != nil {
		return
	}
	if b.changes == nil {
		for _, projectID := range projectIDs {
			b.publisher.BroadcastSPSControllerDelta(projectID, actorID, spsController)
		}
		return
	}

	for _, projectID := range projectIDs {
		b.recordAndPublishWithFields(ctx, projectID, actorID, "project.sps_controller.updated", spsController.ID, changedFields)
	}
}

// BroadcastSPSControllerSystemTypeChange records an exact project event while
// resolving the audience through the owning controller. The system type itself
// has no direct project link, so this keeps the relationship traversal in one
// adapter instead of every global mutation handler.
func (b *FacilityRefreshBroadcaster) BroadcastSPSControllerSystemTypeChange(ctx context.Context, actorID *uuid.UUID, spsControllerID, systemTypeID uuid.UUID, action string, changedFields ...string) {
	if b == nil || b.lookup == nil || b.publisher == nil {
		return
	}
	projectIDs, err := b.lookup.ListProjectIDsBySPSControllerID(ctx, spsControllerID)
	if err != nil {
		return
	}
	for _, projectID := range projectIDs {
		b.recordAndPublishWithFields(ctx, projectID, actorID, "project.sps_controller_system_type."+action, systemTypeID, changedFields)
	}
}

func (b *FacilityRefreshBroadcaster) BroadcastFieldDeviceChange(ctx context.Context, actorID *uuid.UUID, fieldDeviceID uuid.UUID, action string, changedFields ...string) {
	if b == nil || b.lookup == nil {
		return
	}
	lookup, ok := any(b.lookup).(fieldDeviceProjectLookup)
	if !ok {
		return
	}
	projectIDs, err := lookup.ListProjectIDsByFieldDeviceID(ctx, fieldDeviceID)
	if err != nil {
		return
	}
	for _, projectID := range projectIDs {
		b.recordAndPublishWithFields(ctx, projectID, actorID, "project.field_device."+action, fieldDeviceID, changedFields)
	}
}

func (b *FacilityRefreshBroadcaster) recordAndPublish(ctx context.Context, projectID uuid.UUID, actorID *uuid.UUID, eventType string, entityID uuid.UUID) {
	b.recordAndPublishWithFields(ctx, projectID, actorID, eventType, entityID, nil)
}

func (b *FacilityRefreshBroadcaster) recordAndPublishWithFields(ctx context.Context, projectID uuid.UUID, actorID *uuid.UUID, eventType string, entityID uuid.UUID, changedFields []string) {
	if b.changes == nil {
		b.publisher.BroadcastRefreshRequest(projectID, actorID, "", []string{entityID.String()})
		return
	}
	if recorder, ok := b.changes.(interface {
		RecordEventsWithFields(context.Context, uuid.UUID, string, *uuid.UUID, []string, ...string) ([]domainProject.Change, error)
	}); ok {
		events, err := recorder.RecordEventsWithFields(ctx, projectID, eventType, actorID, changedFields, entityID.String())
		if err == nil {
			b.publishDurableChanges(ctx, events)
		}
		return
	}
	recorder, ok := b.changes.(interface {
		RecordEvents(context.Context, uuid.UUID, string, *uuid.UUID, ...string) ([]domainProject.Change, error)
	})
	if !ok {
		_ = b.changes.RecordEvent(ctx, projectID, eventType, actorID, entityID.String())
		return
	}
	events, err := recorder.RecordEvents(ctx, projectID, eventType, actorID, entityID.String())
	if err != nil {
		return
	}
	b.publishDurableChanges(ctx, events)
}

func (b *FacilityRefreshBroadcaster) publishDurableChanges(ctx context.Context, events []domainProject.Change) {
	publisher, ok := any(b.publisher).(durableProjectChangePublisher)
	if !ok {
		return
	}
	for _, event := range events {
		if event.AggregateID == nil {
			continue
		}
		refs := make(map[string]string, len(event.ParentRefs))
		for k, v := range event.ParentRefs {
			refs[k] = v.String()
		}
		_ = publisher.BroadcastProjectChange(ctx, ProjectChange{ProjectID: event.ProjectID, Revision: int64(event.Revision), EventID: event.EventID, AggregateType: event.AggregateType, AggregateID: *event.AggregateID, Action: string(event.Action), ActorID: event.ActorID, ChangedFields: event.ChangedFields, ParentRefs: refs, OccurredAt: event.OccurredAt})
	}
}
