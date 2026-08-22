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

const apparatNrAlreadyUsedMessage = "apparatnummer ist bereits vergeben"

func (s fieldDeviceBacnetSelection) validate() error {
	if s.objectDataID != nil && s.objectsSet {
		return domain.ErrInvalidArgument
	}
	return nil
}

func (w fieldDeviceWriter) create(ctx context.Context, fieldDevice *domainFacility.FieldDevice, selection fieldDeviceBacnetSelection) error {
	return w.service.transaction().run(ctx, func(txCtx context.Context, txService *FieldDeviceService) error {
		return txService.writer().createInTx(txCtx, fieldDevice, selection)
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
	return w.service.transaction().run(ctx, func(txCtx context.Context, txService *FieldDeviceService) error {
		return txService.writer().updateInTx(txCtx, fieldDevice, selection)
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
	return w.service.transaction().run(ctx, func(txCtx context.Context, txService *FieldDeviceService) error {
		return txService.writer().createSpecificationInTx(txCtx, fieldDeviceID, specification)
	})
}

func (w fieldDeviceWriter) createSpecificationInTx(ctx context.Context, fieldDeviceID uuid.UUID, specification *domainFacility.Specification) error {
	return newFieldDeviceSpecificationWriter(w).createInTx(ctx, fieldDeviceID, specification)
}

func (w fieldDeviceWriter) updateSpecificationPatch(ctx context.Context, fieldDeviceID uuid.UUID, patch *domainFacility.SpecificationPatch) (*domainFacility.Specification, error) {
	return runWithFacilityTxResult(ctx, w.service.transaction(), func(txCtx context.Context, txService *FieldDeviceService) (*domainFacility.Specification, error) {
		return txService.writer().updateSpecificationPatchInTx(txCtx, fieldDeviceID, patch)
	})
}

func (w fieldDeviceWriter) updateSpecificationPatchInTx(ctx context.Context, fieldDeviceID uuid.UUID, patch *domainFacility.SpecificationPatch) (*domainFacility.Specification, error) {
	return newFieldDeviceSpecificationWriter(w).updatePatchInTx(ctx, fieldDeviceID, patch)
}

func (w fieldDeviceWriter) applySpecificationPatch(ctx context.Context, fieldDeviceID uuid.UUID, patch *domainFacility.SpecificationPatch) error {
	return newFieldDeviceSpecificationWriter(w).applyPatch(ctx, fieldDeviceID, patch)
}

func (w fieldDeviceWriter) multiCreate(ctx context.Context, items []domainFacility.FieldDeviceCreateItem) *domainFacility.FieldDeviceMultiCreateResult {
	result := &domainFacility.FieldDeviceMultiCreateResult{
		Results:       make([]domainFacility.FieldDeviceCreateResult, len(items)),
		TotalRequests: len(items),
		SuccessCount:  0,
		FailureCount:  0,
	}

	planner := newFieldDeviceCreatePlanner(ctx, w.service)
	createWork := planner.prepare(items, result)

	if len(createWork) == 0 {
		return result
	}

	executeFieldDeviceCreateWork(fieldDeviceCreateExecution{ctx: ctx, writer: w, work: createWork, result: result})

	return result
}

func (w fieldDeviceWriter) bulkUpdate(ctx context.Context, updates []domainFacility.BulkFieldDeviceUpdate) *domainFacility.BulkOperationResult {
	return newFieldDeviceBulkUpdater(ctx, w).run(updates)
}

var errBulkUpdateItem = domain.ErrInvalidArgument

func (w fieldDeviceWriter) validateBulkBaseUpdate(
	ctx context.Context,
	execution bulkUpdateExecution,
) {
	if !w.validateBulkDevice(ctx, execution) {
		return
	}
	if !hasApparatNrConstraintUpdates(execution.update) {
		return
	}
	if w.hasBatchApparatNumberConflict(ctx, execution) {
		return
	}
	w.validatePersistedApparatNumber(ctx, execution)
}

// applyBulkSpecificationPatch mutates the specification owned by fieldDevice
// without writing the aggregate root. bulkUpdate persists the root once after
// every owned child succeeds, which gives the whole item one version change
// and one transaction boundary.
func (w fieldDeviceWriter) applyBulkSpecificationPatch(
	ctx context.Context,
	fieldDevice *domainFacility.FieldDevice,
	patch *domainFacility.SpecificationPatch,
) error {
	return newFieldDeviceSpecificationWriter(w).applyBulkPatch(ctx, fieldDevice, patch)
}

func buildProposedFieldDevice(existing *domainFacility.FieldDevice, update domainFacility.BulkFieldDeviceUpdate) *domainFacility.FieldDevice {
	clone := *existing
	clone.Version = update.BaseVersion.Uint64()
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

func sameApparatNrScope(a, b *domainFacility.FieldDevice) bool {
	if a == nil || b == nil {
		return false
	}
	return a.SPSControllerSystemTypeID == b.SPSControllerSystemTypeID &&
		a.ApparatID == b.ApparatID &&
		a.SystemPartID == b.SystemPartID
}

func isApparatNrAlreadyUsedError(err error) bool {
	ve, ok := domain.AsValidationError(err)
	if !ok {
		return false
	}
	return ve.Fields[fieldDeviceApparatNrField.Key] == apparatNrAlreadyUsedMessage
}

func (w fieldDeviceWriter) addAvailableApparatNrSuggestions(request apparatNumberSuggestionRequest) {
	if request.scope == nil || request.report == nil {
		return
	}
	usedSet, err := w.usedApparatNumberSet(request)
	if err != nil {
		return
	}
	for _, existing := range request.existing {
		if sameApparatNrScope(existing, request.scope) {
			delete(usedSet, existing.ApparatNr)
		}
	}
	for id, proposed := range request.batch {
		if id == request.updateID || !sameApparatNrScope(proposed, request.scope) {
			continue
		}
		if proposed.ApparatNr >= 1 && proposed.ApparatNr <= 99 {
			usedSet[proposed.ApparatNr] = struct{}{}
		}
	}

	available := availableApparatNumbers(usedSet)
	if len(available) == 0 {
		return
	}
	request.report.suggestions[fieldDeviceApparatNrField.Key] = available[0]
	request.report.suggestionOptions[fieldDeviceApparatNrField.Key] = available
}

func (w fieldDeviceWriter) usedApparatNumberSet(request apparatNumberSuggestionRequest) (map[int]struct{}, error) {
	usedNumbers, err := w.service.repo.GetUsedApparatNumbers(
		request.ctx, request.scope.SPSControllerSystemTypeID, request.scope.SystemPartID, request.scope.ApparatID,
	)
	if err != nil {
		return nil, err
	}
	used := make(map[int]struct{}, len(usedNumbers))
	for _, number := range usedNumbers {
		if number >= 1 && number <= 99 {
			used[number] = struct{}{}
		}
	}
	return used, nil
}

func availableApparatNumbers(used map[int]struct{}) []int {
	available := make([]int, 0, 99-len(used))
	for number := 1; number <= 99; number++ {
		if _, exists := used[number]; !exists {
			available = append(available, number)
		}
	}
	return available
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
