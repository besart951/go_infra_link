package facility

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
	"github.com/google/uuid"
)

// MoveToBuilding applies the local part of a ControlCabinet move. Destination
// existence, cabinet-number uniqueness, descendant SPS name regeneration,
// project scope, history, and collaboration require other aggregates or
// infrastructure and remain application concerns.
func (c *ControlCabinet) MoveToBuilding(buildingID uuid.UUID) error {
	if c == nil || buildingID == uuid.Nil {
		return domain.ErrInvalidArgument
	}
	c.BuildingID = buildingID
	return nil
}
