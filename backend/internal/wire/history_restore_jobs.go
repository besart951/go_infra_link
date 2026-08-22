package wire

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	facilityjobs "github.com/besart951/go_infra_link/backend/internal/application/facilityjobs"
	hierarchyrestore "github.com/besart951/go_infra_link/backend/internal/application/hierarchyrestore"
	apptransaction "github.com/besart951/go_infra_link/backend/internal/application/transaction"
	domainHistory "github.com/besart951/go_infra_link/backend/internal/domain/history"
	infratransaction "github.com/besart951/go_infra_link/backend/internal/infrastructure/transaction"
	"github.com/besart951/go_infra_link/backend/internal/repository/historysql"
	facilityservice "github.com/besart951/go_infra_link/backend/internal/service/facility"
	"github.com/google/uuid"
)

type historyRestoreExecution struct {
	job     facilityservice.CopyJob
	payload domainHistory.RestoreControlCabinetJobPayload
	report  func(facilityservice.FacilityJobProgress)
	asOf    time.Time
}

type historyRestoreChunk struct {
	execution historyRestoreExecution
	position  hierarchyrestore.Position
	phase     hierarchyrestore.Phase
	table     string
}

type historyRestoreTaskRegistrar struct {
	steps   facilityjobs.StepStore
	history interface {
		GetEvent(context.Context, uuid.UUID) (*domainHistory.ChangeEvent, error)
	}
}

func registerHistoryRestoreTasks(jobs *facilityservice.CopyJobManager, runtime *RuntimeAdapters, history HistoryRepository) {
	if jobs == nil || history == nil {
		return
	}
	registrar := historyRestoreTaskRegistrar{history: history}
	if runtime != nil {
		registrar.steps = runtime.FacilityJobSteps
	}
	jobs.RegisterTask(domainHistory.TaskRestoreControlCabinet, registrar.handler())
}

func (r historyRestoreTaskRegistrar) handler() facilityservice.FacilityJobTask {
	return func(ctx context.Context, job facilityservice.CopyJob, report func(facilityservice.FacilityJobProgress)) (facilityservice.FacilityJobTaskResult, error) {
		payload, err := decodeHistoryRestorePayload(job.Payload)
		if err != nil {
			return facilityservice.FacilityJobTaskResult{}, err
		}
		asOf, err := r.resolveRestoreTime(ctx, payload)
		if err != nil {
			return facilityservice.FacilityJobTaskResult{}, err
		}
		result, err := r.execute(ctx, historyRestoreExecution{job: job, payload: payload, report: report, asOf: asOf})
		return facilityservice.FacilityJobTaskResult{Result: result}, err
	}
}

func (r historyRestoreTaskRegistrar) execute(ctx context.Context, execution historyRestoreExecution) (json.RawMessage, error) {
	if r.steps == nil {
		return nil, fmt.Errorf("durable facility job step store is unavailable")
	}
	position, err := decodeHistoryRestorePosition(execution.job.Checkpoint)
	if err != nil {
		return nil, err
	}
	phases := hierarchyrestore.Phases()
	for position.PhaseIndex < len(phases) {
		phase := phases[position.PhaseIndex]
		tables := hierarchyrestore.Tables(phase)
		if position.TableIndex >= len(tables) {
			position.PhaseIndex++
			position.TableIndex, position.AfterID = 0, uuid.Nil
			continue
		}
		result, err := r.executeChunk(ctx, historyRestoreChunk{
			execution: execution, position: position, phase: phase, table: tables[position.TableIndex],
		})
		if err != nil {
			return nil, err
		}
		advanceRestorePosition(&position, result)
		if err := reportHistoryRestoreCheckpoint(execution, position, phases); err != nil {
			return nil, err
		}
	}
	if err := r.finalize(ctx, execution, position.Ordinal); err != nil {
		return nil, err
	}
	return encodeHistoryRestoreResult(execution, position)
}

func (r historyRestoreTaskRegistrar) executeChunk(ctx context.Context, chunk historyRestoreChunk) (hierarchyrestore.Result, error) {
	step := facilityjobs.Step{
		Key:        facilityjobs.ItemKey{OwnerID: chunk.execution.job.OwnerID, JobID: chunk.execution.job.ID, Ordinal: chunk.position.Ordinal},
		EntityType: "control_cabinet.restore." + string(chunk.phase) + "." + chunk.table,
		SourceID:   chunk.payloadID(), Input: chunk.execution.job.Payload,
	}
	result, _, err := r.steps.Execute(ctx, step, func(stepCtx context.Context, unit apptransaction.UnitOfWork) (facilityjobs.StepResult, error) {
		db, dbErr := infratransaction.GormDB(unit)
		if dbErr != nil {
			return facilityjobs.StepResult{}, dbErr
		}
		restored, restoreErr := historysql.NewStore(db).RestoreChunk(stepCtx, chunk.command())
		if restoreErr != nil {
			return facilityjobs.StepResult{}, restoreErr
		}
		encoded, encodeErr := json.Marshal(restored)
		return facilityjobs.StepResult{TargetID: chunk.payloadID(), Result: encoded}, encodeErr
	})
	if err != nil {
		return hierarchyrestore.Result{}, err
	}
	var restored hierarchyrestore.Result
	err = json.Unmarshal(result.Result, &restored)
	return restored, err
}

func (c historyRestoreChunk) command() hierarchyrestore.Command {
	return hierarchyrestore.Command{
		ControlCabinetID: c.payloadID(), ProjectID: c.execution.payload.ProjectID,
		AsOf: c.execution.asOf, Phase: c.phase, Table: c.table, AfterID: c.position.AfterID,
		Limit: 500, ActorID: c.execution.job.OwnerID, BatchID: c.execution.job.ID,
	}
}

func (c historyRestoreChunk) payloadID() uuid.UUID {
	return c.execution.payload.ControlCabinetID
}

func (r historyRestoreTaskRegistrar) finalize(ctx context.Context, execution historyRestoreExecution, ordinal int64) error {
	step := facilityjobs.Step{
		Key:        facilityjobs.ItemKey{OwnerID: execution.job.OwnerID, JobID: execution.job.ID, Ordinal: ordinal},
		EntityType: "control_cabinet.restore.finalize", SourceID: execution.payload.ControlCabinetID,
		Input: execution.job.Payload,
	}
	_, _, err := r.steps.Execute(ctx, step, func(stepCtx context.Context, unit apptransaction.UnitOfWork) (facilityjobs.StepResult, error) {
		db, dbErr := infratransaction.GormDB(unit)
		if dbErr != nil {
			return facilityjobs.StepResult{}, dbErr
		}
		if releaseErr := historysql.ReleaseRestoreLifecycle(stepCtx, db, execution.payload.ControlCabinetID); releaseErr != nil {
			return facilityjobs.StepResult{}, releaseErr
		}
		return facilityjobs.StepResult{TargetID: execution.payload.ControlCabinetID, Result: json.RawMessage(`{"finalized":true}`)}, nil
	})
	return err
}

func (r historyRestoreTaskRegistrar) resolveRestoreTime(ctx context.Context, payload domainHistory.RestoreControlCabinetJobPayload) (time.Time, error) {
	if payload.EventID == nil || *payload.EventID == uuid.Nil {
		if payload.AsOf != nil {
			return payload.AsOf.UTC(), nil
		}
		return time.Now().UTC(), nil
	}
	event, err := r.history.GetEvent(ctx, *payload.EventID)
	if err != nil {
		return time.Time{}, err
	}
	return event.OccurredAt.UTC(), nil
}

func advanceRestorePosition(position *hierarchyrestore.Position, result hierarchyrestore.Result) {
	position.Processed += int64(result.Processed)
	position.Restored += int64(result.Restored)
	position.Deleted += int64(result.Deleted)
	position.Skipped += int64(result.Skipped)
	position.Ordinal++
	if result.Done {
		position.TableIndex++
		position.AfterID = uuid.Nil
		return
	}
	position.AfterID = result.NextID
}

func reportHistoryRestoreCheckpoint(execution historyRestoreExecution, position hierarchyrestore.Position, phases []hierarchyrestore.Phase) error {
	encoded, err := json.Marshal(position)
	if err != nil {
		return err
	}
	totalTables := 0
	completedTables := position.TableIndex
	for index, phase := range phases {
		tables := hierarchyrestore.Tables(phase)
		totalTables += len(tables)
		if index < position.PhaseIndex {
			completedTables += len(tables)
		}
	}
	progress := 5 + completedTables*90/max(1, totalTables)
	execution.report(facilityservice.FacilityJobProgress{
		Progress: min(progress, 95), Stage: "restoring", Processed: position.Processed,
		Succeeded: position.Restored + position.Deleted, Failed: 0, Checkpoint: encoded,
	})
	return nil
}

func decodeHistoryRestorePayload(data json.RawMessage) (domainHistory.RestoreControlCabinetJobPayload, error) {
	var payload domainHistory.RestoreControlCabinetJobPayload
	if err := json.Unmarshal(data, &payload); err != nil {
		return payload, fmt.Errorf("decode restore payload: %w", err)
	}
	if payload.ControlCabinetID == uuid.Nil {
		return payload, fmt.Errorf("decode restore payload: control_cabinet_id is required")
	}
	return payload, nil
}

func decodeHistoryRestorePosition(data json.RawMessage) (hierarchyrestore.Position, error) {
	if len(data) == 0 {
		return hierarchyrestore.Position{}, nil
	}
	var position hierarchyrestore.Position
	err := json.Unmarshal(data, &position)
	return position, err
}

func encodeHistoryRestoreResult(execution historyRestoreExecution, position hierarchyrestore.Position) (json.RawMessage, error) {
	return json.Marshal(map[string]any{
		"resource_id": execution.payload.ControlCabinetID,
		"batch_id":    execution.job.ID, "restored_count": position.Restored,
		"deleted_count": position.Deleted, "skipped_count": position.Skipped,
	})
}
