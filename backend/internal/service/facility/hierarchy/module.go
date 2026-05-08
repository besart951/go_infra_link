package hierarchy

import (
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	domainFieldDevice "github.com/besart951/go_infra_link/backend/internal/domain/facility/fielddevice"
	domainHierarchy "github.com/besart951/go_infra_link/backend/internal/domain/facility/hierarchy"
	domainObjectData "github.com/besart951/go_infra_link/backend/internal/domain/facility/objectdata"
)

type Repositories struct {
	Buildings                domainFacility.BuildingRepository
	ControlCabinets          domainFacility.ControlCabinetRepository
	SPSControllers           domainFacility.SPSControllerRepository
	SystemTypes              domainFacility.SystemTypeRepository
	SPSControllerSystemTypes domainHierarchy.SPSControllerSystemTypeStore
	FieldDevices             domainFieldDevice.FieldDeviceStore
	Specifications           domainFieldDevice.SpecificationStore
	BacnetObjects            domainObjectData.BacnetObjectStore
}
