package controlcabinet

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
)

// MoveCommand represents one explicit ControlCabinet parent transition. The
// compatibility PUT remains combined, so the application derives this command
// from authoritative state inside the transaction.
type MoveCommand struct {
	ControlCabinetID uuid.UUID
	FromBuildingID   uuid.UUID
	ToBuildingID     uuid.UUID
}

func newMoveCommand(
	current *domainFacility.ControlCabinet,
	update UpdateCommand,
) (*MoveCommand, error) {
	if update.BuildingID == nil {
		return nil, nil
	}
	if current == nil || current.ID == uuid.Nil || *update.BuildingID == uuid.Nil {
		return nil, domain.ErrInvalidArgument
	}
	if current.BuildingID == *update.BuildingID {
		return nil, nil
	}
	return &MoveCommand{
		ControlCabinetID: current.ID,
		FromBuildingID:   current.BuildingID,
		ToBuildingID:     *update.BuildingID,
	}, nil
}

func (c MoveCommand) applyTo(cabinet *domainFacility.ControlCabinet) error {
	if cabinet == nil || c.ControlCabinetID == uuid.Nil || cabinet.ID != c.ControlCabinetID {
		return domain.ErrInvalidArgument
	}
	return cabinet.MoveToBuilding(c.ToBuildingID)
}
