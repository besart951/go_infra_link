package facility

import (
	"strings"

	domainUser "github.com/besart951/go_infra_link/backend/internal/domain/user"
)

// facilityResourceDefinition is the single HTTP-facing catalog for Facility
// resources. It keeps realtime names, read permissions, and reference-cache
// invalidation together so a new resource cannot silently update only one of
// those projections.
type facilityResourceDefinition struct {
	name                      string
	routeSegment              string
	readPermission            string
	invalidatesReferenceCache bool
}

var facilityResourceCatalog = []facilityResourceDefinition{
	{name: "buildings", routeSegment: "buildings", readPermission: domainUser.PermissionBuildingRead},
	{name: "system_types", routeSegment: "system-types", readPermission: domainUser.PermissionSystemTypeRead},
	{name: "system_parts", routeSegment: "system-parts", readPermission: domainUser.PermissionSystemPartRead, invalidatesReferenceCache: true},
	{name: "apparats", routeSegment: "apparats", readPermission: domainUser.PermissionApparatRead, invalidatesReferenceCache: true},
	{name: "control_cabinets", routeSegment: "control-cabinets", readPermission: domainUser.PermissionControlCabinetRead},
	{name: "sps_controllers", routeSegment: "sps-controllers", readPermission: domainUser.PermissionSPSControllerRead},
	{name: "sps_controller_system_types", routeSegment: "sps-controller-system-types", readPermission: domainUser.PermissionSPSControllerSystemTypeRead},
	{name: "field_devices", routeSegment: "field-devices", readPermission: domainUser.PermissionFieldDeviceRead},
	{name: "bacnet_objects", routeSegment: "bacnet-objects", readPermission: domainUser.PermissionBacnetObjectRead},
	{name: "object_data", routeSegment: "object-data", readPermission: domainUser.PermissionObjectDataRead},
	{name: "state_texts", routeSegment: "state-texts", readPermission: domainUser.PermissionStateTextRead},
	{name: "notification_classes", routeSegment: "notification-classes", readPermission: domainUser.PermissionNotificationClassRead},
	{name: "alarm_definitions", routeSegment: "alarm-definitions", readPermission: domainUser.PermissionAlarmDefinitionRead},
	{name: "alarm_types", routeSegment: "alarm-types", readPermission: domainUser.PermissionAlarmTypeRead},
	{name: "alarm_type_fields", routeSegment: "alarm-type-fields", readPermission: domainUser.PermissionAlarmFieldRead},
	{name: "alarm_fields", routeSegment: "alarm-fields", readPermission: domainUser.PermissionAlarmFieldRead},
	{name: "units", routeSegment: "alarm-units", readPermission: domainUser.PermissionUnitRead},
}

func facilityResourceForRoute(path string) (facilityResourceDefinition, bool) {
	if strings.HasPrefix(path, "/alarm-types/:id/fields") {
		return facilityResourceByName("alarm_type_fields")
	}
	if path == "/imports/field-devices" {
		return facilityResourceByName("field_devices")
	}

	segment := strings.TrimPrefix(path, "/")
	if slash := strings.IndexByte(segment, '/'); slash >= 0 {
		segment = segment[:slash]
	}
	for _, definition := range facilityResourceCatalog {
		if definition.routeSegment == segment {
			return definition, true
		}
	}
	return facilityResourceDefinition{}, false
}

func facilityResourceByName(name string) (facilityResourceDefinition, bool) {
	for _, definition := range facilityResourceCatalog {
		if definition.name == name {
			return definition, true
		}
	}
	return facilityResourceDefinition{}, false
}
