package dto

import "github.com/google/uuid"

type ControlCabinetDeleteImpactResponse struct {
	ControlCabinetID uuid.UUID `json:"control_cabinet_id"`

	SPSControllersCount           int `json:"sps_controllers_count"`
	SPSControllerSystemTypesCount int `json:"sps_controller_system_types_count"`
	FieldDevicesCount             int `json:"field_devices_count"`
	BacnetObjectsCount            int `json:"bacnet_objects_count"`
	SpecificationsCount           int `json:"specifications_count"`
}
