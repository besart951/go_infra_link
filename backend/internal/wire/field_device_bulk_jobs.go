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
	steps facilityjobs.StepStore
}

type fieldDeviceBulkRun[T any] struct {
	job        facilityservice.CopyJob
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

func registerFieldDeviceBulkTasks(jobs *facilityservice.CopyJobManager, runtime *RuntimeAdapters) {
	registrar := fieldDeviceBulkTaskRegistrar{}
	if runtime != nil {
		registrar.steps = runtime.FacilityJobSteps
	}
	jobs.RegisterTask(facilityservice.FacilityJobTaskMultiCreateFieldDevices, registrar.multiCreate)
	jobs.RegisterTask(facilityservice.FacilityJobTaskBulkUpdateFieldDevices, registrar.bulkUpdate)
	jobs.RegisterTask(facilityservice.FacilityJobTaskBulkDeleteFieldDevices, registrar.bulkDelete)
}

func (r fieldDeviceBulkTaskRegistrar) multiCreate(ctx context.Context, job facilityservice.CopyJob, report func(facilityservice.FacilityJobProgress)) (facilityservice.FacilityJobTaskResult, error) {
	var payload facilityservice.FieldDeviceMultiCreateTaskPayload
	if err := decodeBulkPayload(job.Payload, &payload); err != nil {
		return facilityservice.FacilityJobTaskResult{}, err
	}
	return runFieldDeviceBulk(ctx, r, fieldDeviceBulkRun[domainFacility.FieldDeviceCreateItem]{
		job: job, report: report, items: payload.Items, entityType: "field_device_create",
		sourceID: createItemID, mutate: createFieldDeviceItem,
	})
}

func (r fieldDeviceBulkTaskRegistrar) bulkUpdate(ctx context.Context, job facilityservice.CopyJob, report func(facilityservice.FacilityJobProgress)) (facilityservice.FacilityJobTaskResult, error) {
	var payload facilityservice.FieldDeviceBulkUpdateTaskPayload
	if err := decodeBulkPayload(job.Payload, &payload); err != nil {
		return facilityservice.FacilityJobTaskResult{}, err
	}
	return runFieldDeviceBulk(ctx, r, fieldDeviceBulkRun[domainFacility.BulkFieldDeviceUpdate]{
		job: job, report: report, items: payload.Updates, entityType: "field_device_update",
		sourceID: func(item domainFacility.BulkFieldDeviceUpdate) uuid.UUID { return item.ID }, mutate: updateFieldDeviceItem,
	})
}

func (r fieldDeviceBulkTaskRegistrar) bulkDelete(ctx context.Context, job facilityservice.CopyJob, report func(facilityservice.FacilityJobProgress)) (facilityservice.FacilityJobTaskResult, error) {
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
	if len(payload.Commands) > 0 {
		return payload.Commands
	}
	commands := make([]domainFacility.FieldDeviceDeleteCommand, len(payload.IDs))
	for index, id := range payload.IDs {
		commands[index] = domainFacility.FieldDeviceDeleteCommand{ID: id}
	}
	return commands
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
