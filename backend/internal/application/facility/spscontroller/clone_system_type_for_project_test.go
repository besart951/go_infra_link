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
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	domainHistory "github.com/besart951/go_infra_link/backend/internal/domain/history"
	"github.com/google/uuid"
)

type projectSystemTypeCloneBatchKey struct{}

type projectSystemTypeCloneState struct {
	systemTypes       map[uuid.UUID]*domainFacility.SPSControllerSystemType
	projectFieldLinks map[uuid.UUID][]uuid.UUID
	history           []string
}

func (s projectSystemTypeCloneState) clone() projectSystemTypeCloneState {
	systemTypes := make(map[uuid.UUID]*domainFacility.SPSControllerSystemType, len(s.systemTypes))
	for id, systemType := range s.systemTypes {
		systemTypes[id] = cloneSPSControllerSystemType(systemType)
	}
	projectFieldLinks := make(map[uuid.UUID][]uuid.UUID, len(s.projectFieldLinks))
	for projectID, fieldDeviceIDs := range s.projectFieldLinks {
		projectFieldLinks[projectID] = append([]uuid.UUID(nil), fieldDeviceIDs...)
	}
	return projectSystemTypeCloneState{
		systemTypes:       systemTypes,
		projectFieldLinks: projectFieldLinks,
		history:           append([]string(nil), s.history...),
	}
}

type projectSystemTypeCloneUnit struct {
	state *projectSystemTypeCloneState
}

type projectSystemTypeCloneHarness struct {
	committed       projectSystemTypeCloneState
	copyEntity      *domainFacility.SPSControllerSystemType
	copiedFieldIDs  []uuid.UUID
	sourceAccessErr error
	copyErr         error
	commitErr       error
	runnerCalls     int
	accessCalls     int
	copyCalls       int
	projectID       uuid.UUID
	sourceID        uuid.UUID
	historyBatchIDs []uuid.UUID
}

func (h *projectSystemTypeCloneHarness) runner(
	ctx context.Context,
	run func(context.Context, apptransaction.UnitOfWork) error,
) error {
	h.runnerCalls++
	staged := h.committed.clone()
	if err := run(ctx, projectSystemTypeCloneUnit{state: &staged}); err != nil {
		return err
	}
	if h.commitErr != nil {
		return h.commitErr
	}
	h.committed = staged
	return nil
}

func (h *projectSystemTypeCloneHarness) factory(
	unit apptransaction.UnitOfWork,
) (CloneSystemTypeForProjectWorkflow, error) {
	typed, ok := unit.(projectSystemTypeCloneUnit)
	if !ok || typed.state == nil {
		return nil, errors.New("unexpected project SPS system-type clone transaction unit")
	}
	return &projectSystemTypeCloneWorkflowStub{harness: h, state: typed.state}, nil
}

type projectSystemTypeCloneWorkflowStub struct {
	harness *projectSystemTypeCloneHarness
	state   *projectSystemTypeCloneState
}

func (s *projectSystemTypeCloneWorkflowStub) RequireSourceAccess(
	_ context.Context,
	projectID, sourceID uuid.UUID,
) error {
	s.harness.accessCalls++
	s.harness.projectID = projectID
	s.harness.sourceID = sourceID
	return s.harness.sourceAccessErr
}

func (s *projectSystemTypeCloneWorkflowStub) CopySPSControllerSystemType(
	ctx context.Context,
	projectID uuid.UUID,
	sourceID uuid.UUID,
) (*domainFacility.SPSControllerSystemType, error) {
	s.harness.copyCalls++
	s.harness.projectID = projectID
	s.harness.sourceID = sourceID
	if _, ok := s.state.systemTypes[sourceID]; !ok {
		return nil, domain.ErrNotFound
	}
	copyEntity := cloneSPSControllerSystemType(s.harness.copyEntity)
	if copyEntity != nil && copyEntity.ID != uuid.Nil {
		s.state.systemTypes[copyEntity.ID] = cloneSPSControllerSystemType(copyEntity)
		s.state.projectFieldLinks[projectID] = append(
			s.state.projectFieldLinks[projectID],
			s.harness.copiedFieldIDs...,
		)
		s.state.history = append(s.state.history,
			"sps_controller_system_type:create",
			"field_device:create",
			"specification:create",
			"bacnet_object:create",
			"project_field_device:create",
		)
		if batchID, ok := ctx.Value(projectSystemTypeCloneBatchKey{}).(uuid.UUID); ok {
			for range 5 {
				s.harness.historyBatchIDs = append(s.harness.historyBatchIDs, batchID)
			}
		}
	}
	return copyEntity, s.harness.copyErr
}

type projectSystemTypeCloneDispatcherStub struct {
	harness           *projectSystemTypeCloneHarness
	projectID         uuid.UUID
	copyID            uuid.UUID
	commands          []appcollaboration.Command
	err               error
	calledAfterCommit bool
}

func (s *projectSystemTypeCloneDispatcherStub) Dispatch(
	_ context.Context,
	command appcollaboration.Command,
) error {
	s.commands = append(s.commands, command)
	s.calledAfterCommit = s.harness.committed.systemTypes[s.copyID] != nil &&
		reflect.DeepEqual(
			s.harness.committed.projectFieldLinks[s.projectID],
			s.harness.copiedFieldIDs,
		)
	return s.err
}

func TestCloneSystemTypeForProjectCommitsHierarchyAndLinksBeforeTypedDispatch(t *testing.T) {
	projectID := spsTestUUID(301)
	unrelatedProjectID := spsTestUUID(302)
	sourceID := spsTestUUID(303)
	copyID := spsTestUUID(304)
	spsControllerID := spsTestUUID(305)
	systemTypeDefinitionID := spsTestUUID(306)
	fieldDeviceID := spsTestUUID(307)
	unrelatedFieldDeviceID := spsTestUUID(308)
	actorID := spsTestUUID(309)
	operationID := spsTestUUID(310)
	eventID := spsTestUUID(311)
	createdAt := time.Date(2026, time.July, 20, 18, 0, 0, 0, time.UTC)
	occurredAt := createdAt.Add(time.Second)
	number := 2
	documentName := "HLK-02"
	harness := &projectSystemTypeCloneHarness{
		committed: projectSystemTypeCloneState{
			systemTypes: map[uuid.UUID]*domainFacility.SPSControllerSystemType{
				sourceID: {
					Base:            domain.Base{ID: sourceID},
					SPSControllerID: spsControllerID,
					SystemTypeID:    systemTypeDefinitionID,
				},
			},
			projectFieldLinks: map[uuid.UUID][]uuid.UUID{
				unrelatedProjectID: {unrelatedFieldDeviceID},
			},
		},
		copyEntity: &domainFacility.SPSControllerSystemType{
			Base:            domain.Base{ID: copyID, CreatedAt: createdAt, UpdatedAt: createdAt},
			Number:          &number,
			DocumentName:    &documentName,
			SPSControllerID: spsControllerID,
			SystemTypeID:    systemTypeDefinitionID,
		},
		copiedFieldIDs: []uuid.UUID{fieldDeviceID},
	}
	dispatcher := &projectSystemTypeCloneDispatcherStub{
		harness:   harness,
		projectID: projectID,
		copyID:    copyID,
	}
	generatedIDs := []uuid.UUID{operationID, eventID}
	handler := NewCloneSystemTypeForProjectHandler(CloneSystemTypeForProjectDependencies{
		TransactionRunner:   harness.runner,
		TransactionWorkflow: harness.factory,
		HistoryBatch: func(ctx context.Context, batchID uuid.UUID) context.Context {
			return context.WithValue(ctx, projectSystemTypeCloneBatchKey{}, batchID)
		},
		Dispatcher: dispatcher,
		Actor:      func(context.Context) *uuid.UUID { return &actorID },
		NewID: func() uuid.UUID {
			id := generatedIDs[0]
			generatedIDs = generatedIDs[1:]
			return id
		},
		Now: func() time.Time { return occurredAt },
	})

	outcome, err := handler.Execute(context.Background(), CloneSystemTypeForProjectCommand{
		ProjectID:                       projectID,
		SourceSPSControllerSystemTypeID: sourceID,
	})
	if err != nil {
		t.Fatalf("execute project SPS system-type clone: %v", err)
	}
	if harness.runnerCalls != 1 || harness.copyCalls != 1 ||
		harness.projectID != projectID || harness.sourceID != sourceID {
		t.Fatalf("workflow calls: runner=%d copy=%d project=%s source=%s",
			harness.runnerCalls,
			harness.copyCalls,
			harness.projectID,
			harness.sourceID,
		)
	}
	if harness.committed.systemTypes[sourceID] == nil || harness.committed.systemTypes[copyID] == nil ||
		!reflect.DeepEqual(harness.committed.projectFieldLinks[projectID], []uuid.UUID{fieldDeviceID}) ||
		!reflect.DeepEqual(harness.committed.projectFieldLinks[unrelatedProjectID], []uuid.UUID{unrelatedFieldDeviceID}) {
		t.Fatalf("committed project system-type clone: %+v", harness.committed)
	}
	if len(harness.committed.history) != 5 || len(harness.historyBatchIDs) != 5 {
		t.Fatalf("history=%v batches=%v", harness.committed.history, harness.historyBatchIDs)
	}
	for _, batchID := range harness.historyBatchIDs {
		if batchID != operationID {
			t.Fatalf("history batch: got %s, want %s", batchID, operationID)
		}
	}
	if outcome.SPSControllerSystemType == nil || outcome.SPSControllerSystemType.ID != copyID ||
		outcome.Mutation.OperationID != operationID ||
		outcome.Mutation.BatchID == nil || *outcome.Mutation.BatchID != operationID ||
		outcome.Mutation.ActorID == nil || *outcome.Mutation.ActorID != actorID ||
		!reflect.DeepEqual(outcome.Mutation.ProjectIDs, []uuid.UUID{projectID}) {
		t.Fatalf("project system-type clone outcome: %+v", outcome)
	}
	if len(outcome.Mutation.Changes) != 1 {
		t.Fatalf("changes: got %d, want 1", len(outcome.Mutation.Changes))
	}
	change := outcome.Mutation.Changes[0]
	if change.EntityType != "sps_controller_system_type" || change.EntityID != copyID ||
		change.ParentID == nil || *change.ParentID != spsControllerID ||
		change.Action != domainHistory.ActionCreate || len(change.Before) != 0 {
		t.Fatalf("project system-type clone change: %+v", change)
	}
	var snapshot spsControllerSystemTypeSnapshot
	if err := json.Unmarshal(change.After, &snapshot); err != nil {
		t.Fatalf("decode project system-type clone snapshot: %v", err)
	}
	if snapshot.ID != copyID || snapshot.Number == nil || *snapshot.Number != number ||
		snapshot.DocumentName == nil || *snapshot.DocumentName != documentName ||
		snapshot.SystemTypeID != systemTypeDefinitionID {
		t.Fatalf("project system-type clone snapshot: %+v", snapshot)
	}
	if !dispatcher.calledAfterCommit || len(dispatcher.commands) != 1 {
		t.Fatalf("dispatch timing: after=%t commands=%v", dispatcher.calledAfterCommit, dispatcher.commands)
	}
	command, ok := dispatcher.commands[0].(appcollaboration.SPSControllerSystemTypeCloned)
	if !ok || command.ProjectID != projectID ||
		command.SourceSPSControllerSystemTypeID != sourceID ||
		command.SPSControllerSystemTypeID != copyID ||
		command.SPSControllerID != spsControllerID ||
		command.OperationID != operationID || command.EventID != eventID ||
		command.CorrelationID != operationID {
		t.Fatalf("project system-type clone command: %+v", dispatcher.commands[0])
	}
}

func TestCloneSystemTypeForProjectRollsBackBeforeDispatch(t *testing.T) {
	projectID := spsTestUUID(321)
	sourceID := spsTestUUID(322)
	copyID := spsTestUUID(323)
	spsControllerID := spsTestUUID(324)
	systemTypeDefinitionID := spsTestUUID(325)
	copyErr := errors.New("project field-device link failed")
	harness := &projectSystemTypeCloneHarness{
		committed: projectSystemTypeCloneState{
			systemTypes: map[uuid.UUID]*domainFacility.SPSControllerSystemType{
				sourceID: {
					Base:            domain.Base{ID: sourceID},
					SPSControllerID: spsControllerID,
					SystemTypeID:    systemTypeDefinitionID,
				},
			},
			projectFieldLinks: map[uuid.UUID][]uuid.UUID{},
		},
		copyEntity: &domainFacility.SPSControllerSystemType{
			Base:            domain.Base{ID: copyID},
			SPSControllerID: spsControllerID,
			SystemTypeID:    systemTypeDefinitionID,
		},
		copiedFieldIDs: []uuid.UUID{spsTestUUID(326)},
		copyErr:        copyErr,
	}
	dispatcher := &projectSystemTypeCloneDispatcherStub{harness: harness, projectID: projectID, copyID: copyID}
	handler := NewCloneSystemTypeForProjectHandler(CloneSystemTypeForProjectDependencies{
		TransactionRunner:   harness.runner,
		TransactionWorkflow: harness.factory,
		HistoryBatch: func(ctx context.Context, batchID uuid.UUID) context.Context {
			return context.WithValue(ctx, projectSystemTypeCloneBatchKey{}, batchID)
		},
		Dispatcher: dispatcher,
	})

	_, err := handler.Execute(context.Background(), CloneSystemTypeForProjectCommand{
		ProjectID:                       projectID,
		SourceSPSControllerSystemTypeID: sourceID,
	})
	if !errors.Is(err, copyErr) {
		t.Fatalf("expected copy error, got %v", err)
	}
	if harness.committed.systemTypes[copyID] != nil ||
		len(harness.committed.projectFieldLinks[projectID]) != 0 ||
		len(harness.committed.history) != 0 || len(dispatcher.commands) != 0 {
		t.Fatalf("failed clone leaked state: committed=%+v commands=%v", harness.committed, dispatcher.commands)
	}
}

func TestCloneSystemTypeForProjectDoesNotDispatchAfterCommitFailure(t *testing.T) {
	projectID := spsTestUUID(331)
	sourceID := spsTestUUID(332)
	copyID := spsTestUUID(333)
	spsControllerID := spsTestUUID(334)
	systemTypeDefinitionID := spsTestUUID(335)
	commitErr := errors.New("commit failed")
	harness := &projectSystemTypeCloneHarness{
		committed: projectSystemTypeCloneState{
			systemTypes: map[uuid.UUID]*domainFacility.SPSControllerSystemType{
				sourceID: {
					Base:            domain.Base{ID: sourceID},
					SPSControllerID: spsControllerID,
					SystemTypeID:    systemTypeDefinitionID,
				},
			},
			projectFieldLinks: map[uuid.UUID][]uuid.UUID{},
		},
		copyEntity: &domainFacility.SPSControllerSystemType{
			Base:            domain.Base{ID: copyID},
			SPSControllerID: spsControllerID,
			SystemTypeID:    systemTypeDefinitionID,
		},
		copiedFieldIDs: []uuid.UUID{spsTestUUID(336)},
		commitErr:      commitErr,
	}
	dispatcher := &projectSystemTypeCloneDispatcherStub{harness: harness, projectID: projectID, copyID: copyID}
	handler := NewCloneSystemTypeForProjectHandler(CloneSystemTypeForProjectDependencies{
		TransactionRunner:   harness.runner,
		TransactionWorkflow: harness.factory,
		Dispatcher:          dispatcher,
	})

	_, err := handler.Execute(context.Background(), CloneSystemTypeForProjectCommand{
		ProjectID:                       projectID,
		SourceSPSControllerSystemTypeID: sourceID,
	})
	if !errors.Is(err, commitErr) {
		t.Fatalf("expected commit error, got %v", err)
	}
	if harness.committed.systemTypes[copyID] != nil || len(dispatcher.commands) != 0 {
		t.Fatalf("commit failure leaked state: committed=%+v commands=%v", harness.committed, dispatcher.commands)
	}
}

func TestCloneSystemTypeForProjectDispatchFailureIsBestEffort(t *testing.T) {
	projectID := spsTestUUID(341)
	sourceID := spsTestUUID(342)
	copyID := spsTestUUID(343)
	spsControllerID := spsTestUUID(344)
	systemTypeDefinitionID := spsTestUUID(345)
	dispatchErr := errors.New("realtime unavailable")
	harness := &projectSystemTypeCloneHarness{
		committed: projectSystemTypeCloneState{
			systemTypes: map[uuid.UUID]*domainFacility.SPSControllerSystemType{
				sourceID: {
					Base:            domain.Base{ID: sourceID},
					SPSControllerID: spsControllerID,
					SystemTypeID:    systemTypeDefinitionID,
				},
			},
			projectFieldLinks: map[uuid.UUID][]uuid.UUID{},
		},
		copyEntity: &domainFacility.SPSControllerSystemType{
			Base:            domain.Base{ID: copyID},
			SPSControllerID: spsControllerID,
			SystemTypeID:    systemTypeDefinitionID,
		},
	}
	dispatcher := &projectSystemTypeCloneDispatcherStub{
		harness:   harness,
		projectID: projectID,
		copyID:    copyID,
		err:       dispatchErr,
	}
	var reported []error
	handler := NewCloneSystemTypeForProjectHandler(CloneSystemTypeForProjectDependencies{
		TransactionRunner:   harness.runner,
		TransactionWorkflow: harness.factory,
		Dispatcher:          dispatcher,
		ReportError: func(err error) {
			reported = append(reported, err)
		},
	})

	copyEntity, err := handler.CloneSystemTypeForProject(
		context.Background(),
		CloneSystemTypeForProjectCommand{
			ProjectID:                       projectID,
			SourceSPSControllerSystemTypeID: sourceID,
		},
	)
	if err != nil || copyEntity == nil || copyEntity.ID != copyID {
		t.Fatalf("best-effort result: entity=%+v err=%v", copyEntity, err)
	}
	if len(reported) != 1 || !errors.Is(reported[0], dispatchErr) {
		t.Fatalf("reported dispatch errors: %v", reported)
	}
}

func TestCloneSystemTypeForProjectValidatesConfigurationAndIdentifiers(t *testing.T) {
	unconfigured := NewCloneSystemTypeForProjectHandler(CloneSystemTypeForProjectDependencies{})
	_, err := unconfigured.Execute(context.Background(), CloneSystemTypeForProjectCommand{
		ProjectID:                       uuid.New(),
		SourceSPSControllerSystemTypeID: uuid.New(),
	})
	if !errors.Is(err, ErrCloneSystemTypeForProjectTransactionNotConfigured) {
		t.Fatalf("unconfigured error: %v", err)
	}

	harness := &projectSystemTypeCloneHarness{}
	configured := NewCloneSystemTypeForProjectHandler(CloneSystemTypeForProjectDependencies{
		TransactionRunner:   harness.runner,
		TransactionWorkflow: harness.factory,
	})
	for _, command := range []CloneSystemTypeForProjectCommand{
		{SourceSPSControllerSystemTypeID: uuid.New()},
		{ProjectID: uuid.New()},
	} {
		_, err := configured.Execute(context.Background(), command)
		if !errors.Is(err, domain.ErrInvalidArgument) {
			t.Fatalf("command %+v: got %v, want invalid argument", command, err)
		}
	}
	if harness.runnerCalls != 0 {
		t.Fatalf("invalid identifiers started %d transactions", harness.runnerCalls)
	}
}

func TestCloneSystemTypeForProjectRejectsSourceOutsideProjectBeforeMutation(t *testing.T) {
	projectID := spsTestUUID(351)
	sourceID := spsTestUUID(352)
	controllerID := spsTestUUID(353)
	harness := &projectSystemTypeCloneHarness{
		committed: projectSystemTypeCloneState{
			systemTypes: map[uuid.UUID]*domainFacility.SPSControllerSystemType{
				sourceID: {
					Base:            domain.Base{ID: sourceID},
					SPSControllerID: controllerID,
					SystemTypeID:    spsTestUUID(354),
				},
			},
			projectFieldLinks: map[uuid.UUID][]uuid.UUID{},
		},
		sourceAccessErr: domain.ErrNotFound,
	}
	handler := NewCloneSystemTypeForProjectHandler(CloneSystemTypeForProjectDependencies{
		TransactionRunner:   harness.runner,
		TransactionWorkflow: harness.factory,
	})

	_, err := handler.Execute(context.Background(), CloneSystemTypeForProjectCommand{
		ProjectID:                       projectID,
		SourceSPSControllerSystemTypeID: sourceID,
	})
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("source access error: got %v, want %v", err, domain.ErrNotFound)
	}
	if harness.accessCalls != 1 || harness.copyCalls != 0 {
		t.Fatalf("workflow calls: access=%d copy=%d", harness.accessCalls, harness.copyCalls)
	}
	if len(harness.committed.systemTypes) != 1 ||
		len(harness.committed.projectFieldLinks) != 0 || len(harness.committed.history) != 0 {
		t.Fatalf("source denial changed committed state: %+v", harness.committed)
	}
}
