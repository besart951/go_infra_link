package controlcabinet

import (
	"errors"
	"testing"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
)

func TestNewMoveCommandUsesAuthoritativeCurrentBuilding(t *testing.T) {
	cabinetID := uuid.New()
	oldBuildingID := uuid.New()
	newBuildingID := uuid.New()

	command, err := newMoveCommand(
		&domainFacility.ControlCabinet{
			Base:       domain.Base{ID: cabinetID},
			BuildingID: oldBuildingID,
		},
		UpdateCommand{ControlCabinetID: cabinetID, BuildingID: &newBuildingID},
	)
	if err != nil {
		t.Fatalf("derive move: %v", err)
	}
	if command == nil || command.ControlCabinetID != cabinetID ||
		command.FromBuildingID != oldBuildingID || command.ToBuildingID != newBuildingID {
		t.Fatalf("unexpected move command: %+v", command)
	}
}

func TestMoveCommandRejectsMismatchedCabinet(t *testing.T) {
	command := MoveCommand{
		ControlCabinetID: uuid.New(),
		FromBuildingID:   uuid.New(),
		ToBuildingID:     uuid.New(),
	}
	cabinet := &domainFacility.ControlCabinet{Base: domain.Base{ID: uuid.New()}}

	err := command.applyTo(cabinet)

	if !errors.Is(err, domain.ErrInvalidArgument) {
		t.Fatalf("error: got %v, want %v", err, domain.ErrInvalidArgument)
	}
}
