package facility

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
	"github.com/google/uuid"
)

type FieldDevice struct {
	domain.Base
	BMK                       *string
	Description               *string
	ApparatNr                 *int
	SPSControllerSystemTypeID uuid.UUID
	SPSControllerSystemType   SPSControllerSystemType
	SystemPartID              uuid.UUID
	SystemPart                *SystemPart
	SpecificationID           *uuid.UUID
	Specification             *Specification
	ApparatID                 uuid.UUID
	Apparat                   Apparat

	BacnetObjects []BacnetObject
}
