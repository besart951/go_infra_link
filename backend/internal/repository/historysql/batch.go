package historysql

import (
	"context"
	"fmt"
	"strings"
	"time"

	domainHistory "github.com/besart951/go_infra_link/backend/internal/domain/history"
	"github.com/besart951/go_infra_link/backend/internal/service/auditctx"
	"github.com/google/uuid"
)

const historyWriteBatchSize = 500

// RecordMutations persists one correlated history set. All IDs, JSON
// projections, and scopes are prepared before the first insert so a preparation
// failure cannot leave a partial history set even when a caller omitted the
// expected transaction.
func (s *Store) RecordMutations(ctx context.Context, mutations []Mutation) error {
	filtered := make([]Mutation, 0, len(mutations))
	for _, mutation := range mutations {
		if !allowedTable(mutation.EntityTable) {
			return fmt.Errorf("history table not allowed: %s", mutation.EntityTable)
		}
		if mutation.EntityID == uuid.Nil {
			continue
		}
		filtered = append(filtered, mutation)
	}
	if len(filtered) == 0 {
		return nil
	}

	now := time.Now().UTC()
	events := make([]domainHistory.ChangeEvent, len(filtered))
	versions := make([]domainHistory.EntityVersion, len(filtered))
	for i, mutation := range filtered {
		eventID, err := uuid.NewV7()
		if err != nil {
			return err
		}
		versionID, err := uuid.NewV7()
		if err != nil {
			return err
		}
		metadataJSON, err := marshalJSON(mutation.Metadata)
		if err != nil {
			return err
		}
		changeDiff, err := diffJSON(mutation.BeforeJSON, mutation.AfterJSON)
		if err != nil {
			return err
		}

		actorID := mutation.ActorID
		if actorID == nil {
			if actor, ok := auditctx.ActorID(ctx); ok {
				actorID = actor
			}
		}
		batchID := mutation.BatchID
		if batchID == nil {
			if batch, ok := auditctx.BatchID(ctx); ok {
				batchID = batch
			}
		}

		var summary *string
		if trimmed := strings.TrimSpace(mutation.Summary); trimmed != "" {
			summary = &trimmed
		}
		events[i] = domainHistory.ChangeEvent{
			ID:           eventID,
			OccurredAt:   now,
			ActorID:      actorID,
			Action:       mutation.Action,
			EntityTable:  mutation.EntityTable,
			EntityID:     mutation.EntityID,
			BatchID:      batchID,
			Summary:      summary,
			BeforeJSON:   mutation.BeforeJSON,
			AfterJSON:    mutation.AfterJSON,
			DiffJSON:     changeDiff,
			MetadataJSON: metadataJSON,
		}
		versions[i] = domainHistory.EntityVersion{
			ID:            versionID,
			ChangeEventID: eventID,
			EntityTable:   mutation.EntityTable,
			EntityID:      mutation.EntityID,
			VersionAt:     now,
			Action:        mutation.Action,
			SnapshotJSON:  mutation.AfterJSON,
		}
		if mutation.Action == domainHistory.ActionDelete {
			versions[i].SnapshotJSON = nil
		}
	}

	resolved, err := s.resolveMutationScopes(ctx, filtered)
	if err != nil {
		return err
	}
	scopeRows := make([]domainHistory.ChangeEventScope, 0)
	for i, scopes := range resolved {
		for _, scope := range scopes {
			scopeID, err := uuid.NewV7()
			if err != nil {
				return err
			}
			scopeRows = append(scopeRows, domainHistory.ChangeEventScope{
				ID:            scopeID,
				ChangeEventID: events[i].ID,
				ScopeType:     scope.Type,
				ScopeID:       scope.ID,
				OccurredAt:    now,
			})
		}
	}

	if err := s.db.WithContext(ctx).CreateInBatches(events, historyWriteBatchSize).Error; err != nil {
		return err
	}
	if len(scopeRows) > 0 {
		if err := s.db.WithContext(ctx).CreateInBatches(scopeRows, historyWriteBatchSize).Error; err != nil {
			return err
		}
	}
	return s.db.WithContext(ctx).CreateInBatches(versions, historyWriteBatchSize).Error
}
