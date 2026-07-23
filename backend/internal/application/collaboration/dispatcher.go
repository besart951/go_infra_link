package collaboration

import (
	"context"
	"fmt"
)

type CommandDispatcher interface {
	Dispatch(context.Context, Command) error
}

type Dispatcher struct {
	facilityHierarchyRefresh      FacilityHierarchyRefreshHandler
	controlCabinetCreated         ControlCabinetCreatedHandler
	controlCabinetCloned          ControlCabinetClonedHandler
	controlCabinetDeleted         ControlCabinetDeletedHandler
	controlCabinetUpdated         ControlCabinetUpdatedHandler
	controlCabinetMoved           ControlCabinetMovedHandler
	fieldDeviceUpdated            FieldDeviceUpdatedHandler
	fieldDeviceMoved              FieldDeviceMovedHandler
	fieldDeviceDeleted            FieldDeviceDeletedHandler
	fieldDevicesCreated           FieldDevicesCreatedHandler
	bacnetObjectCreated           BacnetObjectCreatedHandler
	bacnetObjectUpdated           BacnetObjectUpdatedHandler
	spsControllerCreated          SPSControllerCreatedHandler
	spsControllerCloned           SPSControllerClonedHandler
	spsControllerSystemTypeCloned SPSControllerSystemTypeClonedHandler
	spsControllerUpdated          SPSControllerUpdatedHandler
	spsControllerMoved            SPSControllerMovedHandler
	spsControllerDeleted          SPSControllerDeletedHandler
}

type DispatcherDependencies struct {
	FacilityHierarchyRefresh      FacilityHierarchyRefreshHandler
	ControlCabinetCreated         ControlCabinetCreatedHandler
	ControlCabinetCloned          ControlCabinetClonedHandler
	ControlCabinetDeleted         ControlCabinetDeletedHandler
	ControlCabinetUpdated         ControlCabinetUpdatedHandler
	ControlCabinetMoved           ControlCabinetMovedHandler
	FieldDeviceUpdated            FieldDeviceUpdatedHandler
	FieldDeviceMoved              FieldDeviceMovedHandler
	FieldDeviceDeleted            FieldDeviceDeletedHandler
	FieldDevicesCreated           FieldDevicesCreatedHandler
	BacnetObjectCreated           BacnetObjectCreatedHandler
	BacnetObjectUpdated           BacnetObjectUpdatedHandler
	SPSControllerCreated          SPSControllerCreatedHandler
	SPSControllerCloned           SPSControllerClonedHandler
	SPSControllerSystemTypeCloned SPSControllerSystemTypeClonedHandler
	SPSControllerUpdated          SPSControllerUpdatedHandler
	SPSControllerMoved            SPSControllerMovedHandler
	SPSControllerDeleted          SPSControllerDeletedHandler
}

func NewDispatcher(deps DispatcherDependencies) *Dispatcher {
	return &Dispatcher{
		facilityHierarchyRefresh:      deps.FacilityHierarchyRefresh,
		controlCabinetCreated:         deps.ControlCabinetCreated,
		controlCabinetCloned:          deps.ControlCabinetCloned,
		controlCabinetDeleted:         deps.ControlCabinetDeleted,
		controlCabinetUpdated:         deps.ControlCabinetUpdated,
		controlCabinetMoved:           deps.ControlCabinetMoved,
		fieldDeviceUpdated:            deps.FieldDeviceUpdated,
		fieldDeviceMoved:              deps.FieldDeviceMoved,
		fieldDeviceDeleted:            deps.FieldDeviceDeleted,
		fieldDevicesCreated:           deps.FieldDevicesCreated,
		bacnetObjectCreated:           deps.BacnetObjectCreated,
		bacnetObjectUpdated:           deps.BacnetObjectUpdated,
		spsControllerCreated:          deps.SPSControllerCreated,
		spsControllerCloned:           deps.SPSControllerCloned,
		spsControllerSystemTypeCloned: deps.SPSControllerSystemTypeCloned,
		spsControllerUpdated:          deps.SPSControllerUpdated,
		spsControllerMoved:            deps.SPSControllerMoved,
		spsControllerDeleted:          deps.SPSControllerDeleted,
	}
}

func (d *Dispatcher) Dispatch(ctx context.Context, command Command) error {
	if d == nil {
		return fmt.Errorf("collaboration dispatcher is not configured")
	}

	switch typed := command.(type) {
	case FacilityHierarchyRefreshRequired:
		if d.facilityHierarchyRefresh == nil {
			return fmt.Errorf("facility hierarchy refresh handler is not configured")
		}
		return d.facilityHierarchyRefresh.HandleFacilityHierarchyRefresh(ctx, typed)
	case ControlCabinetCreated:
		if d.controlCabinetCreated == nil {
			return fmt.Errorf("control cabinet created handler is not configured")
		}
		return d.controlCabinetCreated.HandleControlCabinetCreated(ctx, typed)
	case ControlCabinetCloned:
		if d.controlCabinetCloned == nil {
			return fmt.Errorf("control cabinet cloned handler is not configured")
		}
		return d.controlCabinetCloned.HandleControlCabinetCloned(ctx, typed)
	case ControlCabinetDeleted:
		if d.controlCabinetDeleted == nil {
			return fmt.Errorf("control cabinet deleted handler is not configured")
		}
		return d.controlCabinetDeleted.HandleControlCabinetDeleted(ctx, typed)
	case ControlCabinetUpdated:
		if d.controlCabinetUpdated == nil {
			return fmt.Errorf("control cabinet updated handler is not configured")
		}
		return d.controlCabinetUpdated.HandleControlCabinetUpdated(ctx, typed)
	case ControlCabinetMoved:
		if d.controlCabinetMoved == nil {
			return fmt.Errorf("control cabinet moved handler is not configured")
		}
		return d.controlCabinetMoved.HandleControlCabinetMoved(ctx, typed)
	case FieldDeviceUpdated:
		if d.fieldDeviceUpdated == nil {
			return fmt.Errorf("field device updated handler is not configured")
		}
		return d.fieldDeviceUpdated.HandleFieldDeviceUpdated(ctx, typed)
	case FieldDeviceMoved:
		if d.fieldDeviceMoved == nil {
			return fmt.Errorf("field device moved handler is not configured")
		}
		return d.fieldDeviceMoved.HandleFieldDeviceMoved(ctx, typed)
	case FieldDeviceDeleted:
		if d.fieldDeviceDeleted == nil {
			return fmt.Errorf("field device deleted handler is not configured")
		}
		return d.fieldDeviceDeleted.HandleFieldDeviceDeleted(ctx, typed)
	case FieldDevicesCreated:
		if d.fieldDevicesCreated == nil {
			return fmt.Errorf("field devices created handler is not configured")
		}
		return d.fieldDevicesCreated.HandleFieldDevicesCreated(ctx, typed)
	case BacnetObjectCreated:
		if d.bacnetObjectCreated == nil {
			return fmt.Errorf("BACnet object created handler is not configured")
		}
		return d.bacnetObjectCreated.HandleBacnetObjectCreated(ctx, typed)
	case BacnetObjectUpdated:
		if d.bacnetObjectUpdated == nil {
			return fmt.Errorf("BACnet object updated handler is not configured")
		}
		return d.bacnetObjectUpdated.HandleBacnetObjectUpdated(ctx, typed)
	case SPSControllerCreated:
		if d.spsControllerCreated == nil {
			return fmt.Errorf("SPS controller created handler is not configured")
		}
		return d.spsControllerCreated.HandleSPSControllerCreated(ctx, typed)
	case SPSControllerCloned:
		if d.spsControllerCloned == nil {
			return fmt.Errorf("SPS controller cloned handler is not configured")
		}
		return d.spsControllerCloned.HandleSPSControllerCloned(ctx, typed)
	case SPSControllerSystemTypeCloned:
		if d.spsControllerSystemTypeCloned == nil {
			return fmt.Errorf("SPS controller system type cloned handler is not configured")
		}
		return d.spsControllerSystemTypeCloned.HandleSPSControllerSystemTypeCloned(ctx, typed)
	case SPSControllerUpdated:
		if d.spsControllerUpdated == nil {
			return fmt.Errorf("SPS controller updated handler is not configured")
		}
		return d.spsControllerUpdated.HandleSPSControllerUpdated(ctx, typed)
	case SPSControllerMoved:
		if d.spsControllerMoved == nil {
			return fmt.Errorf("SPS controller moved handler is not configured")
		}
		return d.spsControllerMoved.HandleSPSControllerMoved(ctx, typed)
	case SPSControllerDeleted:
		if d.spsControllerDeleted == nil {
			return fmt.Errorf("SPS controller deleted handler is not configured")
		}
		return d.spsControllerDeleted.HandleSPSControllerDeleted(ctx, typed)
	default:
		return fmt.Errorf("unsupported collaboration command %T", command)
	}
}
