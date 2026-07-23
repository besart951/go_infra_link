package spscontroller

import (
	"errors"
	"testing"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
)

func TestNewMoveCommandClassifiesChangedControlCabinet(t *testing.T) {
	controller := &domainFacility.SPSController{
		Base:             domain.Base{ID: uuid.New()},
		ControlCabinetID: uuid.New(),
	}
	targetID := uuid.New()

	command, err := newMoveCommand(controller, UpdateCommand{
		SPSControllerID:  controller.ID,
		ControlCabinetID: &targetID,
	})
	if err != nil {
		t.Fatalf("build move command: %v", err)
	}
	if command == nil {
		t.Fatal("expected move command")
	}
	if command.FromControlCabinetID != controller.ControlCabinetID ||
		command.ToControlCabinetID != targetID {
		t.Fatalf("unexpected move: %+v", command)
	}
}

func TestNewMoveCommandReturnsNilForUnchangedControlCabinet(t *testing.T) {
	controller := &domainFacility.SPSController{
		Base:             domain.Base{ID: uuid.New()},
		ControlCabinetID: uuid.New(),
	}
	sameID := controller.ControlCabinetID

	command, err := newMoveCommand(controller, UpdateCommand{
		SPSControllerID:  controller.ID,
		ControlCabinetID: &sameID,
	})
	if err != nil {
		t.Fatalf("build unchanged move command: %v", err)
	}
	if command != nil {
		t.Fatalf("expected no move, got %+v", command)
	}
}

func TestNewMoveCommandRejectsInvalidTarget(t *testing.T) {
	controller := &domainFacility.SPSController{
		Base:             domain.Base{ID: uuid.New()},
		ControlCabinetID: uuid.New(),
	}
	invalidID := uuid.Nil

	_, err := newMoveCommand(controller, UpdateCommand{
		SPSControllerID:  controller.ID,
		ControlCabinetID: &invalidID,
	})
	if !errors.Is(err, domain.ErrInvalidArgument) {
		t.Fatalf("error: got %v, want invalid argument", err)
	}
}
