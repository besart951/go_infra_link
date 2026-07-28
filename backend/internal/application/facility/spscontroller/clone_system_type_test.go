package spscontroller

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"testing"
	"time"

	appcollaboration "github.com/besart951/go_infra_link/backend/internal/application/collaboration"
	apptransaction "github.com/besart951/go_infra_link/backend/internal/application/transaction"
	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainCollaboration "github.com/besart951/go_infra_link/backend/internal/domain/collaboration"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	domainHistory "github.com/besart951/go_infra_link/backend/internal/domain/history"
	"github.com/google/uuid"
)

type globalSystemTypeCloneBatchKey struct{}

type globalSystemTypeCloneState struct {
	systemTypes map[uuid.UUID]*domainFacility.SPSControllerSystemType
	history     []string
}

func (s globalSystemTypeCloneState) clone() globalSystemTypeCloneState {
	systemTypes := make(map[uuid.UUID]*domainFacility.SPSControllerSystemType, len(s.systemTypes))
	for id, systemType := range s.systemTypes {
		systemTypes[id] = cloneSPSControllerSystemType(systemType)
	}
	return globalSystemTypeCloneState{
		systemTypes: systemTypes,
		history:     append([]string(nil), s.history...),
	}
}

type globalSystemTypeCloneUnit struct {
	state *globalSystemTypeCloneState
}

type globalSystemTypeCloneHarness struct {
	committed       globalSystemTypeCloneState
	copyEntity      *domainFacility.SPSControllerSystemType
	authoritative   *domainFacility.SPSControllerSystemType
	copyErr         error
	reloadErr       error
	commitErr       error
	projectIDs      []uuid.UUID
	projectIDsErr   error
	outbox          domainCollaboration.OutboxStore
	runnerCalls     int
	copyCalls       int
	reloadCalls     int
	projectIDCalls  int
	sourceID        uuid.UUID
	historyBatchIDs []uuid.UUID
}

func (h *globalSystemTypeCloneHarness) runner(
	ctx context.Context,
	run func(context.Context, apptransaction.UnitOfWork) error,
) error {
	h.runnerCalls++
	staged := h.committed.clone()
	runCtx := ctx
	if h.outbox != nil {
		runCtx = domainCollaboration.WithOutboxStore(ctx, h.outbox)
	}
	if err := run(runCtx, globalSystemTypeCloneUnit{state: &staged}); err != nil {
		return err
	}
	if h.commitErr != nil {
		return h.commitErr
	}
	h.committed = staged
	return nil
}

func (h *globalSystemTypeCloneHarness) factory(
	unit apptransaction.UnitOfWork,
) (CloneSystemTypeWorkflow, error) {
	typed, ok := unit.(globalSystemTypeCloneUnit)
	if !ok || typed.state == nil {
		return nil, errors.New("unexpected global SPS system-type clone transaction unit")
	}
	return &globalSystemTypeCloneWorkflowStub{harness: h, state: typed.state}, nil
}

type globalSystemTypeCloneWorkflowStub struct {
	harness *globalSystemTypeCloneHarness
	state   *globalSystemTypeCloneState
}

func (s *globalSystemTypeCloneWorkflowStub) CopyByID(
	ctx context.Context,
	sourceID uuid.UUID,
) (*domainFacility.SPSControllerSystemType, error) {
	s.harness.copyCalls++
	s.harness.sourceID = sourceID
	if _, ok := s.state.systemTypes[sourceID]; !ok {
		return nil, domain.ErrNotFound
	}
	copyEntity := cloneSPSControllerSystemType(s.harness.copyEntity)
	if copyEntity != nil && copyEntity.ID != uuid.Nil {
		s.state.systemTypes[copyEntity.ID] = cloneSPSControllerSystemType(copyEntity)
		s.state.history = append(s.state.history,
			"sps_controller_system_type:create",
			"field_device:create",
			"specification:create",
			"bacnet_object:create",
		)
		if batchID, ok := ctx.Value(globalSystemTypeCloneBatchKey{}).(uuid.UUID); ok {
			for range 4 {
				s.harness.historyBatchIDs = append(s.harness.historyBatchIDs, batchID)
			}
		}
	}
	return copyEntity, s.harness.copyErr
}

func (s *globalSystemTypeCloneWorkflowStub) GetByID(
	_ context.Context,
	id uuid.UUID,
) (*domainFacility.SPSControllerSystemType, error) {
	s.harness.reloadCalls++
	if s.harness.reloadErr != nil {
		return nil, s.harness.reloadErr
	}
	if s.harness.authoritative != nil {
		authoritative := cloneSPSControllerSystemType(s.harness.authoritative)
		s.state.systemTypes[id] = cloneSPSControllerSystemType(authoritative)
		return authoritative, nil
	}
	entity := s.state.systemTypes[id]
	if entity == nil {
		return nil, domain.ErrNotFound
	}
	return cloneSPSControllerSystemType(entity), nil
}

func (s *globalSystemTypeCloneWorkflowStub) GetOwningProjectIDs(
	_ context.Context,
	_ uuid.UUID,
) ([]uuid.UUID, error) {
	s.harness.projectIDCalls++
	return append([]uuid.UUID(nil), s.harness.projectIDs...), s.harness.projectIDsErr
}

func TestCloneSystemTypeCorrelatesDeepHistoryAndReturnsAuthoritativeRoot(t *testing.T) {
	sourceID := spsTestUUID(401)
	copyID := spsTestUUID(402)
	spsControllerID := spsTestUUID(403)
	systemTypeDefinitionID := spsTestUUID(404)
	actorID := spsTestUUID(405)
	operationID := spsTestUUID(406)
	createdAt := time.Date(2026, time.July, 20, 19, 0, 0, 0, time.UTC)
	occurredAt := createdAt.Add(time.Second)
	number := 2
	copyDocumentName := "copy.pdf"
	authoritativeDocumentName := "authoritative.pdf"
	harness := &globalSystemTypeCloneHarness{
		committed: globalSystemTypeCloneState{
			systemTypes: map[uuid.UUID]*domainFacility.SPSControllerSystemType{
				sourceID: {
					Base:            domain.Base{ID: sourceID},
					SPSControllerID: spsControllerID,
					SystemTypeID:    systemTypeDefinitionID,
				},
			},
		},
		copyEntity: &domainFacility.SPSControllerSystemType{
			Base:            domain.Base{ID: copyID},
			Number:          &number,
			DocumentName:    &copyDocumentName,
			SPSControllerID: spsControllerID,
			SystemTypeID:    systemTypeDefinitionID,
		},
		authoritative: &domainFacility.SPSControllerSystemType{
			Base:              domain.Base{ID: copyID, CreatedAt: createdAt, UpdatedAt: createdAt},
			Number:            &number,
			DocumentName:      &authoritativeDocumentName,
			SPSControllerID:   spsControllerID,
			SystemTypeID:      systemTypeDefinitionID,
			FieldDevicesCount: 3,
		},
	}
	handler := NewCloneSystemTypeHandler(CloneSystemTypeDependencies{
		TransactionRunner:   harness.runner,
		TransactionWorkflow: harness.factory,
		HistoryBatch: func(ctx context.Context, batchID uuid.UUID) context.Context {
			return context.WithValue(ctx, globalSystemTypeCloneBatchKey{}, batchID)
		},
		Actor: func(context.Context) *uuid.UUID { return &actorID },
		NewID: func() uuid.UUID { return operationID },
		Now:   func() time.Time { return occurredAt },
	})

	outcome, err := handler.Execute(context.Background(), CloneSystemTypeCommand{
		SourceSPSControllerSystemTypeID: sourceID,
	})
	if err != nil {
		t.Fatalf("execute global system-type clone: %v", err)
	}
	if harness.runnerCalls != 1 || harness.copyCalls != 1 || harness.reloadCalls != 1 ||
		harness.sourceID != sourceID {
		t.Fatalf("workflow calls: runner=%d copy=%d reload=%d source=%s",
			harness.runnerCalls,
			harness.copyCalls,
			harness.reloadCalls,
			harness.sourceID,
		)
	}
	if harness.committed.systemTypes[sourceID] == nil ||
		harness.committed.systemTypes[sourceID].ID != sourceID ||
		harness.committed.systemTypes[copyID] == nil ||
		harness.committed.systemTypes[copyID].DocumentName == nil ||
		*harness.committed.systemTypes[copyID].DocumentName != authoritativeDocumentName {
		t.Fatalf("committed system types: %+v", harness.committed.systemTypes)
	}
	if !reflect.DeepEqual(harness.committed.history, []string{
		"sps_controller_system_type:create",
		"field_device:create",
		"specification:create",
		"bacnet_object:create",
	}) || len(harness.historyBatchIDs) != 4 {
		t.Fatalf("history=%v batches=%v", harness.committed.history, harness.historyBatchIDs)
	}
	for _, batchID := range harness.historyBatchIDs {
		if batchID != operationID {
			t.Fatalf("history batch: got %s, want %s", batchID, operationID)
		}
	}
	if outcome.SPSControllerSystemType == nil || outcome.SPSControllerSystemType.ID != copyID ||
		outcome.SPSControllerSystemType.DocumentName == nil ||
		*outcome.SPSControllerSystemType.DocumentName != authoritativeDocumentName ||
		outcome.SPSControllerSystemType.FieldDevicesCount != 3 ||
		outcome.Mutation.OperationID != operationID ||
		outcome.Mutation.BatchID == nil || *outcome.Mutation.BatchID != operationID ||
		outcome.Mutation.ActorID == nil || *outcome.Mutation.ActorID != actorID ||
		!outcome.Mutation.OccurredAt.Equal(occurredAt) || len(outcome.Mutation.ProjectIDs) != 0 {
		t.Fatalf("global system-type clone outcome: %+v", outcome)
	}
	if len(outcome.Mutation.Changes) != 1 {
		t.Fatalf("changes: got %d, want 1", len(outcome.Mutation.Changes))
	}
	change := outcome.Mutation.Changes[0]
	if change.EntityType != "sps_controller_system_type" || change.EntityID != copyID ||
		change.ParentID == nil || *change.ParentID != spsControllerID ||
		change.Action != domainHistory.ActionCreate || len(change.Before) != 0 {
		t.Fatalf("global system-type clone change: %+v", change)
	}
	var snapshot spsControllerSystemTypeSnapshot
	if err := json.Unmarshal(change.After, &snapshot); err != nil {
		t.Fatalf("decode system-type clone snapshot: %v", err)
	}
	if snapshot.DocumentName == nil || *snapshot.DocumentName != authoritativeDocumentName ||
		snapshot.Number == nil || *snapshot.Number != number {
		t.Fatalf("system-type clone snapshot: %+v", snapshot)
	}
}

func TestCloneSystemTypeRollsBackCopyReloadAndCommitFailures(t *testing.T) {
	copyErr := errors.New("copy failed")
	reloadErr := errors.New("reload failed")
	commitErr := errors.New("commit failed")
	for _, test := range []struct {
		name      string
		copyErr   error
		reloadErr error
		commitErr error
		want      error
	}{
		{name: "copy", copyErr: copyErr, want: copyErr},
		{name: "reload", reloadErr: reloadErr, want: reloadErr},
		{name: "commit", commitErr: commitErr, want: commitErr},
	} {
		t.Run(test.name, func(t *testing.T) {
			sourceID := spsTestUUID(411)
			copyID := spsTestUUID(412)
			spsControllerID := spsTestUUID(413)
			systemTypeDefinitionID := spsTestUUID(414)
			harness := &globalSystemTypeCloneHarness{
				committed: globalSystemTypeCloneState{
					systemTypes: map[uuid.UUID]*domainFacility.SPSControllerSystemType{
						sourceID: {
							Base:            domain.Base{ID: sourceID},
							SPSControllerID: spsControllerID,
							SystemTypeID:    systemTypeDefinitionID,
						},
					},
				},
				copyEntity: &domainFacility.SPSControllerSystemType{
					Base:            domain.Base{ID: copyID},
					SPSControllerID: spsControllerID,
					SystemTypeID:    systemTypeDefinitionID,
				},
				copyErr:   test.copyErr,
				reloadErr: test.reloadErr,
				commitErr: test.commitErr,
			}
			handler := NewCloneSystemTypeHandler(CloneSystemTypeDependencies{
				TransactionRunner:   harness.runner,
				TransactionWorkflow: harness.factory,
				HistoryBatch: func(ctx context.Context, batchID uuid.UUID) context.Context {
					return context.WithValue(ctx, globalSystemTypeCloneBatchKey{}, batchID)
				},
			})

			_, err := handler.Execute(context.Background(), CloneSystemTypeCommand{
				SourceSPSControllerSystemTypeID: sourceID,
			})
			if !errors.Is(err, test.want) {
				t.Fatalf("error: got %v, want %v", err, test.want)
			}
			if harness.committed.systemTypes[copyID] != nil || len(harness.committed.history) != 0 {
				t.Fatalf("failed clone leaked state: %+v", harness.committed)
			}
		})
	}
}

func TestCloneSystemTypeValidatesConfigurationAndSourceID(t *testing.T) {
	unconfigured := NewCloneSystemTypeHandler(CloneSystemTypeDependencies{})
	_, err := unconfigured.Execute(context.Background(), CloneSystemTypeCommand{
		SourceSPSControllerSystemTypeID: uuid.New(),
	})
	if !errors.Is(err, ErrCloneSystemTypeTransactionNotConfigured) {
		t.Fatalf("unconfigured error: %v", err)
	}

	harness := &globalSystemTypeCloneHarness{}
	configured := NewCloneSystemTypeHandler(CloneSystemTypeDependencies{
		TransactionRunner:   harness.runner,
		TransactionWorkflow: harness.factory,
	})
	_, err = configured.Execute(context.Background(), CloneSystemTypeCommand{})
	if !errors.Is(err, domain.ErrInvalidArgument) {
		t.Fatalf("invalid source error: %v", err)
	}
	if harness.runnerCalls != 0 {
		t.Fatalf("invalid source started %d transactions", harness.runnerCalls)
	}
}

func TestCloneSystemTypeUsesOwningSPSProjectsForDurableRecipients(t *testing.T) {
	sourceID := spsTestUUID(421)
	copyID := spsTestUUID(422)
	spsControllerID := spsTestUUID(423)
	systemTypeDefinitionID := spsTestUUID(424)
	projectOne := spsTestUUID(425)
	projectTwo := spsTestUUID(426)
	operationID := spsTestUUID(427)
	eventOne := spsTestUUID(428)
	eventTwo := spsTestUUID(429)
	occurredAt := time.Date(2026, time.July, 23, 20, 0, 0, 0, time.UTC)
	outbox := &updateOutboxStoreStub{}
	harness := &globalSystemTypeCloneHarness{
		committed: globalSystemTypeCloneState{
			systemTypes: map[uuid.UUID]*domainFacility.SPSControllerSystemType{
				sourceID: {
					Base:            domain.Base{ID: sourceID},
					SPSControllerID: spsControllerID,
					SystemTypeID:    systemTypeDefinitionID,
				},
			},
		},
		copyEntity: &domainFacility.SPSControllerSystemType{
			Base:            domain.Base{ID: copyID},
			SPSControllerID: spsControllerID,
			SystemTypeID:    systemTypeDefinitionID,
		},
		projectIDs: []uuid.UUID{projectTwo, projectOne, projectOne, uuid.Nil},
		outbox:     outbox,
	}
	dispatcher := &updateCommandDispatcherStub{}
	generatedIDs := []uuid.UUID{operationID, eventOne, eventTwo}
	handler := NewCloneSystemTypeHandler(CloneSystemTypeDependencies{
		TransactionRunner:   harness.runner,
		TransactionWorkflow: harness.factory,
		Dispatcher:          dispatcher,
		NewID: func() uuid.UUID {
			id := generatedIDs[0]
			generatedIDs = generatedIDs[1:]
			return id
		},
		Now: func() time.Time { return occurredAt },
	})

	outcome, err := handler.Execute(context.Background(), CloneSystemTypeCommand{
		SourceSPSControllerSystemTypeID: sourceID,
	})
	if err != nil {
		t.Fatalf("execute recipient-aware system-type clone: %v", err)
	}
	if harness.projectIDCalls != 1 {
		t.Fatalf("owning project reads: got %d, want 1", harness.projectIDCalls)
	}
	wantProjects := []uuid.UUID{projectOne, projectTwo}
	if !reflect.DeepEqual(outcome.Mutation.ProjectIDs, wantProjects) {
		t.Fatalf("recipient projects: got %v, want %v", outcome.Mutation.ProjectIDs, wantProjects)
	}
	if len(outbox.events) != 2 || len(dispatcher.commands) != 2 {
		t.Fatalf("durable/live commands: outbox=%d dispatch=%d", len(outbox.events), len(dispatcher.commands))
	}
	for index, raw := range dispatcher.commands {
		command, ok := raw.(appcollaboration.SPSControllerSystemTypeCloned)
		if !ok {
			t.Fatalf("command %d type: %T", index, raw)
		}
		if command.ProjectID != wantProjects[index] ||
			command.SPSControllerID != spsControllerID ||
			command.SPSControllerSystemTypeID != copyID ||
			command.SourceSPSControllerSystemTypeID != sourceID ||
			command.SchemaVersion != appcollaboration.SchemaVersionV2 {
			t.Fatalf("command %d: %+v", index, command)
		}
	}
}
