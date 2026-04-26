package facility

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
)

type projectFacilityCopy struct {
	controlCabinetRepo      domainFacility.ControlCabinetRepository
	buildingRepo            domainFacility.BuildingRepository
	spsControllerRepo       domainFacility.SPSControllerRepository
	systemTypeRepo          domainFacility.SystemTypeRepository
	spsControllerSystemRepo domainFacility.SPSControllerSystemTypeStore
	fieldDeviceRepo         domainFacility.FieldDeviceStore
	specificationRepo       domainFacility.SpecificationStore
	bacnetObjectRepo        domainFacility.BacnetObjectStore
	objectDataRepo          domainFacility.ObjectDataStore
	alarmTypeRepo           domainFacility.AlarmTypeRepository
	bacnetAlarmValueRepo    domainFacility.BacnetObjectAlarmValueRepository
}

func (s *FieldDeviceService) projectFacilityCopy() projectFacilityCopy {
	return projectFacilityCopy{
		fieldDeviceRepo:         s.repo,
		spsControllerSystemRepo: s.spsControllerSystemTypeRepo,
		systemTypeRepo:          s.systemTypeRepo,
		specificationRepo:       s.specificationRepo,
		bacnetObjectRepo:        s.bacnetObjectRepo,
		objectDataRepo:          s.objectDataRepo,
		alarmTypeRepo:           s.alarmTypeRepo,
		bacnetAlarmValueRepo:    s.bacnetAlarmValueRepo,
	}
}

// CopyObjectDataTemplatesForProject creates project-scoped ObjectData copies
// and remaps internal BACnet Object software references.
func CopyObjectDataTemplatesForProject(
	ctx context.Context,
	objectDataRepo domainFacility.ObjectDataStore,
	bacnetObjectRepo domainFacility.BacnetObjectStore,
	projectID uuid.UUID,
) error {
	return projectFacilityCopy{
		objectDataRepo:   objectDataRepo,
		bacnetObjectRepo: bacnetObjectRepo,
	}.copyObjectDataTemplatesForProject(ctx, projectID)
}

func (c projectFacilityCopy) copyControlCabinetByID(ctx context.Context, id uuid.UUID) (*domainFacility.ControlCabinet, error) {
	original, err := domain.GetByID(ctx, c.controlCabinetRepo, id)
	if err != nil {
		return nil, err
	}

	baseNr := ""
	if original.ControlCabinetNr != nil {
		baseNr = strings.TrimSpace(*original.ControlCabinetNr)
	}

	nextNr, err := c.nextAvailableControlCabinetNr(ctx, original.BuildingID, baseNr)
	if err != nil {
		return nil, err
	}

	copyEntity := &domainFacility.ControlCabinet{
		BuildingID:       original.BuildingID,
		ControlCabinetNr: &nextNr,
	}
	if err := c.controlCabinetRepo.Create(ctx, copyEntity); err != nil {
		return nil, err
	}

	if err := c.copySPSControllersForControlCabinet(ctx, original.ID, copyEntity.ID); err != nil {
		return nil, err
	}

	return copyEntity, nil
}

func (c projectFacilityCopy) copySPSControllerByID(ctx context.Context, id uuid.UUID) (*domainFacility.SPSController, error) {
	original, err := domain.GetByID(ctx, c.spsControllerRepo, id)
	if err != nil {
		return nil, err
	}

	controlCabinet, err := domain.GetByID(ctx, c.controlCabinetRepo, original.ControlCabinetID)
	if err != nil {
		return nil, err
	}

	building, err := domain.GetByID(ctx, c.buildingRepo, controlCabinet.BuildingID)
	if err != nil {
		return nil, err
	}

	nextGADevice, err := c.nextAvailableGADevice(ctx, original.ControlCabinetID)
	if err != nil {
		return nil, err
	}
	if nextGADevice == "" {
		return nil, domain.NewValidationError().Add("spscontroller.ga_device", "no available ga_device for control cabinet")
	}

	iwsCode := strings.TrimSpace(building.IWSCode)
	cabinetNr := ""
	if controlCabinet.ControlCabinetNr != nil {
		cabinetNr = strings.TrimSpace(*controlCabinet.ControlCabinetNr)
	}

	deviceName := nextGADevice
	if iwsCode != "" && cabinetNr != "" {
		deviceName = strings.ToUpper(iwsCode + "_" + cabinetNr + "_" + nextGADevice)
	}

	copyEntity := &domainFacility.SPSController{
		ControlCabinetID:  original.ControlCabinetID,
		GADevice:          &nextGADevice,
		DeviceName:        deviceName,
		DeviceDescription: original.DeviceDescription,
		DeviceLocation:    original.DeviceLocation,
		IPAddress:         nil,
		Subnet:            original.Subnet,
		Gateway:           original.Gateway,
		Vlan:              original.Vlan,
	}
	if err := c.spsControllerRepo.Create(ctx, copyEntity); err != nil {
		return nil, err
	}

	if err := c.copySystemTypesAndFieldDevicesForSPSController(ctx, original.ID, copyEntity.ID); err != nil {
		return nil, err
	}

	return copyEntity, nil
}

func (c projectFacilityCopy) copySPSControllerSystemTypeByID(ctx context.Context, id uuid.UUID) (*domainFacility.SPSControllerSystemType, error) {
	original, err := domain.GetByID(ctx, c.spsControllerSystemRepo, id)
	if err != nil {
		return nil, err
	}

	systemType, err := domain.GetByID(ctx, c.systemTypeRepo, original.SystemTypeID)
	if err != nil {
		return nil, err
	}

	existing, err := c.spsControllerSystemRepo.ListBySPSControllerID(ctx, original.SPSControllerID)
	if err != nil {
		return nil, err
	}

	usedNumbers := make(map[int]struct{}, len(existing))
	for _, item := range existing {
		if item.SystemTypeID != original.SystemTypeID || item.Number == nil {
			continue
		}
		usedNumbers[*item.Number] = struct{}{}
	}

	nextNumber, ok := findLowestAvailableNumber(systemType.NumberMin, systemType.NumberMax, usedNumbers)
	if !ok {
		return nil, domain.NewValidationError().Add("spscontroller.system_types", "no available number in the system type range")
	}

	copyNumber := nextNumber
	copyEntity := &domainFacility.SPSControllerSystemType{
		Number:          &copyNumber,
		DocumentName:    original.DocumentName,
		SPSControllerID: original.SPSControllerID,
		SystemTypeID:    original.SystemTypeID,
	}
	if err := c.spsControllerSystemRepo.Create(ctx, copyEntity); err != nil {
		return nil, err
	}

	if err := c.copyFieldDevicesForSystemTypes(ctx, map[uuid.UUID]uuid.UUID{original.ID: copyEntity.ID}); err != nil {
		return nil, err
	}

	return domain.GetByID(ctx, c.spsControllerSystemRepo, copyEntity.ID)
}

func (c projectFacilityCopy) copySPSControllersForControlCabinet(ctx context.Context, originalControlCabinetID, newControlCabinetID uuid.UUID) error {
	newControlCabinet, err := domain.GetByID(ctx, c.controlCabinetRepo, newControlCabinetID)
	if err != nil {
		return err
	}

	building, err := domain.GetByID(ctx, c.buildingRepo, newControlCabinet.BuildingID)
	if err != nil {
		return err
	}

	newCabinetNr := ""
	if newControlCabinet.ControlCabinetNr != nil {
		newCabinetNr = strings.TrimSpace(*newControlCabinet.ControlCabinetNr)
	}
	buildingIWSCode := strings.TrimSpace(building.IWSCode)

	originalSPSControllers, err := c.listSPSControllersByControlCabinetID(ctx, originalControlCabinetID)
	if err != nil {
		return err
	}

	for _, originalSPS := range originalSPSControllers {
		gaDevice := ""
		if originalSPS.GADevice != nil {
			gaDevice = strings.ToUpper(strings.TrimSpace(*originalSPS.GADevice))
		}
		if gaDevice == "" {
			nextGADevice, err := c.nextAvailableGADevice(ctx, newControlCabinetID)
			if err != nil {
				return err
			}
			gaDevice = nextGADevice
		}

		var gaDevicePtr *string
		if gaDevice != "" {
			gaDevicePtr = &gaDevice
		}

		deviceName := strings.TrimSpace(originalSPS.DeviceName)
		if buildingIWSCode != "" && newCabinetNr != "" && gaDevice != "" {
			deviceName = strings.ToUpper(buildingIWSCode + "_" + newCabinetNr + "_" + gaDevice)
		}

		spsCopy := &domainFacility.SPSController{
			ControlCabinetID:  newControlCabinetID,
			GADevice:          gaDevicePtr,
			DeviceName:        deviceName,
			DeviceDescription: originalSPS.DeviceDescription,
			DeviceLocation:    originalSPS.DeviceLocation,
			IPAddress:         nil,
			Subnet:            originalSPS.Subnet,
			Gateway:           originalSPS.Gateway,
			Vlan:              originalSPS.Vlan,
		}
		if err := c.spsControllerRepo.Create(ctx, spsCopy); err != nil {
			return err
		}

		if err := c.copySystemTypesAndFieldDevicesForSPSController(ctx, originalSPS.ID, spsCopy.ID); err != nil {
			return err
		}
	}

	return nil
}

func (c projectFacilityCopy) copySystemTypesAndFieldDevicesForSPSController(ctx context.Context, originalSPSControllerID, newSPSControllerID uuid.UUID) error {
	originalSystemTypes, err := c.listSystemTypesBySPSControllerID(ctx, originalSPSControllerID)
	if err != nil {
		return err
	}

	newSystemTypeMap := make(map[uuid.UUID]uuid.UUID, len(originalSystemTypes))
	if len(originalSystemTypes) > 0 {
		systemTypesToCreate := make([]domainFacility.SPSControllerSystemType, 0, len(originalSystemTypes))
		for _, item := range originalSystemTypes {
			systemTypesToCreate = append(systemTypesToCreate, domainFacility.SPSControllerSystemType{
				Number:       item.Number,
				DocumentName: item.DocumentName,
				SystemTypeID: item.SystemTypeID,
			})
		}

		systemTypeMap, err := c.loadSystemTypeDefinitions(ctx, systemTypesToCreate)
		if err != nil {
			return err
		}
		if err := assignSystemTypeNumbers(systemTypesToCreate, systemTypeMap); err != nil {
			return err
		}

		for idx, item := range systemTypesToCreate {
			newSystemType := &domainFacility.SPSControllerSystemType{
				Number:          item.Number,
				DocumentName:    item.DocumentName,
				SPSControllerID: newSPSControllerID,
				SystemTypeID:    item.SystemTypeID,
			}
			if err := c.spsControllerSystemRepo.Create(ctx, newSystemType); err != nil {
				return err
			}
			newSystemTypeMap[originalSystemTypes[idx].ID] = newSystemType.ID
		}
	}

	return c.copyFieldDevicesForSystemTypes(ctx, newSystemTypeMap)
}

func (c projectFacilityCopy) copyFieldDevicesForSystemTypes(ctx context.Context, newSystemTypeMap map[uuid.UUID]uuid.UUID) error {
	if len(newSystemTypeMap) == 0 {
		return nil
	}

	originalSystemTypeIDs := make([]uuid.UUID, 0, len(newSystemTypeMap))
	for originalSystemTypeID := range newSystemTypeMap {
		originalSystemTypeIDs = append(originalSystemTypeIDs, originalSystemTypeID)
	}

	fieldDeviceIDs, err := c.fieldDeviceRepo.GetIDsBySPSControllerSystemTypeIDs(ctx, originalSystemTypeIDs)
	if err != nil {
		return err
	}
	if len(fieldDeviceIDs) == 0 {
		return nil
	}

	originalFieldDevices, err := c.fieldDeviceRepo.GetByIds(ctx, fieldDeviceIDs)
	if err != nil {
		return err
	}

	return c.copyFieldDevicesWithChildren(ctx, derefSlice(originalFieldDevices), newSystemTypeMap)
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
	values, err := c.buildAlarmValuesForBacnetObjects(ctx, bacnetObjects)
	if err != nil {
		return err
	}
	if len(values) == 0 {
		return nil
	}

	return c.bacnetAlarmValueRepo.BulkCreate(ctx, values, 500)
}

func (c projectFacilityCopy) buildAlarmValuesForBacnetObjects(ctx context.Context, bacnetObjects []*domainFacility.BacnetObject) ([]*domainFacility.BacnetObjectAlarmValue, error) {
	if len(bacnetObjects) == 0 {
		return nil, nil
	}

	alarmTypeCache := make(map[uuid.UUID]*domainFacility.AlarmType)
	values := make([]*domainFacility.BacnetObjectAlarmValue, 0)

	for _, obj := range bacnetObjects {
		if obj == nil || obj.AlarmTypeID == nil {
			continue
		}

		alarmType, ok := alarmTypeCache[*obj.AlarmTypeID]
		if !ok {
			loaded, err := c.alarmTypeRepo.GetWithFields(ctx, *obj.AlarmTypeID)
			if err != nil {
				return nil, err
			}
			if loaded == nil {
				return nil, domain.ErrNotFound
			}
			alarmType = loaded
			alarmTypeCache[*obj.AlarmTypeID] = loaded
		}

		for _, field := range alarmType.Fields {
			value := &domainFacility.BacnetObjectAlarmValue{
				BacnetObjectID:   obj.ID,
				AlarmTypeFieldID: field.ID,
				UnitID:           field.DefaultUnitID,
				Source:           domainFacility.AlarmValueSourceDefault,
			}

			if field.DefaultValueJSON != nil && field.AlarmField != nil {
				applyAlarmDefaultValue(value, field.AlarmField.DataType, *field.DefaultValueJSON)
			}

			values = append(values, value)
		}
	}

	return values, nil
}

func applyAlarmDefaultValue(value *domainFacility.BacnetObjectAlarmValue, dataType string, defaultValueJSON string) {
	if value == nil {
		return
	}

	var decoded any
	if err := json.Unmarshal([]byte(defaultValueJSON), &decoded); err != nil {
		value.ValueString = &defaultValueJSON
		return
	}

	switch strings.ToLower(strings.TrimSpace(dataType)) {
	case "number", "duration":
		if n, ok := toFloat64(decoded); ok {
			value.ValueNumber = &n
		}
	case "integer":
		if n, ok := toInt64(decoded); ok {
			value.ValueInteger = &n
		}
	case "boolean":
		if b, ok := decoded.(bool); ok {
			value.ValueBoolean = &b
		}
	case "string", "enum":
		if s, ok := decoded.(string); ok {
			value.ValueString = &s
		}
	case "state_map", "json":
		if b, err := json.Marshal(decoded); err == nil {
			raw := string(b)
			value.ValueJSON = &raw
		}
	default:
		if b, err := json.Marshal(decoded); err == nil {
			raw := string(b)
			value.ValueJSON = &raw
		}
	}
}

func toFloat64(value any) (float64, bool) {
	switch v := value.(type) {
	case float64:
		return v, true
	case float32:
		return float64(v), true
	case int:
		return float64(v), true
	case int8:
		return float64(v), true
	case int16:
		return float64(v), true
	case int32:
		return float64(v), true
	case int64:
		return float64(v), true
	case uint:
		return float64(v), true
	case uint8:
		return float64(v), true
	case uint16:
		return float64(v), true
	case uint32:
		return float64(v), true
	case uint64:
		return float64(v), true
	default:
		return 0, false
	}
}

func toInt64(value any) (int64, bool) {
	switch v := value.(type) {
	case int:
		return int64(v), true
	case int8:
		return int64(v), true
	case int16:
		return int64(v), true
	case int32:
		return int64(v), true
	case int64:
		return v, true
	case uint:
		return int64(v), true
	case uint8:
		return int64(v), true
	case uint16:
		return int64(v), true
	case uint32:
		return int64(v), true
	case uint64:
		return int64(v), true
	case float64:
		return int64(v), true
	case float32:
		return int64(v), true
	default:
		return 0, false
	}
}

func (c projectFacilityCopy) listSPSControllersByControlCabinetID(ctx context.Context, controlCabinetID uuid.UUID) ([]domainFacility.SPSController, error) {
	result, err := c.spsControllerRepo.GetPaginatedListByControlCabinetID(ctx, controlCabinetID, domain.PaginationParams{
		Page:  1,
		Limit: 500,
	})
	if err != nil {
		return nil, err
	}
	return result.Items, nil
}

func (c projectFacilityCopy) listSystemTypesBySPSControllerID(ctx context.Context, spsControllerID uuid.UUID) ([]domainFacility.SPSControllerSystemType, error) {
	result, err := c.spsControllerSystemRepo.GetPaginatedListBySPSControllerID(ctx, spsControllerID, domain.PaginationParams{
		Page:  1,
		Limit: 500,
	})
	if err != nil {
		return nil, err
	}
	return result.Items, nil
}

func (c projectFacilityCopy) nextAvailableControlCabinetNr(ctx context.Context, buildingID uuid.UUID, base string) (string, error) {
	for i := 1; i <= 9999; i++ {
		candidate := nextIncrementedValue(base, i, 11)
		exists, err := c.controlCabinetRepo.ExistsControlCabinetNr(ctx, buildingID, candidate, nil)
		if err != nil {
			return "", err
		}
		if !exists {
			return candidate, nil
		}
	}

	return "", domain.ErrConflict
}

func (c projectFacilityCopy) nextAvailableGADevice(ctx context.Context, controlCabinetID uuid.UUID) (string, error) {
	devices, err := c.spsControllerRepo.ListGADevicesByControlCabinetID(ctx, controlCabinetID)
	if err != nil {
		return "", err
	}

	used := make(map[string]struct{}, len(devices))
	for _, device := range devices {
		normalized := strings.ToUpper(strings.TrimSpace(device))
		if isValidGADevice(normalized) {
			used[normalized] = struct{}{}
		}
	}

	if next, ok := findLowestAvailableGADevice(used); ok {
		return next, nil
	}
	return "", domain.ErrConflict
}

func (c projectFacilityCopy) loadSystemTypeDefinitions(
	ctx context.Context,
	systemTypes []domainFacility.SPSControllerSystemType,
) (map[uuid.UUID]domainFacility.SystemType, error) {
	if len(systemTypes) == 0 {
		return map[uuid.UUID]domainFacility.SystemType{}, nil
	}

	unique := make(map[uuid.UUID]struct{}, len(systemTypes))
	ids := make([]uuid.UUID, 0, len(systemTypes))
	for _, st := range systemTypes {
		if st.SystemTypeID == uuid.Nil {
			return nil, domain.ErrNotFound
		}
		if _, ok := unique[st.SystemTypeID]; ok {
			continue
		}
		unique[st.SystemTypeID] = struct{}{}
		ids = append(ids, st.SystemTypeID)
	}

	found, err := c.systemTypeRepo.GetByIds(ctx, ids)
	if err != nil {
		return nil, err
	}
	if len(found) != len(ids) {
		return nil, domain.ErrNotFound
	}

	mapOut := make(map[uuid.UUID]domainFacility.SystemType, len(found))
	for _, item := range found {
		mapOut[item.ID] = *item
	}
	return mapOut, nil
}

func loadSystemTypeDefinitions(
	ctx context.Context,
	systemTypeRepo domainFacility.SystemTypeRepository,
	systemTypes []domainFacility.SPSControllerSystemType,
) (map[uuid.UUID]domainFacility.SystemType, error) {
	return projectFacilityCopy{systemTypeRepo: systemTypeRepo}.loadSystemTypeDefinitions(ctx, systemTypes)
}

func assignSystemTypeNumbers(
	systemTypes []domainFacility.SPSControllerSystemType,
	systemTypeMap map[uuid.UUID]domainFacility.SystemType,
) error {
	if len(systemTypes) == 0 {
		return nil
	}

	ve := domain.NewValidationError()
	usedNumbers := make(map[uuid.UUID]map[int]struct{}, len(systemTypes))

	for _, st := range systemTypes {
		systemType, ok := systemTypeMap[st.SystemTypeID]
		if !ok {
			return domain.ErrNotFound
		}
		if st.Number == nil {
			continue
		}
		number := *st.Number
		if number < systemType.NumberMin || number > systemType.NumberMax {
			ve = ve.Add("spscontroller.system_types", "number must be within the system type range")
			continue
		}
		if usedNumbers[st.SystemTypeID] == nil {
			usedNumbers[st.SystemTypeID] = map[int]struct{}{}
		}
		if _, exists := usedNumbers[st.SystemTypeID][number]; exists {
			ve = ve.Add("spscontroller.system_types", "number must be unique per system type")
			continue
		}
		usedNumbers[st.SystemTypeID][number] = struct{}{}
	}

	if len(ve.Fields) > 0 {
		return ve
	}

	for i := range systemTypes {
		if systemTypes[i].Number != nil {
			continue
		}
		systemType, ok := systemTypeMap[systemTypes[i].SystemTypeID]
		if !ok {
			return domain.ErrNotFound
		}
		if usedNumbers[systemTypes[i].SystemTypeID] == nil {
			usedNumbers[systemTypes[i].SystemTypeID] = map[int]struct{}{}
		}
		next, ok := findLowestAvailableNumber(systemType.NumberMin, systemType.NumberMax, usedNumbers[systemTypes[i].SystemTypeID])
		if !ok {
			return domain.NewValidationError().Add("spscontroller.system_types", "no available number in the system type range")
		}
		systemTypes[i].Number = &next
		usedNumbers[systemTypes[i].SystemTypeID][next] = struct{}{}
	}

	return nil
}
