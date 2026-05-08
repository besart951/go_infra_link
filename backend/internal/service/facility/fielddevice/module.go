package fielddevice

import (
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	domainFieldDevice "github.com/besart951/go_infra_link/backend/internal/domain/facility/fielddevice"
	domainHierarchy "github.com/besart951/go_infra_link/backend/internal/domain/facility/hierarchy"
	domainObjectData "github.com/besart951/go_infra_link/backend/internal/domain/facility/objectdata"
)

type Repositories struct {
	FieldDevices             domainFieldDevice.FieldDeviceStore
	SPSControllerSystemTypes domainHierarchy.SPSControllerSystemTypeStore
	SystemTypes              domainFacility.SystemTypeRepository
	Apparats                 domainFacility.ApparatRepository
	SystemParts              domainFacility.SystemPartRepository
	Specifications           domainFieldDevice.SpecificationStore
	BacnetObjects            domainObjectData.BacnetObjectStore
	ObjectData               domainObjectData.ObjectDataStore
	AlarmTypes               domainFacility.AlarmTypeRepository
	BacnetObjectAlarmValues  domainFacility.BacnetObjectAlarmValueRepository
}
