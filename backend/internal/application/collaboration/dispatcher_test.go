package collaboration

import (
	"context"
	"testing"

	"github.com/google/uuid"
)

type refreshHandlerSpy struct {
	commands []FacilityHierarchyRefreshRequired
}

type controlCabinetUpdatedHandlerSpy struct {
	commands []ControlCabinetUpdated
}

type controlCabinetCreatedHandlerSpy struct {
	commands []ControlCabinetCreated
}

type controlCabinetClonedHandlerSpy struct {
	commands []ControlCabinetCloned
}

type controlCabinetDeletedHandlerSpy struct {
	commands []ControlCabinetDeleted
}

func (s *controlCabinetDeletedHandlerSpy) HandleControlCabinetDeleted(
	_ context.Context,
	command ControlCabinetDeleted,
) error {
	s.commands = append(s.commands, command)
	return nil
}

func (s *controlCabinetClonedHandlerSpy) HandleControlCabinetCloned(
	_ context.Context,
	command ControlCabinetCloned,
) error {
	s.commands = append(s.commands, command)
	return nil
}

func (s *controlCabinetCreatedHandlerSpy) HandleControlCabinetCreated(
	_ context.Context,
	command ControlCabinetCreated,
) error {
	s.commands = append(s.commands, command)
	return nil
}

func (s *controlCabinetUpdatedHandlerSpy) HandleControlCabinetUpdated(
	_ context.Context,
	command ControlCabinetUpdated,
) error {
	s.commands = append(s.commands, command)
	return nil
}

type controlCabinetMovedHandlerSpy struct {
	commands []ControlCabinetMoved
}

func (s *controlCabinetMovedHandlerSpy) HandleControlCabinetMoved(
	_ context.Context,
	command ControlCabinetMoved,
) error {
	s.commands = append(s.commands, command)
	return nil
}

func (s *refreshHandlerSpy) HandleFacilityHierarchyRefresh(
	_ context.Context,
	command FacilityHierarchyRefreshRequired,
) error {
	s.commands = append(s.commands, command)
	return nil
}

type fieldDeviceUpdatedHandlerSpy struct {
	commands []FieldDeviceUpdated
}

func (s *fieldDeviceUpdatedHandlerSpy) HandleFieldDeviceUpdated(
	_ context.Context,
	command FieldDeviceUpdated,
) error {
	s.commands = append(s.commands, command)
	return nil
}

type fieldDeviceMovedHandlerSpy struct {
	commands []FieldDeviceMoved
}

type fieldDeviceDeletedHandlerSpy struct {
	commands []FieldDeviceDeleted
}

type fieldDevicesCreatedHandlerSpy struct {
	commands []FieldDevicesCreated
}

func (s *fieldDevicesCreatedHandlerSpy) HandleFieldDevicesCreated(
	_ context.Context,
	command FieldDevicesCreated,
) error {
	s.commands = append(s.commands, command)
	return nil
}

func (s *fieldDeviceDeletedHandlerSpy) HandleFieldDeviceDeleted(
	_ context.Context,
	command FieldDeviceDeleted,
) error {
	s.commands = append(s.commands, command)
	return nil
}

type bacnetObjectUpdatedHandlerSpy struct {
	commands []BacnetObjectUpdated
}

type bacnetObjectCreatedHandlerSpy struct {
	commands []BacnetObjectCreated
}

func (s *bacnetObjectCreatedHandlerSpy) HandleBacnetObjectCreated(
	_ context.Context,
	command BacnetObjectCreated,
) error {
	s.commands = append(s.commands, command)
	return nil
}

func (s *bacnetObjectUpdatedHandlerSpy) HandleBacnetObjectUpdated(
	_ context.Context,
	command BacnetObjectUpdated,
) error {
	s.commands = append(s.commands, command)
	return nil
}

type spsControllerUpdatedHandlerSpy struct {
	commands []SPSControllerUpdated
}

type spsControllerCreatedHandlerSpy struct {
	commands []SPSControllerCreated
}

type spsControllerClonedHandlerSpy struct {
	commands []SPSControllerCloned
}

type spsControllerSystemTypeClonedHandlerSpy struct {
	commands []SPSControllerSystemTypeCloned
}

func (s *spsControllerCreatedHandlerSpy) HandleSPSControllerCreated(
	_ context.Context,
	command SPSControllerCreated,
) error {
	s.commands = append(s.commands, command)
	return nil
}

func (s *spsControllerClonedHandlerSpy) HandleSPSControllerCloned(
	_ context.Context,
	command SPSControllerCloned,
) error {
	s.commands = append(s.commands, command)
	return nil
}

func (s *spsControllerSystemTypeClonedHandlerSpy) HandleSPSControllerSystemTypeCloned(
	_ context.Context,
	command SPSControllerSystemTypeCloned,
) error {
	s.commands = append(s.commands, command)
	return nil
}

func (s *spsControllerUpdatedHandlerSpy) HandleSPSControllerUpdated(
	_ context.Context,
	command SPSControllerUpdated,
) error {
	s.commands = append(s.commands, command)
	return nil
}

type spsControllerMovedHandlerSpy struct {
	commands []SPSControllerMoved
}

type spsControllerDeletedHandlerSpy struct {
	commands []SPSControllerDeleted
}

func (s *spsControllerMovedHandlerSpy) HandleSPSControllerMoved(
	_ context.Context,
	command SPSControllerMoved,
) error {
	s.commands = append(s.commands, command)
	return nil
}

func (s *spsControllerDeletedHandlerSpy) HandleSPSControllerDeleted(
	_ context.Context,
	command SPSControllerDeleted,
) error {
	s.commands = append(s.commands, command)
	return nil
}

func (s *fieldDeviceMovedHandlerSpy) HandleFieldDeviceMoved(
	_ context.Context,
	command FieldDeviceMoved,
) error {
	s.commands = append(s.commands, command)
	return nil
}

func TestDispatcherRoutesFacilityHierarchyRefreshThroughSingleHandler(t *testing.T) {
	handler := &refreshHandlerSpy{}
	dispatcher := NewDispatcher(DispatcherDependencies{
		FacilityHierarchyRefresh: handler,
	})
	command := FacilityHierarchyRefreshRequired{
		Envelope: Envelope{
			EventID:   uuid.New(),
			ProjectID: uuid.New(),
		},
		Scope: FacilityScopeFieldDevice,
	}

	if err := dispatcher.Dispatch(context.Background(), command); err != nil {
		t.Fatalf("Dispatch returned error: %v", err)
	}
	if len(handler.commands) != 1 {
		t.Fatalf("expected one routed command, got %d", len(handler.commands))
	}
	if handler.commands[0].EventID != command.EventID {
		t.Fatalf("expected event %s, got %s", command.EventID, handler.commands[0].EventID)
	}
}

func TestDispatcherRoutesControlCabinetUpdatedThroughSingleHandler(t *testing.T) {
	handler := &controlCabinetUpdatedHandlerSpy{}
	dispatcher := NewDispatcher(DispatcherDependencies{
		ControlCabinetUpdated: handler,
	})
	command := ControlCabinetUpdated{
		Envelope:       Envelope{EventID: uuid.New(), ProjectID: uuid.New()},
		ControlCabinet: ControlCabinetState{ID: uuid.New(), BuildingID: uuid.New()},
	}

	if err := dispatcher.Dispatch(context.Background(), command); err != nil {
		t.Fatalf("Dispatch returned error: %v", err)
	}
	if len(handler.commands) != 1 ||
		handler.commands[0].ControlCabinet.ID != command.ControlCabinet.ID {
		t.Fatalf("unexpected routed command: %+v", handler.commands)
	}
}

func TestDispatcherRoutesControlCabinetCreatedThroughSingleHandler(t *testing.T) {
	handler := &controlCabinetCreatedHandlerSpy{}
	dispatcher := NewDispatcher(DispatcherDependencies{
		ControlCabinetCreated: handler,
	})
	command := ControlCabinetCreated{
		Envelope:       Envelope{EventID: uuid.New(), ProjectID: uuid.New()},
		ControlCabinet: ControlCabinetState{ID: uuid.New(), BuildingID: uuid.New()},
	}

	if err := dispatcher.Dispatch(context.Background(), command); err != nil {
		t.Fatalf("Dispatch returned error: %v", err)
	}
	if len(handler.commands) != 1 ||
		handler.commands[0].ControlCabinet.ID != command.ControlCabinet.ID {
		t.Fatalf("unexpected routed command: %+v", handler.commands)
	}
}

func TestDispatcherRoutesControlCabinetClonedThroughSingleHandler(t *testing.T) {
	handler := &controlCabinetClonedHandlerSpy{}
	dispatcher := NewDispatcher(DispatcherDependencies{
		ControlCabinetCloned: handler,
	})
	command := ControlCabinetCloned{
		Envelope:               Envelope{EventID: uuid.New(), ProjectID: uuid.New()},
		SourceControlCabinetID: uuid.New(),
		ControlCabinet:         ControlCabinetState{ID: uuid.New(), BuildingID: uuid.New()},
	}

	if err := dispatcher.Dispatch(context.Background(), command); err != nil {
		t.Fatalf("Dispatch returned error: %v", err)
	}
	if len(handler.commands) != 1 ||
		handler.commands[0].SourceControlCabinetID != command.SourceControlCabinetID ||
		handler.commands[0].ControlCabinet.ID != command.ControlCabinet.ID {
		t.Fatalf("unexpected routed command: %+v", handler.commands)
	}
}

func TestDispatcherRoutesControlCabinetDeletedThroughSingleHandler(t *testing.T) {
	handler := &controlCabinetDeletedHandlerSpy{}
	dispatcher := NewDispatcher(DispatcherDependencies{
		ControlCabinetDeleted: handler,
	})
	command := ControlCabinetDeleted{
		Envelope:         Envelope{EventID: uuid.New(), ProjectID: uuid.New()},
		ControlCabinetID: uuid.New(),
		BuildingID:       uuid.New(),
	}

	if err := dispatcher.Dispatch(context.Background(), command); err != nil {
		t.Fatalf("Dispatch returned error: %v", err)
	}
	if len(handler.commands) != 1 ||
		handler.commands[0].ControlCabinetID != command.ControlCabinetID ||
		handler.commands[0].BuildingID != command.BuildingID {
		t.Fatalf("unexpected routed command: %+v", handler.commands)
	}
}

func TestDispatcherRoutesControlCabinetMovedThroughSingleHandler(t *testing.T) {
	handler := &controlCabinetMovedHandlerSpy{}
	dispatcher := NewDispatcher(DispatcherDependencies{
		ControlCabinetMoved: handler,
	})
	command := ControlCabinetMoved{
		Envelope:       Envelope{EventID: uuid.New(), ProjectID: uuid.New()},
		ControlCabinet: ControlCabinetState{ID: uuid.New(), BuildingID: uuid.New()},
		FromBuildingID: uuid.New(),
		ToBuildingID:   uuid.New(),
	}

	if err := dispatcher.Dispatch(context.Background(), command); err != nil {
		t.Fatalf("Dispatch returned error: %v", err)
	}
	if len(handler.commands) != 1 ||
		handler.commands[0].FromBuildingID != command.FromBuildingID ||
		handler.commands[0].ToBuildingID != command.ToBuildingID {
		t.Fatalf("unexpected routed command: %+v", handler.commands)
	}
}

func TestDispatcherRoutesFieldDeviceUpdatedThroughSingleHandler(t *testing.T) {
	handler := &fieldDeviceUpdatedHandlerSpy{}
	dispatcher := NewDispatcher(DispatcherDependencies{
		FieldDeviceUpdated: handler,
	})
	command := FieldDeviceUpdated{
		Envelope: Envelope{
			EventID:   uuid.New(),
			ProjectID: uuid.New(),
		},
		FieldDeviceID: uuid.New(),
	}

	if err := dispatcher.Dispatch(context.Background(), command); err != nil {
		t.Fatalf("Dispatch returned error: %v", err)
	}
	if len(handler.commands) != 1 {
		t.Fatalf("expected one routed command, got %d", len(handler.commands))
	}
	if handler.commands[0].FieldDeviceID != command.FieldDeviceID {
		t.Fatalf(
			"expected field device %s, got %s",
			command.FieldDeviceID,
			handler.commands[0].FieldDeviceID,
		)
	}
}

func TestDispatcherRoutesFieldDeviceMovedThroughSingleHandler(t *testing.T) {
	handler := &fieldDeviceMovedHandlerSpy{}
	dispatcher := NewDispatcher(DispatcherDependencies{
		FieldDeviceMoved: handler,
	})
	command := FieldDeviceMoved{
		Envelope: Envelope{
			EventID:   uuid.New(),
			ProjectID: uuid.New(),
		},
		FieldDeviceID:                 uuid.New(),
		FromSPSControllerSystemTypeID: uuid.New(),
		ToSPSControllerSystemTypeID:   uuid.New(),
	}

	if err := dispatcher.Dispatch(context.Background(), command); err != nil {
		t.Fatalf("Dispatch returned error: %v", err)
	}
	if len(handler.commands) != 1 {
		t.Fatalf("expected one routed command, got %d", len(handler.commands))
	}
	if handler.commands[0].FromSPSControllerSystemTypeID !=
		command.FromSPSControllerSystemTypeID ||
		handler.commands[0].ToSPSControllerSystemTypeID !=
			command.ToSPSControllerSystemTypeID {
		t.Fatalf("unexpected move command: %+v", handler.commands[0])
	}
}

func TestDispatcherRoutesFieldDeviceDeletedThroughSingleHandler(t *testing.T) {
	handler := &fieldDeviceDeletedHandlerSpy{}
	dispatcher := NewDispatcher(DispatcherDependencies{
		FieldDeviceDeleted: handler,
	})
	command := FieldDeviceDeleted{
		Envelope:                  Envelope{EventID: uuid.New(), ProjectID: uuid.New()},
		FieldDeviceID:             uuid.New(),
		SPSControllerSystemTypeID: uuid.New(),
	}

	if err := dispatcher.Dispatch(context.Background(), command); err != nil {
		t.Fatalf("Dispatch returned error: %v", err)
	}
	if len(handler.commands) != 1 ||
		handler.commands[0].FieldDeviceID != command.FieldDeviceID ||
		handler.commands[0].SPSControllerSystemTypeID != command.SPSControllerSystemTypeID {
		t.Fatalf("unexpected routed command: %+v", handler.commands)
	}
}

func TestDispatcherRoutesFieldDevicesCreatedThroughSingleHandler(t *testing.T) {
	handler := &fieldDevicesCreatedHandlerSpy{}
	dispatcher := NewDispatcher(DispatcherDependencies{
		FieldDevicesCreated: handler,
	})
	command := FieldDevicesCreated{
		Envelope: Envelope{EventID: uuid.New(), ProjectID: uuid.New()},
		FieldDevices: []FieldDeviceState{{
			ID:                        uuid.New(),
			SPSControllerSystemTypeID: uuid.New(),
		}},
	}

	if err := dispatcher.Dispatch(context.Background(), command); err != nil {
		t.Fatalf("Dispatch returned error: %v", err)
	}
	if len(handler.commands) != 1 ||
		len(handler.commands[0].FieldDevices) != 1 ||
		handler.commands[0].FieldDevices[0].ID != command.FieldDevices[0].ID {
		t.Fatalf("unexpected routed command: %+v", handler.commands)
	}
}

func TestDispatcherRoutesBacnetObjectUpdatedThroughSingleHandler(t *testing.T) {
	handler := &bacnetObjectUpdatedHandlerSpy{}
	dispatcher := NewDispatcher(DispatcherDependencies{
		BacnetObjectUpdated: handler,
	})
	command := BacnetObjectUpdated{
		Envelope:       Envelope{EventID: uuid.New(), ProjectID: uuid.New()},
		BacnetObjectID: uuid.New(),
		FieldDeviceIDs: []uuid.UUID{uuid.New(), uuid.New()},
	}

	if err := dispatcher.Dispatch(context.Background(), command); err != nil {
		t.Fatalf("Dispatch returned error: %v", err)
	}
	if len(handler.commands) != 1 ||
		handler.commands[0].BacnetObjectID != command.BacnetObjectID ||
		len(handler.commands[0].FieldDeviceIDs) != 2 {
		t.Fatalf("unexpected routed command: %+v", handler.commands)
	}
}

func TestDispatcherRoutesBacnetObjectCreatedThroughSingleHandler(t *testing.T) {
	handler := &bacnetObjectCreatedHandlerSpy{}
	dispatcher := NewDispatcher(DispatcherDependencies{
		BacnetObjectCreated: handler,
	})
	command := BacnetObjectCreated{
		Envelope:       Envelope{EventID: uuid.New(), ProjectID: uuid.New()},
		BacnetObjectID: uuid.New(),
		FieldDeviceID:  uuid.New(),
	}

	if err := dispatcher.Dispatch(context.Background(), command); err != nil {
		t.Fatalf("Dispatch returned error: %v", err)
	}
	if len(handler.commands) != 1 ||
		handler.commands[0].BacnetObjectID != command.BacnetObjectID ||
		handler.commands[0].FieldDeviceID != command.FieldDeviceID {
		t.Fatalf("unexpected routed command: %+v", handler.commands)
	}
}

func TestDispatcherRoutesSPSControllerUpdatedThroughSingleHandler(t *testing.T) {
	handler := &spsControllerUpdatedHandlerSpy{}
	dispatcher := NewDispatcher(DispatcherDependencies{
		SPSControllerUpdated: handler,
	})
	command := SPSControllerUpdated{
		Envelope:        Envelope{EventID: uuid.New(), ProjectID: uuid.New()},
		SPSControllerID: uuid.New(),
	}

	if err := dispatcher.Dispatch(context.Background(), command); err != nil {
		t.Fatalf("Dispatch returned error: %v", err)
	}
	if len(handler.commands) != 1 ||
		handler.commands[0].SPSControllerID != command.SPSControllerID {
		t.Fatalf("unexpected routed command: %+v", handler.commands)
	}
}

func TestDispatcherRoutesSPSControllerCreatedThroughSingleHandler(t *testing.T) {
	handler := &spsControllerCreatedHandlerSpy{}
	dispatcher := NewDispatcher(DispatcherDependencies{
		SPSControllerCreated: handler,
	})
	command := SPSControllerCreated{
		Envelope:      Envelope{EventID: uuid.New(), ProjectID: uuid.New()},
		SPSController: SPSControllerState{ID: uuid.New(), ControlCabinetID: uuid.New()},
	}

	if err := dispatcher.Dispatch(context.Background(), command); err != nil {
		t.Fatalf("Dispatch returned error: %v", err)
	}
	if len(handler.commands) != 1 ||
		handler.commands[0].SPSController.ID != command.SPSController.ID {
		t.Fatalf("unexpected routed command: %+v", handler.commands)
	}
}

func TestDispatcherRoutesSPSControllerMovedThroughSingleHandler(t *testing.T) {
	handler := &spsControllerMovedHandlerSpy{}
	dispatcher := NewDispatcher(DispatcherDependencies{
		SPSControllerMoved: handler,
	})
	command := SPSControllerMoved{
		Envelope:             Envelope{EventID: uuid.New(), ProjectID: uuid.New()},
		SPSControllerID:      uuid.New(),
		FromControlCabinetID: uuid.New(),
		ToControlCabinetID:   uuid.New(),
	}

	if err := dispatcher.Dispatch(context.Background(), command); err != nil {
		t.Fatalf("Dispatch returned error: %v", err)
	}
	if len(handler.commands) != 1 ||
		handler.commands[0].FromControlCabinetID != command.FromControlCabinetID ||
		handler.commands[0].ToControlCabinetID != command.ToControlCabinetID {
		t.Fatalf("unexpected routed command: %+v", handler.commands)
	}
}

func TestDispatcherRoutesSPSControllerClonedThroughSingleHandler(t *testing.T) {
	handler := &spsControllerClonedHandlerSpy{}
	dispatcher := NewDispatcher(DispatcherDependencies{
		SPSControllerCloned: handler,
	})
	command := SPSControllerCloned{
		Envelope:              Envelope{EventID: uuid.New(), ProjectID: uuid.New()},
		SourceSPSControllerID: uuid.New(),
		SPSController:         SPSControllerState{ID: uuid.New(), ControlCabinetID: uuid.New()},
	}

	if err := dispatcher.Dispatch(context.Background(), command); err != nil {
		t.Fatalf("Dispatch returned error: %v", err)
	}
	if len(handler.commands) != 1 ||
		handler.commands[0].SourceSPSControllerID != command.SourceSPSControllerID ||
		handler.commands[0].SPSController.ID != command.SPSController.ID {
		t.Fatalf("unexpected routed command: %+v", handler.commands)
	}
}

func TestDispatcherRoutesSPSControllerSystemTypeClonedThroughSingleHandler(t *testing.T) {
	handler := &spsControllerSystemTypeClonedHandlerSpy{}
	dispatcher := NewDispatcher(DispatcherDependencies{
		SPSControllerSystemTypeCloned: handler,
	})
	command := SPSControllerSystemTypeCloned{
		Envelope:                        Envelope{EventID: uuid.New(), ProjectID: uuid.New()},
		SourceSPSControllerSystemTypeID: uuid.New(),
		SPSControllerSystemTypeID:       uuid.New(),
		SPSControllerID:                 uuid.New(),
	}

	if err := dispatcher.Dispatch(context.Background(), command); err != nil {
		t.Fatalf("Dispatch returned error: %v", err)
	}
	if len(handler.commands) != 1 ||
		handler.commands[0].SourceSPSControllerSystemTypeID != command.SourceSPSControllerSystemTypeID ||
		handler.commands[0].SPSControllerSystemTypeID != command.SPSControllerSystemTypeID ||
		handler.commands[0].SPSControllerID != command.SPSControllerID {
		t.Fatalf("unexpected routed command: %+v", handler.commands)
	}
}

func TestDispatcherRoutesSPSControllerDeletedThroughSingleHandler(t *testing.T) {
	handler := &spsControllerDeletedHandlerSpy{}
	dispatcher := NewDispatcher(DispatcherDependencies{
		SPSControllerDeleted: handler,
	})
	command := SPSControllerDeleted{
		Envelope:         Envelope{EventID: uuid.New(), ProjectID: uuid.New()},
		SPSControllerID:  uuid.New(),
		ControlCabinetID: uuid.New(),
	}

	if err := dispatcher.Dispatch(context.Background(), command); err != nil {
		t.Fatalf("Dispatch returned error: %v", err)
	}
	if len(handler.commands) != 1 ||
		handler.commands[0].SPSControllerID != command.SPSControllerID ||
		handler.commands[0].ControlCabinetID != command.ControlCabinetID {
		t.Fatalf("unexpected routed command: %+v", handler.commands)
	}
}

func TestDispatcherRejectsMissingTypedHandler(t *testing.T) {
	dispatcher := NewDispatcher(DispatcherDependencies{})

	err := dispatcher.Dispatch(context.Background(), FacilityHierarchyRefreshRequired{})
	if err == nil {
		t.Fatal("expected missing handler error")
	}
}
