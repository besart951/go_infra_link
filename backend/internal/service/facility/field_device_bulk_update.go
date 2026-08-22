package facility

import (
	"context"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/besart951/go_infra_link/backend/internal/service/changecapture"
	"github.com/google/uuid"
)

type bulkUpdateReport struct {
	errors            map[string]string
	suggestions       map[string]int
	suggestionOptions map[string][]int
}

type bulkUpdateExecution struct {
	update   domainFacility.BulkFieldDeviceUpdate
	proposed *domainFacility.FieldDevice
	batchIDs []uuid.UUID
	existing map[uuid.UUID]*domainFacility.FieldDevice
	batch    map[uuid.UUID]*domainFacility.FieldDevice
	report   *bulkUpdateReport
}

type apparatNumberSuggestionRequest struct {
	ctx      context.Context
	scope    *domainFacility.FieldDevice
	updateID uuid.UUID
	existing map[uuid.UUID]*domainFacility.FieldDevice
	batch    map[uuid.UUID]*domainFacility.FieldDevice
	report   *bulkUpdateReport
}

type bulkUpdateFailure struct {
	result *domainFacility.BulkOperationResult
	item   *domainFacility.BulkOperationResultItem
	report *bulkUpdateReport
	cause  error
}

type bulkUpdatePersistence struct {
	execution bulkUpdateExecution
	phases    int
}

type fieldDeviceBulkUpdater struct {
	ctx      context.Context
	writer   fieldDeviceWriter
	ids      []uuid.UUID
	existing map[uuid.UUID]*domainFacility.FieldDevice
	proposed map[uuid.UUID]*domainFacility.FieldDevice
	result   *domainFacility.BulkOperationResult
}

type fieldDeviceNumberSwapStore interface {
	LockNumberSwapRows(context.Context, []uuid.UUID) ([]*domainFacility.FieldDevice, error)
	DeferNumberConstraint(context.Context) error
}

func newFieldDeviceBulkUpdater(ctx context.Context, writer fieldDeviceWriter) *fieldDeviceBulkUpdater {
	return &fieldDeviceBulkUpdater{ctx: ctx, writer: writer}
}

func (u *fieldDeviceBulkUpdater) run(updates []domainFacility.BulkFieldDeviceUpdate) *domainFacility.BulkOperationResult {
	u.result = &domainFacility.BulkOperationResult{
		Results: make([]domainFacility.BulkOperationResultItem, len(updates)), TotalCount: len(updates),
	}
	u.ids = bulkUpdateIDs(updates)
	if err := u.loadCandidates(updates); err != nil {
		u.failAll(err)
		return u.result
	}
	for _, group := range planFieldDeviceUpdateGroups(updates, u.existing, u.proposed) {
		u.updateGroup(group, updates)
	}
	return u.result
}

func (u *fieldDeviceBulkUpdater) updateGroup(group fieldDeviceUpdateGroup, updates []domainFacility.BulkFieldDeviceUpdate) {
	if len(group.Indexes) == 1 {
		u.updateOne(group.Indexes[0], updates[group.Indexes[0]], group.ID)
		return
	}
	u.updateDependentGroup(group, updates)
}

func bulkUpdateIDs(updates []domainFacility.BulkFieldDeviceUpdate) []uuid.UUID {
	ids := make([]uuid.UUID, len(updates))
	for index := range updates {
		ids[index] = updates[index].ID
	}
	return ids
}

func (u *fieldDeviceBulkUpdater) loadCandidates(updates []domainFacility.BulkFieldDeviceUpdate) error {
	items, err := u.writer.service.repo.GetByIds(u.ctx, u.ids)
	if err != nil {
		return err
	}
	u.existing = make(map[uuid.UUID]*domainFacility.FieldDevice, len(items))
	for _, item := range items {
		u.existing[item.ID] = item
	}
	u.proposed = make(map[uuid.UUID]*domainFacility.FieldDevice, len(items))
	for _, update := range updates {
		if existing := u.existing[update.ID]; existing != nil {
			u.proposed[update.ID] = buildProposedFieldDevice(existing, update)
		}
	}
	return nil
}

func (u *fieldDeviceBulkUpdater) failAll(err error) {
	for index := range u.result.Results {
		u.result.Results[index].Error = "failed to fetch existing items: " + err.Error()
		u.result.FailureCount++
	}
}

func (u *fieldDeviceBulkUpdater) updateOne(index int, update domainFacility.BulkFieldDeviceUpdate, groupID uuid.UUID) {
	resultItem := &u.result.Results[index]
	prepareBulkResultItem(resultItem, update.ID)
	resultItem.DependencyGroupID = groupID
	proposed, ok := u.resolveCandidate(update, resultItem)
	if !ok {
		u.result.FailureCount++
		return
	}
	report := newBulkUpdateReport()
	err := u.execute(update, proposed, report)
	if err == nil {
		completeBulkUpdateSuccess(u.result, resultItem, proposed)
		return
	}
	completeBulkUpdateFailure(bulkUpdateFailure{
		result: u.result, item: resultItem, report: report, cause: err,
	})
}

func (u *fieldDeviceBulkUpdater) updateDependentGroup(group fieldDeviceUpdateGroup, updates []domainFacility.BulkFieldDeviceUpdate) {
	executions, ok := u.prepareGroupExecutions(group, updates)
	if !ok {
		return
	}
	err := u.writer.service.transaction().run(u.ctx, func(txCtx context.Context, service *FieldDeviceService) error {
		if err := lockNumberSwapGroup(txCtx, service, executions); err != nil {
			return err
		}
		for _, execution := range executions {
			if err := runBulkUpdatePhases(txCtx, service, execution); err != nil {
				return err
			}
		}
		return nil
	})
	err = mapFieldDeviceNumberConflict(err)
	u.completeGroup(group, executions, err)
}

func lockNumberSwapGroup(ctx context.Context, service *FieldDeviceService, executions []bulkUpdateExecution) error {
	store, ok := service.repo.(fieldDeviceNumberSwapStore)
	if !ok {
		return domain.ErrInvalidArgument
	}
	ids := make([]uuid.UUID, len(executions))
	for index := range executions {
		ids[index] = executions[index].update.ID
	}
	locked, err := store.LockNumberSwapRows(ctx, ids)
	if err != nil {
		return err
	}
	if err := validateLockedSwapVersions(locked, executions); err != nil {
		return err
	}
	return store.DeferNumberConstraint(ctx)
}

func validateLockedSwapVersions(locked []*domainFacility.FieldDevice, executions []bulkUpdateExecution) error {
	versions := make(map[uuid.UUID]uint64, len(locked))
	for _, item := range locked {
		versions[item.ID] = item.Version
	}
	for _, execution := range executions {
		if execution.update.BaseVersion == 0 || versions[execution.update.ID] != execution.update.BaseVersion.Uint64() {
			return domain.ErrConflict
		}
	}
	return nil
}

func (u *fieldDeviceBulkUpdater) prepareGroupExecutions(group fieldDeviceUpdateGroup, updates []domainFacility.BulkFieldDeviceUpdate) ([]bulkUpdateExecution, bool) {
	executions := make([]bulkUpdateExecution, 0, len(group.Indexes))
	valid := true
	for _, index := range group.Indexes {
		item := &u.result.Results[index]
		prepareBulkResultItem(item, updates[index].ID)
		item.DependencyGroupID = group.ID
		proposed, ok := u.resolveCandidate(updates[index], item)
		if !ok {
			u.result.FailureCount++
			valid = false
			continue
		}
		executions = append(executions, bulkUpdateExecution{
			update: updates[index], proposed: proposed, batchIDs: u.ids,
			existing: u.existing, batch: u.proposed, report: newBulkUpdateReport(),
		})
	}
	if !valid {
		u.failUnresolvedGroupMembers(group, "dependency group validation failed")
	}
	return executions, valid
}

func (u *fieldDeviceBulkUpdater) failUnresolvedGroupMembers(group fieldDeviceUpdateGroup, message string) {
	for _, index := range group.Indexes {
		item := &u.result.Results[index]
		if item.Error != "" {
			continue
		}
		item.Error = message
		item.Fields["fielddevice"] = message
		u.result.FailureCount++
	}
}

func (u *fieldDeviceBulkUpdater) completeGroup(group fieldDeviceUpdateGroup, executions []bulkUpdateExecution, err error) {
	byID := make(map[uuid.UUID]bulkUpdateExecution, len(executions))
	for _, execution := range executions {
		byID[execution.update.ID] = execution
	}
	for _, index := range group.Indexes {
		execution := byID[u.result.Results[index].ID]
		if err == nil {
			completeBulkUpdateSuccess(u.result, &u.result.Results[index], execution.proposed)
			continue
		}
		completeBulkUpdateFailure(bulkUpdateFailure{
			result: u.result, item: &u.result.Results[index], report: execution.report, cause: err,
		})
	}
}

func (u *fieldDeviceBulkUpdater) resolveCandidate(update domainFacility.BulkFieldDeviceUpdate, result *domainFacility.BulkOperationResultItem) (*domainFacility.FieldDevice, bool) {
	existing := u.existing[update.ID]
	proposed := u.proposed[update.ID]
	if existing == nil || proposed == nil {
		result.Error = "field device not found"
		return nil, false
	}
	if update.BaseVersion == 0 {
		result.Error = domain.ErrInvalidArgument.Error()
		result.Fields["base_version"] = "required"
		return nil, false
	}
	if existing.Version != update.BaseVersion.Uint64() {
		result.Error = domain.ErrConflict.Error()
		result.Fields["fielddevice"] = "write_conflict"
		result.Version = existing.Version
		result.FieldDevice = existing
		return nil, false
	}
	return proposed, true
}

func (u *fieldDeviceBulkUpdater) execute(update domainFacility.BulkFieldDeviceUpdate, proposed *domainFacility.FieldDevice, report *bulkUpdateReport) error {
	execution := bulkUpdateExecution{
		update: update, proposed: proposed, batchIDs: u.ids,
		existing: u.existing, batch: u.proposed, report: report,
	}
	err := u.writer.service.transaction().run(u.ctx, func(txCtx context.Context, txService *FieldDeviceService) error {
		return runBulkUpdatePhases(txCtx, txService, execution)
	})
	err = mapFieldDeviceNumberConflict(err)
	if err != nil && len(report.errors) == 0 {
		report.errors["fielddevice"] = err.Error()
	}
	return err
}

func runBulkUpdatePhases(ctx context.Context, service *FieldDeviceService, execution bulkUpdateExecution) error {
	phases := 0
	if hasBaseFieldDeviceUpdates(execution.update) {
		phases++
		service.writer().validateBulkBaseUpdate(ctx, execution)
		if len(execution.report.errors) > 0 {
			return errBulkUpdateItem
		}
	}
	if execution.update.Specification != nil && execution.update.Specification.HasChanges() {
		phases++
		if err := service.writer().applyBulkSpecificationPatch(ctx, execution.proposed, execution.update.Specification); err != nil {
			addBulkUpdateError(execution.report.errors, "specification", "failed to update specification: ", err)
			return errBulkUpdateItem
		}
	}
	if execution.update.BacnetObjects != nil {
		phases++
		if err := service.patchBacnetObjects(ctx, execution.proposed.ID, *execution.update.BacnetObjects); err != nil {
			addBulkUpdateError(execution.report.errors, "bacnet_objects", "failed to update BACnet objects: ", err)
			return errBulkUpdateItem
		}
	}
	return persistBulkUpdate(ctx, service, bulkUpdatePersistence{execution: execution, phases: phases})
}

func persistBulkUpdate(ctx context.Context, service *FieldDeviceService, persistence bulkUpdatePersistence) error {
	if persistence.phases == 0 {
		persistence.execution.report.errors["fielddevice"] = "no changes provided"
		return errBulkUpdateItem
	}
	if err := service.repo.Update(ctx, persistence.execution.proposed); err != nil {
		addBulkUpdateError(persistence.execution.report.errors, "fielddevice", "", err)
		return errBulkUpdateItem
	}
	if err := service.recordFieldDeviceChange(ctx, changecapture.ActionUpdated, persistence.execution.proposed.ID); err != nil {
		persistence.execution.report.errors["fielddevice"] = err.Error()
		return errBulkUpdateItem
	}
	return nil
}

func (w fieldDeviceWriter) validateBulkDevice(ctx context.Context, execution bulkUpdateExecution) bool {
	if err := w.service.validateRequiredFields(execution.proposed); err != nil {
		addBulkUpdateError(execution.report.errors, "fielddevice", "", err)
		return false
	}
	if err := w.service.ensureParentsExist(ctx, execution.proposed); err != nil {
		if err == domain.ErrNotFound {
			execution.report.errors["fielddevice"] = "one or more parent entities not found"
		} else {
			execution.report.errors["fielddevice"] = err.Error()
		}
		return false
	}
	return true
}

func (w fieldDeviceWriter) hasBatchApparatNumberConflict(ctx context.Context, execution bulkUpdateExecution) bool {
	for otherID, proposed := range execution.batch {
		if otherID == execution.update.ID || !isApparatNrConflict(execution.proposed, proposed) {
			continue
		}
		execution.report.errors["fielddevice.apparat_nr"] = apparatNrAlreadyUsedMessage
		w.addAvailableApparatNrSuggestions(newApparatNumberSuggestionRequest(ctx, execution))
		return true
	}
	return false
}

func (w fieldDeviceWriter) validatePersistedApparatNumber(ctx context.Context, execution bulkUpdateExecution) {
	err := w.service.ensureApparatNrAvailableWithExclusions(ctx, execution.proposed, execution.batchIDs)
	if err == nil {
		return
	}
	if isApparatNrAlreadyUsedError(err) {
		w.addAvailableApparatNrSuggestions(newApparatNumberSuggestionRequest(ctx, execution))
	}
	addBulkUpdateError(execution.report.errors, "fielddevice.apparat_nr", "", err)
}

func newApparatNumberSuggestionRequest(ctx context.Context, execution bulkUpdateExecution) apparatNumberSuggestionRequest {
	return apparatNumberSuggestionRequest{
		ctx: ctx, scope: execution.proposed, updateID: execution.update.ID,
		existing: execution.existing, batch: execution.batch, report: execution.report,
	}
}

func newBulkUpdateReport() *bulkUpdateReport {
	return &bulkUpdateReport{
		errors: make(map[string]string), suggestions: make(map[string]int),
		suggestionOptions: make(map[string][]int),
	}
}

func prepareBulkResultItem(result *domainFacility.BulkOperationResultItem, id uuid.UUID) {
	result.ID = id
	result.Fields = make(map[string]string)
}

func completeBulkUpdateSuccess(result *domainFacility.BulkOperationResult, item *domainFacility.BulkOperationResultItem, proposed *domainFacility.FieldDevice) {
	item.Success = true
	item.Version = proposed.Version
	item.FieldDevice = proposed
	result.SuccessCount++
}

func completeBulkUpdateFailure(failure bulkUpdateFailure) {
	failure.item.Fields = failure.report.errors
	failure.item.Suggestions = failure.report.suggestions
	failure.item.SuggestionOptions = failure.report.suggestionOptions
	for _, message := range failure.report.errors {
		failure.item.Error = message
		break
	}
	if failure.item.Error == "" {
		failure.item.Error = failure.cause.Error()
	}
	failure.result.FailureCount++
}
