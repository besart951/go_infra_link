package facility

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
	"github.com/google/uuid"
)

type SPSControllerSystemType struct {
	domain.Base
	Number          *int
	DocumentName    *string
	SPSControllerID uuid.UUID
	SPSController   SPSController
	SystemTypeID    uuid.UUID
	SystemType      SystemType

	FieldDevices []FieldDevice
}
