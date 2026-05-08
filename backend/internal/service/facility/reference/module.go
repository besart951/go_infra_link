package reference

import (
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	domainObjectData "github.com/besart951/go_infra_link/backend/internal/domain/facility/objectdata"
)

type Repositories struct {
	SystemTypes         domainFacility.SystemTypeRepository
	SystemParts         domainFacility.SystemPartRepository
	Apparats            domainFacility.ApparatRepository
	ObjectData          domainObjectData.ObjectDataStore
	StateTexts          domainFacility.StateTextRepository
	NotificationClasses domainFacility.NotificationClassRepository
}
