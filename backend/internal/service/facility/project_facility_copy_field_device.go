package facility

import (
	"context"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
)

func (c projectFacilityCopy) copyFieldDevicesForSystemTypes(ctx context.Context, newSystemTypeMap map[uuid.UUID]uuid.UUID) error {
	if len(newSystemTypeMap) == 0 {
		return nil
	}

	type systemTypeCopyPair struct {
		originalID uuid.UUID
		copyID     uuid.UUID
	}

	pairs := make([]systemTypeCopyPair, 0, len(newSystemTypeMap))
	for originalSystemTypeID, copySystemTypeID := range newSystemTypeMap {
		pairs = append(pairs, systemTypeCopyPair{originalID: originalSystemTypeID, copyID: copySystemTypeID})
	}

	for start := 0; start < len(pairs); start += copyFieldDeviceSystemTypeChunkSize {
		end := start + copyFieldDeviceSystemTypeChunkSize
		if end > len(pairs) {
			end = len(pairs)
		}

		chunkMap := make(map[uuid.UUID]uuid.UUID, end-start)
		originalSystemTypeIDs := make([]uuid.UUID, 0, end-start)
		for _, pair := range pairs[start:end] {
			chunkMap[pair.originalID] = pair.copyID
			originalSystemTypeIDs = append(originalSystemTypeIDs, pair.originalID)
		}

		fieldDeviceIDs, err := c.fieldDeviceRepo.GetIDsBySPSControllerSystemTypeIDs(ctx, originalSystemTypeIDs)
		if err != nil {
			return err
		}
		if len(fieldDeviceIDs) == 0 {
			continue
		}

		originalFieldDevices, err := c.fieldDeviceRepo.GetByIds(ctx, fieldDeviceIDs)
		if err != nil {
			return err
		}
		if err := c.copyFieldDevicesWithChildren(ctx, derefSlice(originalFieldDevices), chunkMap); err != nil {
			return err
		}
	}

	return nil
}

func (c projectFacilityCopy) copyFieldDevicesWithChildren(
	ctx context.Context,
	originalFieldDevices []domainFacility.FieldDevice,
	newSystemTypeMap map[uuid.UUID]uuid.UUID,
) error {
	_, err := c.copyFieldDevicesWithChildrenDetailed(ctx, originalFieldDevices, newSystemTypeMap)
	return err
}

func (c projectFacilityCopy) copyFieldDevicesWithChildrenDetailed(
	ctx context.Context,
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

	if err := c.fieldDeviceRepo.BulkCreate(ctx, fieldDeviceCopies, 0); err != nil {
		return nil, err
	}
	for originalID, copyDevice := range originalToCopy {
		originalToCopyID[originalID] = copyDevice.ID
	}

	specs, err := c.specificationRepo.GetByFieldDeviceIDs(ctx, originalIDs)
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
		if err := c.specificationRepo.BulkCreate(ctx, specCopies, 0); err != nil {
			return nil, err
		}
	}

	bacnetObjects, err := c.bacnetObjectRepo.GetByFieldDeviceIDs(ctx, originalIDs)
	if err != nil {
		return nil, err
	}
	if len(bacnetObjects) > 0 {
		if err := c.copyBacnetObjectsWithFieldDeviceMap(ctx, bacnetObjects, originalToCopyID); err != nil {
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

func (c projectFacilityCopy) copyBacnetObjectsWithFieldDeviceMap(
	ctx context.Context,
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
	if err := c.bacnetObjectRepo.BulkCreate(ctx, boCopies, 0); err != nil {
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
		if err := c.bacnetObjectRepo.Update(ctx, newObj); err != nil {
			return err
		}
	}

	return nil
}
