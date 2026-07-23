package collaboration

import "context"

type ProjectCollaborationPort interface {
	PublishFacilityHierarchyRefresh(context.Context, FacilityHierarchyRefreshRequired) error
	PublishControlCabinetCreated(context.Context, ControlCabinetCreated) error
	PublishControlCabinetCloned(context.Context, ControlCabinetCloned) error
	PublishControlCabinetDeleted(context.Context, ControlCabinetDeleted) error
	PublishControlCabinetUpdated(context.Context, ControlCabinetUpdated) error
	PublishControlCabinetMoved(context.Context, ControlCabinetMoved) error
	PublishFieldDeviceUpdated(context.Context, FieldDeviceUpdated) error
	PublishFieldDeviceMoved(context.Context, FieldDeviceMoved) error
	PublishFieldDeviceDeleted(context.Context, FieldDeviceDeleted) error
	PublishFieldDevicesCreated(context.Context, FieldDevicesCreated) error
	PublishBacnetObjectCreated(context.Context, BacnetObjectCreated) error
	PublishBacnetObjectUpdated(context.Context, BacnetObjectUpdated) error
	PublishSPSControllerCreated(context.Context, SPSControllerCreated) error
	PublishSPSControllerCloned(context.Context, SPSControllerCloned) error
	PublishSPSControllerSystemTypeCloned(context.Context, SPSControllerSystemTypeCloned) error
	PublishSPSControllerUpdated(context.Context, SPSControllerUpdated) error
	PublishSPSControllerMoved(context.Context, SPSControllerMoved) error
	PublishSPSControllerDeleted(context.Context, SPSControllerDeleted) error
}
