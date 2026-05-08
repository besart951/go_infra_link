package alarm

import (
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	domainObjectData "github.com/besart951/go_infra_link/backend/internal/domain/facility/objectdata"
)

type Repositories struct {
	AlarmDefinitions        domainFacility.AlarmDefinitionRepository
	Units                   domainFacility.UnitRepository
	AlarmFields             domainFacility.AlarmFieldRepository
	AlarmTypes              domainFacility.AlarmTypeRepository
	AlarmTypeFields         domainFacility.AlarmTypeFieldRepository
	BacnetObjectAlarmValues domainFacility.BacnetObjectAlarmValueRepository
	BacnetObjects           domainObjectData.BacnetObjectStore
}
