package spscontroller

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
)

// MoveCommand represents one explicit SPSController parent transition. The
// compatibility PUT remains combined, so the application derives this command
// from authoritative state inside the transaction.
type MoveCommand struct {
	SPSControllerID      uuid.UUID
	FromControlCabinetID uuid.UUID
	ToControlCabinetID   uuid.UUID
}

func newMoveCommand(
	current *domainFacility.SPSController,
	update UpdateCommand,
) (*MoveCommand, error) {
	if update.ControlCabinetID == nil {
		return nil, nil
	}
	if current == nil || current.ID == uuid.Nil || *update.ControlCabinetID == uuid.Nil {
		return nil, domain.ErrInvalidArgument
	}
	if current.ControlCabinetID == *update.ControlCabinetID {
		return nil, nil
	}
	return &MoveCommand{
		SPSControllerID:      current.ID,
		FromControlCabinetID: current.ControlCabinetID,
		ToControlCabinetID:   *update.ControlCabinetID,
	}, nil
}

func (c MoveCommand) applyTo(controller *domainFacility.SPSController) error {
	if controller == nil || c.SPSControllerID == uuid.Nil || controller.ID != c.SPSControllerID {
		return domain.ErrInvalidArgument
	}
	return controller.MoveToControlCabinet(c.ToControlCabinetID)
}
