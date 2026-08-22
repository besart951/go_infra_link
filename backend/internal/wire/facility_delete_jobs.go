package wire

import (
	"context"
	"encoding/json"
	"fmt"

	facilityjobs "github.com/besart951/go_infra_link/backend/internal/application/facilityjobs"
	hierarchydelete "github.com/besart951/go_infra_link/backend/internal/application/hierarchydelete"
	apptransaction "github.com/besart951/go_infra_link/backend/internal/application/transaction"
	infratransaction "github.com/besart951/go_infra_link/backend/internal/infrastructure/transaction"
	"github.com/besart951/go_infra_link/backend/internal/repository/facilitysql"
	facilityservice "github.com/besart951/go_infra_link/backend/internal/service/facility"
	"github.com/google/uuid"
)

type facilityDeleteOperation struct {
	task     string
	resource string
	rootKind hierarchydelete.RootKind
}

type facilityDeleteCheckpoint struct {
	StageIndex int   `json:"stage_index"`
	Ordinal    int64 `json:"ordinal"`
	Processed  int64 `json:"processed"`
}

type facilityDeleteExecution struct {
	job       facilityservice.CopyJob
	payload   facilityservice.FacilityDeleteTaskPayload
	operation facilityDeleteOperation
	report    func(facilityservice.FacilityJobProgress)
}

type facilityDeleteNotification struct {
	ownerID  uuid.UUID
	resource string
	sourceID uuid.UUID
}

type facilityDeleteChunk struct {
	execution facilityDeleteExecution
	stage     hierarchydelete.Stage
	ordinal   int64
}

type facilityDeleteTaskRegistrar struct {
	steps   facilityjobs.StepStore
	runtime *RuntimeAdapters
}

func registerFacilityDeleteTasks(jobs *facilityservice.CopyJobManager, runtime *RuntimeAdapters) {
	registrar := facilityDeleteTaskRegistrar{runtime: runtime}
	if runtime != nil {
		registrar.steps = runtime.FacilityJobSteps
	}
	for _, operation := range facilityDeleteOperations() {
		jobs.RegisterTask(operation.task, registrar.handler(operation))
	}
}

func facilityDeleteOperations() []facilityDeleteOperation {
	return []facilityDeleteOperation{
		{facilityservice.FacilityJobTaskDeleteControlCabinet, "control_cabinets", hierarchydelete.RootControlCabinet},
		{facilityservice.FacilityJobTaskDeleteSPSController, "sps_controllers", hierarchydelete.RootSPSController},
		{facilityservice.FacilityJobTaskDeleteSPSControllerSystemType, "sps_controller_system_types", hierarchydelete.RootSPSControllerSystemType},
	}
}

func (r facilityDeleteTaskRegistrar) handler(operation facilityDeleteOperation) facilityservice.FacilityJobTask {
	return func(ctx context.Context, job facilityservice.CopyJob, report func(facilityservice.FacilityJobProgress)) (facilityservice.FacilityJobTaskResult, error) {
		payload, err := decodeFacilityDeletePayload(job.Payload)
		if err != nil {
			return facilityservice.FacilityJobTaskResult{}, err
		}
		execution := facilityDeleteExecution{job: job, payload: payload, operation: operation, report: report}
		result, err := r.execute(ctx, execution)
		if err != nil {
			return facilityservice.FacilityJobTaskResult{}, err
		}
		r.publish(ctx, facilityDeleteNotification{ownerID: job.OwnerID, resource: operation.resource, sourceID: payload.SourceID})
		return facilityservice.FacilityJobTaskResult{Result: result}, nil
	}
}

func (r facilityDeleteTaskRegistrar) execute(ctx context.Context, execution facilityDeleteExecution) (json.RawMessage, error) {
	if r.steps == nil {
		return nil, fmt.Errorf("durable facility job step store is unavailable")
	}
	checkpoint, err := decodeFacilityDeleteCheckpoint(execution.job.Checkpoint)
	if err != nil {
		return nil, err
	}
	stages := hierarchydelete.Stages(execution.operation.rootKind)
	for checkpoint.StageIndex < len(stages) {
		result, err := r.executeChunk(ctx, facilityDeleteChunk{
			execution: execution, stage: stages[checkpoint.StageIndex], ordinal: checkpoint.Ordinal,
		})
		if err != nil {
			return nil, err
		}
		checkpoint.Processed += int64(result.Deleted)
		checkpoint.Ordinal++
		if result.Done {
			checkpoint.StageIndex++
		}
		if err := reportFacilityDeleteCheckpoint(execution, checkpoint, stages); err != nil {
			return nil, err
		}
	}
	return json.Marshal(map[string]string{"deleted_resource_id": execution.payload.SourceID.String()})
}

func (r facilityDeleteTaskRegistrar) executeChunk(ctx context.Context, chunk facilityDeleteChunk) (hierarchydelete.Result, error) {
	step := facilityDeleteStep(chunk.execution, chunk.stage, chunk.ordinal)
	result, _, err := r.steps.Execute(ctx, step, func(stepCtx context.Context, unit apptransaction.UnitOfWork) (facilityjobs.StepResult, error) {
		db, dbErr := infratransaction.GormDB(unit)
		if dbErr != nil {
			return facilityjobs.StepResult{}, dbErr
		}
		deleted, deleteErr := facilitysql.NewHierarchyDeleteStore(db).DeleteChunk(stepCtx, hierarchydelete.Command{
			RootKind: chunk.execution.operation.rootKind, RootID: chunk.execution.payload.SourceID,
			Stage: chunk.stage, Limit: 500, ActorID: chunk.execution.job.OwnerID, BatchID: chunk.execution.job.ID,
		})
		if deleteErr != nil {
			return facilityjobs.StepResult{}, deleteErr
		}
		encoded, encodeErr := json.Marshal(deleted)
		return facilityjobs.StepResult{TargetID: chunk.execution.payload.SourceID, Result: encoded}, encodeErr
	})
	if err != nil {
		return hierarchydelete.Result{}, err
	}
	var deleted hierarchydelete.Result
	err = json.Unmarshal(result.Result, &deleted)
	return deleted, err
}

func facilityDeleteStep(execution facilityDeleteExecution, stage hierarchydelete.Stage, ordinal int64) facilityjobs.Step {
	return facilityjobs.Step{
		Key:        facilityjobs.ItemKey{OwnerID: execution.job.OwnerID, JobID: execution.job.ID, Ordinal: ordinal},
		EntityType: string(execution.operation.rootKind) + ".delete." + string(stage),
		SourceID:   execution.payload.SourceID, Input: execution.job.Payload,
	}
}

func reportFacilityDeleteCheckpoint(execution facilityDeleteExecution, checkpoint facilityDeleteCheckpoint, stages []hierarchydelete.Stage) error {
	encoded, err := json.Marshal(checkpoint)
	if err != nil {
		return err
	}
	progress := 10 + checkpoint.StageIndex*80/max(1, len(stages))
	stage := "finalizing"
	if checkpoint.StageIndex < len(stages) {
		stage = "deleting_" + string(stages[checkpoint.StageIndex])
	}
	execution.report(facilityservice.FacilityJobProgress{
		Progress: min(progress, 95), Stage: stage, Processed: checkpoint.Processed,
		Succeeded: checkpoint.Processed, Checkpoint: encoded,
	})
	return nil
}

func decodeFacilityDeleteCheckpoint(data json.RawMessage) (facilityDeleteCheckpoint, error) {
	if len(data) == 0 {
		return facilityDeleteCheckpoint{}, nil
	}
	var checkpoint facilityDeleteCheckpoint
	if err := json.Unmarshal(data, &checkpoint); err != nil {
		return checkpoint, fmt.Errorf("decode delete checkpoint: %w", err)
	}
	if checkpoint.StageIndex < 0 || checkpoint.Ordinal < 0 || checkpoint.Processed < 0 {
		return checkpoint, fmt.Errorf("decode delete checkpoint: negative value")
	}
	return checkpoint, nil
}

func decodeFacilityDeletePayload(data json.RawMessage) (facilityservice.FacilityDeleteTaskPayload, error) {
	var payload facilityservice.FacilityDeleteTaskPayload
	if err := json.Unmarshal(data, &payload); err != nil {
		return payload, fmt.Errorf("decode delete task payload: %w", err)
	}
	if payload.SourceID == uuid.Nil {
		return payload, fmt.Errorf("decode delete task payload: source_id is required")
	}
	return payload, nil
}

func (r facilityDeleteTaskRegistrar) publish(ctx context.Context, notification facilityDeleteNotification) {
	if r.runtime == nil || r.runtime.FacilityReferenceData == nil {
		return
	}
	r.runtime.FacilityReferenceData.BroadcastFacilityChange(
		ctx, notification.resource, "deleted", []uuid.UUID{notification.sourceID}, &notification.ownerID,
	)
}
