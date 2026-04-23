package facility

import (
	"context"
	"strconv"
	"strings"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
)

type BacnetObjectService struct {
	repo                  domainFacility.BacnetObjectStore
	fieldDeviceRepo       domainFacility.FieldDeviceStore
	objectDataRepo        domainFacility.ObjectDataStore
	objectDataBacnetStore domainFacility.ObjectDataBacnetObjectStore
	alarmDefinitionRepo   domainFacility.AlarmDefinitionRepository
	alarmTypeRepo         domainFacility.AlarmTypeRepository
	tx                    txCoordinator
}

func (s *BacnetObjectService) resolveAlarmBindingForTemplate(ctx context.Context, bacnetObject *domainFacility.BacnetObject) error {
	if bacnetObject == nil {
		return nil
	}

	if bacnetObject.AlarmTypeID != nil {
		if _, err := domain.GetByID(ctx, s.alarmTypeRepo, *bacnetObject.AlarmTypeID); err != nil {
			return err
		}
		bacnetObject.AlarmDefinitionID = nil
		return nil
	}

	if bacnetObject.AlarmDefinitionID == nil {
		return nil
	}

	defs, err := s.alarmDefinitionRepo.GetByIds(ctx, []uuid.UUID{*bacnetObject.AlarmDefinitionID})
	if err != nil {
		return err
	}
	if len(defs) == 0 || defs[0].AlarmTypeID == nil {
		return domain.NewValidationError().Add("objectdata.bacnetobject.alarm_type_id", "alarm_type_id is required")
	}

	bacnetObject.AlarmTypeID = defs[0].AlarmTypeID
	bacnetObject.AlarmDefinitionID = nil
	if _, err := domain.GetByID(ctx, s.alarmTypeRepo, *bacnetObject.AlarmTypeID); err != nil {
		return err
	}
	return nil
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
			return domain.NewValidationError().Add("fielddevice.bacnetobject.textfix", "textfix must be unique within the field device")
		}
	}
	return nil
}

func (s *BacnetObjectService) ensureSoftwareUniqueForObjectData(ctx context.Context, objectDataID uuid.UUID, softwareType domainFacility.BacnetSoftwareType, softwareNumber uint16, excludeID *uuid.UUID) error {
	ids, err := s.objectDataRepo.GetBacnetObjectIDs(ctx, objectDataID)
	if err != nil {
		return err
	}
	if len(ids) == 0 {
		return nil
	}
	items, err := s.repo.GetByIds(ctx, ids)
	if err != nil {
		return err
	}
	for _, it := range items {
		if excludeID != nil && it.ID == *excludeID {
			continue
		}
		if strings.EqualFold(string(it.SoftwareType), string(softwareType)) && it.SoftwareNumber == softwareNumber {
			return domain.NewValidationError().Add("objectdata.bacnetobject.software", "software_type + software_number must be unique within the object data")
		}
	}
	return nil
}

func (s *BacnetObjectService) validateRequiredFields(bacnetObject *domainFacility.BacnetObject, prefix string) error {
	ve := domain.NewValidationError()

	if strings.TrimSpace(bacnetObject.TextFix) == "" {
		ve = ve.Add(prefix+".textfix", "textfix is required")
	}
	if strings.TrimSpace(string(bacnetObject.SoftwareType)) == "" {
		ve = ve.Add(prefix+".software_type", "software_type is required")
	}

	if len(ve.Fields) > 0 {
		return ve
	}
	return nil
}

func NewBacnetObjectService(
	repo domainFacility.BacnetObjectStore,
	fieldDeviceRepo domainFacility.FieldDeviceStore,
	objectDataRepo domainFacility.ObjectDataStore,
	objectDataBacnetStore domainFacility.ObjectDataBacnetObjectStore,
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

func (s *BacnetObjectService) GetByID(ctx context.Context, id uuid.UUID) (*domainFacility.BacnetObject, error) {
	return domain.GetByID(ctx, s.repo, id)
}

func (s *BacnetObjectService) GetByIDs(ctx context.Context, ids []uuid.UUID) ([]*domainFacility.BacnetObject, error) {
	return s.repo.GetByIds(ctx, ids)
}

// CreateWithParent creates a bacnet object either for a field device (fieldDeviceID)
// or for an object data template (objectDataID). Exactly one must be provided.
func (s *BacnetObjectService) CreateWithParent(ctx context.Context, bacnetObject *domainFacility.BacnetObject, fieldDeviceID *uuid.UUID, objectDataID *uuid.UUID) error {
	return runWithFacilityTx(s.tx, s, func(services *Services) *BacnetObjectService {
		return services.BacnetObject
	}, func(txService *BacnetObjectService) error {
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
			if _, err := domain.GetByID(ctx, txService.fieldDeviceRepo, *fieldDeviceID); err != nil {
				return err
			}
			if err := txService.ensureTextFixUniqueForFieldDevice(ctx, *fieldDeviceID, bacnetObject.TextFix, nil); err != nil {
				return err
			}
			if err := txService.resolveAlarmBindingForTemplate(ctx, bacnetObject); err != nil {
				return err
			}
			bacnetObject.FieldDeviceID = fieldDeviceID
			return txService.repo.Create(ctx, bacnetObject)
		}

		od, err := domain.GetByID(ctx, txService.objectDataRepo, *objectDataID)
		if err != nil {
			return err
		}
		if !od.IsActive {
			return domain.ErrNotFound
		}

		if err := txService.ensureSoftwareUniqueForObjectData(ctx, *objectDataID, bacnetObject.SoftwareType, bacnetObject.SoftwareNumber, nil); err != nil {
			return err
		}
		if err := txService.resolveAlarmBindingForTemplate(ctx, bacnetObject); err != nil {
			return err
		}

		bacnetObject.FieldDeviceID = nil
		if err := txService.repo.Create(ctx, bacnetObject); err != nil {
			return err
		}
		return txService.objectDataBacnetStore.Add(ctx, *objectDataID, bacnetObject.ID)
	})
}

// Update updates a bacnet object. If objectDataID is provided, it will also attach
// the bacnet object to that object data (template) after validating the object data.
func (s *BacnetObjectService) Update(ctx context.Context, bacnetObject *domainFacility.BacnetObject, objectDataID *uuid.UUID) error {
	return runWithFacilityTx(s.tx, s, func(services *Services) *BacnetObjectService {
		return services.BacnetObject
	}, func(txService *BacnetObjectService) error {
		bacnetObject.TextFix = normalizeBacnetTextFix(bacnetObject.TextFix)

		if bacnetObject.FieldDeviceID != nil {
			if err := txService.validateRequiredFields(bacnetObject, "fielddevice.bacnetobject"); err != nil {
				return err
			}
		} else if objectDataID != nil {
			if err := txService.validateRequiredFields(bacnetObject, "objectdata.bacnetobject"); err != nil {
				return err
			}
		}

		if _, err := domain.GetByID(ctx, txService.repo, bacnetObject.ID); err != nil {
			return err
		}
		if bacnetObject.FieldDeviceID != nil {
			if err := txService.ensureTextFixUniqueForFieldDevice(ctx, *bacnetObject.FieldDeviceID, bacnetObject.TextFix, &bacnetObject.ID); err != nil {
				return err
			}
		}

		if objectDataID != nil {
			if err := txService.ensureSoftwareUniqueForObjectData(ctx, *objectDataID, bacnetObject.SoftwareType, bacnetObject.SoftwareNumber, &bacnetObject.ID); err != nil {
				return err
			}
		}

		if err := txService.resolveAlarmBindingForTemplate(ctx, bacnetObject); err != nil {
			return err
		}

		if err := txService.repo.Update(ctx, bacnetObject); err != nil {
			return err
		}

		if objectDataID != nil {
			od, err := domain.GetByID(ctx, txService.objectDataRepo, *objectDataID)
			if err != nil {
				return err
			}
			if !od.IsActive {
				return domain.ErrNotFound
			}
			return txService.objectDataBacnetStore.Add(ctx, *objectDataID, bacnetObject.ID)
		}

		return nil
	})
}

// ReplaceForObjectData replaces all bacnet objects for an object data template.
// Existing links are removed and the provided list is created and attached.
func (s *BacnetObjectService) ReplaceForObjectData(ctx context.Context, objectDataID uuid.UUID, inputs []domainFacility.BacnetObject) error {
	return runWithFacilityTx(s.tx, s, func(services *Services) *BacnetObjectService {
		return services.BacnetObject
	}, func(txService *BacnetObjectService) error {
		od, err := domain.GetByID(ctx, txService.objectDataRepo, objectDataID)
		if err != nil {
			return err
		}
		if !od.IsActive {
			return domain.ErrNotFound
		}

		seen := map[string]struct{}{}
		for i := range inputs {
			bo := &inputs[i]
			bo.TextFix = normalizeBacnetTextFix(bo.TextFix)
			if err := txService.validateRequiredFields(bo, "objectdata.bacnetobject"); err != nil {
				return err
			}
			if err := txService.resolveAlarmBindingForTemplate(ctx, bo); err != nil {
				return err
			}
			softwareKey := strings.ToLower(strings.TrimSpace(string(bo.SoftwareType))) + ":" + strconv.FormatUint(uint64(bo.SoftwareNumber), 10)
			if _, exists := seen[softwareKey]; exists {
				return domain.NewValidationError().Add("objectdata.bacnetobject.software", "software_type + software_number must be unique within the object data")
			}
			seen[softwareKey] = struct{}{}
		}

		if err := txService.objectDataBacnetStore.DeleteByObjectDataID(ctx, objectDataID); err != nil {
			return err
		}

		for i := range inputs {
			bo := inputs[i]
			bo.FieldDeviceID = nil
			if err := txService.repo.Create(ctx, &bo); err != nil {
				return err
			}
			if err := txService.objectDataBacnetStore.Add(ctx, objectDataID, bo.ID); err != nil {
				return err
			}
		}

		return nil
	})
}
