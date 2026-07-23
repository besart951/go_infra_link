package realtime

import (
	"context"
	"encoding/json"

	appcollaboration "github.com/besart951/go_infra_link/backend/internal/application/collaboration"
	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
)

type projectCollaborationRefreshPublisher interface {
	BroadcastRefreshRequest(
		projectID uuid.UUID,
		actorID *uuid.UUID,
		scope string,
		entityIDs []string,
	)
	BroadcastControlCabinetDelta(
		projectID uuid.UUID,
		actorID *uuid.UUID,
		controlCabinet domainFacility.ControlCabinet,
	)
	BroadcastSPSControllerDelta(
		projectID uuid.UUID,
		actorID *uuid.UUID,
		spsController domainFacility.SPSController,
	)
	BroadcastFieldDeviceDelta(
		projectID uuid.UUID,
		actorID *uuid.UUID,
		fieldDevices []map[string]any,
	)
}

const (
	collaborationFieldDeviceDeltaBudgetBytes   = 24 * 1024
	collaborationFieldDeviceDeltaEnvelopeSlack = 256
)

func (a *CollaborationCommandAdapter) PublishControlCabinetUpdated(
	ctx context.Context,
	command appcollaboration.ControlCabinetUpdated,
) error {
	return a.publishControlCabinetDelta(ctx, command.Envelope, command.ControlCabinet)
}

func (a *CollaborationCommandAdapter) PublishControlCabinetCreated(
	ctx context.Context,
	command appcollaboration.ControlCabinetCreated,
) error {
	return a.publishControlCabinetDelta(ctx, command.Envelope, command.ControlCabinet)
}

func (a *CollaborationCommandAdapter) PublishControlCabinetCloned(
	ctx context.Context,
	command appcollaboration.ControlCabinetCloned,
) error {
	return a.publishControlCabinetDelta(ctx, command.Envelope, command.ControlCabinet)
}

func (a *CollaborationCommandAdapter) PublishControlCabinetDeleted(
	ctx context.Context,
	command appcollaboration.ControlCabinetDeleted,
) error {
	return a.publishControlCabinetRefresh(ctx, command.Envelope, command.ControlCabinetID)
}

func (a *CollaborationCommandAdapter) PublishControlCabinetMoved(
	ctx context.Context,
	command appcollaboration.ControlCabinetMoved,
) error {
	return a.publishControlCabinetDelta(ctx, command.Envelope, command.ControlCabinet)
}

type CollaborationCommandAdapter struct {
	publisher projectCollaborationRefreshPublisher
}

func NewCollaborationCommandAdapter(
	publisher projectCollaborationRefreshPublisher,
) *CollaborationCommandAdapter {
	return &CollaborationCommandAdapter{publisher: publisher}
}

func (a *CollaborationCommandAdapter) PublishFacilityHierarchyRefresh(
	_ context.Context,
	command appcollaboration.FacilityHierarchyRefreshRequired,
) error {
	if a == nil || a.publisher == nil {
		return nil
	}

	entityIDs := make([]string, 0, len(command.EntityIDs))
	if !command.FullRefresh {
		for _, id := range command.EntityIDs {
			if id != uuid.Nil {
				entityIDs = append(entityIDs, id.String())
			}
		}
	}

	a.publisher.BroadcastRefreshRequest(
		command.ProjectID,
		command.ActorID,
		string(command.Scope),
		entityIDs,
	)
	return nil
}

func (a *CollaborationCommandAdapter) PublishFieldDeviceUpdated(
	ctx context.Context,
	command appcollaboration.FieldDeviceUpdated,
) error {
	return a.publishFieldDeviceRefresh(ctx, command.Envelope, []uuid.UUID{command.FieldDeviceID})
}

func (a *CollaborationCommandAdapter) PublishFieldDeviceMoved(
	ctx context.Context,
	command appcollaboration.FieldDeviceMoved,
) error {
	return a.publishFieldDeviceRefresh(ctx, command.Envelope, []uuid.UUID{command.FieldDeviceID})
}

func (a *CollaborationCommandAdapter) PublishFieldDeviceDeleted(
	ctx context.Context,
	command appcollaboration.FieldDeviceDeleted,
) error {
	return a.publishFieldDeviceRefresh(ctx, command.Envelope, []uuid.UUID{command.FieldDeviceID})
}

func (a *CollaborationCommandAdapter) PublishFieldDevicesCreated(
	_ context.Context,
	command appcollaboration.FieldDevicesCreated,
) error {
	if a == nil || a.publisher == nil || len(command.FieldDevices) == 0 {
		return nil
	}

	fieldDevices := fieldDeviceDeltaPayload(command.FieldDevices)
	if len(fieldDevices) > projectCollaborationMaxFieldDeviceDeltas ||
		!fieldDeviceDeltaFitsV1Budget(command.Envelope, fieldDevices) {
		a.publisher.BroadcastRefreshRequest(
			command.ProjectID,
			command.ActorID,
			string(appcollaboration.FacilityScopeFieldDevice),
			nil,
		)
		return nil
	}

	a.publisher.BroadcastFieldDeviceDelta(
		command.ProjectID,
		command.ActorID,
		fieldDevices,
	)
	return nil
}

func fieldDeviceDeltaPayload(states []appcollaboration.FieldDeviceState) []map[string]any {
	items := make([]map[string]any, 0, len(states))
	for _, state := range states {
		apparatNumber := state.ApparatNumber
		systemPartID := state.SystemPartID
		items = append(items, map[string]any{
			"id":                            state.ID,
			"bmk":                           clonePointer(state.BMK),
			"description":                   clonePointer(state.Description),
			"text_fix":                      clonePointer(state.TextFix),
			"apparat_nr":                    &apparatNumber,
			"sps_controller_system_type_id": state.SPSControllerSystemTypeID,
			"system_part_id":                &systemPartID,
			"specification_id":              clonePointer(state.SpecificationID),
			"apparat_id":                    state.ApparatID,
			"created_at":                    state.CreatedAt,
			"updated_at":                    state.UpdatedAt,
		})
	}
	return items
}

func fieldDeviceDeltaFitsV1Budget(
	envelope appcollaboration.Envelope,
	fieldDevices []map[string]any,
) bool {
	actorID := ""
	if envelope.ActorID != nil {
		actorID = envelope.ActorID.String()
	}
	encoded, err := json.Marshal(projectCollaborationEntityDeltaMessage{
		Type:         projectCollaborationMessageEntityDelta,
		ProjectID:    envelope.ProjectID,
		Scope:        projectCollaborationRefreshScopeFieldDevice,
		ActorID:      actorID,
		FieldDevices: fieldDevices,
		At:           envelope.OccurredAt,
	})
	return err == nil &&
		len(encoded)+collaborationFieldDeviceDeltaEnvelopeSlack <=
			collaborationFieldDeviceDeltaBudgetBytes
}

func (a *CollaborationCommandAdapter) PublishBacnetObjectUpdated(
	ctx context.Context,
	command appcollaboration.BacnetObjectUpdated,
) error {
	return a.publishFieldDeviceRefresh(ctx, command.Envelope, command.FieldDeviceIDs)
}

func (a *CollaborationCommandAdapter) PublishBacnetObjectCreated(
	ctx context.Context,
	command appcollaboration.BacnetObjectCreated,
) error {
	return a.publishFieldDeviceRefresh(ctx, command.Envelope, []uuid.UUID{command.FieldDeviceID})
}

func (a *CollaborationCommandAdapter) PublishSPSControllerUpdated(
	ctx context.Context,
	command appcollaboration.SPSControllerUpdated,
) error {
	return a.publishSPSControllerRefresh(ctx, command.Envelope, command.SPSControllerID)
}

func (a *CollaborationCommandAdapter) PublishSPSControllerCreated(
	ctx context.Context,
	command appcollaboration.SPSControllerCreated,
) error {
	return a.publishSPSControllerDelta(ctx, command.Envelope, command.SPSController)
}

func (a *CollaborationCommandAdapter) PublishSPSControllerCloned(
	ctx context.Context,
	command appcollaboration.SPSControllerCloned,
) error {
	return a.publishSPSControllerDelta(ctx, command.Envelope, command.SPSController)
}

func (a *CollaborationCommandAdapter) PublishSPSControllerSystemTypeCloned(
	ctx context.Context,
	command appcollaboration.SPSControllerSystemTypeCloned,
) error {
	return a.publishSPSControllerRefresh(ctx, command.Envelope, command.SPSControllerID)
}

func (a *CollaborationCommandAdapter) publishSPSControllerDelta(
	_ context.Context,
	envelope appcollaboration.Envelope,
	state appcollaboration.SPSControllerState,
) error {
	if a == nil || a.publisher == nil {
		return nil
	}
	a.publisher.BroadcastSPSControllerDelta(
		envelope.ProjectID,
		envelope.ActorID,
		domainFacility.SPSController{
			Base: domain.Base{
				ID:        state.ID,
				CreatedAt: state.CreatedAt,
				UpdatedAt: state.UpdatedAt,
			},
			ControlCabinetID:  state.ControlCabinetID,
			GADevice:          clonePointer(state.GADevice),
			DeviceName:        state.DeviceName,
			DeviceDescription: clonePointer(state.DeviceDescription),
			DeviceLocation:    clonePointer(state.DeviceLocation),
			IPAddress:         clonePointer(state.IPAddress),
			Subnet:            clonePointer(state.Subnet),
			Gateway:           clonePointer(state.Gateway),
			Vlan:              clonePointer(state.VLAN),
		},
	)
	return nil
}

func clonePointer[T any](value *T) *T {
	if value == nil {
		return nil
	}
	clone := *value
	return &clone
}

func (a *CollaborationCommandAdapter) PublishSPSControllerMoved(
	ctx context.Context,
	command appcollaboration.SPSControllerMoved,
) error {
	return a.publishSPSControllerRefresh(ctx, command.Envelope, command.SPSControllerID)
}

func (a *CollaborationCommandAdapter) PublishSPSControllerDeleted(
	ctx context.Context,
	command appcollaboration.SPSControllerDeleted,
) error {
	return a.publishSPSControllerRefresh(ctx, command.Envelope, command.SPSControllerID)
}

func (a *CollaborationCommandAdapter) publishFieldDeviceRefresh(
	_ context.Context,
	envelope appcollaboration.Envelope,
	fieldDeviceIDs []uuid.UUID,
) error {
	if a == nil || a.publisher == nil {
		return nil
	}

	entityIDs := make([]string, 0, len(fieldDeviceIDs))
	for _, fieldDeviceID := range fieldDeviceIDs {
		if fieldDeviceID != uuid.Nil {
			entityIDs = append(entityIDs, fieldDeviceID.String())
		}
	}
	a.publisher.BroadcastRefreshRequest(
		envelope.ProjectID,
		envelope.ActorID,
		string(appcollaboration.FacilityScopeFieldDevice),
		entityIDs,
	)
	return nil
}

func (a *CollaborationCommandAdapter) publishControlCabinetDelta(
	_ context.Context,
	envelope appcollaboration.Envelope,
	state appcollaboration.ControlCabinetState,
) error {
	if a == nil || a.publisher == nil {
		return nil
	}

	var cabinetNumber *string
	if state.ControlCabinetNr != nil {
		value := *state.ControlCabinetNr
		cabinetNumber = &value
	}
	a.publisher.BroadcastControlCabinetDelta(
		envelope.ProjectID,
		envelope.ActorID,
		domainFacility.ControlCabinet{
			Base: domain.Base{
				ID:        state.ID,
				CreatedAt: state.CreatedAt,
				UpdatedAt: state.UpdatedAt,
			},
			BuildingID:       state.BuildingID,
			ControlCabinetNr: cabinetNumber,
		},
	)
	return nil
}

func (a *CollaborationCommandAdapter) publishControlCabinetRefresh(
	_ context.Context,
	envelope appcollaboration.Envelope,
	controlCabinetID uuid.UUID,
) error {
	if a == nil || a.publisher == nil {
		return nil
	}

	entityIDs := []string(nil)
	if controlCabinetID != uuid.Nil {
		entityIDs = []string{controlCabinetID.String()}
	}
	a.publisher.BroadcastRefreshRequest(
		envelope.ProjectID,
		envelope.ActorID,
		string(appcollaboration.FacilityScopeControlCabinet),
		entityIDs,
	)
	return nil
}

func (a *CollaborationCommandAdapter) publishSPSControllerRefresh(
	_ context.Context,
	envelope appcollaboration.Envelope,
	spsControllerID uuid.UUID,
) error {
	if a == nil || a.publisher == nil {
		return nil
	}

	entityIDs := []string(nil)
	if spsControllerID != uuid.Nil {
		entityIDs = []string{spsControllerID.String()}
	}
	a.publisher.BroadcastRefreshRequest(
		envelope.ProjectID,
		envelope.ActorID,
		string(appcollaboration.FacilityScopeSPSController),
		entityIDs,
	)
	return nil
}
