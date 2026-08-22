package wire

import (
	"context"
	"encoding/json"
	"fmt"

	facilityjobs "github.com/besart951/go_infra_link/backend/internal/application/facilityjobs"
	apptransaction "github.com/besart951/go_infra_link/backend/internal/application/transaction"
	domainproject "github.com/besart951/go_infra_link/backend/internal/domain/project"
	infrarealtime "github.com/besart951/go_infra_link/backend/internal/infrastructure/realtime"
	facilityservice "github.com/besart951/go_infra_link/backend/internal/service/facility"
	projectservice "github.com/besart951/go_infra_link/backend/internal/service/project"
	"github.com/google/uuid"
)

type projectCopyOperation struct {
	task       string
	eventType  string
	entityType string
	copy       func(context.Context, *projectservice.Services, projectCopyCommand) (uuid.UUID, error)
}

type projectCopyCommand struct {
	ProjectID uuid.UUID
	SourceID  uuid.UUID
}

type projectCopyCompletion struct {
	ownerID   uuid.UUID
	projectID uuid.UUID
	eventType string
	resultID  uuid.UUID
}

type projectCopyCheckpoint struct {
	ResultID              uuid.UUID `json:"result_id"`
	ProjectChangeRecorded bool      `json:"project_change_recorded"`
}

type projectCopyExecution struct {
	job       facilityservice.CopyJob
	operation projectCopyOperation
	report    func(facilityservice.FacilityJobProgress)
}

type projectCopyState struct {
	execution  projectCopyExecution
	payload    domainproject.FacilityCopyJobPayload
	checkpoint projectCopyCheckpoint
}

type projectCopyTaskRegistrar struct {
	jobs          *facilityservice.CopyJobManager
	project       *projectservice.Services
	collaboration *infrarealtime.ProjectCollaborationHub
	steps         facilityjobs.StepStore
}

func registerProjectCopyTasks(jobs *facilityservice.CopyJobManager, services *Services, runtime *RuntimeAdapters) {
	if jobs == nil || services == nil || services.Project == nil {
		return
	}
	registrar := projectCopyTaskRegistrar{jobs: jobs, project: services.Project}
	if runtime != nil {
		registrar.collaboration = runtime.ProjectCollaboration
		registrar.steps = runtime.FacilityJobSteps
	}
	registrar.registerOperations()
}

func (r projectCopyTaskRegistrar) registerOperations() {
	operations := []projectCopyOperation{
		{domainproject.TaskCopyProjectControlCabinet, "project.control_cabinet.copied", "control_cabinet", copyProjectControlCabinet},
		{domainproject.TaskCopyProjectSPSController, "project.sps_controller.copied", "sps_controller", copyProjectSPSController},
		{domainproject.TaskCopyProjectSPSControllerSystemType, "project.sps_controller_system_type.copied", "sps_controller_system_type", copyProjectSPSControllerSystemType},
	}
	for _, operation := range operations {
		r.jobs.RegisterTask(operation.task, r.taskHandler(operation))
	}
}

func (r projectCopyTaskRegistrar) taskHandler(operation projectCopyOperation) facilityservice.FacilityJobTask {
	return func(ctx context.Context, job facilityservice.CopyJob, report func(facilityservice.FacilityJobProgress)) (facilityservice.FacilityJobTaskResult, error) {
		return r.executeProjectCopy(ctx, projectCopyExecution{job: job, operation: operation, report: report})
	}
}

func (r projectCopyTaskRegistrar) executeProjectCopy(ctx context.Context, execution projectCopyExecution) (facilityservice.FacilityJobTaskResult, error) {
	payload, err := decodeProjectCopyPayload(execution.job.Payload)
	if err != nil {
		return facilityservice.FacilityJobTaskResult{}, err
	}
	if r.steps != nil {
		return r.executePersistedProjectStep(ctx, execution, payload)
	}
	checkpoint, err := r.resumeOrCopy(ctx, execution, payload)
	if err != nil {
		return facilityservice.FacilityJobTaskResult{}, err
	}
	checkpoint, err = r.ensureProjectChange(ctx, projectCopyState{
		execution: execution, payload: payload, checkpoint: checkpoint,
	})
	if err != nil {
		return facilityservice.FacilityJobTaskResult{}, err
	}
	execution.report(facilityservice.FacilityJobProgress{Progress: 95, Stage: "finalizing", Processed: 1, Succeeded: 1})
	result, err := json.Marshal(map[string]string{"resource_id": checkpoint.ResultID.String()})
	return facilityservice.FacilityJobTaskResult{Result: result}, err
}

func (r projectCopyTaskRegistrar) executePersistedProjectStep(
	ctx context.Context,
	execution projectCopyExecution,
	payload domainproject.FacilityCopyJobPayload,
) (facilityservice.FacilityJobTaskResult, error) {
	if checkpoint, resumed, err := projectCopyCheckpointFromData(execution.job.Checkpoint); err != nil || resumed {
		return completedProjectCopyResult(checkpoint.ResultID, err)
	}
	step := projectCopyStep(execution, payload)
	var changes []domainproject.Change
	result, _, err := r.steps.Execute(ctx, step, func(stepCtx context.Context, unit apptransaction.UnitOfWork) (facilityjobs.StepResult, error) {
		services, buildErr := projectServicesFromUnit(unit)
		if buildErr != nil {
			return facilityjobs.StepResult{}, buildErr
		}
		return executeProjectMutation(stepCtx, projectMutationRequest{
			services: services, execution: execution, payload: payload, changes: &changes,
		})
	})
	if err != nil {
		return facilityservice.FacilityJobTaskResult{}, err
	}
	for _, change := range changes {
		r.publishProjectChange(ctx, change)
	}
	execution.report(facilityservice.FacilityJobProgress{Progress: 95, Stage: "finalizing", Processed: 1, Succeeded: 1})
	return facilityservice.FacilityJobTaskResult{Result: result.Result}, nil
}

type projectMutationRequest struct {
	services  *projectservice.Services
	execution projectCopyExecution
	payload   domainproject.FacilityCopyJobPayload
	changes   *[]domainproject.Change
}

func executeProjectMutation(ctx context.Context, request projectMutationRequest) (facilityjobs.StepResult, error) {
	command := projectCopyCommand{ProjectID: request.payload.ProjectID, SourceID: request.payload.SourceID}
	resultID, err := request.execution.operation.copy(ctx, request.services, command)
	if err != nil {
		return facilityjobs.StepResult{}, err
	}
	changes, err := recordProjectCopyChange(ctx, request.services, projectCopyCompletion{
		ownerID: request.execution.job.OwnerID, projectID: request.payload.ProjectID,
		eventType: request.execution.operation.eventType, resultID: resultID,
	})
	if err != nil {
		return facilityjobs.StepResult{}, err
	}
	*request.changes = changes
	result, err := json.Marshal(map[string]string{"resource_id": resultID.String()})
	return facilityjobs.StepResult{TargetID: resultID, Result: result}, err
}

func projectCopyStep(execution projectCopyExecution, payload domainproject.FacilityCopyJobPayload) facilityjobs.Step {
	return facilityjobs.Step{
		Key:        facilityjobs.ItemKey{OwnerID: execution.job.OwnerID, JobID: execution.job.ID},
		EntityType: execution.operation.entityType, SourceID: payload.SourceID, Input: execution.job.Payload,
	}
}

func projectServicesFromUnit(unit apptransaction.UnitOfWork) (*projectservice.Services, error) {
	repos, err := repositoriesFromUnit(unit)
	if err != nil {
		return nil, err
	}
	facilityServices := facilityservice.NewServices(buildFacilityRepositories(repos))
	return projectservice.NewServices(buildProjectDependencies(repos, facilityServices)), nil
}

func completedProjectCopyResult(resultID uuid.UUID, cause error) (facilityservice.FacilityJobTaskResult, error) {
	if cause != nil {
		return facilityservice.FacilityJobTaskResult{}, cause
	}
	result, err := json.Marshal(map[string]string{"resource_id": resultID.String()})
	return facilityservice.FacilityJobTaskResult{Result: result}, err
}

func (r projectCopyTaskRegistrar) resumeOrCopy(ctx context.Context, execution projectCopyExecution, payload domainproject.FacilityCopyJobPayload) (projectCopyCheckpoint, error) {
	checkpoint, resumed, err := projectCopyCheckpointFromData(execution.job.Checkpoint)
	if err != nil || resumed {
		return checkpoint, err
	}
	return r.copyAndCheckpoint(ctx, execution, payload)
}

func (r projectCopyTaskRegistrar) ensureProjectChange(ctx context.Context, state projectCopyState) (projectCopyCheckpoint, error) {
	if state.checkpoint.ProjectChangeRecorded {
		return state.checkpoint, nil
	}
	completion := projectCopyCompletion{
		ownerID: state.execution.job.OwnerID, projectID: state.payload.ProjectID,
		eventType: state.execution.operation.eventType, resultID: state.checkpoint.ResultID,
	}
	if err := r.recordProjectChange(ctx, completion); err != nil {
		return state.checkpoint, err
	}
	state.checkpoint.ProjectChangeRecorded = true
	progress, err := checkpointProgress(94, state.checkpoint)
	if err == nil {
		state.execution.report(progress)
	}
	return state.checkpoint, err
}

func (r projectCopyTaskRegistrar) copyAndCheckpoint(ctx context.Context, execution projectCopyExecution, payload domainproject.FacilityCopyJobPayload) (projectCopyCheckpoint, error) {
	execution.report(facilityservice.FacilityJobProgress{Progress: 10, Stage: "copying_root"})
	command := projectCopyCommand{ProjectID: payload.ProjectID, SourceID: payload.SourceID}
	resultID, err := execution.operation.copy(ctx, r.project, command)
	if err != nil {
		return projectCopyCheckpoint{}, err
	}
	checkpoint := projectCopyCheckpoint{ResultID: resultID}
	progress, err := checkpointProgress(90, checkpoint)
	if err != nil {
		return projectCopyCheckpoint{}, err
	}
	execution.report(progress)
	return checkpoint, nil
}

func projectCopyCheckpointFromData(data json.RawMessage) (projectCopyCheckpoint, bool, error) {
	if len(data) == 0 {
		return projectCopyCheckpoint{}, false, nil
	}
	var checkpoint projectCopyCheckpoint
	if err := json.Unmarshal(data, &checkpoint); err != nil {
		return projectCopyCheckpoint{}, false, fmt.Errorf("decode project copy checkpoint: %w", err)
	}
	if checkpoint.ResultID == uuid.Nil {
		return projectCopyCheckpoint{}, false, fmt.Errorf("decode project copy checkpoint: result_id is required")
	}
	return checkpoint, true, nil
}

func checkpointProgress(progress int, checkpoint projectCopyCheckpoint) (facilityservice.FacilityJobProgress, error) {
	encoded, err := json.Marshal(checkpoint)
	return facilityservice.FacilityJobProgress{Progress: progress, Stage: "finalizing", Checkpoint: encoded}, err
}

func decodeProjectCopyPayload(data json.RawMessage) (domainproject.FacilityCopyJobPayload, error) {
	var payload domainproject.FacilityCopyJobPayload
	if err := json.Unmarshal(data, &payload); err != nil {
		return payload, fmt.Errorf("decode project copy task payload: %w", err)
	}
	if payload.ProjectID == uuid.Nil || payload.SourceID == uuid.Nil {
		return payload, fmt.Errorf("decode project copy task payload: project_id and source_id are required")
	}
	return payload, nil
}

func (r projectCopyTaskRegistrar) recordProjectChange(ctx context.Context, completion projectCopyCompletion) error {
	changes, err := recordProjectCopyChange(ctx, r.project, completion)
	if err != nil {
		return err
	}
	for _, change := range changes {
		r.publishProjectChange(ctx, change)
	}
	return nil
}

func recordProjectCopyChange(ctx context.Context, services *projectservice.Services, completion projectCopyCompletion) ([]domainproject.Change, error) {
	if services.Changes == nil {
		return nil, nil
	}
	changes, err := services.Changes.RecordEvents(
		ctx, completion.projectID, completion.eventType, &completion.ownerID, completion.resultID.String(),
	)
	if err != nil {
		return nil, fmt.Errorf("record copied project resource: %w", err)
	}
	return changes, nil
}

func (r projectCopyTaskRegistrar) publishProjectChange(ctx context.Context, change domainproject.Change) {
	if r.collaboration == nil {
		return
	}
	realtimeChange, ok := infrarealtime.ProjectChangeFromDomain(change)
	if ok {
		_ = r.collaboration.BroadcastProjectChange(ctx, realtimeChange)
	}
}

func copyProjectControlCabinet(ctx context.Context, services *projectservice.Services, command projectCopyCommand) (uuid.UUID, error) {
	item, err := services.FacilityLink.CopyControlCabinet(ctx, command.ProjectID, command.SourceID)
	if err != nil {
		return uuid.Nil, err
	}
	return item.ID, nil
}

func copyProjectSPSController(ctx context.Context, services *projectservice.Services, command projectCopyCommand) (uuid.UUID, error) {
	item, err := services.FacilityLink.CopySPSController(ctx, command.ProjectID, command.SourceID)
	if err != nil {
		return uuid.Nil, err
	}
	return item.ID, nil
}

func copyProjectSPSControllerSystemType(ctx context.Context, services *projectservice.Services, command projectCopyCommand) (uuid.UUID, error) {
	item, err := services.FacilityLink.CopySPSControllerSystemType(ctx, command.ProjectID, command.SourceID)
	if err != nil {
		return uuid.Nil, err
	}
	return item.ID, nil
}
