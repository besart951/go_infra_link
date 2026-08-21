package facility

import (
	"context"
	"fmt"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
)

// CopyByID creates an independent ObjectData template in the same scope. Its
// Apparat references remain shared while BACnet templates receive fresh IDs
// and internal software references are remapped.
func (s *ObjectDataService) CopyByID(ctx context.Context, id uuid.UUID) (*domainFacility.ObjectData, error) {
	return runWithFacilityTxResult(ctx, s.transaction(), func(txCtx context.Context, txService *ObjectDataService) (*domainFacility.ObjectData, error) {
		original, err := domain.GetByID(txCtx, txService.extRepo, id)
		if err != nil {
			return nil, err
		}
		description, err := txService.nextCopyDescription(txCtx, original)
		if err != nil {
			return nil, err
		}
		apparatIDs, err := txService.GetApparatIDs(txCtx, original.ID)
		if err != nil {
			return nil, err
		}
		apparats, err := txService.template().loadApparats(txCtx, apparatIDs)
		if err != nil {
			return nil, err
		}

		copyTemplate := *original
		copyTemplate.Base = domain.Base{}
		copyTemplate.Description = description
		copyTemplate.Project = nil
		copyTemplate.BacnetObjects = nil
		copyTemplate.Apparats = apparats
		if err := txService.extRepo.Create(txCtx, &copyTemplate); err != nil {
			return nil, err
		}
		if err := copyObjectDataBacnetTemplates(txCtx, txService, original.ID, copyTemplate.ID); err != nil {
			return nil, err
		}
		return domain.GetByID(txCtx, txService.extRepo, copyTemplate.ID)
	})
}

func (s *ObjectDataService) nextCopyDescription(ctx context.Context, source *domainFacility.ObjectData) (string, error) {
	for increment := 1; increment <= 10_000; increment++ {
		candidate := nextIncrementedValue(source.Description, increment, 250)
		exists, err := s.extRepo.ExistsByDescription(ctx, source.ProjectID, candidate, nil)
		if err != nil {
			return "", err
		}
		if !exists {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("object data copy: no conflict-free description available")
}

func copyObjectDataBacnetTemplates(ctx context.Context, service *ObjectDataService, sourceID, targetID uuid.UUID) error {
	ids, err := service.extRepo.GetBacnetObjectIDs(ctx, sourceID)
	if err != nil {
		return err
	}
	if len(ids) == 0 {
		return nil
	}
	originals, err := service.bacnetObjectRepo.GetByIds(ctx, ids)
	if err != nil {
		return err
	}

	copies := make([]*domainFacility.BacnetObject, 0, len(originals))
	copyBySourceID := make(map[uuid.UUID]*domainFacility.BacnetObject, len(originals))
	originalByID := make(map[uuid.UUID]*domainFacility.BacnetObject, len(originals))
	for _, original := range originals {
		if original == nil {
			continue
		}
		copyObject := *original
		copyObject.Base = domain.Base{}
		copyObject.FieldDeviceID = nil
		copyObject.FieldDevice = nil
		copyObject.SoftwareReferenceID = nil
		copyObject.SoftwareReference = nil
		copyObject.StateText = nil
		copyObject.NotificationClass = nil
		copyObject.AlarmType = nil
		copies = append(copies, &copyObject)
		copyBySourceID[original.ID] = &copyObject
		originalByID[original.ID] = original
	}
	if err := service.bacnetObjectRepo.BulkCreate(ctx, copies, 500); err != nil {
		return err
	}
	for _, copyObject := range copies {
		if err := service.objectDataBacnetStore.Add(ctx, targetID, copyObject.ID); err != nil {
			return err
		}
	}
	for sourceObjectID, copyObject := range copyBySourceID {
		original := originalByID[sourceObjectID]
		if original.SoftwareReferenceID == nil {
			continue
		}
		target, ok := copyBySourceID[*original.SoftwareReferenceID]
		if !ok {
			continue
		}
		targetObjectID := target.ID
		copyObject.SoftwareReferenceID = &targetObjectID
		if err := service.bacnetObjectRepo.Update(ctx, copyObject); err != nil {
			return err
		}
	}
	return nil
}
