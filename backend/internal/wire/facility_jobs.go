package wire

import (
	"context"
	"encoding/json"
	"fmt"

	facilityjobs "github.com/besart951/go_infra_link/backend/internal/application/facilityjobs"
	apptransaction "github.com/besart951/go_infra_link/backend/internal/application/transaction"
	facilityservice "github.com/besart951/go_infra_link/backend/internal/service/facility"
	"github.com/google/uuid"
)

type facilityCopyOperation struct {
	task       string
	resource   string
	entityType string
	copy       func(context.Context, *facilityservice.Services, uuid.UUID) (uuid.UUID, error)
}

type facilityCopyTaskRegistrar struct {
	jobs     *facilityservice.CopyJobManager
	services *facilityservice.Services
	runtime  *RuntimeAdapters
}

func registerFacilityCopyTasks(jobs *facilityservice.CopyJobManager, services *Services, runtime *RuntimeAdapters) {
	if jobs == nil || services == nil || services.Facility == nil {
		return
	}
	registrar := facilityCopyTaskRegistrar{jobs: jobs, services: services.Facility, runtime: runtime}
	for _, operation := range facilityCopyOperations() {
		jobs.RegisterTask(operation.task, registrar.taskHandler(operation))
	}
	registerFieldDeviceBulkTasks(jobs, runtime)
	registerFacilityDeleteTasks(jobs, runtime)
}

func facilityCopyOperations() []facilityCopyOperation {
	return []facilityCopyOperation{
		{facilityservice.FacilityJobTaskCopyControlCabinet, "control_cabinets", "control_cabinet", copyControlCabinet},
		{facilityservice.FacilityJobTaskCopySPSController, "sps_controllers", "sps_controller", copySPSController},
		{facilityservice.FacilityJobTaskCopySPSControllerSystemType, "sps_controller_system_types", "sps_controller_system_type", copySPSControllerSystemType},
		{facilityservice.FacilityJobTaskCopyFieldDevice, "field_devices", "field_device", copyFieldDevice},
		{facilityservice.FacilityJobTaskCopyObjectData, "object_data", "object_data", copyObjectData},
	}
}

func (r facilityCopyTaskRegistrar) taskHandler(operation facilityCopyOperation) facilityservice.FacilityJobTask {
	return func(ctx context.Context, job facilityservice.CopyJob, report func(facilityservice.FacilityJobProgress)) (facilityservice.FacilityJobTaskResult, error) {
		payload, err := decodeFacilityCopyPayload(job.Payload)
		if err != nil {
			return facilityservice.FacilityJobTaskResult{}, err
		}
		report(facilityservice.FacilityJobProgress{Progress: 10, Stage: "copying_root"})
		result, err := r.executeStep(ctx, facilityCopyExecution{job: job, payload: payload, operation: operation})
		if err != nil {
			return facilityservice.FacilityJobTaskResult{}, err
		}
		report(facilityservice.FacilityJobProgress{Progress: 95, Stage: "finalizing", Processed: 1, Succeeded: 1})
		r.publish(ctx, facilityCopyNotification{
			ownerID: job.OwnerID, resource: operation.resource, resultID: result.TargetID,
		})
		return facilityservice.FacilityJobTaskResult{Result: result.Result}, nil
	}
}

type facilityCopyExecution struct {
	job       facilityservice.CopyJob
	payload   facilityservice.FacilityCopyTaskPayload
	operation facilityCopyOperation
}

type facilityCopyNotification struct {
	ownerID  uuid.UUID
	resource string
	resultID uuid.UUID
}

func (r facilityCopyTaskRegistrar) executeStep(ctx context.Context, execution facilityCopyExecution) (facilityjobs.StepResult, error) {
	if r.runtime == nil || r.runtime.FacilityJobSteps == nil {
		return runFacilityCopy(ctx, r.services, execution)
	}
	step := facilityjobs.Step{
		Key:        facilityjobs.ItemKey{OwnerID: execution.job.OwnerID, JobID: execution.job.ID},
		EntityType: execution.operation.entityType, SourceID: execution.payload.SourceID, Input: execution.job.Payload,
	}
	result, _, err := r.runtime.FacilityJobSteps.Execute(ctx, step, func(stepCtx context.Context, unit apptransaction.UnitOfWork) (facilityjobs.StepResult, error) {
		repos, buildErr := repositoriesFromUnit(unit)
		if buildErr != nil {
			return facilityjobs.StepResult{}, buildErr
		}
		services := facilityservice.NewServices(buildFacilityRepositories(repos))
		return runFacilityCopy(stepCtx, services, execution)
	})
	return result, err
}

func runFacilityCopy(ctx context.Context, services *facilityservice.Services, execution facilityCopyExecution) (facilityjobs.StepResult, error) {
	resultID, err := execution.operation.copy(ctx, services, execution.payload.SourceID)
	if err != nil {
		return facilityjobs.StepResult{}, err
	}
	result, err := json.Marshal(map[string]string{"resource_id": resultID.String()})
	return facilityjobs.StepResult{TargetID: resultID, Result: result}, err
}

func decodeFacilityCopyPayload(data json.RawMessage) (facilityservice.FacilityCopyTaskPayload, error) {
	var payload facilityservice.FacilityCopyTaskPayload
	if err := json.Unmarshal(data, &payload); err != nil {
		return payload, fmt.Errorf("decode copy task payload: %w", err)
	}
	if payload.SourceID == uuid.Nil {
		return payload, fmt.Errorf("decode copy task payload: source_id is required")
	}
	return payload, nil
}

func (r facilityCopyTaskRegistrar) publish(ctx context.Context, notification facilityCopyNotification) {
	if r.runtime == nil || r.runtime.FacilityReferenceData == nil {
		return
	}
	r.runtime.FacilityReferenceData.BroadcastFacilityChange(
		ctx, notification.resource, "copied", []uuid.UUID{notification.resultID}, &notification.ownerID,
	)
}

func copyControlCabinet(ctx context.Context, services *facilityservice.Services, id uuid.UUID) (uuid.UUID, error) {
	item, err := services.ControlCabinet.CopyByID(ctx, id)
	if err != nil {
		return uuid.Nil, err
	}
	return item.ID, nil
}

func copySPSController(ctx context.Context, services *facilityservice.Services, id uuid.UUID) (uuid.UUID, error) {
	item, err := services.SPSController.CopyByID(ctx, id)
	if err != nil {
		return uuid.Nil, err
	}
	return item.ID, nil
}

func copySPSControllerSystemType(ctx context.Context, services *facilityservice.Services, id uuid.UUID) (uuid.UUID, error) {
	item, err := services.SPSControllerSystemType.CopyByID(ctx, id)
	if err != nil {
		return uuid.Nil, err
	}
	return item.ID, nil
}

func copyFieldDevice(ctx context.Context, services *facilityservice.Services, id uuid.UUID) (uuid.UUID, error) {
	item, err := services.FieldDevice.CopyByID(ctx, id)
	if err != nil {
		return uuid.Nil, err
	}
	return item.ID, nil
}

func copyObjectData(ctx context.Context, services *facilityservice.Services, id uuid.UUID) (uuid.UUID, error) {
	item, err := services.ObjectData.CopyByID(ctx, id)
	if err != nil {
		return uuid.Nil, err
	}
	return item.ID, nil
}
