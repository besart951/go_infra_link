package collaboration

import (
	"context"
	"fmt"
)

type FacilityHierarchyRefreshHandler interface {
	HandleFacilityHierarchyRefresh(context.Context, FacilityHierarchyRefreshRequired) error
}

type ControlCabinetUpdatedHandler interface {
	HandleControlCabinetUpdated(context.Context, ControlCabinetUpdated) error
}

type ControlCabinetCreatedHandler interface {
	HandleControlCabinetCreated(context.Context, ControlCabinetCreated) error
}

type ControlCabinetClonedHandler interface {
	HandleControlCabinetCloned(context.Context, ControlCabinetCloned) error
}

type ControlCabinetDeletedHandler interface {
	HandleControlCabinetDeleted(context.Context, ControlCabinetDeleted) error
}

type ControlCabinetMovedHandler interface {
	HandleControlCabinetMoved(context.Context, ControlCabinetMoved) error
}

type FieldDeviceUpdatedHandler interface {
	HandleFieldDeviceUpdated(context.Context, FieldDeviceUpdated) error
}

type FieldDeviceMovedHandler interface {
	HandleFieldDeviceMoved(context.Context, FieldDeviceMoved) error
}

type FieldDeviceDeletedHandler interface {
	HandleFieldDeviceDeleted(context.Context, FieldDeviceDeleted) error
}

type FieldDevicesCreatedHandler interface {
	HandleFieldDevicesCreated(context.Context, FieldDevicesCreated) error
}

type BacnetObjectUpdatedHandler interface {
	HandleBacnetObjectUpdated(context.Context, BacnetObjectUpdated) error
}

type BacnetObjectCreatedHandler interface {
	HandleBacnetObjectCreated(context.Context, BacnetObjectCreated) error
}

type SPSControllerUpdatedHandler interface {
	HandleSPSControllerUpdated(context.Context, SPSControllerUpdated) error
}

type SPSControllerCreatedHandler interface {
	HandleSPSControllerCreated(context.Context, SPSControllerCreated) error
}

type SPSControllerClonedHandler interface {
	HandleSPSControllerCloned(context.Context, SPSControllerCloned) error
}

type SPSControllerSystemTypeClonedHandler interface {
	HandleSPSControllerSystemTypeCloned(context.Context, SPSControllerSystemTypeCloned) error
}

type SPSControllerMovedHandler interface {
	HandleSPSControllerMoved(context.Context, SPSControllerMoved) error
}

type SPSControllerDeletedHandler interface {
	HandleSPSControllerDeleted(context.Context, SPSControllerDeleted) error
}

type ProjectCommandHandler struct {
	port ProjectCollaborationPort
}

func NewProjectCommandHandler(port ProjectCollaborationPort) *ProjectCommandHandler {
	return &ProjectCommandHandler{port: port}
}

func (h *ProjectCommandHandler) HandleFacilityHierarchyRefresh(
	ctx context.Context,
	command FacilityHierarchyRefreshRequired,
) error {
	if h == nil || h.port == nil {
		return fmt.Errorf("project collaboration port is not configured")
	}
	return h.port.PublishFacilityHierarchyRefresh(ctx, command)
}

func (h *ProjectCommandHandler) HandleControlCabinetUpdated(
	ctx context.Context,
	command ControlCabinetUpdated,
) error {
	if h == nil || h.port == nil {
		return fmt.Errorf("project collaboration port is not configured")
	}
	return h.port.PublishControlCabinetUpdated(ctx, command)
}

func (h *ProjectCommandHandler) HandleControlCabinetCreated(
	ctx context.Context,
	command ControlCabinetCreated,
) error {
	if h == nil || h.port == nil {
		return fmt.Errorf("project collaboration port is not configured")
	}
	return h.port.PublishControlCabinetCreated(ctx, command)
}

func (h *ProjectCommandHandler) HandleControlCabinetCloned(
	ctx context.Context,
	command ControlCabinetCloned,
) error {
	if h == nil || h.port == nil {
		return fmt.Errorf("project collaboration port is not configured")
	}
	return h.port.PublishControlCabinetCloned(ctx, command)
}

func (h *ProjectCommandHandler) HandleControlCabinetDeleted(
	ctx context.Context,
	command ControlCabinetDeleted,
) error {
	if h == nil || h.port == nil {
		return fmt.Errorf("project collaboration port is not configured")
	}
	return h.port.PublishControlCabinetDeleted(ctx, command)
}

func (h *ProjectCommandHandler) HandleControlCabinetMoved(
	ctx context.Context,
	command ControlCabinetMoved,
) error {
	if h == nil || h.port == nil {
		return fmt.Errorf("project collaboration port is not configured")
	}
	return h.port.PublishControlCabinetMoved(ctx, command)
}

func (h *ProjectCommandHandler) HandleFieldDeviceUpdated(
	ctx context.Context,
	command FieldDeviceUpdated,
) error {
	if h == nil || h.port == nil {
		return fmt.Errorf("project collaboration port is not configured")
	}
	return h.port.PublishFieldDeviceUpdated(ctx, command)
}

func (h *ProjectCommandHandler) HandleFieldDeviceMoved(
	ctx context.Context,
	command FieldDeviceMoved,
) error {
	if h == nil || h.port == nil {
		return fmt.Errorf("project collaboration port is not configured")
	}
	return h.port.PublishFieldDeviceMoved(ctx, command)
}

func (h *ProjectCommandHandler) HandleFieldDeviceDeleted(
	ctx context.Context,
	command FieldDeviceDeleted,
) error {
	if h == nil || h.port == nil {
		return fmt.Errorf("project collaboration port is not configured")
	}
	return h.port.PublishFieldDeviceDeleted(ctx, command)
}

func (h *ProjectCommandHandler) HandleFieldDevicesCreated(
	ctx context.Context,
	command FieldDevicesCreated,
) error {
	if h == nil || h.port == nil {
		return fmt.Errorf("project collaboration port is not configured")
	}
	return h.port.PublishFieldDevicesCreated(ctx, command)
}

func (h *ProjectCommandHandler) HandleBacnetObjectUpdated(
	ctx context.Context,
	command BacnetObjectUpdated,
) error {
	if h == nil || h.port == nil {
		return fmt.Errorf("project collaboration port is not configured")
	}
	return h.port.PublishBacnetObjectUpdated(ctx, command)
}

func (h *ProjectCommandHandler) HandleBacnetObjectCreated(
	ctx context.Context,
	command BacnetObjectCreated,
) error {
	if h == nil || h.port == nil {
		return fmt.Errorf("project collaboration port is not configured")
	}
	return h.port.PublishBacnetObjectCreated(ctx, command)
}

func (h *ProjectCommandHandler) HandleSPSControllerUpdated(
	ctx context.Context,
	command SPSControllerUpdated,
) error {
	if h == nil || h.port == nil {
		return fmt.Errorf("project collaboration port is not configured")
	}
	return h.port.PublishSPSControllerUpdated(ctx, command)
}

func (h *ProjectCommandHandler) HandleSPSControllerCreated(
	ctx context.Context,
	command SPSControllerCreated,
) error {
	if h == nil || h.port == nil {
		return fmt.Errorf("project collaboration port is not configured")
	}
	return h.port.PublishSPSControllerCreated(ctx, command)
}

func (h *ProjectCommandHandler) HandleSPSControllerCloned(
	ctx context.Context,
	command SPSControllerCloned,
) error {
	if h == nil || h.port == nil {
		return fmt.Errorf("project collaboration port is not configured")
	}
	return h.port.PublishSPSControllerCloned(ctx, command)
}

func (h *ProjectCommandHandler) HandleSPSControllerSystemTypeCloned(
	ctx context.Context,
	command SPSControllerSystemTypeCloned,
) error {
	if h == nil || h.port == nil {
		return fmt.Errorf("project collaboration port is not configured")
	}
	return h.port.PublishSPSControllerSystemTypeCloned(ctx, command)
}

func (h *ProjectCommandHandler) HandleSPSControllerMoved(
	ctx context.Context,
	command SPSControllerMoved,
) error {
	if h == nil || h.port == nil {
		return fmt.Errorf("project collaboration port is not configured")
	}
	return h.port.PublishSPSControllerMoved(ctx, command)
}

func (h *ProjectCommandHandler) HandleSPSControllerDeleted(
	ctx context.Context,
	command SPSControllerDeleted,
) error {
	if h == nil || h.port == nil {
		return fmt.Errorf("project collaboration port is not configured")
	}
	return h.port.PublishSPSControllerDeleted(ctx, command)
}
