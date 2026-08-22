package historysql

import (
	"context"

	domainHistory "github.com/besart951/go_infra_link/backend/internal/domain/history"
)

type mutationV2Write struct {
	event   *domainHistory.ChangeEvent
	scopes  []domainHistory.ChangeEventScope
	version *domainHistory.EntityVersion
}

func (s *Store) recordMutationV2(ctx context.Context, write mutationV2Write) error {
	if s.db.Dialector != nil && s.db.Dialector.Name() != "postgres" && !s.db.Migrator().HasTable(&historyV2StateRecord{}) {
		return nil
	}
	if err := s.db.WithContext(ctx).Create(changeEventV2FromDomain(write.event)).Error; err != nil {
		return err
	}
	if len(write.scopes) > 0 {
		records := make([]changeEventScopeV2Record, len(write.scopes))
		for i := range write.scopes {
			records[i] = changeEventScopeV2FromDomain(write.scopes[i])
		}
		if err := s.db.WithContext(ctx).CreateInBatches(records, 500).Error; err != nil {
			return err
		}
	}
	return s.db.WithContext(ctx).Create(entityVersionV2FromDomain(write.version)).Error
}

func changeEventV2FromDomain(event *domainHistory.ChangeEvent) changeEventV2Record {
	return changeEventV2Record{
		ID: event.ID, OccurredAt: event.OccurredAt, ActorID: event.ActorID,
		Action: event.Action, EntityTable: event.EntityTable, EntityID: event.EntityID,
		BatchID: event.BatchID, Summary: event.Summary, BeforeJSON: event.BeforeJSON,
		AfterJSON: event.AfterJSON, DiffJSON: event.DiffJSON, MetadataJSON: event.MetadataJSON,
	}
}

func changeEventScopeV2FromDomain(scope domainHistory.ChangeEventScope) changeEventScopeV2Record {
	return changeEventScopeV2Record{
		ID: scope.ID, OccurredAt: scope.OccurredAt, ChangeEventID: scope.ChangeEventID,
		ScopeType: scope.ScopeType, ScopeID: scope.ScopeID,
	}
}

func entityVersionV2FromDomain(version *domainHistory.EntityVersion) entityVersionV2Record {
	return entityVersionV2Record{
		ID: version.ID, VersionAt: version.VersionAt, ChangeEventID: version.ChangeEventID,
		EntityTable: version.EntityTable, EntityID: version.EntityID,
		Action: version.Action, SnapshotJSON: version.SnapshotJSON,
	}
}
