package bacnetobject

import (
	"context"
	"fmt"
	"time"

	appcollaboration "github.com/besart951/go_infra_link/backend/internal/application/collaboration"
	"github.com/google/uuid"
)

type collaborationDependencies struct {
	projectLinks     ProjectLinkReader
	objectDataOwners ObjectDataOwnerReader
	dispatcher       appcollaboration.CommandDispatcher
	newID            IDGenerator
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
