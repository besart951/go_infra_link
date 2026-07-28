package fielddevice

import (
	"encoding/json"
	"sort"
	"time"

	"github.com/besart951/go_infra_link/backend/internal/application/facility/mutation"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	domainHistory "github.com/besart951/go_infra_link/backend/internal/domain/history"
	"github.com/google/uuid"
)

type fieldDeviceSnapshot struct {
	ID                        uuid.UUID  `json:"id"`
	CreatedAt                 time.Time  `json:"created_at"`
	UpdatedAt                 time.Time  `json:"updated_at"`
	BMK                       *string    `json:"bmk"`
	Description               *string    `json:"description"`
	ApparatNr                 int        `json:"apparat_nr"`
	TextIndividuell           *string    `json:"text_individuell"`
	SPSControllerSystemTypeID uuid.UUID  `json:"sps_controller_system_type_id"`
	SystemPartID              uuid.UUID  `json:"system_part_id"`
	SpecificationID           *uuid.UUID `json:"specification_id"`
	ApparatID                 uuid.UUID  `json:"apparat_id"`
}

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

func buildUpdateChanges(
	before *domainFacility.FieldDevice,
	after *domainFacility.FieldDevice,
	beforeBacnet []domainFacility.BacnetObject,
	afterBacnet []domainFacility.BacnetObject,
	bacnetReplaced bool,
) ([]mutation.EntityChange, error) {
	beforeJSON, err := marshalSnapshot(toFieldDeviceSnapshot(before))
	if err != nil {
		return nil, err
	}
	afterJSON, err := marshalSnapshot(toFieldDeviceSnapshot(after))
	if err != nil {
		return nil, err
	}

	parentID := after.SPSControllerSystemTypeID
	changes := []mutation.EntityChange{{
		EntityType:    mutation.EntityTypeFieldDevice,
		EntityID:      after.ID,
		ParentID:      &parentID,
		Action:        domainHistory.ActionUpdate,
		Before:        beforeJSON,
		After:         afterJSON,
		ChangedFields: changedFieldDeviceFields(before, after, bacnetReplaced),
		Revision:      revisionPointer(after.Revision),
	}}

	if !bacnetReplaced {
		return changes, nil
	}

	sort.Slice(beforeBacnet, func(i, j int) bool {
		return beforeBacnet[i].ID.String() < beforeBacnet[j].ID.String()
	})
	for i := range beforeBacnet {
		snapshotJSON, marshalErr := marshalSnapshot(toBacnetObjectSnapshot(&beforeBacnet[i]))
		if marshalErr != nil {
			return nil, marshalErr
		}
		fieldDeviceID := after.ID
		changes = append(changes, mutation.EntityChange{
			EntityType: mutation.EntityTypeBacnetObject,
			EntityID:   beforeBacnet[i].ID,
			ParentID:   &fieldDeviceID,
			Action:     domainHistory.ActionDelete,
			Before:     snapshotJSON,
		})
	}

	sort.Slice(afterBacnet, func(i, j int) bool {
		return afterBacnet[i].ID.String() < afterBacnet[j].ID.String()
	})
	for i := range afterBacnet {
		snapshotJSON, marshalErr := marshalSnapshot(toBacnetObjectSnapshot(&afterBacnet[i]))
		if marshalErr != nil {
			return nil, marshalErr
		}
		fieldDeviceID := after.ID
		changes = append(changes, mutation.EntityChange{
			EntityType: mutation.EntityTypeBacnetObject,
			EntityID:   afterBacnet[i].ID,
			ParentID:   &fieldDeviceID,
			Action:     domainHistory.ActionCreate,
			After:      snapshotJSON,
		})
	}

	return changes, nil
}

func revisionPointer(revision uint64) *uint64 {
	if revision == 0 {
		return nil
	}
	value := revision
	return &value
}

func buildDeleteChange(
	before *domainFacility.FieldDevice,
) (mutation.EntityChange, error) {
	beforeJSON, err := marshalSnapshot(toFieldDeviceSnapshot(before))
	if err != nil {
		return mutation.EntityChange{}, err
	}
	parentID := before.SPSControllerSystemTypeID
	return mutation.EntityChange{
		EntityType: mutation.EntityTypeFieldDevice,
		EntityID:   before.ID,
		ParentID:   &parentID,
		Action:     domainHistory.ActionDelete,
		Before:     beforeJSON,
	}, nil
}

func changedFieldDeviceFields(
	before *domainFacility.FieldDevice,
	after *domainFacility.FieldDevice,
	bacnetReplaced bool,
) []mutation.FieldName {
	fields := make([]mutation.FieldName, 0, 8)
	if !equalPointers(before.BMK, after.BMK) {
		fields = append(fields, mutation.FieldNameBMK)
	}
	if !equalPointers(before.Description, after.Description) {
		fields = append(fields, mutation.FieldNameDescription)
	}
	if !equalPointers(before.TextIndividuell, after.TextIndividuell) {
		fields = append(fields, mutation.FieldNameTextFix)
	}
	if before.ApparatNr != after.ApparatNr {
		fields = append(fields, mutation.FieldNameApparatNumber)
	}
	if before.ApparatID != after.ApparatID {
		fields = append(fields, mutation.FieldNameApparat)
	}
	if before.SystemPartID != after.SystemPartID {
		fields = append(fields, mutation.FieldNameSystemPart)
	}
	if before.SPSControllerSystemTypeID != after.SPSControllerSystemTypeID {
		fields = append(fields, mutation.FieldNameSystemType)
	}
	if bacnetReplaced {
		fields = append(fields, mutation.FieldNameBacnetObjects)
	}
	return fields
}

func toFieldDeviceSnapshot(fieldDevice *domainFacility.FieldDevice) fieldDeviceSnapshot {
	if fieldDevice == nil {
		return fieldDeviceSnapshot{}
	}
	return fieldDeviceSnapshot{
		ID:                        fieldDevice.ID,
		CreatedAt:                 fieldDevice.CreatedAt,
		UpdatedAt:                 fieldDevice.UpdatedAt,
		BMK:                       clonePointer(fieldDevice.BMK),
		Description:               clonePointer(fieldDevice.Description),
		ApparatNr:                 fieldDevice.ApparatNr,
		TextIndividuell:           clonePointer(fieldDevice.TextIndividuell),
		SPSControllerSystemTypeID: fieldDevice.SPSControllerSystemTypeID,
		SystemPartID:              fieldDevice.SystemPartID,
		SpecificationID:           clonePointer(fieldDevice.SpecificationID),
		ApparatID:                 fieldDevice.ApparatID,
	}
}

func toBacnetObjectSnapshot(object *domainFacility.BacnetObject) bacnetObjectSnapshot {
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

func marshalSnapshot(snapshot any) (json.RawMessage, error) {
	encoded, err := json.Marshal(snapshot)
	if err != nil {
		return nil, err
	}
	return json.RawMessage(encoded), nil
}

func cloneFieldDevice(fieldDevice *domainFacility.FieldDevice) *domainFacility.FieldDevice {
	if fieldDevice == nil {
		return nil
	}
	clone := *fieldDevice
	clone.BMK = clonePointer(fieldDevice.BMK)
	clone.Description = clonePointer(fieldDevice.Description)
	clone.TextIndividuell = clonePointer(fieldDevice.TextIndividuell)
	clone.SpecificationID = clonePointer(fieldDevice.SpecificationID)
	if fieldDevice.Specification != nil {
		specification := *fieldDevice.Specification
		clone.Specification = &specification
	}
	clone.BacnetObjects = append([]domainFacility.BacnetObject(nil), fieldDevice.BacnetObjects...)
	return &clone
}

func cloneBacnetObjectSelection(
	objects *[]domainFacility.BacnetObject,
) *[]domainFacility.BacnetObject {
	if objects == nil {
		return nil
	}
	clones := make([]domainFacility.BacnetObject, len(*objects))
	for i := range *objects {
		clones[i] = *cloneBacnetObject(&(*objects)[i])
	}
	return &clones
}

func cloneBacnetObject(object *domainFacility.BacnetObject) *domainFacility.BacnetObject {
	if object == nil {
		return nil
	}
	clone := *object
	clone.Description = clonePointer(object.Description)
	clone.TextIndividual = clonePointer(object.TextIndividual)
	clone.FieldDeviceID = clonePointer(object.FieldDeviceID)
	clone.SoftwareReferenceID = clonePointer(object.SoftwareReferenceID)
	clone.StateTextID = clonePointer(object.StateTextID)
	clone.NotificationClassID = clonePointer(object.NotificationClassID)
	clone.AlarmTypeID = clonePointer(object.AlarmTypeID)
	return &clone
}

func clonePointer[T any](value *T) *T {
	if value == nil {
		return nil
	}
	clone := *value
	return &clone
}

func equalPointers[T comparable](left *T, right *T) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	return *left == *right
}
