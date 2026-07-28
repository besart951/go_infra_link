package bacnetobject

import (
	"encoding/json"
	"time"

	"github.com/besart951/go_infra_link/backend/internal/application/facility/mutation"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	domainHistory "github.com/besart951/go_infra_link/backend/internal/domain/history"
	"github.com/google/uuid"
)

type bacnetObjectSnapshot struct {
	ID                  uuid.UUID                         `json:"id"`
	CreatedAt           time.Time                         `json:"created_at"`
	UpdatedAt           time.Time                         `json:"updated_at"`
	TextFix             string                            `json:"text_fix"`
	Description         *string                           `json:"description"`
	GMSVisible          bool                              `json:"gms_visible"`
	Optional            bool                              `json:"optional"`
	TextIndividual      *string                           `json:"text_individual"`
	SoftwareType        domainFacility.BacnetSoftwareType `json:"software_type"`
	SoftwareNumber      uint16                            `json:"software_number"`
	HardwareType        domainFacility.BacnetHardwareType `json:"hardware_type"`
	HardwareQuantity    uint8                             `json:"hardware_quantity"`
	FieldDeviceID       *uuid.UUID                        `json:"field_device_id"`
	SoftwareReferenceID *uuid.UUID                        `json:"software_reference_id"`
	StateTextID         *uuid.UUID                        `json:"state_text_id"`
	NotificationClassID *uuid.UUID                        `json:"notification_class_id"`
	AlarmTypeID         *uuid.UUID                        `json:"alarm_type_id"`
}

func buildCreateChange(
	after *domainFacility.BacnetObject,
) (mutation.EntityChange, error) {
	return buildCreateChangeForParent(after, after.FieldDeviceID)
}

func buildCreateChangeForParent(
	after *domainFacility.BacnetObject,
	parentID *uuid.UUID,
) (mutation.EntityChange, error) {
	afterJSON, err := json.Marshal(toSnapshot(after))
	if err != nil {
		return mutation.EntityChange{}, err
	}
	return mutation.EntityChange{
		EntityType: mutation.EntityTypeBacnetObject,
		EntityID:   after.ID,
		ParentID:   clonePointer(parentID),
		Action:     domainHistory.ActionCreate,
		After:      json.RawMessage(afterJSON),
	}, nil
}

func buildUpdateChange(
	before *domainFacility.BacnetObject,
	after *domainFacility.BacnetObject,
	objectDataID *uuid.UUID,
) (mutation.EntityChange, error) {
	beforeJSON, err := json.Marshal(toSnapshot(before))
	if err != nil {
		return mutation.EntityChange{}, err
	}
	afterJSON, err := json.Marshal(toSnapshot(after))
	if err != nil {
		return mutation.EntityChange{}, err
	}
	parentID := clonePointer(after.FieldDeviceID)
	if parentID == nil {
		parentID = clonePointer(objectDataID)
	}
	return mutation.EntityChange{
		EntityType:    mutation.EntityTypeBacnetObject,
		EntityID:      after.ID,
		ParentID:      parentID,
		Action:        domainHistory.ActionUpdate,
		Before:        json.RawMessage(beforeJSON),
		After:         json.RawMessage(afterJSON),
		ChangedFields: changedFields(before, after),
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
	before *domainFacility.BacnetObject,
	after *domainFacility.BacnetObject,
) []mutation.FieldName {
	fields := make([]mutation.FieldName, 0, 15)
	if before.TextFix != after.TextFix {
		fields = append(fields, mutation.FieldNameTextFix)
	}
	if !equalPointers(before.Description, after.Description) {
		fields = append(fields, mutation.FieldNameDescription)
	}
	if before.GMSVisible != after.GMSVisible {
		fields = append(fields, mutation.FieldNameGMSVisible)
	}
	if before.Optional != after.Optional {
		fields = append(fields, mutation.FieldNameOptional)
	}
	if !equalPointers(before.TextIndividual, after.TextIndividual) {
		fields = append(fields, mutation.FieldNameTextIndividual)
	}
	if before.SoftwareType != after.SoftwareType {
		fields = append(fields, mutation.FieldNameSoftwareType)
	}
	if before.SoftwareNumber != after.SoftwareNumber {
		fields = append(fields, mutation.FieldNameSoftwareNumber)
	}
	if before.HardwareType != after.HardwareType {
		fields = append(fields, mutation.FieldNameHardwareType)
	}
	if before.HardwareQuantity != after.HardwareQuantity {
		fields = append(fields, mutation.FieldNameHardwareQuantity)
	}
	if !equalPointers(before.FieldDeviceID, after.FieldDeviceID) {
		fields = append(fields, mutation.FieldNameFieldDevice)
	}
	if !equalPointers(before.SoftwareReferenceID, after.SoftwareReferenceID) {
		fields = append(fields, mutation.FieldNameSoftwareReference)
	}
	if !equalPointers(before.StateTextID, after.StateTextID) {
		fields = append(fields, mutation.FieldNameStateText)
	}
	if !equalPointers(before.NotificationClassID, after.NotificationClassID) {
		fields = append(fields, mutation.FieldNameNotificationClass)
	}
	if !equalPointers(before.AlarmTypeID, after.AlarmTypeID) {
		fields = append(fields, mutation.FieldNameAlarmType)
	}
	return fields
}

func toSnapshot(object *domainFacility.BacnetObject) bacnetObjectSnapshot {
	if object == nil {
		return bacnetObjectSnapshot{}
	}
	return bacnetObjectSnapshot{
		ID:                  object.ID,
		CreatedAt:           object.CreatedAt,
		UpdatedAt:           object.UpdatedAt,
		TextFix:             object.TextFix,
		Description:         clonePointer(object.Description),
		GMSVisible:          object.GMSVisible,
		Optional:            object.Optional,
		TextIndividual:      clonePointer(object.TextIndividual),
		SoftwareType:        object.SoftwareType,
		SoftwareNumber:      object.SoftwareNumber,
		HardwareType:        object.HardwareType,
		HardwareQuantity:    object.HardwareQuantity,
		FieldDeviceID:       clonePointer(object.FieldDeviceID),
		SoftwareReferenceID: clonePointer(object.SoftwareReferenceID),
		StateTextID:         clonePointer(object.StateTextID),
		NotificationClassID: clonePointer(object.NotificationClassID),
		AlarmTypeID:         clonePointer(object.AlarmTypeID),
	}
}

func equalPointers[T comparable](left, right *T) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	return *left == *right
}
