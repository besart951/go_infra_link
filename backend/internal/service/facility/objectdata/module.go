package objectdata

import (
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	domainFieldDevice "github.com/besart951/go_infra_link/backend/internal/domain/facility/fielddevice"
	domainObjectData "github.com/besart951/go_infra_link/backend/internal/domain/facility/objectdata"
)

type Repositories struct {
	ObjectData              domainObjectData.ObjectDataStore
	BacnetObjects           domainObjectData.BacnetObjectStore
	ObjectDataBacnetObjects domainObjectData.ObjectDataBacnetObjectStore
	Apparats                domainFacility.ApparatRepository
	FieldDevices            domainFieldDevice.FieldDeviceStore
	AlarmDefinitions        domainFacility.AlarmDefinitionRepository
	AlarmTypes              domainFacility.AlarmTypeRepository
}
