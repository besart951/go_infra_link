package fielddevice

import (
	"errors"
	"testing"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
)

func TestNewMoveCommandBuildsCompleteTargetFromPartialCompatibilityUpdate(t *testing.T) {
	fieldDevice := &domainFacility.FieldDevice{
		Base:                      domain.Base{ID: uuid.New()},
		SPSControllerSystemTypeID: uuid.New(),
		SystemPartID:              uuid.New(),
		ApparatID:                 uuid.New(),
		ApparatNr:                 7,
	}
	newSystemTypeID := uuid.New()
	newApparatNumber := 8

	command, err := newMoveCommand(fieldDevice, UpdateCommand{
		FieldDeviceID:             fieldDevice.ID,
		SPSControllerSystemTypeID: &newSystemTypeID,
		ApparatNr:                 &newApparatNumber,
	})
	if err != nil {
		t.Fatalf("build move command: %v", err)
	}
	if command == nil {
		t.Fatal("expected a move command")
	}
	if command.From != fieldDevice.Placement() {
		t.Fatalf("from placement: got %+v, want %+v", command.From, fieldDevice.Placement())
	}
	if command.To.SPSControllerSystemTypeID != newSystemTypeID ||
		command.To.SystemPartID != fieldDevice.SystemPartID ||
		command.To.ApparatID != fieldDevice.ApparatID ||
		command.To.ApparatNumber != newApparatNumber {
		t.Fatalf("unexpected target placement: %+v", command.To)
	}
	if !command.movesParent() {
		t.Fatal("expected parent move")
	}
}

func TestNewMoveCommandReturnsNilWhenPlacementIsUnchanged(t *testing.T) {
	fieldDevice := &domainFacility.FieldDevice{
		Base:                      domain.Base{ID: uuid.New()},
		SPSControllerSystemTypeID: uuid.New(),
		SystemPartID:              uuid.New(),
		ApparatID:                 uuid.New(),
		ApparatNr:                 7,
	}
	sameSystemPartID := fieldDevice.SystemPartID

	command, err := newMoveCommand(fieldDevice, UpdateCommand{
		FieldDeviceID: fieldDevice.ID,
		SystemPartID:  &sameSystemPartID,
	})
	if err != nil {
		t.Fatalf("build unchanged move command: %v", err)
	}
	if command != nil {
		t.Fatalf("expected no move command, got %+v", command)
	}
}

func TestNewMoveCommandRejectsInvalidTargetPlacement(t *testing.T) {
	fieldDevice := &domainFacility.FieldDevice{
		Base:                      domain.Base{ID: uuid.New()},
		SPSControllerSystemTypeID: uuid.New(),
		SystemPartID:              uuid.New(),
		ApparatID:                 uuid.New(),
		ApparatNr:                 7,
	}
	invalidNumber := 100

	_, err := newMoveCommand(fieldDevice, UpdateCommand{
		FieldDeviceID: fieldDevice.ID,
		ApparatNr:     &invalidNumber,
	})
	if !errors.Is(err, domain.ErrInvalidArgument) {
		t.Fatalf("error: got %v, want invalid argument", err)
	}
}
