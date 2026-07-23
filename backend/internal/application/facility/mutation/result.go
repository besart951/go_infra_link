package mutation

import (
	"encoding/json"
	"time"

	domainHistory "github.com/besart951/go_infra_link/backend/internal/domain/history"
	"github.com/google/uuid"
)

type EntityType string

const (
	EntityTypeControlCabinet          EntityType = "control_cabinet"
	EntityTypeSPSController           EntityType = "sps_controller"
	EntityTypeSPSControllerSystemType EntityType = "sps_controller_system_type"
	EntityTypeFieldDevice             EntityType = "field_device"
	EntityTypeProjectControlCabinet   EntityType = "project_control_cabinet"
	EntityTypeProjectSPSController    EntityType = "project_sps_controller"
	EntityTypeProjectFieldDevice      EntityType = "project_field_device"
	EntityTypeBacnetObject            EntityType = "bacnet_object"
	EntityTypeBacnetAlarmValue        EntityType = "bacnet_object_alarm_value"
	EntityTypeObjectData              EntityType = "object_data"
)

type FieldName string

const (
	FieldNameBMK               FieldName = "bmk"
	FieldNameDescription       FieldName = "description"
	FieldNameTextFix           FieldName = "text_fix"
	FieldNameApparatNumber     FieldName = "apparat_nr"
	FieldNameApparat           FieldName = "apparat_id"
	FieldNameSystemPart        FieldName = "system_part_id"
	FieldNameSystemType        FieldName = "sps_controller_system_type_id"
	FieldNameSpecification     FieldName = "specification"
	FieldNameBacnetObjects     FieldName = "bacnet_objects"
	FieldNameControlCabinet    FieldName = "control_cabinet_id"
	FieldNameSPSController     FieldName = "sps_controller_id"
	FieldNameGADevice          FieldName = "ga_device"
	FieldNameDeviceName        FieldName = "device_name"
	FieldNameDeviceLocation    FieldName = "device_location"
	FieldNameIPAddress         FieldName = "ip_address"
	FieldNameSubnet            FieldName = "subnet"
	FieldNameGateway           FieldName = "gateway"
	FieldNameVLAN              FieldName = "vlan"
	FieldNameSystemTypes       FieldName = "system_types"
	FieldNameBuilding          FieldName = "building_id"
	FieldNameCabinetNumber     FieldName = "control_cabinet_nr"
	FieldNameGMSVisible        FieldName = "gms_visible"
	FieldNameOptional          FieldName = "optional"
	FieldNameTextIndividual    FieldName = "text_individual"
	FieldNameSoftwareType      FieldName = "software_type"
	FieldNameSoftwareNumber    FieldName = "software_number"
	FieldNameHardwareType      FieldName = "hardware_type"
	FieldNameHardwareQuantity  FieldName = "hardware_quantity"
	FieldNameFieldDevice       FieldName = "field_device_id"
	FieldNameSoftwareReference FieldName = "software_reference_id"
	FieldNameStateText         FieldName = "state_text_id"
	FieldNameNotificationClass FieldName = "notification_class_id"
	FieldNameAlarmType         FieldName = "alarm_type_id"
	FieldNameProject           FieldName = "project_id"
	FieldNameIsActive          FieldName = "is_active"
)

// Result is the application representation of one completed mutation
// operation. Before and After deliberately use JSON at the history/transport
// seam; domain behavior must not depend on their serialized shape.
type Result struct {
	OperationID uuid.UUID
	BatchID     *uuid.UUID
	ActorID     *uuid.UUID
	OccurredAt  time.Time
	ProjectIDs  []uuid.UUID
	Changes     []EntityChange
}

type EntityChange struct {
	EntityType    EntityType
	EntityID      uuid.UUID
	ParentID      *uuid.UUID
	Action        domainHistory.Action
	Before        json.RawMessage
	After         json.RawMessage
	ChangedFields []FieldName
	Revision      *uint64
}
