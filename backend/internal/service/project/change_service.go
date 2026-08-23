package project

import (
	"context"

	domainProject "github.com/besart951/go_infra_link/backend/internal/domain/project"
	"github.com/google/uuid"
)

type ChangeService struct {
	store domainProject.ChangeStore
}

func NewChangeService(store domainProject.ChangeStore) *ChangeService {
	return &ChangeService{store: store}
}

func (s *ChangeService) ListAfter(ctx context.Context, projectID uuid.UUID, afterRevision uint64, limit int) (*domainProject.ChangePage, error) {
	return s.store.ListAfter(ctx, projectID, afterRevision, limit)
}

// RecordEvent converts the existing semantic project event vocabulary into one
// durable aggregate event per entity. It is deliberately independent of the
// WebSocket collaboration payload.
func (s *ChangeService) RecordEvent(ctx context.Context, projectID uuid.UUID, eventType string, actorID *uuid.UUID, entityIDs ...string) error {
	_, err := s.RecordEvents(ctx, projectID, eventType, actorID, entityIDs...)
	return err
}

// RecordEvents persists semantic changes and returns their allocated project
// revisions so the realtime transport can publish the durable cursor exactly.
func (s *ChangeService) RecordEvents(ctx context.Context, projectID uuid.UUID, eventType string, actorID *uuid.UUID, entityIDs ...string) ([]domainProject.Change, error) {
	return s.RecordEventsWithFields(ctx, projectID, eventType, actorID, nil, entityIDs...)
}

// RecordEventsWithFields records one durable project change per entity. A nil
// changedFields value preserves the legacy semantic defaults; a non-nil value
// is the precise set supplied by the mutation request contract.
func (s *ChangeService) RecordEventsWithFields(ctx context.Context, projectID uuid.UUID, eventType string, actorID *uuid.UUID, changedFields []string, entityIDs ...string) ([]domainProject.Change, error) {
	semantics, err := changeSemantics(eventType)
	if err != nil {
		return nil, err
	}
	if changedFields == nil {
		changedFields = append([]string(nil), semantics.changedFields...)
	}
	ids := parseEntityIDs(entityIDs)
	inputs := make([]domainProject.NewChange, 0, max(1, len(ids)))
	if len(ids) == 0 {
		if semantics.aggregate == domainProject.AggregateProject {
			ids = append(ids, projectID)
		} else {
			inputs = append(inputs, newChange(projectID, semantics.aggregate, nil, semantics.action, actorID, changedFields))
		}
	}
	for i := range ids {
		id := ids[i]
		inputs = append(inputs, newChange(projectID, semantics.aggregate, &id, semantics.action, actorID, changedFields))
	}
	if batch, ok := s.store.(domainProject.BatchChangeStore); ok {
		return batch.AppendBatch(ctx, inputs)
	}
	changes := make([]domainProject.Change, 0, len(inputs))
	for _, input := range inputs {
		change, err := s.store.Append(ctx, input)
		if err != nil {
			return nil, err
		}
		changes = append(changes, *change)
	}
	return changes, nil
}

func newChange(projectID uuid.UUID, aggregateType domainProject.AggregateKind, aggregateID *uuid.UUID, action domainProject.ChangeAction, actorID *uuid.UUID, changedFields []string) domainProject.NewChange {
	return domainProject.NewChange{
		ProjectID:     projectID,
		AggregateType: aggregateType,
		AggregateID:   aggregateID,
		Action:        action,
		ActorID:       actorID,
		ChangedFields: changedFields,
	}
}

func parseEntityIDs(values []string) []uuid.UUID {
	ids := make([]uuid.UUID, 0, len(values))
	for _, value := range values {
		id, err := uuid.Parse(value)
		if err == nil {
			ids = append(ids, id)
		}
	}
	return ids
}
