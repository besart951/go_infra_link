package facility

import (
	"errors"
	"testing"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	"github.com/google/uuid"
)

func TestControlCabinetMoveToBuilding(t *testing.T) {
	cabinet := &ControlCabinet{BuildingID: uuid.New()}
	destinationID := uuid.New()

	if err := cabinet.MoveToBuilding(destinationID); err != nil {
		t.Fatalf("move cabinet: %v", err)
	}
	if cabinet.BuildingID != destinationID {
		t.Fatalf("building ID: got %s, want %s", cabinet.BuildingID, destinationID)
	}
}

func TestControlCabinetMoveToBuildingRejectsNilDestination(t *testing.T) {
	cabinet := &ControlCabinet{BuildingID: uuid.New()}
	originalID := cabinet.BuildingID

	err := cabinet.MoveToBuilding(uuid.Nil)

	if !errors.Is(err, domain.ErrInvalidArgument) {
		t.Fatalf("error: got %v, want %v", err, domain.ErrInvalidArgument)
	}
	if cabinet.BuildingID != originalID {
		t.Fatalf("invalid move changed building: got %s, want %s", cabinet.BuildingID, originalID)
	}
}
