package facility

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
)

// copySpecificationForFieldDevice copies the specification from the original field device to the new field device.
func copySpecificationForFieldDevice(
	specificationRepo domainFacility.SpecificationStore,
	originalFieldDeviceID,
	newFieldDeviceID uuid.UUID,
) error {
	specs, err := specificationRepo.GetByFieldDeviceIDs([]uuid.UUID{originalFieldDeviceID})
	if err != nil {
		return err
	}
	if len(specs) == 0 {
		return nil
	}

	originalSpec := specs[0]
	newSpec := cloneSpecificationForCopy(*originalSpec, newFieldDeviceID)
	return specificationRepo.Create(newSpec)
}

// copyBacnetObjectsForFieldDevice copies all BACnet objects from the original field device
// to the new field device and rewires in-device software references to the copied objects.
func copyBacnetObjectsForFieldDevice(
	bacnetObjectRepo domainFacility.BacnetObjectStore,
	originalFieldDeviceID,
	newFieldDeviceID uuid.UUID,
) error {
	bacnetObjects, err := bacnetObjectRepo.GetByFieldDeviceIDs([]uuid.UUID{originalFieldDeviceID})
	if err != nil {
		return err
	}
	if len(bacnetObjects) == 0 {
		return nil
	}

	return copyBacnetObjectsWithFieldDeviceMap(
		bacnetObjectRepo,
		bacnetObjects,
		map[uuid.UUID]uuid.UUID{originalFieldDeviceID: newFieldDeviceID},
	)
}

// copyFieldDevicesWithChildren bulk-copies field devices along with specifications and BACnet objects.
// It uses the system type map to link each copied field device to the new system type ID.
func copyFieldDevicesWithChildren(
	fieldDeviceRepo domainFacility.FieldDeviceStore,
	specificationRepo domainFacility.SpecificationStore,
	bacnetObjectRepo domainFacility.BacnetObjectStore,
	originalFieldDevices []domainFacility.FieldDevice,
	newSystemTypeMap map[uuid.UUID]uuid.UUID,
) error {
	_, err := copyFieldDevicesWithChildrenDetailed(
		fieldDeviceRepo,
		specificationRepo,
		bacnetObjectRepo,
		originalFieldDevices,
		newSystemTypeMap,
	)
	return err
}

func copyFieldDevicesWithChildrenDetailed(
	fieldDeviceRepo domainFacility.FieldDeviceStore,
	specificationRepo domainFacility.SpecificationStore,
	bacnetObjectRepo domainFacility.BacnetObjectStore,
	originalFieldDevices []domainFacility.FieldDevice,
	newSystemTypeMap map[uuid.UUID]uuid.UUID,
) ([]*domainFacility.FieldDevice, error) {
	if len(originalFieldDevices) == 0 {
		return nil, nil
	}

	fieldDeviceCopies := make([]*domainFacility.FieldDevice, 0, len(originalFieldDevices))
	originalIDs := make([]uuid.UUID, 0, len(originalFieldDevices))
	originalToCopy := make(map[uuid.UUID]*domainFacility.FieldDevice, len(originalFieldDevices))
	originalToCopyID := make(map[uuid.UUID]uuid.UUID, len(originalFieldDevices))

	for _, originalFD := range originalFieldDevices {
		newSystemTypeID, ok := newSystemTypeMap[originalFD.SPSControllerSystemTypeID]
		if !ok {
			continue
		}

		fieldDeviceCopy := cloneFieldDeviceForCopy(originalFD, newSystemTypeID)
		fieldDeviceCopies = append(fieldDeviceCopies, fieldDeviceCopy)
		originalIDs = append(originalIDs, originalFD.ID)
		originalToCopy[originalFD.ID] = fieldDeviceCopy
	}

	if len(fieldDeviceCopies) == 0 {
		return nil, nil
	}

	if err := fieldDeviceRepo.BulkCreate(fieldDeviceCopies, 0); err != nil {
		return nil, err
	}
	for originalID, copyDevice := range originalToCopy {
		originalToCopyID[originalID] = copyDevice.ID
	}

	specs, err := specificationRepo.GetByFieldDeviceIDs(originalIDs)
	if err != nil {
		return nil, err
	}
	if len(specs) > 0 {
		specCopies := make([]*domainFacility.Specification, 0, len(specs))
		for _, originalSpec := range specs {
			if originalSpec.FieldDeviceID == nil {
				continue
			}
			copyDevice, ok := originalToCopy[*originalSpec.FieldDeviceID]
			if !ok {
				continue
			}
			specCopies = append(specCopies, cloneSpecificationForCopy(*originalSpec, copyDevice.ID))
		}
		if err := specificationRepo.BulkCreate(specCopies, 0); err != nil {
			return nil, err
		}
	}

	bacnetObjects, err := bacnetObjectRepo.GetByFieldDeviceIDs(originalIDs)
	if err != nil {
		return nil, err
	}
	if len(bacnetObjects) > 0 {
		if err := copyBacnetObjectsWithFieldDeviceMap(bacnetObjectRepo, bacnetObjects, originalToCopyID); err != nil {
			return nil, err
		}
	}

	return fieldDeviceCopies, nil
}

func cloneFieldDeviceForCopy(originalFD domainFacility.FieldDevice, newSystemTypeID uuid.UUID) *domainFacility.FieldDevice {
	clone := originalFD
	clone.Base = domain.Base{}
	clone.SPSControllerSystemTypeID = newSystemTypeID
	clone.SPSControllerSystemType = domainFacility.SPSControllerSystemType{}
	clone.SpecificationID = nil
	clone.Specification = nil
	clone.BacnetObjects = nil
	clone.SystemPart = domainFacility.SystemPart{}
	clone.Apparat = domainFacility.Apparat{}
	return &clone
}

func cloneSpecificationForCopy(originalSpec domainFacility.Specification, newFieldDeviceID uuid.UUID) *domainFacility.Specification {
	clone := originalSpec
	clone.Base = domain.Base{}
	clone.FieldDeviceID = &newFieldDeviceID
	return &clone
}

func cloneBacnetObjectForCopy(originalObj domainFacility.BacnetObject, newFieldDeviceID uuid.UUID) *domainFacility.BacnetObject {
	clone := originalObj
	clone.Base = domain.Base{}
	clone.FieldDeviceID = &newFieldDeviceID
	clone.FieldDevice = nil
	clone.SoftwareReference = nil
	clone.StateText = nil
	clone.NotificationClass = nil
	clone.AlarmType = nil
	clone.AlarmDefinitionID = nil
	return &clone
}

func copyBacnetObjectsWithFieldDeviceMap(
	bacnetObjectRepo domainFacility.BacnetObjectStore,
	originalObjects []*domainFacility.BacnetObject,
	fieldDeviceIDMap map[uuid.UUID]uuid.UUID,
) error {
	if len(originalObjects) == 0 || len(fieldDeviceIDMap) == 0 {
		return nil
	}

	boCopies := make([]*domainFacility.BacnetObject, 0, len(originalObjects))
	oldToNew := make(map[uuid.UUID]*domainFacility.BacnetObject, len(originalObjects))
	originalByID := make(map[uuid.UUID]*domainFacility.BacnetObject, len(originalObjects))

	for _, originalObj := range originalObjects {
		if originalObj == nil || originalObj.FieldDeviceID == nil {
			continue
		}
		newFieldDeviceID, ok := fieldDeviceIDMap[*originalObj.FieldDeviceID]
		if !ok {
			continue
		}

		newObj := cloneBacnetObjectForCopy(*originalObj, newFieldDeviceID)
		boCopies = append(boCopies, newObj)
		oldToNew[originalObj.ID] = newObj
		originalByID[originalObj.ID] = originalObj
	}

	if len(boCopies) == 0 {
		return nil
	}
	if err := bacnetObjectRepo.BulkCreate(boCopies, 0); err != nil {
		return err
	}

	for originalID, newObj := range oldToNew {
		originalObj := originalByID[originalID]
		if originalObj == nil || originalObj.SoftwareReferenceID == nil {
			continue
		}
		target, ok := oldToNew[*originalObj.SoftwareReferenceID]
		if !ok {
			continue
		}

		targetID := target.ID
		newObj.SoftwareReferenceID = &targetID
		if err := bacnetObjectRepo.Update(newObj); err != nil {
			return err
		}
	}

	return nil
}
