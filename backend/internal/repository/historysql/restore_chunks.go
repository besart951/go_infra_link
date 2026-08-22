package historysql

import (
	"context"
	"fmt"
	"time"

	hierarchyrestore "github.com/besart951/go_infra_link/backend/internal/application/hierarchyrestore"
	domainHistory "github.com/besart951/go_infra_link/backend/internal/domain/history"
	"github.com/besart951/go_infra_link/backend/internal/service/auditctx"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

const hierarchyRestoreChunkLimit = 500

type restoreChunkState struct {
	command  hierarchyrestore.Command
	versions map[uuid.UUID]domainHistory.EntityVersion
}

type restoreItem struct {
	state restoreChunkState
	id    uuid.UUID
}

type restoreHistoryEffect struct {
	command hierarchyrestore.Command
	id      uuid.UUID
	before  domainHistory.JSONB
	after   domainHistory.JSONB
	effect  string
}

type restoreEffect func(context.Context, *Store, restoreItem) (hierarchyrestore.Result, error)

var restoreEffects = map[hierarchyrestore.Phase]restoreEffect{
	hierarchyrestore.PhaseDelete:  applyRestoreDelete,
	hierarchyrestore.PhaseRestore: applyRestoreUpsert,
}

func (s *Store) RestoreChunk(ctx context.Context, command hierarchyrestore.Command) (hierarchyrestore.Result, error) {
	if !allowedTable(command.Table) || restoreEffects[command.Phase] == nil {
		return hierarchyrestore.Result{}, fmt.Errorf("invalid hierarchy restore stage")
	}
	if command.Limit <= 0 || command.Limit > hierarchyRestoreChunkLimit {
		command.Limit = hierarchyRestoreChunkLimit
	}
	ids, hasMore, err := s.restoreTargetIDs(ctx, command)
	if err != nil || len(ids) == 0 {
		return hierarchyrestore.Result{Done: len(ids) == 0}, err
	}
	versions, err := s.latestRestoreVersions(ctx, command, ids)
	if err != nil {
		return hierarchyrestore.Result{}, err
	}
	result, err := s.applyRestoreEffects(ctx, restoreChunkState{command: command, versions: versions}, ids)
	result.NextID = ids[len(ids)-1]
	result.Done = !hasMore
	return result, err
}

func (s *Store) restoreTargetIDs(ctx context.Context, command hierarchyrestore.Command) ([]uuid.UUID, bool, error) {
	query := s.db.WithContext(ctx).Table("change_events ce").
		Select("DISTINCT ce.entity_id").
		Joins("JOIN change_event_scopes cabinet_scope ON cabinet_scope.change_event_id = ce.id AND cabinet_scope.scope_type = ? AND cabinet_scope.scope_id = ?", scopeControlCabinet, command.ControlCabinetID).
		Where("ce.entity_table = ? AND ce.entity_id > ?", command.Table, command.AfterID)
	if command.ProjectID != nil {
		query = query.Joins("JOIN change_event_scopes project_scope ON project_scope.change_event_id = ce.id AND project_scope.scope_type = ? AND project_scope.scope_id = ?", scopeProject, *command.ProjectID)
	}
	var ids []uuid.UUID
	err := query.Order("ce.entity_id ASC").Limit(command.Limit+1).Pluck("ce.entity_id", &ids).Error
	if err != nil {
		return nil, false, err
	}
	hasMore := len(ids) > command.Limit
	if hasMore {
		ids = ids[:command.Limit]
	}
	return ids, hasMore, nil
}

func (s *Store) latestRestoreVersions(ctx context.Context, command hierarchyrestore.Command, ids []uuid.UUID) (map[uuid.UUID]domainHistory.EntityVersion, error) {
	var rows []domainHistory.EntityVersion
	err := s.db.WithContext(ctx).Raw(`
		SELECT ranked.id, ranked.change_event_id, ranked.entity_table, ranked.entity_id,
		       ranked.version_at, ranked.action, ranked.snapshot_json
		FROM (
			SELECT ev.*, ROW_NUMBER() OVER (
				PARTITION BY ev.entity_table, ev.entity_id
				ORDER BY ev.version_at DESC, ev.id DESC
			) AS row_number
			FROM entity_versions ev
			WHERE ev.entity_table = ? AND ev.entity_id IN ? AND ev.version_at <= ?
		) ranked
		WHERE ranked.row_number = 1
	`, command.Table, ids, command.AsOf).Scan(&rows).Error
	versions := make(map[uuid.UUID]domainHistory.EntityVersion, len(rows))
	for _, row := range rows {
		versions[row.EntityID] = row
	}
	return versions, err
}

func (s *Store) applyRestoreEffects(ctx context.Context, state restoreChunkState, ids []uuid.UUID) (hierarchyrestore.Result, error) {
	result := hierarchyrestore.Result{Processed: len(ids)}
	effect := restoreEffects[state.command.Phase]
	for _, id := range ids {
		item, err := effect(ctx, s, restoreItem{state: state, id: id})
		if err != nil {
			return hierarchyrestore.Result{}, err
		}
		result.Restored += item.Restored
		result.Deleted += item.Deleted
		result.Skipped += item.Skipped
	}
	return result, nil
}

func applyRestoreDelete(ctx context.Context, store *Store, item restoreItem) (hierarchyrestore.Result, error) {
	version, exists := item.state.versions[item.id]
	if exists && len(version.SnapshotJSON) > 0 {
		return hierarchyrestore.Result{Skipped: 1}, nil
	}
	before, found, err := store.LoadRow(ctx, item.state.command.Table, item.id)
	if err != nil || !found {
		return hierarchyrestore.Result{Skipped: 1}, err
	}
	if err := deleteRow(ctx, store.db, item.state.command.Table, item.id); err != nil {
		return hierarchyrestore.Result{}, err
	}
	err = store.recordRestoreEffect(ctx, restoreHistoryEffect{
		command: item.state.command, id: item.id, before: before, effect: "delete",
	})
	return hierarchyrestore.Result{Deleted: 1}, err
}

func applyRestoreUpsert(ctx context.Context, store *Store, item restoreItem) (hierarchyrestore.Result, error) {
	version, exists := item.state.versions[item.id]
	if !exists || len(version.SnapshotJSON) == 0 {
		return hierarchyrestore.Result{Skipped: 1}, nil
	}
	before, _, err := store.LoadRow(ctx, item.state.command.Table, item.id)
	if err != nil {
		return hierarchyrestore.Result{}, err
	}
	if err := upsertRow(ctx, store.db, item.state.command.Table, version.SnapshotJSON); err != nil {
		return hierarchyrestore.Result{}, err
	}
	err = store.recordRestoreEffect(ctx, restoreHistoryEffect{
		command: item.state.command, id: item.id, before: before,
		after: version.SnapshotJSON, effect: "upsert",
	})
	return hierarchyrestore.Result{Restored: 1}, err
}

func (s *Store) recordRestoreEffect(ctx context.Context, effect restoreHistoryEffect) error {
	ctx = auditctx.WithActorID(ctx, effect.command.ActorID)
	return s.RecordMutation(ctx, Mutation{
		Action: domainHistory.ActionRestore, EntityTable: effect.command.Table, EntityID: effect.id,
		BeforeJSON: effect.before, AfterJSON: effect.after, BatchID: &effect.command.BatchID,
		Summary: "hierarchy restored from history",
		Metadata: map[string]any{
			"control_cabinet_id": effect.command.ControlCabinetID.String(),
			"restore_as_of":      effect.command.AsOf.UTC().Format(time.RFC3339Nano), "restore_effect": effect.effect,
		},
	})
}

func ReleaseRestoreLifecycle(ctx context.Context, db *gorm.DB, cabinetID uuid.UUID) error {
	return db.WithContext(ctx).Exec(
		"DELETE FROM facility_aggregate_lifecycle WHERE kind = ? AND resource_id = ?",
		"control_cabinet", cabinetID,
	).Error
}
