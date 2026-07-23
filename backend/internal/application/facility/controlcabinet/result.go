package controlcabinet

import (
	"encoding/json"
	"time"

	appcollaboration "github.com/besart951/go_infra_link/backend/internal/application/collaboration"
	"github.com/besart951/go_infra_link/backend/internal/application/facility/mutation"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	domainHistory "github.com/besart951/go_infra_link/backend/internal/domain/history"
	"github.com/google/uuid"
)

type controlCabinetSnapshot struct {
	ID               uuid.UUID `json:"id"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
	BuildingID       uuid.UUID `json:"building_id"`
	ControlCabinetNr *string   `json:"control_cabinet_nr"`
}

func buildCreateChange(
	after *domainFacility.ControlCabinet,
) (mutation.EntityChange, error) {
	afterJSON, err := json.Marshal(toSnapshot(after))
	if err != nil {
		return mutation.EntityChange{}, err
	}
	parentID := after.BuildingID
	return mutation.EntityChange{
		EntityType: mutation.EntityTypeControlCabinet,
		EntityID:   after.ID,
		ParentID:   &parentID,
		Action:     domainHistory.ActionCreate,
		After:      json.RawMessage(afterJSON),
	}, nil
}

func buildDeleteChange(
	before *domainFacility.ControlCabinet,
) (mutation.EntityChange, error) {
	beforeJSON, err := json.Marshal(toSnapshot(before))
	if err != nil {
		return mutation.EntityChange{}, err
	}
	parentID := before.BuildingID
	return mutation.EntityChange{
		EntityType: mutation.EntityTypeControlCabinet,
		EntityID:   before.ID,
		ParentID:   &parentID,
		Action:     domainHistory.ActionDelete,
		Before:     json.RawMessage(beforeJSON),
	}, nil
}

func buildUpdateChange(
	before *domainFacility.ControlCabinet,
	after *domainFacility.ControlCabinet,
) (mutation.EntityChange, error) {
	beforeJSON, err := json.Marshal(toSnapshot(before))
	if err != nil {
		return mutation.EntityChange{}, err
	}
	afterJSON, err := json.Marshal(toSnapshot(after))
	if err != nil {
		return mutation.EntityChange{}, err
	}
	parentID := after.BuildingID
	return mutation.EntityChange{
		EntityType:    mutation.EntityTypeControlCabinet,
		EntityID:      after.ID,
		ParentID:      &parentID,
		Action:        domainHistory.ActionUpdate,
		Before:        json.RawMessage(beforeJSON),
		After:         json.RawMessage(afterJSON),
		ChangedFields: changedFields(before, after),
	}, nil
}

func changedFields(
	before *domainFacility.ControlCabinet,
	after *domainFacility.ControlCabinet,
) []mutation.FieldName {
	fields := make([]mutation.FieldName, 0, 2)
	if before.BuildingID != after.BuildingID {
		fields = append(fields, mutation.FieldNameBuilding)
	}
	if !equalPointers(before.ControlCabinetNr, after.ControlCabinetNr) {
		fields = append(fields, mutation.FieldNameCabinetNumber)
	}
	return fields
}

func toSnapshot(cabinet *domainFacility.ControlCabinet) controlCabinetSnapshot {
	if cabinet == nil {
		return controlCabinetSnapshot{}
	}
	return controlCabinetSnapshot{
		ID:               cabinet.ID,
		CreatedAt:        cabinet.CreatedAt,
		UpdatedAt:        cabinet.UpdatedAt,
		BuildingID:       cabinet.BuildingID,
		ControlCabinetNr: clonePointer(cabinet.ControlCabinetNr),
	}
}

func toCollaborationState(
	cabinet *domainFacility.ControlCabinet,
) appcollaboration.ControlCabinetState {
	if cabinet == nil {
		return appcollaboration.ControlCabinetState{}
	}
	return appcollaboration.ControlCabinetState{
		ID:               cabinet.ID,
		BuildingID:       cabinet.BuildingID,
		ControlCabinetNr: clonePointer(cabinet.ControlCabinetNr),
		CreatedAt:        cabinet.CreatedAt,
		UpdatedAt:        cabinet.UpdatedAt,
	}
}

func equalPointers[T comparable](left, right *T) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	return *left == *right
}
