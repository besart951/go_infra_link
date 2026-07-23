package facility

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
	"github.com/google/uuid"
)

// MoveToControlCabinet applies the local part of an SPSController move. Name
// generation, destination existence, uniqueness, project scope, history, and
// collaboration remain application concerns because they require other
// aggregates or infrastructure.
func (s *SPSController) MoveToControlCabinet(controlCabinetID uuid.UUID) error {
	if s == nil || controlCabinetID == uuid.Nil {
		return domain.ErrInvalidArgument
	}
	s.ControlCabinetID = controlCabinetID
	return nil
}
