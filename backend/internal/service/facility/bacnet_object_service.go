package facility

import (
	"context"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	domainFieldDevice "github.com/besart951/go_infra_link/backend/internal/domain/facility/fielddevice"
	domainObjectData "github.com/besart951/go_infra_link/backend/internal/domain/facility/objectdata"
	"github.com/google/uuid"
)

type BacnetObjectService struct {
	repo                  domainObjectData.BacnetObjectStore
	fieldDeviceRepo       domainFieldDevice.FieldDeviceStore
	objectDataRepo        domainObjectData.ObjectDataStore
	objectDataBacnetStore domainObjectData.ObjectDataBacnetObjectStore
	alarmDefinitionRepo   domainFacility.AlarmDefinitionRepository
	alarmTypeRepo         domainFacility.AlarmTypeRepository
	tx                    txCoordinator
}

func (s *BacnetObjectService) resolveAlarmBindingForTemplate(ctx context.Context, bacnetObject *domainFacility.BacnetObject) error {
	return s.objectDataTemplate().resolveAlarmBinding(ctx, bacnetObject)
}

func (s *BacnetObjectService) ensureTextFixUniqueForFieldDevice(ctx context.Context, fieldDeviceID uuid.UUID, textFix string, excludeID *uuid.UUID) error {
	items, err := s.repo.GetByFieldDeviceIDs(ctx, []uuid.UUID{fieldDeviceID})
	if err != nil {
		return err
	}
	for _, it := range items {
		if excludeID != nil && it.ID == *excludeID {
			continue
		}
		if it.TextFix == textFix {
			return domain.NewValidationError().AddCode("fielddevice.bacnetobject.text_fix", "unique", "text_fix must be unique within the field device")
		}
	}
	return nil
}

func (s *BacnetObjectService) validateRequiredFields(bacnetObject *domainFacility.BacnetObject, prefix string) error {
	return bacnetObject.Validate(prefix)
}

func NewBacnetObjectService(
	repo domainObjectData.BacnetObjectStore,
	fieldDeviceRepo domainFieldDevice.FieldDeviceStore,
	objectDataRepo domainObjectData.ObjectDataStore,
	objectDataBacnetStore domainObjectData.ObjectDataBacnetObjectStore,
	alarmDefinitionRepo domainFacility.AlarmDefinitionRepository,
	alarmTypeRepo domainFacility.AlarmTypeRepository,
) *BacnetObjectService {
	return &BacnetObjectService{
		repo:                  repo,
		fieldDeviceRepo:       fieldDeviceRepo,
		objectDataRepo:        objectDataRepo,
		objectDataBacnetStore: objectDataBacnetStore,
		alarmDefinitionRepo:   alarmDefinitionRepo,
		alarmTypeRepo:         alarmTypeRepo,
	}
}

func (s *BacnetObjectService) bindTransactions(tx txCoordinator) {
	s.tx = tx
}

func (s *BacnetObjectService) transaction() facilityTx[*BacnetObjectService] {
	return newFacilityTx(s.tx, s, func(services *Services) *BacnetObjectService {
		return services.BacnetObject
	})
}

func (s *BacnetObjectService) objectDataTemplate() objectDataTemplate {
	return objectDataTemplate{
		objectDataRepo:        s.objectDataRepo,
		bacnetObjectRepo:      s.repo,
		objectDataBacnetStore: s.objectDataBacnetStore,
		alarmDefinitionRepo:   s.alarmDefinitionRepo,
		alarmTypeRepo:         s.alarmTypeRepo,
	}
}

func (s *BacnetObjectService) GetByID(ctx context.Context, id uuid.UUID) (*domainFacility.BacnetObject, error) {
	return domain.GetByID(ctx, s.repo, id)
}

func (s *BacnetObjectService) GetByIDs(ctx context.Context, ids []uuid.UUID) ([]*domainFacility.BacnetObject, error) {
	return s.repo.GetByIds(ctx, ids)
}

func (s *BacnetObjectService) DeleteByID(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return domain.ErrInvalidArgument
	}
	if _, err := s.GetByID(ctx, id); err != nil {
		return err
	}
	return s.repo.DeleteByIds(ctx, []uuid.UUID{id})
}

// CreateWithParent creates a bacnet object either for a field device (fieldDeviceID)
// or for an object data template (objectDataID). Exactly one must be provided.
func (s *BacnetObjectService) CreateWithParent(ctx context.Context, bacnetObject *domainFacility.BacnetObject, fieldDeviceID *uuid.UUID, objectDataID *uuid.UUID) error {
	return s.transaction().run(ctx, func(txCtx context.Context, txService *BacnetObjectService) error {
		if (fieldDeviceID == nil && objectDataID == nil) || (fieldDeviceID != nil && objectDataID != nil) {
			return domain.ErrInvalidArgument
		}

		bacnetObject.TextFix = normalizeBacnetTextFix(bacnetObject.TextFix)

		if fieldDeviceID != nil {
			if err := txService.validateRequiredFields(bacnetObject, "fielddevice.bacnetobject"); err != nil {
				return err
			}
		}
		if objectDataID != nil {
			if err := txService.validateRequiredFields(bacnetObject, "objectdata.bacnetobject"); err != nil {
				return err
			}
		}

		if fieldDeviceID != nil {
			if _, err := domain.GetByID(txCtx, txService.fieldDeviceRepo, *fieldDeviceID); err != nil {
				return err
			}
			if err := txService.ensureTextFixUniqueForFieldDevice(txCtx, *fieldDeviceID, bacnetObject.TextFix, nil); err != nil {
				return err
			}
			if err := txService.resolveAlarmBindingForTemplate(txCtx, bacnetObject); err != nil {
				return err
			}
			bacnetObject.FieldDeviceID = fieldDeviceID
			return txService.repo.Create(txCtx, bacnetObject)
		}

		return txService.objectDataTemplate().createBacnetObject(txCtx, *objectDataID, bacnetObject)
	})
}

// Update updates a bacnet object. If objectDataID is provided, it will also attach
// the bacnet object to that object data (template) after validating the object data.
func (s *BacnetObjectService) Update(ctx context.Context, bacnetObject *domainFacility.BacnetObject, objectDataID *uuid.UUID) error {
	return s.transaction().run(ctx, func(txCtx context.Context, txService *BacnetObjectService) error {
		if objectDataID != nil {
			return txService.objectDataTemplate().updateBacnetObject(txCtx, *objectDataID, bacnetObject)
		}

		bacnetObject.TextFix = normalizeBacnetTextFix(bacnetObject.TextFix)

		if bacnetObject.FieldDeviceID != nil {
			if err := txService.validateRequiredFields(bacnetObject, "fielddevice.bacnetobject"); err != nil {
				return err
			}
		}

		if _, err := domain.GetByID(txCtx, txService.repo, bacnetObject.ID); err != nil {
			return err
		}
		if bacnetObject.FieldDeviceID != nil {
			if err := txService.ensureTextFixUniqueForFieldDevice(txCtx, *bacnetObject.FieldDeviceID, bacnetObject.TextFix, &bacnetObject.ID); err != nil {
				return err
			}
		}

		if err := txService.resolveAlarmBindingForTemplate(txCtx, bacnetObject); err != nil {
			return err
		}

		if err := txService.repo.Update(txCtx, bacnetObject); err != nil {
			return err
		}

		return nil
	})
}

// ReplaceForObjectData replaces all bacnet objects for an object data template.
// Existing links are removed and the provided list is created and attached.
func (s *BacnetObjectService) ReplaceForObjectData(ctx context.Context, objectDataID uuid.UUID, inputs []domainFacility.BacnetObject) error {
	return s.transaction().run(ctx, func(txCtx context.Context, txService *BacnetObjectService) error {
		return txService.objectDataTemplate().replaceBacnetObjects(txCtx, objectDataID, inputs)
	})
}
