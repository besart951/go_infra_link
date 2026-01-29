package facility

import "github.com/google/uuid"

// ControlCabinetDeleteImpact summarizes what would be deleted when removing a control cabinet.
// Counts include only non-deleted records.
type ControlCabinetDeleteImpact struct {
	ControlCabinetID uuid.UUID `json:"control_cabinet_id"`

	SPSControllersCount           int `json:"sps_controllers_count"`
	SPSControllerSystemTypesCount int `json:"sps_controller_system_types_count"`
	FieldDevicesCount             int `json:"field_devices_count"`
	BacnetObjectsCount            int `json:"bacnet_objects_count"`
	SpecificationsCount           int `json:"specifications_count"`
}
