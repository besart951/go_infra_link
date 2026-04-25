package facility

import (
	"context"
	"strconv"
	"strings"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
)

type objectDataTemplate struct {
	objectDataRepo        domainFacility.ObjectDataStore
	bacnetObjectRepo      domainFacility.BacnetObjectStore
	objectDataBacnetStore domainFacility.ObjectDataBacnetObjectStore
	apparatRepo           domainFacility.ApparatRepository
	alarmDefinitionRepo   domainFacility.AlarmDefinitionRepository
	alarmTypeRepo         domainFacility.AlarmTypeRepository
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

	bacnetObject.FieldDeviceID = nil
	if err := m.bacnetObjectRepo.Create(ctx, bacnetObject); err != nil {
		return err
	}
	return m.objectDataBacnetStore.Add(ctx, objectDataID, bacnetObject.ID)
}

func (m objectDataTemplate) updateBacnetObject(ctx context.Context, objectDataID uuid.UUID, bacnetObject *domainFacility.BacnetObject) error {
	if bacnetObject == nil {
		return domain.ErrInvalidArgument
	}
	if err := m.ensureActive(ctx, objectDataID); err != nil {
		return err
	}
	if _, err := domain.GetByID(ctx, m.bacnetObjectRepo, bacnetObject.ID); err != nil {
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

	bacnetObject.FieldDeviceID = nil
	if err := m.bacnetObjectRepo.Update(ctx, bacnetObject); err != nil {
		return err
	}
	return m.objectDataBacnetStore.Add(ctx, objectDataID, bacnetObject.ID)
}

func (m objectDataTemplate) replaceBacnetObjects(ctx context.Context, objectDataID uuid.UUID, inputs []domainFacility.BacnetObject) error {
	if err := m.ensureActive(ctx, objectDataID); err != nil {
		return err
	}
	if err := m.prepareBacnetObjects(ctx, inputs); err != nil {
		return err
	}

	if err := m.objectDataBacnetStore.DeleteByObjectDataID(ctx, objectDataID); err != nil {
		return err
	}

	for i := range inputs {
		bo := inputs[i]
		bo.FieldDeviceID = nil
		if err := m.bacnetObjectRepo.Create(ctx, &bo); err != nil {
			return err
		}
		if err := m.objectDataBacnetStore.Add(ctx, objectDataID, bo.ID); err != nil {
			return err
		}
	}

	return nil
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
	ve := domain.NewValidationError()

	if strings.TrimSpace(bacnetObject.TextFix) == "" {
		ve = ve.Add("objectdata.bacnetobject.textfix", "textfix is required")
	}
	if strings.TrimSpace(string(bacnetObject.SoftwareType)) == "" {
		ve = ve.Add("objectdata.bacnetobject.software_type", "software_type is required")
	}

	if len(ve.Fields) > 0 {
		return ve
	}
	return nil
}

func (m objectDataTemplate) resolveAlarmBinding(ctx context.Context, bacnetObject *domainFacility.BacnetObject) error {
	if bacnetObject == nil {
		return nil
	}

	if bacnetObject.AlarmTypeID != nil {
		if _, err := domain.GetByID(ctx, m.alarmTypeRepo, *bacnetObject.AlarmTypeID); err != nil {
			return err
		}
		bacnetObject.AlarmDefinitionID = nil
		return nil
	}

	if bacnetObject.AlarmDefinitionID == nil {
		return nil
	}

	defs, err := m.alarmDefinitionRepo.GetByIds(ctx, []uuid.UUID{*bacnetObject.AlarmDefinitionID})
	if err != nil {
		return err
	}
	if len(defs) == 0 || defs[0].AlarmTypeID == nil {
		return domain.NewValidationError().Add("objectdata.bacnetobject.alarm_type_id", "alarm_type_id is required")
	}

	bacnetObject.AlarmTypeID = defs[0].AlarmTypeID
	bacnetObject.AlarmDefinitionID = nil
	if _, err := domain.GetByID(ctx, m.alarmTypeRepo, *bacnetObject.AlarmTypeID); err != nil {
		return err
	}
	return nil
}

func (m objectDataTemplate) ensureSoftwareUnique(ctx context.Context, objectDataID uuid.UUID, softwareType domainFacility.BacnetSoftwareType, softwareNumber uint16, excludeID *uuid.UUID) error {
	ids, err := m.objectDataRepo.GetBacnetObjectIDs(ctx, objectDataID)
	if err != nil {
		return err
	}
	if len(ids) == 0 {
		return nil
	}
	items, err := m.bacnetObjectRepo.GetByIds(ctx, ids)
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
