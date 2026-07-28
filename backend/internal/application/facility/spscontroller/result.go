package spscontroller

import (
	"encoding/json"
	"time"

	appcollaboration "github.com/besart951/go_infra_link/backend/internal/application/collaboration"
	"github.com/besart951/go_infra_link/backend/internal/application/facility/mutation"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	domainHistory "github.com/besart951/go_infra_link/backend/internal/domain/history"
	"github.com/google/uuid"
)

type spsControllerSnapshot struct {
	ID                uuid.UUID `json:"id"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
	ControlCabinetID  uuid.UUID `json:"control_cabinet_id"`
	GADevice          *string   `json:"ga_device"`
	DeviceName        string    `json:"device_name"`
	DeviceDescription *string   `json:"device_description"`
	DeviceLocation    *string   `json:"device_location"`
	IPAddress         *string   `json:"ip_address"`
	Subnet            *string   `json:"subnet"`
	Gateway           *string   `json:"gateway"`
	VLAN              *string   `json:"vlan"`
}

func buildCreateChange(
	after *domainFacility.SPSController,
) (mutation.EntityChange, error) {
	afterJSON, err := json.Marshal(toSnapshot(after))
	if err != nil {
		return mutation.EntityChange{}, err
	}
	parentID := after.ControlCabinetID
	return mutation.EntityChange{
		EntityType: mutation.EntityTypeSPSController,
		EntityID:   after.ID,
		ParentID:   &parentID,
		Action:     domainHistory.ActionCreate,
		After:      json.RawMessage(afterJSON),
	}, nil
}

func buildDeleteChange(
	before *domainFacility.SPSController,
) (mutation.EntityChange, error) {
	beforeJSON, err := json.Marshal(toSnapshot(before))
	if err != nil {
		return mutation.EntityChange{}, err
	}
	parentID := before.ControlCabinetID
	return mutation.EntityChange{
		EntityType: mutation.EntityTypeSPSController,
		EntityID:   before.ID,
		ParentID:   &parentID,
		Action:     domainHistory.ActionDelete,
		Before:     json.RawMessage(beforeJSON),
	}, nil
}

func buildUpdateChange(
	before *domainFacility.SPSController,
	after *domainFacility.SPSController,
	systemTypesReplaced bool,
) (mutation.EntityChange, error) {
	beforeJSON, err := json.Marshal(toSnapshot(before))
	if err != nil {
		return mutation.EntityChange{}, err
	}
	afterJSON, err := json.Marshal(toSnapshot(after))
	if err != nil {
		return mutation.EntityChange{}, err
	}
	parentID := after.ControlCabinetID
	return mutation.EntityChange{
		EntityType:    mutation.EntityTypeSPSController,
		EntityID:      after.ID,
		ParentID:      &parentID,
		Action:        domainHistory.ActionUpdate,
		Before:        json.RawMessage(beforeJSON),
		After:         json.RawMessage(afterJSON),
		ChangedFields: changedFields(before, after, systemTypesReplaced),
		Revision:      revisionPointer(after.Revision),
	}, nil
}

func revisionPointer(revision uint64) *uint64 {
	if revision == 0 {
		return nil
	}
	value := revision
	return &value
}

func changedFields(
	before *domainFacility.SPSController,
	after *domainFacility.SPSController,
	systemTypesReplaced bool,
) []mutation.FieldName {
	fields := make([]mutation.FieldName, 0, 10)
	if before.ControlCabinetID != after.ControlCabinetID {
		fields = append(fields, mutation.FieldNameControlCabinet)
	}
	if !equalPointers(before.GADevice, after.GADevice) {
		fields = append(fields, mutation.FieldNameGADevice)
	}
	if before.DeviceName != after.DeviceName {
		fields = append(fields, mutation.FieldNameDeviceName)
	}
	if !equalPointers(before.DeviceDescription, after.DeviceDescription) {
		fields = append(fields, mutation.FieldNameDescription)
	}
	if !equalPointers(before.DeviceLocation, after.DeviceLocation) {
		fields = append(fields, mutation.FieldNameDeviceLocation)
	}
	if !equalPointers(before.IPAddress, after.IPAddress) {
		fields = append(fields, mutation.FieldNameIPAddress)
	}
	if !equalPointers(before.Subnet, after.Subnet) {
		fields = append(fields, mutation.FieldNameSubnet)
	}
	if !equalPointers(before.Gateway, after.Gateway) {
		fields = append(fields, mutation.FieldNameGateway)
	}
	if !equalPointers(before.Vlan, after.Vlan) {
		fields = append(fields, mutation.FieldNameVLAN)
	}
	if systemTypesReplaced {
		fields = append(fields, mutation.FieldNameSystemTypes)
	}
	return fields
}

func toSnapshot(controller *domainFacility.SPSController) spsControllerSnapshot {
	if controller == nil {
		return spsControllerSnapshot{}
	}
	return spsControllerSnapshot{
		ID:                controller.ID,
		CreatedAt:         controller.CreatedAt,
		UpdatedAt:         controller.UpdatedAt,
		ControlCabinetID:  controller.ControlCabinetID,
		GADevice:          clonePointer(controller.GADevice),
		DeviceName:        controller.DeviceName,
		DeviceDescription: clonePointer(controller.DeviceDescription),
		DeviceLocation:    clonePointer(controller.DeviceLocation),
		IPAddress:         clonePointer(controller.IPAddress),
		Subnet:            clonePointer(controller.Subnet),
		Gateway:           clonePointer(controller.Gateway),
		VLAN:              clonePointer(controller.Vlan),
	}
}

func toCollaborationState(
	controller *domainFacility.SPSController,
) appcollaboration.SPSControllerState {
	if controller == nil {
		return appcollaboration.SPSControllerState{}
	}
	return appcollaboration.SPSControllerState{
		ID:                controller.ID,
		Revision:          controller.Revision,
		ControlCabinetID:  controller.ControlCabinetID,
		GADevice:          clonePointer(controller.GADevice),
		DeviceName:        controller.DeviceName,
		DeviceDescription: clonePointer(controller.DeviceDescription),
		DeviceLocation:    clonePointer(controller.DeviceLocation),
		IPAddress:         clonePointer(controller.IPAddress),
		Subnet:            clonePointer(controller.Subnet),
		Gateway:           clonePointer(controller.Gateway),
		VLAN:              clonePointer(controller.Vlan),
		CreatedAt:         controller.CreatedAt,
		UpdatedAt:         controller.UpdatedAt,
	}
}

func equalPointers[T comparable](left, right *T) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	return *left == *right
}
