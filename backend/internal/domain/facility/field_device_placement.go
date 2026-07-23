package facility

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
	"github.com/google/uuid"
)

// FieldDevicePlacement is the local placement and uniqueness scope of one
// FieldDevice. Parent/reference existence and cross-entity uniqueness remain
// application concerns because they require repositories.
type FieldDevicePlacement struct {
	SPSControllerSystemTypeID uuid.UUID
	SystemPartID              uuid.UUID
	ApparatID                 uuid.UUID
	ApparatNumber             int
}

func NewFieldDevicePlacement(
	spsControllerSystemTypeID uuid.UUID,
	systemPartID uuid.UUID,
	apparatID uuid.UUID,
	apparatNumber int,
) (FieldDevicePlacement, error) {
	if spsControllerSystemTypeID == uuid.Nil ||
		systemPartID == uuid.Nil ||
		apparatID == uuid.Nil ||
		apparatNumber < 1 ||
		apparatNumber > 99 {
		return FieldDevicePlacement{}, domain.ErrInvalidArgument
	}
	return FieldDevicePlacement{
		SPSControllerSystemTypeID: spsControllerSystemTypeID,
		SystemPartID:              systemPartID,
		ApparatID:                 apparatID,
		ApparatNumber:             apparatNumber,
	}, nil
}

func (f *FieldDevice) Placement() FieldDevicePlacement {
	if f == nil {
		return FieldDevicePlacement{}
	}
	return FieldDevicePlacement{
		SPSControllerSystemTypeID: f.SPSControllerSystemTypeID,
		SystemPartID:              f.SystemPartID,
		ApparatID:                 f.ApparatID,
		ApparatNumber:             f.ApparatNr,
	}
}

// MoveTo applies a validated placement. It deliberately performs no database
// access, project reconciliation, history write, or collaboration dispatch.
func (f *FieldDevice) MoveTo(placement FieldDevicePlacement) error {
	if f == nil {
		return domain.ErrInvalidArgument
	}
	validated, err := NewFieldDevicePlacement(
		placement.SPSControllerSystemTypeID,
		placement.SystemPartID,
		placement.ApparatID,
		placement.ApparatNumber,
	)
	if err != nil {
		return err
	}
	f.SPSControllerSystemTypeID = validated.SPSControllerSystemTypeID
	f.SystemPartID = validated.SystemPartID
	f.ApparatID = validated.ApparatID
	f.ApparatNr = validated.ApparatNumber
	return nil
}
