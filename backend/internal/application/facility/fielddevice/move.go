package fielddevice

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
)

// MoveCommand represents one explicit FieldDevice placement transition. The
// application update Adapter builds it from the compatibility PUT payload and
// the authoritative state loaded inside the transaction.
type MoveCommand struct {
	FieldDeviceID uuid.UUID
	From          domainFacility.FieldDevicePlacement
	To            domainFacility.FieldDevicePlacement
}

func newMoveCommand(
	current *domainFacility.FieldDevice,
	update UpdateCommand,
) (*MoveCommand, error) {
	if !update.hasPlacementInput() {
		return nil, nil
	}
	if current == nil || current.ID == uuid.Nil {
		return nil, domain.ErrInvalidArgument
	}

	from := current.Placement()
	targetSystemTypeID := from.SPSControllerSystemTypeID
	if update.SPSControllerSystemTypeID != nil {
		targetSystemTypeID = *update.SPSControllerSystemTypeID
	}
	targetSystemPartID := from.SystemPartID
	if update.SystemPartID != nil {
		targetSystemPartID = *update.SystemPartID
	}
	targetApparatID := from.ApparatID
	if update.ApparatID != nil {
		targetApparatID = *update.ApparatID
	}
	targetApparatNumber := from.ApparatNumber
	if update.ApparatNr != nil {
		targetApparatNumber = *update.ApparatNr
	}

	to, err := domainFacility.NewFieldDevicePlacement(
		targetSystemTypeID,
		targetSystemPartID,
		targetApparatID,
		targetApparatNumber,
	)
	if err != nil {
		return nil, err
	}
	if from == to {
		return nil, nil
	}
	return &MoveCommand{
		FieldDeviceID: current.ID,
		From:          from,
		To:            to,
	}, nil
}

func (c UpdateCommand) hasPlacementInput() bool {
	return c.SPSControllerSystemTypeID != nil ||
		c.SystemPartID != nil ||
		c.ApparatID != nil ||
		c.ApparatNr != nil
}

func (c MoveCommand) applyTo(fieldDevice *domainFacility.FieldDevice) error {
	if fieldDevice == nil || c.FieldDeviceID == uuid.Nil || fieldDevice.ID != c.FieldDeviceID {
		return domain.ErrInvalidArgument
	}
	return fieldDevice.MoveTo(c.To)
}

func (c MoveCommand) movesParent() bool {
	return c.From.SPSControllerSystemTypeID != c.To.SPSControllerSystemTypeID
}
