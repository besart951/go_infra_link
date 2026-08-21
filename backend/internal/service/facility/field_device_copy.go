package facility

import (
	"context"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/besart951/go_infra_link/backend/internal/service/changecapture"
	"github.com/google/uuid"
)

// CopyByID creates an independent FieldDevice aggregate while preserving
// references to shared master data. Every owned child is copied in the same
// transaction and internal BACnet references are remapped to the new IDs.
func (s *FieldDeviceService) CopyByID(ctx context.Context, id uuid.UUID) (*domainFacility.FieldDevice, error) {
	return runWithFacilityTxResult(ctx, s.transaction(), func(txCtx context.Context, txService *FieldDeviceService) (*domainFacility.FieldDevice, error) {
		original, err := domain.GetByID(txCtx, txService.repo, id)
		if err != nil {
			return nil, err
		}

		available, err := txService.ListAvailableApparatNumbers(
			txCtx,
			original.SPSControllerSystemTypeID,
			original.SystemPartID,
			original.ApparatID,
		)
		if err != nil {
			return nil, err
		}
		if len(available) == 0 {
			return nil, domain.NewValidationError().Add("fielddevice.apparat_nr", "no free apparat number is available")
		}

		copyDevice := cloneFieldDeviceForCopy(*original, original.SPSControllerSystemTypeID)
		copyDevice.ApparatNr = available[0]
		if err := txService.Validate(txCtx, copyDevice, nil); err != nil {
			return nil, err
		}
		if err := txService.repo.Create(txCtx, copyDevice); err != nil {
			return nil, err
		}

		if err := copyFieldDeviceSpecification(txCtx, txService, original.ID, copyDevice.ID); err != nil {
			return nil, err
		}
		if err := copyFieldDeviceBacnetObjects(txCtx, txService, original.ID, copyDevice.ID); err != nil {
			return nil, err
		}
		if err := txService.recordFieldDeviceChange(txCtx, changecapture.ActionCreated, copyDevice.ID); err != nil {
			return nil, err
		}
		return domain.GetByID(txCtx, txService.repo, copyDevice.ID)
	})
}

func copyFieldDeviceSpecification(ctx context.Context, service *FieldDeviceService, sourceID, targetID uuid.UUID) error {
	specifications, err := service.specificationRepo.GetByFieldDeviceIDs(ctx, []uuid.UUID{sourceID})
	if err != nil {
		return err
	}
	if len(specifications) == 0 || specifications[0] == nil {
		return nil
	}
	return service.specificationRepo.Create(ctx, cloneSpecificationForCopy(*specifications[0], targetID))
}

func copyFieldDeviceBacnetObjects(ctx context.Context, service *FieldDeviceService, sourceID, targetID uuid.UUID) error {
	originals, err := service.bacnetObjectRepo.GetByFieldDeviceIDs(ctx, []uuid.UUID{sourceID})
	if err != nil {
		return err
	}
	if len(originals) == 0 {
		return nil
	}

	copies := make([]*domainFacility.BacnetObject, 0, len(originals))
	copyBySourceID := make(map[uuid.UUID]*domainFacility.BacnetObject, len(originals))
	originalByID := make(map[uuid.UUID]*domainFacility.BacnetObject, len(originals))
	for _, original := range originals {
		if original == nil {
			continue
		}
		copyObject := cloneBacnetObjectForCopy(*original, targetID)
		copies = append(copies, copyObject)
		copyBySourceID[original.ID] = copyObject
		originalByID[original.ID] = original
	}
	if err := service.bacnetObjectRepo.BulkCreate(ctx, copies, 500); err != nil {
		return err
	}

	for sourceID, copyObject := range copyBySourceID {
		original := originalByID[sourceID]
		if original.SoftwareReferenceID != nil {
			if target, ok := copyBySourceID[*original.SoftwareReferenceID]; ok {
				targetID := target.ID
				copyObject.SoftwareReferenceID = &targetID
				if err := service.bacnetObjectRepo.Update(ctx, copyObject); err != nil {
					return err
				}
			}
		}
		if err := copyBacnetAlarmValues(ctx, service, sourceID, copyObject.ID); err != nil {
			return err
		}
	}
	return nil
}

func copyBacnetAlarmValues(ctx context.Context, service *FieldDeviceService, sourceID, targetID uuid.UUID) error {
	if service.bacnetAlarmValueRepo == nil {
		return nil
	}
	values, err := service.bacnetAlarmValueRepo.GetByBacnetObjectID(ctx, sourceID)
	if err != nil {
		return err
	}
	copies := make([]*domainFacility.BacnetObjectAlarmValue, 0, len(values))
	for i := range values {
		copyValue := values[i]
		copyValue.Base = domain.Base{}
		copyValue.BacnetObjectID = targetID
		copyValue.BacnetObject = nil
		copyValue.AlarmTypeField = nil
		copyValue.Unit = nil
		copies = append(copies, &copyValue)
	}
	if len(copies) == 0 {
		return nil
	}
	return service.bacnetAlarmValueRepo.BulkCreate(ctx, copies, 500)
}
