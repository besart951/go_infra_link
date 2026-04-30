package facility

import (
	"context"
	"maps"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/besart951/go_infra_link/backend/internal/service/changecapture"
	"github.com/google/uuid"
)

type fieldDeviceWriter struct {
	service *FieldDeviceService
}

type fieldDeviceBacnetSelection struct {
	objectDataID *uuid.UUID
	objects      []domainFacility.BacnetObject
	objectsSet   bool
}

func (s fieldDeviceBacnetSelection) validate() error {
	if s.objectDataID != nil && s.objectsSet {
		return domain.ErrInvalidArgument
	}
	return nil
}

func (w fieldDeviceWriter) create(ctx context.Context, fieldDevice *domainFacility.FieldDevice, selection fieldDeviceBacnetSelection) error {
	return w.service.transaction().run(func(txService *FieldDeviceService) error {
		return txService.writer().createInTx(ctx, fieldDevice, selection)
	})
}

func (w fieldDeviceWriter) createInTx(ctx context.Context, fieldDevice *domainFacility.FieldDevice, selection fieldDeviceBacnetSelection) error {
	if err := selection.validate(); err != nil {
		return err
	}
	if err := w.service.Validate(ctx, fieldDevice, nil); err != nil {
		return err
	}
	if err := w.service.repo.Create(ctx, fieldDevice); err != nil {
		return err
	}
	if err := w.applyBacnetSelection(ctx, fieldDevice.ID, selection); err != nil {
		return err
	}
	return w.service.recordFieldDeviceChange(ctx, changecapture.ActionCreated, fieldDevice.ID)
}

func (w fieldDeviceWriter) updateBase(ctx context.Context, fieldDevice *domainFacility.FieldDevice) error {
	if err := w.service.Validate(ctx, fieldDevice, &fieldDevice.ID); err != nil {
		return err
	}
	if err := w.service.repo.Update(ctx, fieldDevice); err != nil {
		return err
	}
	return w.service.recordFieldDeviceChange(ctx, changecapture.ActionUpdated, fieldDevice.ID)
}

func (w fieldDeviceWriter) update(ctx context.Context, fieldDevice *domainFacility.FieldDevice, selection fieldDeviceBacnetSelection) error {
	return w.service.transaction().run(func(txService *FieldDeviceService) error {
		return txService.writer().updateInTx(ctx, fieldDevice, selection)
	})
}

func (w fieldDeviceWriter) updateInTx(ctx context.Context, fieldDevice *domainFacility.FieldDevice, selection fieldDeviceBacnetSelection) error {
	if err := selection.validate(); err != nil {
		return err
	}
	if err := w.service.Validate(ctx, fieldDevice, &fieldDevice.ID); err != nil {
		return err
	}
	if err := w.service.repo.Update(ctx, fieldDevice); err != nil {
		return err
	}
	if err := w.applyBacnetSelection(ctx, fieldDevice.ID, selection); err != nil {
		return err
	}
	return w.service.recordFieldDeviceChange(ctx, changecapture.ActionUpdated, fieldDevice.ID)
}

func (w fieldDeviceWriter) applyBacnetSelection(ctx context.Context, fieldDeviceID uuid.UUID, selection fieldDeviceBacnetSelection) error {
	if selection.objectDataID != nil {
		return w.service.replaceBacnetObjectsFromObjectData(ctx, fieldDeviceID, *selection.objectDataID)
	}
	if selection.objectsSet {
		return w.service.replaceBacnetObjects(ctx, fieldDeviceID, selection.objects)
	}
	return nil
}

func (w fieldDeviceWriter) createSpecification(ctx context.Context, fieldDeviceID uuid.UUID, specification *domainFacility.Specification) error {
	return w.service.transaction().run(func(txService *FieldDeviceService) error {
		return txService.writer().createSpecificationInTx(ctx, fieldDeviceID, specification)
	})
}

func (w fieldDeviceWriter) createSpecificationInTx(ctx context.Context, fieldDeviceID uuid.UUID, specification *domainFacility.Specification) error {
	fieldDevice, err := domain.GetByID(ctx, w.service.repo, fieldDeviceID)
	if err != nil {
		return err
	}

	existing, err := w.service.specificationRepo.GetByFieldDeviceIDs(ctx, []uuid.UUID{fieldDeviceID})
	if err != nil {
		return err
	}
	if len(existing) > 0 {
		return domain.ErrConflict
	}

	specification.SpecificationSupplier = normalizeOptionalString(specification.SpecificationSupplier)
	specification.SpecificationBrand = normalizeOptionalString(specification.SpecificationBrand)
	specification.SpecificationType = normalizeOptionalString(specification.SpecificationType)
	specification.AdditionalInfoMotorValve = normalizeOptionalString(specification.AdditionalInfoMotorValve)
	specification.AdditionalInformationInstallationLocation = normalizeOptionalString(specification.AdditionalInformationInstallationLocation)
	specification.ElectricalConnectionACDC = normalizeOptionalString(specification.ElectricalConnectionACDC)
	if err := w.service.validateSpecification(specification); err != nil {
		return err
	}

	id := fieldDeviceID
	specification.FieldDeviceID = &id
	if err := w.service.specificationRepo.Create(ctx, specification); err != nil {
		return err
	}

	fieldDevice.SpecificationID = &specification.ID
	return w.service.repo.Update(ctx, fieldDevice)
}

func (w fieldDeviceWriter) updateSpecificationPatch(ctx context.Context, fieldDeviceID uuid.UUID, patch *domainFacility.SpecificationPatch) (*domainFacility.Specification, error) {
	if _, err := domain.GetByID(ctx, w.service.repo, fieldDeviceID); err != nil {
		return nil, err
	}

	specs, err := w.service.specificationRepo.GetByFieldDeviceIDs(ctx, []uuid.UUID{fieldDeviceID})
	if err != nil {
		return nil, err
	}
	if len(specs) == 0 {
		return nil, domain.ErrNotFound
	}
	spec := specs[0]

	applySpecificationPatch(spec, patch)
	if err := w.service.validateSpecification(spec); err != nil {
		return nil, err
	}

	if err := w.service.specificationRepo.Update(ctx, spec); err != nil {
		return nil, err
	}
	return spec, nil
}

func (w fieldDeviceWriter) applySpecificationPatch(ctx context.Context, fieldDeviceID uuid.UUID, patch *domainFacility.SpecificationPatch) error {
	if patch == nil || !patch.HasChanges() {
		return nil
	}

	if _, err := domain.GetByID(ctx, w.service.repo, fieldDeviceID); err != nil {
		return err
	}

	specs, err := w.service.specificationRepo.GetByFieldDeviceIDs(ctx, []uuid.UUID{fieldDeviceID})
	if err != nil {
		return err
	}

	if len(specs) == 0 {
		if !patch.HasNonNilValues() {
			return nil
		}

		spec := &domainFacility.Specification{
			SpecificationSupplier:                     normalizeOptionalString(patch.SpecificationSupplier),
			SpecificationBrand:                        normalizeOptionalString(patch.SpecificationBrand),
			SpecificationType:                         normalizeOptionalString(patch.SpecificationType),
			AdditionalInfoMotorValve:                  normalizeOptionalString(patch.AdditionalInfoMotorValve),
			AdditionalInfoSize:                        patch.AdditionalInfoSize,
			AdditionalInformationInstallationLocation: normalizeOptionalString(patch.AdditionalInformationInstallationLocation),
			ElectricalConnectionPH:                    patch.ElectricalConnectionPH,
			ElectricalConnectionACDC:                  normalizeOptionalString(patch.ElectricalConnectionACDC),
			ElectricalConnectionAmperage:              patch.ElectricalConnectionAmperage,
			ElectricalConnectionPower:                 patch.ElectricalConnectionPower,
			ElectricalConnectionRotation:              patch.ElectricalConnectionRotation,
		}
		return w.createSpecification(ctx, fieldDeviceID, spec)
	}

	spec := specs[0]
	applySpecificationPatch(spec, patch)

	if err := w.service.validateSpecification(spec); err != nil {
		return err
	}
	return w.service.specificationRepo.Update(ctx, spec)
}

func (w fieldDeviceWriter) multiCreate(ctx context.Context, items []domainFacility.FieldDeviceCreateItem) *domainFacility.FieldDeviceMultiCreateResult {
	result := &domainFacility.FieldDeviceMultiCreateResult{
		Results:       make([]domainFacility.FieldDeviceCreateResult, len(items)),
		TotalRequests: len(items),
		SuccessCount:  0,
		FailureCount:  0,
	}

	stsCache := make(map[uuid.UUID]*domainFacility.SPSControllerSystemType)
	systemTypeCache := make(map[uuid.UUID]bool)
	apparatCache := make(map[uuid.UUID]bool)
	systemPartCache := make(map[uuid.UUID]bool)
	usedApparatNumbersCache := make(map[string]map[int]struct{})

	getApparatNumbersKey := func(spsControllerSystemTypeID uuid.UUID, systemPartID uuid.UUID, apparatID uuid.UUID) string {
		return spsControllerSystemTypeID.String() + "|" + apparatID.String() + "|" + systemPartID.String()
	}

	ensureParentsExistCached := func(fieldDevice *domainFacility.FieldDevice) error {
		sts, ok := stsCache[fieldDevice.SPSControllerSystemTypeID]
		if !ok {
			loaded, err := domain.GetByID(ctx, w.service.spsControllerSystemTypeRepo, fieldDevice.SPSControllerSystemTypeID)
			if err != nil {
				stsCache[fieldDevice.SPSControllerSystemTypeID] = nil
				return err
			}
			sts = loaded
			stsCache[fieldDevice.SPSControllerSystemTypeID] = loaded
		}
		if sts == nil {
			return domain.ErrNotFound
		}

		if _, ok := systemTypeCache[sts.SystemTypeID]; !ok {
			if _, err := domain.GetByID(ctx, w.service.systemTypeRepo, sts.SystemTypeID); err != nil {
				return err
			}
			systemTypeCache[sts.SystemTypeID] = true
		}

		if _, ok := apparatCache[fieldDevice.ApparatID]; !ok {
			if _, err := domain.GetByID(ctx, w.service.apparatRepo, fieldDevice.ApparatID); err != nil {
				return err
			}
			apparatCache[fieldDevice.ApparatID] = true
		}

		if _, ok := systemPartCache[fieldDevice.SystemPartID]; !ok {
			if _, err := domain.GetByID(ctx, w.service.systemPartRepo, fieldDevice.SystemPartID); err != nil {
				return err
			}
			systemPartCache[fieldDevice.SystemPartID] = true
		}

		return nil
	}

	type createWorkItem struct {
		index int
		item  domainFacility.FieldDeviceCreateItem
	}
	createWork := make([]createWorkItem, 0, len(items))

	for i, item := range items {
		createResult := &result.Results[i]
		createResult.Index = i
		createResult.Success = false

		if item.FieldDevice == nil {
			createResult.Error = "field device is required"
			createResult.ErrorField = "fielddevice"
			result.FailureCount++
			continue
		}

		if item.ObjectDataID != nil && len(item.BacnetObjects) > 0 {
			createResult.Error = "object_data_id and bacnet_objects are mutually exclusive"
			createResult.ErrorField = "fielddevice"
			result.FailureCount++
			continue
		}

		if err := w.service.validateRequiredFields(item.FieldDevice); err != nil {
			setFieldDeviceCreateError(createResult, err, "fielddevice")
			result.FailureCount++
			continue
		}

		if err := ensureParentsExistCached(item.FieldDevice); err != nil {
			if err == domain.ErrNotFound {
				createResult.Error = "one or more parent entities (SPS controller, apparat, system part) not found"
				createResult.ErrorField = "fielddevice"
			} else {
				createResult.Error = err.Error()
				createResult.ErrorField = "fielddevice"
			}
			result.FailureCount++
			continue
		}

		if item.FieldDevice.ApparatNr == 0 {
			createResult.Error = "apparat_nr is required"
			createResult.ErrorField = "fielddevice.apparat_nr"
			result.FailureCount++
			continue
		}
		if item.FieldDevice.ApparatNr < 1 || item.FieldDevice.ApparatNr > 99 {
			createResult.Error = "apparat_nr must be between 1 and 99"
			createResult.ErrorField = "fielddevice.apparat_nr"
			result.FailureCount++
			continue
		}

		key := getApparatNumbersKey(
			item.FieldDevice.SPSControllerSystemTypeID,
			item.FieldDevice.SystemPartID,
			item.FieldDevice.ApparatID,
		)
		usedSet, ok := usedApparatNumbersCache[key]
		if !ok {
			usedNumbers, err := w.service.repo.GetUsedApparatNumbers(ctx,
				item.FieldDevice.SPSControllerSystemTypeID,
				item.FieldDevice.SystemPartID,
				item.FieldDevice.ApparatID,
			)
			if err != nil {
				createResult.Error = err.Error()
				createResult.ErrorField = "fielddevice.apparat_nr"
				result.FailureCount++
				continue
			}
			usedSet = make(map[int]struct{}, len(usedNumbers))
			for _, n := range usedNumbers {
				usedSet[n] = struct{}{}
			}
			usedApparatNumbersCache[key] = usedSet
		}
		if _, exists := usedSet[item.FieldDevice.ApparatNr]; exists {
			createResult.Error = "apparatnummer ist bereits vergeben"
			createResult.ErrorField = "fielddevice.apparat_nr"
			result.FailureCount++
			continue
		}
		usedSet[item.FieldDevice.ApparatNr] = struct{}{}
		createWork = append(createWork, createWorkItem{index: i, item: item})
	}

	if len(createWork) == 0 {
		return result
	}

	for _, work := range createWork {
		createResult := &result.Results[work.index]
		selection := fieldDeviceBacnetSelection{
			objectDataID: work.item.ObjectDataID,
			objects:      work.item.BacnetObjects,
			objectsSet:   len(work.item.BacnetObjects) > 0,
		}
		if err := w.create(ctx, work.item.FieldDevice, selection); err != nil {
			setFieldDeviceCreateError(createResult, err, "fielddevice")
			result.FailureCount++
			continue
		}

		createResult.Success = true
		createResult.FieldDevice = work.item.FieldDevice
		result.SuccessCount++
	}

	return result
}

func (w fieldDeviceWriter) bulkUpdate(ctx context.Context, updates []domainFacility.BulkFieldDeviceUpdate) *domainFacility.BulkOperationResult {
	result := &domainFacility.BulkOperationResult{
		Results:      make([]domainFacility.BulkOperationResultItem, len(updates)),
		TotalCount:   len(updates),
		SuccessCount: 0,
		FailureCount: 0,
	}

	ids := make([]uuid.UUID, 0, len(updates))
	for _, u := range updates {
		ids = append(ids, u.ID)
	}

	existingItems, err := w.service.repo.GetByIds(ctx, ids)
	if err != nil {
		for i := range result.Results {
			result.Results[i].Error = "failed to fetch existing items: " + err.Error()
			result.FailureCount++
		}
		return result
	}

	existingMap := make(map[uuid.UUID]*domainFacility.FieldDevice)
	for _, item := range existingItems {
		existingMap[item.ID] = item
	}

	proposedMap := make(map[uuid.UUID]*domainFacility.FieldDevice)
	for _, update := range updates {
		existing, ok := existingMap[update.ID]
		if !ok {
			continue
		}
		proposedMap[update.ID] = buildProposedFieldDevice(existing, update)
	}

	for i, update := range updates {
		resultItem := &result.Results[i]
		resultItem.ID = update.ID
		resultItem.Success = false
		resultItem.Fields = make(map[string]string)

		if _, ok := existingMap[update.ID]; !ok {
			resultItem.Error = "field device not found"
			result.FailureCount++
			continue
		}

		proposed, ok := proposedMap[update.ID]
		if !ok {
			resultItem.Error = "field device not found"
			result.FailureCount++
			continue
		}

		phaseErrors := make(map[string]string)
		totalPhases := 0

		if hasBaseFieldDeviceUpdates(update) {
			totalPhases++
			w.applyBulkBaseUpdate(ctx, proposed, update, ids, proposedMap, phaseErrors)
		}

		if update.Specification != nil && update.Specification.HasChanges() {
			totalPhases++
			if err := w.applySpecificationPatch(ctx, proposed.ID, update.Specification); err != nil {
				phaseErrors["specification"] = "failed to update specification: " + err.Error()
			}
		}

		if update.BacnetObjects != nil {
			totalPhases++
			if err := w.service.patchBacnetObjects(ctx, proposed.ID, *update.BacnetObjects); err != nil {
				addBulkUpdateError(phaseErrors, "bacnet_objects", "failed to update BACnet objects: ", err)
			}
		}

		if len(phaseErrors) == 0 && totalPhases > 0 {
			resultItem.Success = true
			result.SuccessCount++
		} else {
			resultItem.Fields = phaseErrors
			for _, msg := range phaseErrors {
				resultItem.Error = msg
				break
			}
			result.FailureCount++
		}
	}

	return result
}

func (w fieldDeviceWriter) applyBulkBaseUpdate(
	ctx context.Context,
	proposed *domainFacility.FieldDevice,
	update domainFacility.BulkFieldDeviceUpdate,
	batchIDs []uuid.UUID,
	proposedMap map[uuid.UUID]*domainFacility.FieldDevice,
	phaseErrors map[string]string,
) {
	if err := w.service.validateRequiredFields(proposed); err != nil {
		addBulkUpdateError(phaseErrors, "fielddevice", "", err)
		return
	}

	if err := w.service.ensureParentsExist(ctx, proposed); err != nil {
		if err == domain.ErrNotFound {
			phaseErrors["fielddevice"] = "one or more parent entities not found"
		} else {
			phaseErrors["fielddevice"] = err.Error()
		}
		return
	}

	if hasApparatNrConstraintUpdates(update) {
		for otherID, otherProposed := range proposedMap {
			if otherID == update.ID {
				continue
			}
			if isApparatNrConflict(proposed, otherProposed) {
				phaseErrors["fielddevice.apparat_nr"] = "apparatnummer ist bereits vergeben"
				return
			}
		}

		if err := w.service.ensureApparatNrAvailableWithExclusions(ctx, proposed, batchIDs); err != nil {
			addBulkUpdateError(phaseErrors, "fielddevice.apparat_nr", "", err)
			return
		}
	}

	if err := w.service.repo.Update(ctx, proposed); err != nil {
		phaseErrors["fielddevice"] = err.Error()
		return
	}
	if err := w.service.recordFieldDeviceChange(ctx, changecapture.ActionUpdated, proposed.ID); err != nil {
		phaseErrors["fielddevice"] = err.Error()
	}
}

func buildProposedFieldDevice(existing *domainFacility.FieldDevice, update domainFacility.BulkFieldDeviceUpdate) *domainFacility.FieldDevice {
	clone := *existing
	if update.HasBMKUpdate() {
		clone.BMK = normalizeOptionalString(update.BMK)
	}
	if update.HasDescriptionUpdate() {
		clone.Description = normalizeOptionalString(update.Description)
	}
	if update.HasTextIndividuellUpdate() {
		clone.TextIndividuell = normalizeOptionalString(update.TextIndividuell)
	}
	if update.ApparatNr != nil {
		clone.ApparatNr = *update.ApparatNr
	}
	if update.ApparatID != nil {
		clone.ApparatID = *update.ApparatID
	}
	if update.SystemPartID != nil {
		clone.SystemPartID = *update.SystemPartID
	}
	return &clone
}

func hasBaseFieldDeviceUpdates(update domainFacility.BulkFieldDeviceUpdate) bool {
	return update.HasBMKUpdate() ||
		update.HasDescriptionUpdate() ||
		update.HasTextIndividuellUpdate() ||
		update.ApparatNr != nil ||
		update.ApparatID != nil ||
		update.SystemPartID != nil
}

func hasApparatNrConstraintUpdates(update domainFacility.BulkFieldDeviceUpdate) bool {
	return update.ApparatNr != nil || update.ApparatID != nil || update.SystemPartID != nil
}

func isApparatNrConflict(a, b *domainFacility.FieldDevice) bool {
	if a.SPSControllerSystemTypeID != b.SPSControllerSystemTypeID {
		return false
	}
	if a.ApparatID != b.ApparatID {
		return false
	}
	if a.ApparatNr != b.ApparatNr {
		return false
	}
	return a.SystemPartID == b.SystemPartID
}

func addBulkUpdateError(fields map[string]string, fallbackField string, prefix string, err error) {
	if ve, ok := domain.AsValidationError(err); ok {
		maps.Copy(fields, ve.Fields)
		return
	}
	fields[fallbackField] = prefix + err.Error()
}

func setFieldDeviceCreateError(resultItem *domainFacility.FieldDeviceCreateResult, err error, defaultField string) {
	if ve, ok := domain.AsValidationError(err); ok {
		for field, msg := range ve.Fields {
			resultItem.Error = msg
			resultItem.ErrorField = field
			return
		}
	}
	resultItem.Error = err.Error()
	resultItem.ErrorField = defaultField
}
