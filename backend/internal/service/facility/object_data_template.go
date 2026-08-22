package facility

import (
	"context"
	"strconv"
	"strings"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	domainObjectData "github.com/besart951/go_infra_link/backend/internal/domain/facility/objectdata"
	"github.com/google/uuid"
)

type objectDataTemplate struct {
	objectDataRepo      domainObjectData.ObjectDataStore
	templateStore       domainObjectData.BacnetObjectTemplateStore
	apparatRepo         domainFacility.ApparatRepository
	alarmDefinitionRepo domainFacility.AlarmDefinitionRepository
	alarmTypeRepo       domainFacility.AlarmTypeRepository
}

func (m objectDataTemplate) create(ctx context.Context, input domainFacility.ObjectDataTemplateCreate) (*domainFacility.ObjectData, error) {
	objectData := &domainFacility.ObjectData{
		Description: input.Description,
		Version:     input.Version,
		ProjectID:   input.ProjectID,
	}
	if input.IsActive != nil {
		objectData.IsActive = *input.IsActive
	}

	apparats, err := m.loadApparats(ctx, input.ApparatIDs)
	if err != nil {
		return nil, err
	}
	objectData.Apparats = apparats

	if err := m.ensureDescriptionUnique(ctx, objectData, nil); err != nil {
		return nil, err
	}
	if err := m.objectDataRepo.Create(ctx, objectData); err != nil {
		return nil, err
	}
	if len(input.BacnetObjects) > 0 {
		if err := m.replaceBacnetObjects(ctx, objectData.ID, input.BacnetObjects); err != nil {
			return nil, err
		}
	}

	return domain.GetByID(ctx, m.objectDataRepo, objectData.ID)
}

func (m objectDataTemplate) update(ctx context.Context, id uuid.UUID, input domainFacility.ObjectDataTemplateUpdate) (*domainFacility.ObjectData, error) {
	objectData, err := domain.GetByID(ctx, m.objectDataRepo, id)
	if err != nil {
		return nil, err
	}
	objectData.Base.Version = input.BaseVersion

	if input.Description != nil {
		objectData.Description = *input.Description
	}
	if input.Version != nil {
		objectData.Version = *input.Version
	}
	if input.IsActive != nil {
		objectData.IsActive = *input.IsActive
	}
	if input.ProjectID != nil {
		objectData.ProjectID = input.ProjectID
	}
	if input.ApparatIDs != nil {
		apparats, err := m.loadApparats(ctx, *input.ApparatIDs)
		if err != nil {
			return nil, err
		}
		objectData.Apparats = apparats
	}

	if err := m.ensureDescriptionUnique(ctx, objectData, &objectData.ID); err != nil {
		return nil, err
	}
	if err := m.objectDataRepo.Update(ctx, objectData); err != nil {
		return nil, err
	}
	if input.BacnetObjects != nil {
		if err := m.replaceBacnetObjects(ctx, objectData.ID, *input.BacnetObjects); err != nil {
			return nil, err
		}
	}

	return domain.GetByID(ctx, m.objectDataRepo, objectData.ID)
}

func (m objectDataTemplate) createBacnetObject(ctx context.Context, objectDataID uuid.UUID, bacnetObject *domainFacility.BacnetObject) error {
	if bacnetObject == nil {
		return domain.ErrInvalidArgument
	}
	if err := m.ensureActive(ctx, objectDataID); err != nil {
		return err
	}

	bacnetObject.TextFix = normalizeBacnetTextFix(bacnetObject.TextFix)
	if err := m.validateBacnetObject(bacnetObject); err != nil {
		return err
	}
	if err := m.ensureSoftwareUnique(ctx, objectDataID, bacnetObject.SoftwareType, bacnetObject.SoftwareNumber, nil); err != nil {
		return err
	}
	if err := m.resolveAlarmBinding(ctx, bacnetObject); err != nil {
		return err
	}

	template := bacnetObjectToTemplate(objectDataID, *bacnetObject)
	if err := m.templateStore.Create(ctx, &template); err != nil {
		return err
	}
	*bacnetObject = *templateToBacnetObject(template)
	return nil
}

func (m objectDataTemplate) updateBacnetObject(ctx context.Context, objectDataID uuid.UUID, bacnetObject *domainFacility.BacnetObject) error {
	if bacnetObject == nil {
		return domain.ErrInvalidArgument
	}
	if err := m.ensureActive(ctx, objectDataID); err != nil {
		return err
	}
	if _, err := m.templateStore.GetByID(ctx, bacnetObject.ID); err != nil {
		return err
	}

	bacnetObject.TextFix = normalizeBacnetTextFix(bacnetObject.TextFix)
	if err := m.validateBacnetObject(bacnetObject); err != nil {
		return err
	}
	if err := m.ensureSoftwareUnique(ctx, objectDataID, bacnetObject.SoftwareType, bacnetObject.SoftwareNumber, &bacnetObject.ID); err != nil {
		return err
	}
	if err := m.resolveAlarmBinding(ctx, bacnetObject); err != nil {
		return err
	}

	template := bacnetObjectToTemplate(objectDataID, *bacnetObject)
	if err := m.templateStore.Update(ctx, &template); err != nil {
		return err
	}
	*bacnetObject = *templateToBacnetObject(template)
	return nil
}

func (m objectDataTemplate) replaceBacnetObjects(ctx context.Context, objectDataID uuid.UUID, inputs []domainFacility.BacnetObject) error {
	if err := m.ensureActive(ctx, objectDataID); err != nil {
		return err
	}
	if err := m.prepareBacnetObjects(ctx, inputs); err != nil {
		return err
	}

	templates := make([]domainObjectData.BacnetObjectTemplate, len(inputs))
	for i := range inputs {
		templates[i] = bacnetObjectToTemplate(objectDataID, inputs[i])
	}
	return m.templateStore.Replace(ctx, objectDataID, templates)
}

func (m objectDataTemplate) prepareBacnetObjects(ctx context.Context, inputs []domainFacility.BacnetObject) error {
	seen := map[string]struct{}{}
	for i := range inputs {
		bo := &inputs[i]
		bo.TextFix = normalizeBacnetTextFix(bo.TextFix)
		if err := m.validateBacnetObject(bo); err != nil {
			return err
		}
		if err := m.resolveAlarmBinding(ctx, bo); err != nil {
			return err
		}

		softwareKey := strings.ToLower(strings.TrimSpace(string(bo.SoftwareType))) + ":" + strconv.FormatUint(uint64(bo.SoftwareNumber), 10)
		if _, exists := seen[softwareKey]; exists {
			return domain.NewValidationError().Add("objectdata.bacnetobject.software", "software_type + software_number must be unique within the object data")
		}
		seen[softwareKey] = struct{}{}
	}
	return nil
}

func (m objectDataTemplate) validateBacnetObject(bacnetObject *domainFacility.BacnetObject) error {
	return bacnetObject.Validate("objectdata.bacnetobject")
}

func (m objectDataTemplate) resolveAlarmBinding(ctx context.Context, bacnetObject *domainFacility.BacnetObject) error {
	if bacnetObject == nil {
		return nil
	}

	if bacnetObject.AlarmDefinitionID != nil {
		defs, err := m.alarmDefinitionRepo.GetByIds(ctx, []uuid.UUID{*bacnetObject.AlarmDefinitionID})
		if err != nil {
			return err
		}
		if len(defs) == 0 || defs[0].AlarmTypeID == nil {
			return domain.NewValidationError().Add("objectdata.bacnetobject.alarm_type_id", "alarm_type_id is required")
		}

		if bacnetObject.AlarmTypeID != nil && *bacnetObject.AlarmTypeID != *defs[0].AlarmTypeID {
			return domain.NewValidationError().Add("objectdata.bacnetobject.alarm_type_id", "alarm_type_id conflicts with alarm_definition_id")
		}

		bacnetObject.AlarmTypeID = defs[0].AlarmTypeID
		bacnetObject.AlarmDefinitionID = nil
		if _, err := domain.GetByID(ctx, m.alarmTypeRepo, *bacnetObject.AlarmTypeID); err != nil {
			return err
		}
		return nil
	}

	if bacnetObject.AlarmTypeID != nil {
		if _, err := domain.GetByID(ctx, m.alarmTypeRepo, *bacnetObject.AlarmTypeID); err != nil {
			return err
		}
		bacnetObject.AlarmDefinitionID = nil
	}

	return nil
}

func (m objectDataTemplate) ensureSoftwareUnique(ctx context.Context, objectDataID uuid.UUID, softwareType domainFacility.BacnetSoftwareType, softwareNumber uint16, excludeID *uuid.UUID) error {
	items, err := m.templateStore.ListByObjectDataID(ctx, objectDataID)
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

func bacnetObjectToTemplate(objectDataID uuid.UUID, item domainFacility.BacnetObject) domainObjectData.BacnetObjectTemplate {
	template := domainObjectData.BacnetObjectTemplate{
		ID: item.ID, CreatedAt: item.CreatedAt, UpdatedAt: item.UpdatedAt, Version: item.Base.Version,
		ObjectDataID: objectDataID, TextFix: item.TextFix, Description: item.Description,
		GMSVisible: item.GMSVisible, Optional: item.Optional, TextIndividual: item.TextIndividual,
		SoftwareType: item.SoftwareType, SoftwareNumber: item.SoftwareNumber,
		HardwareType: item.HardwareType, HardwareQuantity: item.HardwareQuantity,
		SoftwareReferenceID: item.SoftwareReferenceID, StateTextID: item.StateTextID,
		NotificationClassID: item.NotificationClassID, AlarmTypeID: item.AlarmTypeID,
	}
	template.AlarmValues = make([]domainObjectData.BacnetObjectTemplateAlarmValue, len(item.AlarmValues))
	for i := range item.AlarmValues {
		value := item.AlarmValues[i]
		template.AlarmValues[i] = domainObjectData.BacnetObjectTemplateAlarmValue{
			ID: value.ID, CreatedAt: value.CreatedAt, UpdatedAt: value.UpdatedAt, Version: value.Version,
			TemplateID: item.ID, AlarmTypeFieldID: value.AlarmTypeFieldID,
			ValueNumber: value.ValueNumber, ValueInteger: value.ValueInteger, ValueBoolean: value.ValueBoolean,
			ValueString: value.ValueString, ValueJSON: value.ValueJSON, UnitID: value.UnitID, Source: value.Source,
		}
	}
	return template
}

func templateToBacnetObject(template domainObjectData.BacnetObjectTemplate) *domainFacility.BacnetObject {
	item := &domainFacility.BacnetObject{
		Base:    domain.Base{ID: template.ID, CreatedAt: template.CreatedAt, UpdatedAt: template.UpdatedAt, Version: template.Version},
		TextFix: template.TextFix, Description: template.Description, GMSVisible: template.GMSVisible,
		Optional: template.Optional, TextIndividual: template.TextIndividual,
		SoftwareType: template.SoftwareType, SoftwareNumber: template.SoftwareNumber,
		HardwareType: template.HardwareType, HardwareQuantity: template.HardwareQuantity,
		SoftwareReferenceID: template.SoftwareReferenceID, StateTextID: template.StateTextID,
		NotificationClassID: template.NotificationClassID, AlarmTypeID: template.AlarmTypeID,
	}
	item.AlarmValues = make([]domainFacility.BacnetObjectAlarmValue, len(template.AlarmValues))
	for i := range template.AlarmValues {
		value := template.AlarmValues[i]
		item.AlarmValues[i] = domainFacility.BacnetObjectAlarmValue{
			Base:           domain.Base{ID: value.ID, CreatedAt: value.CreatedAt, UpdatedAt: value.UpdatedAt, Version: value.Version},
			BacnetObjectID: template.ID, AlarmTypeFieldID: value.AlarmTypeFieldID,
			ValueNumber: value.ValueNumber, ValueInteger: value.ValueInteger, ValueBoolean: value.ValueBoolean,
			ValueString: value.ValueString, ValueJSON: value.ValueJSON, UnitID: value.UnitID, Source: value.Source,
		}
	}
	return item
}

func (m objectDataTemplate) ensureActive(ctx context.Context, objectDataID uuid.UUID) error {
	od, err := domain.GetByID(ctx, m.objectDataRepo, objectDataID)
	if err != nil {
		return err
	}
	if !od.IsActive {
		return domain.ErrNotFound
	}
	return nil
}

func (m objectDataTemplate) ensureDescriptionUnique(ctx context.Context, objectData *domainFacility.ObjectData, excludeID *uuid.UUID) error {
	description := strings.TrimSpace(objectData.Description)
	if description == "" {
		return nil
	}
	exists, err := m.objectDataRepo.ExistsByDescription(ctx, objectData.ProjectID, description, excludeID)
	if err != nil {
		return err
	}
	if exists {
		return domain.NewValidationError().Add("objectdata.description", "description must be unique")
	}
	return nil
}

func (m objectDataTemplate) loadApparats(ctx context.Context, ids []uuid.UUID) ([]*domainFacility.Apparat, error) {
	uniqueIDs := uniqueUUIDs(ids)
	if len(uniqueIDs) == 0 {
		return []*domainFacility.Apparat{}, nil
	}

	apparats, err := m.apparatRepo.GetByIds(ctx, uniqueIDs)
	if err != nil {
		return nil, err
	}

	found := make(map[uuid.UUID]struct{}, len(apparats))
	for _, apparat := range apparats {
		if apparat != nil {
			found[apparat.ID] = struct{}{}
		}
	}
	for _, id := range uniqueIDs {
		if _, ok := found[id]; !ok {
			return nil, domain.ErrNotFound
		}
	}

	return apparats, nil
}
