package facility

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
)

// copySpecificationForFieldDevice copies the specification from the original field device to the new field device
func copySpecificationForFieldDevice(
	specificationRepo domainFacility.SpecificationStore,
	originalFieldDeviceID,
	newFieldDeviceID uuid.UUID,
) error {
	// Get specifications for original field device
	specs, err := specificationRepo.GetByFieldDeviceIDs([]uuid.UUID{originalFieldDeviceID})
	if err != nil {
		return err
	}

	// If no specification exists, nothing to copy
	if len(specs) == 0 {
		return nil
	}

	// Copy the specification
	originalSpec := specs[0]
	newSpec := cloneSpecificationForCopy(*originalSpec, newFieldDeviceID)

	return specificationRepo.Create(newSpec)
}

// copyBacnetObjectsForFieldDevice copies all BACnet objects from the original field device to the new field device
func copyBacnetObjectsForFieldDevice(
	bacnetObjectRepo domainFacility.BacnetObjectStore,
	originalFieldDeviceID,
	newFieldDeviceID uuid.UUID,
) error {
	// Get BACnet objects for original field device
	bacnetObjects, err := bacnetObjectRepo.GetByFieldDeviceIDs([]uuid.UUID{originalFieldDeviceID})
	if err != nil {
		return err
	}

	// If no BACnet objects exist, nothing to copy
	if len(bacnetObjects) == 0 {
		return nil
	}

	// Copy each BACnet object
	for _, originalObj := range bacnetObjects {
		newObj := cloneBacnetObjectForCopy(*originalObj, newFieldDeviceID)

		if err := bacnetObjectRepo.Create(newObj); err != nil {
			return err
		}
	}

	return nil
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
	if len(originalFieldDevices) == 0 {
		return nil
	}

	fieldDeviceCopies := make([]*domainFacility.FieldDevice, 0, len(originalFieldDevices))
	originalIDs := make([]uuid.UUID, 0, len(originalFieldDevices))
	originalToCopy := make(map[uuid.UUID]*domainFacility.FieldDevice, len(originalFieldDevices))

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
		return nil
	}

	if err := fieldDeviceRepo.BulkCreate(fieldDeviceCopies, 0); err != nil {
		return err
	}

	specs, err := specificationRepo.GetByFieldDeviceIDs(originalIDs)
	if err != nil {
		return err
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
			return err
		}
	}

	bacnetObjects, err := bacnetObjectRepo.GetByFieldDeviceIDs(originalIDs)
	if err != nil {
		return err
	}

	if len(bacnetObjects) > 0 {
		boCopies := make([]*domainFacility.BacnetObject, 0, len(bacnetObjects))
		for _, originalObj := range bacnetObjects {
			if originalObj.FieldDeviceID == nil {
				continue
			}
			copyDevice, ok := originalToCopy[*originalObj.FieldDeviceID]
			if !ok {
				continue
			}
			boCopies = append(boCopies, cloneBacnetObjectForCopy(*originalObj, copyDevice.ID))
		}

		if err := bacnetObjectRepo.BulkCreate(boCopies, 0); err != nil {
			return err
		}
	}

	return nil
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
