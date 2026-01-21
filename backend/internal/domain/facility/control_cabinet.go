package facility

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
	"github.com/google/uuid"
)

type ControlCabinet struct {
	domain.Base
	BuildingID       uuid.UUID
	Building         Building
	ControlCabinetNr *string

	SPSControllers []SPSController
}
