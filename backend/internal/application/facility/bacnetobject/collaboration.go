package bacnetobject

import (
	"context"
	"fmt"
	"sort"
	"time"

	appcollaboration "github.com/besart951/go_infra_link/backend/internal/application/collaboration"
	domainObjectData "github.com/besart951/go_infra_link/backend/internal/domain/facility/objectdata"
	"github.com/google/uuid"
)

type collaborationDependencies struct {
	projectLinks     ProjectLinkReader
	objectDataOwners ObjectDataOwnerReader
	dispatcher       appcollaboration.CommandDispatcher
	newID            IDGenerator
}

type transactionalCollaborationResolver interface {
	ProjectLinkReader
	ObjectDataOwnerReader
}

// enqueueTransactionalMutation captures the exact project recipient set and
// persists one version-2 event per project before the surrounding mutation
// commits. ObjectData-only ownership uses the typed object_data scope; mixed
// direct/template ownership falls back to a project reconciliation because one
// narrower scope cannot authoritatively cover both views.
func enqueueTransactionalMutation(
	ctx context.Context,
	resolver transactionalCollaborationResolver,
	bacnetObjectID uuid.UUID,
	revision uint64,
	fieldDeviceIDs []uuid.UUID,
	operationID uuid.UUID,
	actorID *uuid.UUID,
	occurredAt time.Time,
	newID IDGenerator,
) ([]uuid.UUID, error) {
	if resolver == nil || !appcollaboration.OutboxConfigured(ctx) {
		return nil, nil
	}
	groupedFieldDevices := make(map[uuid.UUID][]uuid.UUID)
	if len(fieldDeviceIDs) > 0 {
		links, err := resolver.GetByFieldDeviceIDs(ctx, fieldDeviceIDs)
		if err != nil {
			return nil, fmt.Errorf("resolve BACnet FieldDevice projects for outbox: %w", err)
		}
		groupedFieldDevices = groupLinkedFieldDevices(links, fieldDeviceIDs)
	}
	owners, err := resolver.GetByBacnetObjectIDs(ctx, []uuid.UUID{bacnetObjectID})
	if err != nil {
		return nil, fmt.Errorf("resolve BACnet ObjectData projects for outbox: %w", err)
	}
	objectDataIDs := objectDataIDsByProject(bacnetObjectID, owners)
	objectDataProjects := make(map[uuid.UUID]struct{}, len(objectDataIDs))
	for projectID := range objectDataIDs {
		objectDataProjects[projectID] = struct{}{}
	}
	projectIDs := sortedUpdateProjectIDs(groupedFieldDevices, objectDataProjects)
	for _, projectID := range projectIDs {
		envelope := appcollaboration.Envelope{
			SchemaVersion: appcollaboration.SchemaVersionV2,
			EventID:       newID(),
			OperationID:   operationID,
			CorrelationID: operationID,
			ProjectID:     projectID,
			ActorID:       actorID,
			OccurredAt:    occurredAt,
			EntityRevisions: map[string]uint64{
				bacnetObjectID.String(): revision,
			},
		}
		var event appcollaboration.Command
		_, hasDirectOwner := groupedFieldDevices[projectID]
		ids, hasObjectDataOwner := objectDataIDs[projectID]
		switch {
		case hasDirectOwner && hasObjectDataOwner:
			event = appcollaboration.FacilityHierarchyRefreshRequired{
				Envelope: envelope, Scope: appcollaboration.FacilityScopeProject, FullRefresh: true,
			}
		case hasObjectDataOwner:
			event = appcollaboration.FacilityHierarchyRefreshRequired{
				Envelope: envelope, Scope: appcollaboration.FacilityScopeObjectData,
				EntityIDs: append([]uuid.UUID(nil), ids...),
			}
		default:
			event = appcollaboration.BacnetObjectUpdated{
				Envelope: envelope, BacnetObjectID: bacnetObjectID,
				FieldDeviceIDs: append([]uuid.UUID(nil), groupedFieldDevices[projectID]...),
			}
		}
		if _, err := appcollaboration.EnqueueCommand(ctx, event); err != nil {
			return nil, fmt.Errorf(
				"enqueue BACnet collaboration event for project %s: %w",
				projectID,
				err,
			)
		}
	}
	return projectIDs, nil
}

func objectDataIDsByProject(
	bacnetObjectID uuid.UUID,
	owners []domainObjectData.BacnetObjectOwner,
) map[uuid.UUID][]uuid.UUID {
	sets := make(map[uuid.UUID]map[uuid.UUID]struct{})
	for _, owner := range owners {
		if owner.BacnetObjectID != bacnetObjectID || owner.ObjectDataID == uuid.Nil ||
			owner.ProjectID == nil || *owner.ProjectID == uuid.Nil {
			continue
		}
		if sets[*owner.ProjectID] == nil {
			sets[*owner.ProjectID] = make(map[uuid.UUID]struct{})
		}
		sets[*owner.ProjectID][owner.ObjectDataID] = struct{}{}
	}
	result := make(map[uuid.UUID][]uuid.UUID, len(sets))
	for projectID, ids := range sets {
		values := make([]uuid.UUID, 0, len(ids))
		for id := range ids {
			values = append(values, id)
		}
		sort.Slice(values, func(i, j int) bool { return values[i].String() < values[j].String() })
		result[projectID] = values
	}
	return result
}

// dispatchCommittedMutation is the single post-commit policy shared by
// BACnet-root and BACnet-child mutations. Direct-only projects receive a
// targeted FieldDevice refresh. A persisted ObjectData association broadens
// that project's command to the existing version-one project refresh.
func dispatchCommittedMutation(
	ctx context.Context,
	dependencies collaborationDependencies,
	bacnetObjectID uuid.UUID,
	fieldDeviceIDs []uuid.UUID,
	operationID uuid.UUID,
	actorID *uuid.UUID,
	occurredAt time.Time,
) ([]uuid.UUID, []error) {
	if dependencies.dispatcher == nil {
		return nil, nil
	}

	dispatchCtx := context.WithoutCancel(ctx)
	errors := make([]error, 0, 2)
	groupedFieldDevices := make(map[uuid.UUID][]uuid.UUID)
	if len(fieldDeviceIDs) > 0 && dependencies.projectLinks != nil {
		links, err := dependencies.projectLinks.GetByFieldDeviceIDs(
			dispatchCtx,
			fieldDeviceIDs,
		)
		if err != nil {
			errors = append(errors, fmt.Errorf(
				"resolve BACnet object FieldDevice collaboration projects: %w",
				err,
			))
		} else {
			groupedFieldDevices = groupLinkedFieldDevices(links, fieldDeviceIDs)
		}
	}

	objectDataProjects := make(map[uuid.UUID]struct{})
	if dependencies.objectDataOwners != nil {
		owners, err := dependencies.objectDataOwners.GetByBacnetObjectIDs(
			dispatchCtx,
			[]uuid.UUID{bacnetObjectID},
		)
		if err != nil {
			errors = append(errors, fmt.Errorf(
				"resolve BACnet object ObjectData collaboration projects: %w",
				err,
			))
		} else {
			objectDataProjects = objectDataProjectIDs(bacnetObjectID, owners)
		}
	}

	projectIDs := sortedUpdateProjectIDs(groupedFieldDevices, objectDataProjects)
	for _, projectID := range projectIDs {
		envelope := appcollaboration.Envelope{
			SchemaVersion: appcollaboration.SchemaVersionV1,
			EventID:       dependencies.newID(),
			OperationID:   operationID,
			CorrelationID: operationID,
			ProjectID:     projectID,
			ActorID:       actorID,
			OccurredAt:    occurredAt,
		}
		if _, ok := objectDataProjects[projectID]; ok {
			command := appcollaboration.FacilityHierarchyRefreshRequired{
				Envelope:    envelope,
				Scope:       appcollaboration.FacilityScopeProject,
				FullRefresh: true,
			}
			if err := dependencies.dispatcher.Dispatch(dispatchCtx, command); err != nil {
				errors = append(errors, fmt.Errorf(
					"dispatch ObjectData-owned BACnet object mutation for project %s: %w",
					projectID,
					err,
				))
			}
			continue
		}

		command := appcollaboration.BacnetObjectUpdated{
			Envelope:       envelope,
			BacnetObjectID: bacnetObjectID,
			FieldDeviceIDs: append([]uuid.UUID(nil), groupedFieldDevices[projectID]...),
		}
		if err := dependencies.dispatcher.Dispatch(dispatchCtx, command); err != nil {
			errors = append(errors, fmt.Errorf(
				"dispatch BACnet object mutation for project %s: %w",
				projectID,
				err,
			))
		}
	}

	return projectIDs, errors
}
