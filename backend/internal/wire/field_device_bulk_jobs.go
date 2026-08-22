package wire

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	facilityjobs "github.com/besart951/go_infra_link/backend/internal/application/facilityjobs"
	apptransaction "github.com/besart951/go_infra_link/backend/internal/application/transaction"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	facilityservice "github.com/besart951/go_infra_link/backend/internal/service/facility"
	"github.com/google/uuid"
)

const facilityBulkChunkSize = 500

type fieldDeviceBulkTaskRegistrar struct {
	steps    facilityjobs.StepStore
	plans    facilityjobs.FieldDeviceUpdatePlanStore
	services *Services
}

type fieldDeviceBulkRun[T any] struct {
	job        facilityservice.FacilityJob
	report     func(facilityservice.FacilityJobProgress)
	items      []T
	entityType string
	sourceID   func(T) uuid.UUID
	mutate     func(context.Context, *facilityservice.Services, T) (facilityjobs.StepResult, error)
}

type fieldDeviceBulkItemExecution[T any] struct {
	ctx       context.Context
	registrar fieldDeviceBulkTaskRegistrar
	run       fieldDeviceBulkRun[T]
	index     int
	item      T
}

type fieldDeviceUpdateGroupRun struct {
	ctx    context.Context
	job    facilityservice.FacilityJob
	report func(facilityservice.FacilityJobProgress)
	total  int
	groups []facilityservice.FieldDeviceBulkUpdateGroup
}

type fieldDeviceUpdateGroupExecution struct {
	ctx     context.Context
	job     facilityservice.FacilityJob
	ordinal int64
	group   facilityservice.FieldDeviceBulkUpdateGroup
}

func registerFieldDeviceBulkTasks(jobs *facilityservice.FacilityJobManager, runtime *RuntimeAdapters, services *Services) {
	registrar := fieldDeviceBulkTaskRegistrar{services: services}
	if runtime != nil {
		registrar.steps = runtime.FacilityJobSteps
		registrar.plans = runtime.FieldDeviceUpdatePlans
	}
	jobs.RegisterTask(facilityservice.FacilityJobTaskMultiCreateFieldDevices, facilityservice.FacilityJobHandlerFunc(registrar.multiCreate))
	jobs.RegisterTask(facilityservice.FacilityJobTaskBulkUpdateFieldDevices, facilityservice.FacilityJobHandlerFunc(registrar.bulkUpdate))
	jobs.RegisterTask(facilityservice.FacilityJobTaskBulkDeleteFieldDevices, facilityservice.FacilityJobHandlerFunc(registrar.bulkDelete))
}

func (r fieldDeviceBulkTaskRegistrar) multiCreate(ctx context.Context, execution facilityservice.FacilityJobExecution) (facilityservice.FacilityJobTaskResult, error) {
	job, report := execution.Job, execution.Reporter.Report
	var payload facilityservice.FieldDeviceMultiCreateTaskPayload
	if err := decodeBulkPayload(job.Payload, &payload); err != nil {
		return facilityservice.FacilityJobTaskResult{}, err
	}
	return runFieldDeviceBulk(ctx, r, fieldDeviceBulkRun[domainFacility.FieldDeviceCreateItem]{
		job: job, report: report, items: payload.Items, entityType: "field_device_create",
		sourceID: createItemID, mutate: createFieldDeviceItem,
	})
}

func (r fieldDeviceBulkTaskRegistrar) bulkUpdate(ctx context.Context, execution facilityservice.FacilityJobExecution) (facilityservice.FacilityJobTaskResult, error) {
	job, report := execution.Job, execution.Reporter.Report
	var payload facilityservice.FieldDeviceBulkUpdateTaskPayload
	if err := decodeBulkPayload(job.Payload, &payload); err != nil {
		return facilityservice.FacilityJobTaskResult{}, err
	}
	if r.steps == nil {
		return facilityservice.FacilityJobTaskResult{}, errors.New("durable facility job step store is unavailable")
	}
	groups, err := r.prepareUpdateGroups(ctx, job, payload.Updates)
	if err != nil {
		return facilityservice.FacilityJobTaskResult{}, err
	}
	return r.runUpdateGroups(fieldDeviceUpdateGroupRun{
		ctx: ctx, job: job, report: report, total: len(payload.Updates), groups: groups,
	})
}

func (r fieldDeviceBulkTaskRegistrar) prepareUpdateGroups(ctx context.Context, job facilityservice.FacilityJob, updates []domainFacility.BulkFieldDeviceUpdate) ([]facilityservice.FieldDeviceBulkUpdateGroup, error) {
	items, err := r.steps.ListItems(ctx, job.OwnerID, job.ID)
	if err != nil {
		return nil, err
	}
	if len(items) > 0 {
		return decodePreparedUpdateGroups(items)
	}
	groups, err := r.loadUpdatePlan(ctx, job)
	if err != nil {
		return nil, err
	}
	if len(groups) > 0 {
		return groups, r.prepareUpdateGroupSteps(ctx, job, groups)
	}
	if err := r.saveUnplannedUpdates(ctx, job, updates); err != nil {
		return nil, err
	}
	if err := r.plans.Plan(ctx, job.OwnerID, job.ID); err != nil {
		return nil, err
	}
	groups, err = r.loadUpdatePlan(ctx, job)
	if err != nil {
		return nil, err
	}
	return groups, r.prepareUpdateGroupSteps(ctx, job, groups)
}

func (r fieldDeviceBulkTaskRegistrar) prepareUpdateGroupSteps(ctx context.Context, job facilityservice.FacilityJob, groups []facilityservice.FieldDeviceBulkUpdateGroup) error {
	steps, err := updateGroupSteps(job, groups)
	if err != nil {
		return err
	}
	return r.steps.Prepare(ctx, steps)
}

func (r fieldDeviceBulkTaskRegistrar) loadUpdatePlan(ctx context.Context, job facilityservice.FacilityJob) ([]facilityservice.FieldDeviceBulkUpdateGroup, error) {
	if r.plans == nil {
		return nil, errors.New("field device bulk plan store is unavailable")
	}
	items, err := r.plans.List(ctx, job.OwnerID, job.ID)
	if err != nil || len(items) == 0 {
		return nil, err
	}
	return updateGroupsFromPlan(items)
}

func (r fieldDeviceBulkTaskRegistrar) saveUnplannedUpdates(ctx context.Context, job facilityservice.FacilityJob, updates []domainFacility.BulkFieldDeviceUpdate) error {
	if r.plans == nil {
		return errors.New("field device bulk plan store is unavailable")
	}
	for start := 0; start < len(updates); start += facilityBulkChunkSize {
		end := min(start+facilityBulkChunkSize, len(updates))
		items, err := unplannedUpdateItems(job, updates[start:end], start)
		if err != nil {
			return err
		}
		if err := r.plans.Save(ctx, items); err != nil {
			return err
		}
	}
	return nil
}

func unplannedUpdateItems(job facilityservice.FacilityJob, updates []domainFacility.BulkFieldDeviceUpdate, offset int) ([]facilityjobs.FieldDeviceUpdatePlanItem, error) {
	items := make([]facilityjobs.FieldDeviceUpdatePlanItem, len(updates))
	for index, update := range updates {
		command, err := json.Marshal(update)
		if err != nil {
			return nil, err
		}
		ordinal := offset + index
		items[index] = facilityjobs.FieldDeviceUpdatePlanItem{
			OwnerID: job.OwnerID, JobID: job.ID, Ordinal: int64(ordinal), GroupOrdinal: int64(ordinal),
			DependencyGroupID: uuid.NewMD5(uuid.NameSpaceOID, []byte(update.ID.String())),
			FieldDeviceID:     update.ID, Command: command,
		}
	}
	return items, nil
}

func updateGroupsFromPlan(items []facilityjobs.FieldDeviceUpdatePlanItem) ([]facilityservice.FieldDeviceBulkUpdateGroup, error) {
	groups := make([]facilityservice.FieldDeviceBulkUpdateGroup, 0)
	for _, item := range items {
		if len(groups) == 0 || groups[len(groups)-1].ID != item.DependencyGroupID {
			groups = append(groups, facilityservice.FieldDeviceBulkUpdateGroup{ID: item.DependencyGroupID})
		}
		var update domainFacility.BulkFieldDeviceUpdate
		if err := json.Unmarshal(item.Command, &update); err != nil {
			return nil, fmt.Errorf("decode field device update plan item: %w", err)
		}
		group := &groups[len(groups)-1]
		group.Indexes = append(group.Indexes, int(item.Ordinal))
		group.Updates = append(group.Updates, update)
	}
	return groups, nil
}

func decodePreparedUpdateGroups(items []facilityjobs.Item) ([]facilityservice.FieldDeviceBulkUpdateGroup, error) {
	groups := make([]facilityservice.FieldDeviceBulkUpdateGroup, len(items))
	for index, item := range items {
		if err := json.Unmarshal(item.Input, &groups[index]); err != nil {
			return nil, fmt.Errorf("decode persisted field device update group: %w", err)
		}
	}
	return groups, nil
}

func updateGroupSteps(job facilityservice.FacilityJob, groups []facilityservice.FieldDeviceBulkUpdateGroup) ([]facilityjobs.Step, error) {
	steps := make([]facilityjobs.Step, len(groups))
	for index, group := range groups {
		input, err := json.Marshal(group)
		if err != nil {
			return nil, err
		}
		steps[index] = updateGroupStep(fieldDeviceUpdateGroupExecution{
			job: job, ordinal: int64(index), group: group,
		}, input)
	}
	return steps, nil
}

func updateGroupStep(execution fieldDeviceUpdateGroupExecution, input json.RawMessage) facilityjobs.Step {
	return facilityjobs.Step{
		Key: facilityjobs.ItemKey{
			OwnerID: execution.job.OwnerID, JobID: execution.job.ID, Ordinal: execution.ordinal,
		},
		EntityType: "field_device_update_group", SourceID: execution.group.ID, Input: input,
	}
}

func (r fieldDeviceBulkTaskRegistrar) runUpdateGroups(run fieldDeviceUpdateGroupRun) (facilityservice.FacilityJobTaskResult, error) {
	result := facilityservice.FieldDeviceBulkJobResult{TotalCount: run.total}
	processed := 0
	for ordinal, group := range run.groups {
		err := r.executeUpdateGroup(fieldDeviceUpdateGroupExecution{
			ctx: run.ctx, job: run.job, ordinal: int64(ordinal), group: group,
		})
		processed += len(group.Updates)
		applyUpdateGroupResult(&result, group, err)
		reportFieldDeviceBulkProgress(run.report, processed, result)
	}
	encoded, err := json.Marshal(result)
	return facilityservice.FacilityJobTaskResult{Result: encoded}, err
}

func (r fieldDeviceBulkTaskRegistrar) executeUpdateGroup(execution fieldDeviceUpdateGroupExecution) error {
	input, err := json.Marshal(execution.group)
	if err != nil {
		return err
	}
	step := updateGroupStep(execution, input)
	_, _, err = r.steps.Execute(execution.ctx, step, func(itemCtx context.Context, unit apptransaction.UnitOfWork) (facilityjobs.StepResult, error) {
		services, buildErr := facilityServicesFromUnit(unit)
		if buildErr != nil {
			return facilityjobs.StepResult{}, buildErr
		}
		return updateFieldDeviceGroup(itemCtx, services, execution.group)
	})
	return err
}

func updateFieldDeviceGroup(ctx context.Context, services *facilityservice.Services, group facilityservice.FieldDeviceBulkUpdateGroup) (facilityjobs.StepResult, error) {
	result := services.FieldDevice.BulkUpdate(ctx, group.Updates)
	if result == nil || result.FailureCount > 0 || result.SuccessCount != len(group.Updates) {
		return facilityjobs.StepResult{}, bulkOperationError(result)
	}
	encoded, err := json.Marshal(result)
	return facilityjobs.StepResult{Result: encoded}, err
}

func applyUpdateGroupResult(result *facilityservice.FieldDeviceBulkJobResult, group facilityservice.FieldDeviceBulkUpdateGroup, cause error) {
	if cause == nil {
		result.SuccessCount += len(group.Updates)
		return
	}
	result.FailureCount += len(group.Updates)
	for index, update := range group.Updates {
		result.Failures = append(result.Failures, facilityservice.FieldDeviceBulkJobFailure{
			Ordinal: group.Indexes[index], SourceID: update.ID, DependencyGroupID: group.ID, Error: cause.Error(),
		})
	}
}

func (r fieldDeviceBulkTaskRegistrar) bulkDelete(ctx context.Context, execution facilityservice.FacilityJobExecution) (facilityservice.FacilityJobTaskResult, error) {
	job, report := execution.Job, execution.Reporter.Report
	var payload facilityservice.FieldDeviceBulkDeleteTaskPayload
	if err := decodeBulkPayload(job.Payload, &payload); err != nil {
		return facilityservice.FacilityJobTaskResult{}, err
	}
	commands := deleteTaskCommands(payload)
	return runFieldDeviceBulk(ctx, r, fieldDeviceBulkRun[domainFacility.FieldDeviceDeleteCommand]{
		job: job, report: report, items: commands, entityType: "field_device_delete",
		sourceID: func(command domainFacility.FieldDeviceDeleteCommand) uuid.UUID { return command.ID }, mutate: deleteFieldDeviceItem,
	})
}

func deleteTaskCommands(payload facilityservice.FieldDeviceBulkDeleteTaskPayload) []domainFacility.FieldDeviceDeleteCommand {
	return payload.Commands
}

func runFieldDeviceBulk[T any](ctx context.Context, registrar fieldDeviceBulkTaskRegistrar, run fieldDeviceBulkRun[T]) (facilityservice.FacilityJobTaskResult, error) {
	if registrar.steps == nil {
		return facilityservice.FacilityJobTaskResult{}, errors.New("durable facility job step store is unavailable")
	}
	result := facilityservice.FieldDeviceBulkJobResult{TotalCount: len(run.items)}
	for index, item := range run.items {
		err := executeFieldDeviceBulkItem(fieldDeviceBulkItemExecution[T]{
			ctx: ctx, registrar: registrar, run: run, index: index, item: item,
		})
		if err != nil {
			result.FailureCount++
			result.Failures = append(result.Failures, facilityservice.FieldDeviceBulkJobFailure{
				Ordinal: index, SourceID: run.sourceID(item), Error: err.Error(),
			})
		} else {
			result.SuccessCount++
		}
		reportFieldDeviceBulkProgress(run.report, index+1, result)
	}
	encoded, err := json.Marshal(result)
	return facilityservice.FacilityJobTaskResult{Result: encoded}, err
}

func executeFieldDeviceBulkItem[T any](execution fieldDeviceBulkItemExecution[T]) error {
	input, err := json.Marshal(execution.item)
	if err != nil {
		return err
	}
	step := facilityjobs.Step{
		Key: facilityjobs.ItemKey{
			OwnerID: execution.run.job.OwnerID, JobID: execution.run.job.ID, Ordinal: int64(execution.index),
		},
		EntityType: execution.run.entityType, SourceID: execution.run.sourceID(execution.item), Input: input,
	}
	_, _, err = execution.registrar.steps.Execute(execution.ctx, step, func(itemCtx context.Context, unit apptransaction.UnitOfWork) (facilityjobs.StepResult, error) {
		services, buildErr := facilityServicesFromUnit(unit)
		if buildErr != nil {
			return facilityjobs.StepResult{}, buildErr
		}
		return execution.run.mutate(itemCtx, services, execution.item)
	})
	return err
}

func facilityServicesFromUnit(unit apptransaction.UnitOfWork) (*facilityservice.Services, error) {
	repos, err := repositoriesFromUnit(unit)
	if err != nil {
		return nil, err
	}
	return facilityservice.NewServices(buildFacilityRepositories(repos)), nil
}

func reportFieldDeviceBulkProgress(report func(facilityservice.FacilityJobProgress), processed int, result facilityservice.FieldDeviceBulkJobResult) {
	if processed%facilityBulkChunkSize != 0 && processed != result.TotalCount {
		return
	}
	total := int64(result.TotalCount)
	progress := 5
	if total > 0 {
		progress += int(90 * int64(processed) / total)
	}
	report(facilityservice.FacilityJobProgress{
		Progress: progress, Stage: "processing_items", Processed: int64(processed), Total: &total,
		Succeeded: int64(result.SuccessCount), Failed: int64(result.FailureCount),
	})
}

func createItemID(item domainFacility.FieldDeviceCreateItem) uuid.UUID {
	if item.FieldDevice == nil {
		return uuid.Nil
	}
	return item.FieldDevice.ID
}

func createFieldDeviceItem(ctx context.Context, services *facilityservice.Services, item domainFacility.FieldDeviceCreateItem) (facilityjobs.StepResult, error) {
	result := services.FieldDevice.MultiCreate(ctx, []domainFacility.FieldDeviceCreateItem{item})
	if result == nil || len(result.Results) != 1 || !result.Results[0].Success {
		return facilityjobs.StepResult{}, bulkCreateError(result)
	}
	created := result.Results[0].FieldDevice
	if created == nil {
		return facilityjobs.StepResult{}, errors.New("field device create returned no entity")
	}
	encoded, err := json.Marshal(result.Results[0])
	return facilityjobs.StepResult{TargetID: created.ID, Result: encoded}, err
}

func updateFieldDeviceItem(ctx context.Context, services *facilityservice.Services, item domainFacility.BulkFieldDeviceUpdate) (facilityjobs.StepResult, error) {
	result := services.FieldDevice.BulkUpdate(ctx, []domainFacility.BulkFieldDeviceUpdate{item})
	if result == nil || len(result.Results) != 1 || !result.Results[0].Success {
		return facilityjobs.StepResult{}, bulkOperationError(result)
	}
	encoded, err := json.Marshal(result.Results[0])
	return facilityjobs.StepResult{TargetID: item.ID, Result: encoded}, err
}

func deleteFieldDeviceItem(ctx context.Context, services *facilityservice.Services, command domainFacility.FieldDeviceDeleteCommand) (facilityjobs.StepResult, error) {
	result := services.FieldDevice.BulkDeleteCommands(ctx, []domainFacility.FieldDeviceDeleteCommand{command})
	if result == nil || len(result.Results) != 1 || !result.Results[0].Success {
		return facilityjobs.StepResult{}, bulkOperationError(result)
	}
	encoded, err := json.Marshal(result.Results[0])
	return facilityjobs.StepResult{TargetID: command.ID, Result: encoded}, err
}

func bulkCreateError(result *domainFacility.FieldDeviceMultiCreateResult) error {
	if result != nil && len(result.Results) == 1 && result.Results[0].Error != "" {
		return errors.New(result.Results[0].Error)
	}
	return errors.New("field device create failed")
}

func bulkOperationError(result *domainFacility.BulkOperationResult) error {
	if result != nil && len(result.Results) == 1 && result.Results[0].Error != "" {
		return errors.New(result.Results[0].Error)
	}
	return errors.New("field device bulk item failed")
}

func decodeBulkPayload(data json.RawMessage, target any) error {
	if err := json.Unmarshal(data, target); err != nil {
		return fmt.Errorf("decode field device bulk task: %w", err)
	}
	return nil
}
