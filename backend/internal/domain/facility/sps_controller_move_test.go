package facility

import (
	"errors"
	"testing"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	"github.com/google/uuid"
)

func TestSPSControllerMoveToControlCabinet(t *testing.T) {
	controller := &SPSController{ControlCabinetID: uuid.New()}
	targetID := uuid.New()

	if err := controller.MoveToControlCabinet(targetID); err != nil {
		t.Fatalf("move SPS controller: %v", err)
	}
	if controller.ControlCabinetID != targetID {
		t.Fatalf("control cabinet: got %s, want %s", controller.ControlCabinetID, targetID)
	}
}

func TestSPSControllerMoveToControlCabinetRejectsInvalidTarget(t *testing.T) {
	controller := &SPSController{ControlCabinetID: uuid.New()}
	originalID := controller.ControlCabinetID

	err := controller.MoveToControlCabinet(uuid.Nil)

	if !errors.Is(err, domain.ErrInvalidArgument) {
		t.Fatalf("error: got %v, want invalid argument", err)
	}
	if controller.ControlCabinetID != originalID {
		t.Fatalf("failed move changed parent to %s", controller.ControlCabinetID)
	}
}
