package facility

import (
	"context"
	"fmt"
	"time"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	domainObjectData "github.com/besart951/go_infra_link/backend/internal/domain/facility/objectdata"
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
	originals, err := service.templateStore.ListByObjectDataID(ctx, sourceID)
	if err != nil {
		return err
	}
	ids := make(map[uuid.UUID]uuid.UUID, len(originals))
	for _, original := range originals {
		ids[original.ID] = uuid.New()
	}
	copies := make([]domainObjectData.BacnetObjectTemplate, len(originals))
	for i := range originals {
		copies[i] = copyBacnetTemplate(originals[i], targetID, ids)
	}
	return service.templateStore.Replace(ctx, targetID, copies)
}

func copyBacnetTemplate(source domainObjectData.BacnetObjectTemplate, ownerID uuid.UUID, ids map[uuid.UUID]uuid.UUID) domainObjectData.BacnetObjectTemplate {
	copyTemplate := source
	copyTemplate.ID = ids[source.ID]
	copyTemplate.ObjectDataID = ownerID
	copyTemplate.CreatedAt, copyTemplate.UpdatedAt, copyTemplate.Version = time.Time{}, time.Time{}, 0
	copyTemplate.SoftwareReferenceID = remapOptionalID(source.SoftwareReferenceID, ids)
	copyTemplate.AlarmValues = copyTemplateAlarmValues(source.AlarmValues, copyTemplate.ID)
	return copyTemplate
}

func copyTemplateAlarmValues(values []domainObjectData.BacnetObjectTemplateAlarmValue, templateID uuid.UUID) []domainObjectData.BacnetObjectTemplateAlarmValue {
	copies := make([]domainObjectData.BacnetObjectTemplateAlarmValue, len(values))
	for i := range values {
		copies[i] = values[i]
		copies[i].ID = uuid.Nil
		copies[i].TemplateID = templateID
		copies[i].CreatedAt, copies[i].UpdatedAt, copies[i].Version = time.Time{}, time.Time{}, 0
	}
	return copies
}

func remapOptionalID(source *uuid.UUID, ids map[uuid.UUID]uuid.UUID) *uuid.UUID {
	if source == nil {
		return nil
	}
	target, ok := ids[*source]
	if !ok {
		return nil
	}
	return &target
}
