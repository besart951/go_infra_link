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
	repo                domainObjectData.BacnetObjectStore
	fieldDeviceRepo     domainFieldDevice.FieldDeviceStore
	objectDataRepo      domainObjectData.ObjectDataStore
	templateStore       domainObjectData.BacnetObjectTemplateStore
	alarmValueRepo      domainFacility.BacnetObjectAlarmValueRepository
	alarmDefinitionRepo domainFacility.AlarmDefinitionRepository
	alarmTypeRepo       domainFacility.AlarmTypeRepository
	tx                  txCoordinator
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

type BacnetObjectDependencies struct {
	Objects          domainObjectData.BacnetObjectStore
	FieldDevices     domainFieldDevice.FieldDeviceStore
	ObjectData       domainObjectData.ObjectDataStore
	Templates        domainObjectData.BacnetObjectTemplateStore
	AlarmValues      domainFacility.BacnetObjectAlarmValueRepository
	AlarmDefinitions domainFacility.AlarmDefinitionRepository
	AlarmTypes       domainFacility.AlarmTypeRepository
}

func NewBacnetObjectService(deps BacnetObjectDependencies) *BacnetObjectService {
	return &BacnetObjectService{
		repo:                deps.Objects,
		fieldDeviceRepo:     deps.FieldDevices,
		objectDataRepo:      deps.ObjectData,
		templateStore:       deps.Templates,
		alarmValueRepo:      deps.AlarmValues,
		alarmDefinitionRepo: deps.AlarmDefinitions,
		alarmTypeRepo:       deps.AlarmTypes,
	}
}

func (s *BacnetObjectService) ReplaceAlarmValues(ctx context.Context, id uuid.UUID, version uint64, values []domainFacility.BacnetObjectAlarmValue) (uint64, error) {
	return runWithFacilityTxResult(ctx, s.transaction(), func(txCtx context.Context, txService *BacnetObjectService) (uint64, error) {
		return txService.replaceAlarmValues(txCtx, id, version, values)
	})
}

func (s *BacnetObjectService) replaceAlarmValues(ctx context.Context, id uuid.UUID, version uint64, values []domainFacility.BacnetObjectAlarmValue) (uint64, error) {
	if id == uuid.Nil || version == 0 {
		return 0, domain.ErrInvalidArgument
	}
	instances, err := s.repo.GetByIds(ctx, []uuid.UUID{id})
	if err != nil {
		return 0, err
	}
	if len(instances) > 0 {
		return s.replaceInstanceAlarmValues(ctx, instances[0], version, values)
	}
	return s.replaceTemplateAlarmValues(ctx, id, version, values)
}

func (s *BacnetObjectService) replaceInstanceAlarmValues(ctx context.Context, object *domainFacility.BacnetObject, version uint64, values []domainFacility.BacnetObjectAlarmValue) (uint64, error) {
	object.Version = version
	if err := s.repo.Update(ctx, object); err != nil {
		return 0, err
	}
	if err := s.alarmValueRepo.ReplaceForBacnetObject(ctx, object.ID, values); err != nil {
		return 0, err
	}
	return object.Version, nil
}

func (s *BacnetObjectService) replaceTemplateAlarmValues(ctx context.Context, id uuid.UUID, version uint64, values []domainFacility.BacnetObjectAlarmValue) (uint64, error) {
	template, err := s.templateStore.GetByID(ctx, id)
	if err != nil {
		return 0, err
	}
	template.Version = version
	template.AlarmValues = templateAlarmValues(id, values)
	if err := s.templateStore.Update(ctx, template); err != nil {
		return 0, err
	}
	return template.Version, nil
}

func templateAlarmValues(templateID uuid.UUID, values []domainFacility.BacnetObjectAlarmValue) []domainObjectData.BacnetObjectTemplateAlarmValue {
	result := make([]domainObjectData.BacnetObjectTemplateAlarmValue, len(values))
	for index, value := range values {
		result[index] = domainObjectData.BacnetObjectTemplateAlarmValue{
			ID: value.ID, TemplateID: templateID, AlarmTypeFieldID: value.AlarmTypeFieldID,
			ValueNumber: value.ValueNumber, ValueInteger: value.ValueInteger, ValueBoolean: value.ValueBoolean,
			ValueString: value.ValueString, ValueJSON: value.ValueJSON, UnitID: value.UnitID, Source: value.Source,
		}
	}
	return result
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
		objectDataRepo:      s.objectDataRepo,
		templateStore:       s.templateStore,
		alarmDefinitionRepo: s.alarmDefinitionRepo,
		alarmTypeRepo:       s.alarmTypeRepo,
	}
}

func (s *BacnetObjectService) GetByID(ctx context.Context, id uuid.UUID) (*domainFacility.BacnetObject, error) {
	items, err := s.repo.GetByIds(ctx, []uuid.UUID{id})
	if err == nil && len(items) > 0 {
		return items[0], nil
	}
	template, templateErr := s.templateStore.GetByID(ctx, id)
	if templateErr != nil {
		return nil, templateErr
	}
	return templateToBacnetObject(*template), nil
}

func (s *BacnetObjectService) GetByIDs(ctx context.Context, ids []uuid.UUID) ([]*domainFacility.BacnetObject, error) {
	instances, err := s.repo.GetByIds(ctx, ids)
	if err != nil {
		return nil, err
	}
	found := make(map[uuid.UUID]struct{}, len(instances))
	for _, item := range instances {
		found[item.ID] = struct{}{}
	}
	missing := make([]uuid.UUID, 0, len(ids)-len(found))
	for _, id := range ids {
		if _, ok := found[id]; !ok {
			missing = append(missing, id)
		}
	}
	templates, err := s.templateStore.GetByIDs(ctx, missing)
	if err != nil {
		return nil, err
	}
	for _, template := range templates {
		instances = append(instances, templateToBacnetObject(template))
	}
	return instances, nil
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

func (s *BacnetObjectService) DeleteAtVersion(ctx context.Context, id uuid.UUID, version uint64) error {
	if id == uuid.Nil || version == 0 {
		return domain.ErrInvalidArgument
	}
	instances, err := s.repo.GetByIds(ctx, []uuid.UUID{id})
	if err != nil {
		return err
	}
	if len(instances) == 0 {
		return s.templateStore.DeleteAtVersion(ctx, id, version)
	}
	deleter, ok := s.repo.(interface {
		DeleteAtVersion(context.Context, uuid.UUID, uint64) error
	})
	if !ok {
		return domain.ErrInvalidArgument
	}
	return deleter.DeleteAtVersion(ctx, id, version)
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
