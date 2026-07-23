package historycapture

import (
	"context"

	domainHistory "github.com/besart951/go_infra_link/backend/internal/domain/history"
	"github.com/google/uuid"
)

// ChangeStore is the consumer-oriented Interface needed by transactional
// repository decorators. historysql.Store is the production Implementation.
type ChangeStore interface {
	LoadRow(context.Context, string, uuid.UUID) (domainHistory.JSONB, bool, error)
	LoadRows(context.Context, string, []uuid.UUID) (map[uuid.UUID]domainHistory.JSONB, error)
	LoadRowsWhere(context.Context, string, string, ...any) (map[uuid.UUID]domainHistory.JSONB, error)
	RecordCreate(context.Context, string, uuid.UUID) error
	RecordCreates(context.Context, string, []uuid.UUID) error
	RecordUpdate(context.Context, string, uuid.UUID, domainHistory.JSONB) error
	RecordUpdates(context.Context, string, map[uuid.UUID]domainHistory.JSONB) error
	RecordDelete(context.Context, string, uuid.UUID, domainHistory.JSONB) error
	RecordDeletes(context.Context, string, map[uuid.UUID]domainHistory.JSONB) error
}
