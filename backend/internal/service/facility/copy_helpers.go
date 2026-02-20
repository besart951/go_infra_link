package facility

import (
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
	newSpec := &domainFacility.Specification{
		FieldDeviceID:                             &newFieldDeviceID,
		SpecificationSupplier:                     originalSpec.SpecificationSupplier,
		SpecificationBrand:                        originalSpec.SpecificationBrand,
		SpecificationType:                         originalSpec.SpecificationType,
		AdditionalInfoMotorValve:                  originalSpec.AdditionalInfoMotorValve,
		AdditionalInfoSize:                        originalSpec.AdditionalInfoSize,
		AdditionalInformationInstallationLocation: originalSpec.AdditionalInformationInstallationLocation,
		ElectricalConnectionPH:                    originalSpec.ElectricalConnectionPH,
		ElectricalConnectionACDC:                  originalSpec.ElectricalConnectionACDC,
		ElectricalConnectionAmperage:              originalSpec.ElectricalConnectionAmperage,
		ElectricalConnectionPower:                 originalSpec.ElectricalConnectionPower,
		ElectricalConnectionRotation:              originalSpec.ElectricalConnectionRotation,
	}

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
		newObj := &domainFacility.BacnetObject{
			TextFix:             originalObj.TextFix,
			Description:         originalObj.Description,
			GMSVisible:          originalObj.GMSVisible,
			Optional:            originalObj.Optional,
			TextIndividual:      originalObj.TextIndividual,
			SoftwareType:        originalObj.SoftwareType,
			SoftwareNumber:      originalObj.SoftwareNumber,
			HardwareType:        originalObj.HardwareType,
			HardwareQuantity:    originalObj.HardwareQuantity,
			FieldDeviceID:       &newFieldDeviceID,
			SoftwareReferenceID: originalObj.SoftwareReferenceID,
			StateTextID:         originalObj.StateTextID,
			NotificationClassID: originalObj.NotificationClassID,
			AlarmTypeID:         originalObj.AlarmTypeID,
		}

		if err := bacnetObjectRepo.Create(newObj); err != nil {
			return err
		}
	}

	return nil
}
