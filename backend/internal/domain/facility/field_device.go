package facility

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
	"github.com/besart951/go_infra_link/backend/internal/domain/project"
	"github.com/google/uuid"
)

type FieldDevice struct {
	domain.Base
	BMK                       *string
	Description               *string
	ApparatNr                 *int
	SPSControllerSystemTypeID uuid.UUID
	SPSControllerSystemType   SPSControllerSystemType
	SystemPartID              *uuid.UUID
	SystemPart                *SystemPart
	SpecificationID           *uuid.UUID
	Specification             *Specification
	ProjectID                 *uuid.UUID
	Project                   *project.Project
	ApparatID                 uuid.UUID
	Apparat                   Apparat

	BacnetObjects []BacnetObject
}
