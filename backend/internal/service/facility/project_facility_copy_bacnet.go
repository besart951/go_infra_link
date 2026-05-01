package facility

import (
	"context"
	"fmt"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
)

func (c projectFacilityCopy) replaceFieldDeviceBacnetObjects(ctx context.Context, fieldDeviceID uuid.UUID, bacnetObjects []domainFacility.BacnetObject) error {
	ve := domain.NewValidationError()
	for i, obj := range bacnetObjects {
		obj.TextFix = normalizeBacnetTextFix(obj.TextFix)
		bacnetObjects[i].TextFix = obj.TextFix
		if obj.TextFix == "" {
			ve = ve.Add(fmt.Sprintf("bacnet_objects.%d.text_fix", i), "text_fix is required")
			continue
		}
	}
	if len(ve.Fields) > 0 {
		return ve
	}

	if err := c.bacnetObjectRepo.DeleteByFieldDeviceIDs(ctx, []uuid.UUID{fieldDeviceID}); err != nil {
		return err
	}
	if len(bacnetObjects) == 0 {
		return nil
	}

	objects := make([]*domainFacility.BacnetObject, 0, len(bacnetObjects))
	for i := range bacnetObjects {
		id := fieldDeviceID
		bacnetObjects[i].FieldDeviceID = &id
		objects = append(objects, &bacnetObjects[i])
	}

	if err := c.bacnetObjectRepo.BulkCreate(ctx, objects, 200); err != nil {
		return err
	}

	return c.createAlarmValuesForBacnetObjects(ctx, objects)
}

func (c projectFacilityCopy) replaceFieldDeviceBacnetObjectsFromObjectData(ctx context.Context, fieldDeviceID uuid.UUID, objectDataID uuid.UUID) error {
	od, err := domain.GetByID(ctx, c.objectDataRepo, objectDataID)
	if err != nil {
		return err
	}
	if !od.IsActive {
		return domain.ErrNotFound
	}

	ids, err := c.objectDataRepo.GetBacnetObjectIDs(ctx, objectDataID)
	if err != nil {
		return err
	}
	if len(ids) == 0 {
		return c.bacnetObjectRepo.DeleteByFieldDeviceIDs(ctx, []uuid.UUID{fieldDeviceID})
	}

	templates, err := c.bacnetObjectRepo.GetByIds(ctx, ids)
	if err != nil {
		return err
	}
	if len(templates) != len(ids) {
		return domain.ErrNotFound
	}

	templateToClone := make(map[uuid.UUID]*domainFacility.BacnetObject, len(templates))
	templateRef := make(map[uuid.UUID]*uuid.UUID, len(templates))
	clones := make([]*domainFacility.BacnetObject, 0, len(templates))

	for _, t := range templates {
		textFix := normalizeBacnetTextFix(t.TextFix)
		if textFix == "" {
			return domain.NewValidationError().Add("bacnet_objects.text_fix", "text_fix is required")
		}

		clone := cloneBacnetObjectForFieldDeviceTemplate(*t, fieldDeviceID, textFix)
		templateToClone[t.ID] = clone
		templateRef[t.ID] = t.SoftwareReferenceID
		clones = append(clones, clone)
	}

	if err := c.bacnetObjectRepo.DeleteByFieldDeviceIDs(ctx, []uuid.UUID{fieldDeviceID}); err != nil {
		return err
	}

	if err := c.bacnetObjectRepo.BulkCreate(ctx, clones, 200); err != nil {
		return err
	}
	if err := c.createAlarmValuesForBacnetObjects(ctx, clones); err != nil {
		return err
	}

	return c.remapSoftwareReferences(ctx, templateToClone, templateRef)
}

func cloneBacnetObjectForFieldDeviceTemplate(original domainFacility.BacnetObject, fieldDeviceID uuid.UUID, textFix string) *domainFacility.BacnetObject {
	return &domainFacility.BacnetObject{
		TextFix:             textFix,
		Description:         original.Description,
		GMSVisible:          original.GMSVisible,
		Optional:            original.Optional,
		TextIndividual:      original.TextIndividual,
		SoftwareType:        original.SoftwareType,
		SoftwareNumber:      original.SoftwareNumber,
		HardwareType:        original.HardwareType,
		HardwareQuantity:    original.HardwareQuantity,
		FieldDeviceID:       &fieldDeviceID,
		SoftwareReferenceID: nil,
		StateTextID:         original.StateTextID,
		NotificationClassID: original.NotificationClassID,
		AlarmTypeID:         original.AlarmTypeID,
	}
}

func (c projectFacilityCopy) copyObjectDataTemplatesForProject(ctx context.Context, projectID uuid.UUID) error {
	templates, err := c.objectDataRepo.GetTemplates(ctx)
	if err != nil {
		return err
	}

	for _, tmpl := range templates {
		if tmpl == nil {
			continue
		}
		if err := c.copyObjectDataTemplateForProject(ctx, tmpl, projectID); err != nil {
			return err
		}
	}

	return nil
}

func (c projectFacilityCopy) copyObjectDataTemplateForProject(ctx context.Context, tmpl *domainFacility.ObjectData, projectID uuid.UUID) error {
	copyEntity := *tmpl
	copyEntity.ID = uuid.Nil
	copyEntity.ProjectID = &projectID
	copyEntity.BacnetObjects = nil

	if err := c.objectDataRepo.Create(ctx, &copyEntity); err != nil {
		return err
	}

	if len(tmpl.BacnetObjects) == 0 {
		return nil
	}

	newBacnetObjects, err := c.copyBacnetObjectsForObjectData(ctx, tmpl.BacnetObjects)
	if err != nil {
		return err
	}

	copyEntity.BacnetObjects = newBacnetObjects
	return c.objectDataRepo.Update(ctx, &copyEntity)
}

func (c projectFacilityCopy) copyBacnetObjectsForObjectData(ctx context.Context, templates []*domainFacility.BacnetObject) ([]*domainFacility.BacnetObject, error) {
	oldToNew := make(map[uuid.UUID]*domainFacility.BacnetObject, len(templates))
	oldRefs := make(map[uuid.UUID]*uuid.UUID, len(templates))

	for _, bo := range templates {
		if bo == nil {
			continue
		}
		newBO := cloneBacnetObjectForObjectDataTemplate(*bo)
		if err := c.bacnetObjectRepo.Create(ctx, newBO); err != nil {
			return nil, err
		}
		oldToNew[bo.ID] = newBO
		oldRefs[bo.ID] = bo.SoftwareReferenceID
	}

	if err := c.remapSoftwareReferences(ctx, oldToNew, oldRefs); err != nil {
		return nil, err
	}

	newBacnetObjects := make([]*domainFacility.BacnetObject, 0, len(oldToNew))
	for _, newBO := range oldToNew {
		newBacnetObjects = append(newBacnetObjects, newBO)
	}

	return newBacnetObjects, nil
}

func cloneBacnetObjectForObjectDataTemplate(original domainFacility.BacnetObject) *domainFacility.BacnetObject {
	return &domainFacility.BacnetObject{
		TextFix:             original.TextFix,
		Description:         original.Description,
		GMSVisible:          original.GMSVisible,
		Optional:            original.Optional,
		TextIndividual:      original.TextIndividual,
		SoftwareType:        original.SoftwareType,
		SoftwareNumber:      original.SoftwareNumber,
		HardwareType:        original.HardwareType,
		HardwareQuantity:    original.HardwareQuantity,
		StateTextID:         original.StateTextID,
		NotificationClassID: original.NotificationClassID,
		AlarmTypeID:         original.AlarmTypeID,
	}
}

func (c projectFacilityCopy) createAlarmValuesForBacnetObjects(ctx context.Context, bacnetObjects []*domainFacility.BacnetObject) error {
	values, err := newAlarmValueMaterializer(c.alarmTypeRepo).buildDefaultValues(ctx, bacnetObjects)
	if err != nil {
		return err
	}
	if len(values) == 0 {
		return nil
	}

	return c.bacnetAlarmValueRepo.BulkCreate(ctx, values, 500)
}
