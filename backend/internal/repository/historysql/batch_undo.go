package historysql

import (
	"context"
	"sort"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainHistory "github.com/besart951/go_infra_link/backend/internal/domain/history"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type batchUndoTarget struct {
	first   domainHistory.ChangeEvent
	last    domainHistory.ChangeEvent
	current domainHistory.JSONB
}

func (s *Store) UndoBatch(ctx context.Context, batchID uuid.UUID) (*domainHistory.RestoreResult, error) {
	events, err := s.batchEvents(ctx, batchID)
	if err != nil {
		return nil, err
	}
	targets := groupBatchUndoTargets(events)
	result := &domainHistory.RestoreResult{BatchID: uuid.New()}
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return s.WithDB(tx).undoBatchInTransaction(ctx, targets, result)
	})
	return result, err
}

func (s *Store) batchEvents(ctx context.Context, batchID uuid.UUID) ([]domainHistory.ChangeEvent, error) {
	var events []domainHistory.ChangeEvent
	err := s.db.WithContext(ctx).Where("batch_id = ?", batchID).
		Order("occurred_at ASC, id ASC").Find(&events).Error
	if err == nil && len(events) == 0 {
		err = domain.ErrNotFound
	}
	return events, err
}

func groupBatchUndoTargets(events []domainHistory.ChangeEvent) []batchUndoTarget {
	byEntity := make(map[string]batchUndoTarget)
	for _, event := range events {
		key := targetKey(event.EntityTable, event.EntityID)
		target, exists := byEntity[key]
		if !exists {
			target.first = event
		}
		target.last = event
		byEntity[key] = target
	}
	keys := make([]string, 0, len(byEntity))
	for key := range byEntity {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	targets := make([]batchUndoTarget, len(keys))
	for index, key := range keys {
		targets[index] = byEntity[key]
	}
	return targets
}

func (s *Store) undoBatchInTransaction(ctx context.Context, targets []batchUndoTarget, result *domainHistory.RestoreResult) error {
	if err := s.preflightBatchUndo(ctx, targets); err != nil {
		return err
	}
	for index := range targets {
		if err := s.applyBatchUndoTarget(ctx, targets[index], result); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) preflightBatchUndo(ctx context.Context, targets []batchUndoTarget) error {
	for index := range targets {
		target := &targets[index]
		if err := s.lockUndoEntity(ctx, target.last.EntityTable, target.last.EntityID); err != nil {
			return err
		}
		current, _, err := s.LoadRow(ctx, target.last.EntityTable, target.last.EntityID)
		if err != nil {
			return err
		}
		target.current = current
		if err := s.checkUndoConflict(ctx, undoPreflight{event: &target.last, mode: domainHistory.RestoreModeBefore, current: current}); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) applyBatchUndoTarget(ctx context.Context, target batchUndoTarget, result *domainHistory.RestoreResult) error {
	if len(target.first.BeforeJSON) == 0 {
		if err := deleteRow(ctx, s.db, target.first.EntityTable, target.first.EntityID); err != nil {
			return err
		}
		result.DeletedCount++
	} else {
		if err := upsertRow(ctx, s.db, target.first.EntityTable, target.first.BeforeJSON); err != nil {
			return err
		}
		result.RestoredCount++
	}
	return s.RecordMutation(ctx, Mutation{
		Action: domainHistory.ActionRestore, EntityTable: target.first.EntityTable,
		EntityID: target.first.EntityID, BeforeJSON: target.current, AfterJSON: target.first.BeforeJSON,
		BatchID: &result.BatchID, Summary: "batch undone from history",
		Metadata: map[string]any{"source_batch_id": target.first.BatchID},
	})
}
