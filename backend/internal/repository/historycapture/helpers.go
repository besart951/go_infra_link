package historycapture

import (
	"context"
	"reflect"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainHistory "github.com/besart951/go_infra_link/backend/internal/domain/history"
	"github.com/besart951/go_infra_link/backend/internal/repository/historysql"
	"github.com/google/uuid"
)

type audit[T any] struct {
	table string
	store *historysql.Store
}

func newAudit[T any](table string, store *historysql.Store) audit[T] {
	return audit[T]{table: table, store: store}
}

func (a audit[T]) create(ctx context.Context, next domain.Creator[T], entity *T) error {
	if err := next.Create(ctx, entity); err != nil {
		return err
	}
	return a.recordCreated(ctx, idOf(entity))
}

func (a audit[T]) update(ctx context.Context, next domain.Updater[T], entity *T) error {
	id := idOf(entity)
	before, _, err := a.store.LoadRow(ctx, a.table, id)
	if err != nil {
		return err
	}
	if err := next.Update(ctx, entity); err != nil {
		return err
	}
	return a.store.RecordUpdate(ctx, a.table, id, before)
}

func (a audit[T]) deleteByIds(ctx context.Context, next domain.Deleter[T], ids []uuid.UUID) error {
	if len(ids) == 0 {
		return nil
	}
	before, err := a.store.LoadRows(ctx, a.table, ids)
	if err != nil {
		return err
	}
	if err := next.DeleteByIds(ctx, ids); err != nil {
		return err
	}
	return a.recordDeletedRows(ctx, before)
}

func (a audit[T]) bulkCreate(ctx context.Context, create func(context.Context) error, ids func() []uuid.UUID) error {
	if err := create(ctx); err != nil {
		return err
	}
	for _, id := range ids() {
		if err := a.recordCreated(ctx, id); err != nil {
			return err
		}
	}
	return nil
}

func (a audit[T]) deleteRows(ctx context.Context, load func(context.Context) (map[uuid.UUID]domainHistory.JSONB, error), delete func(context.Context) error) error {
	before, err := load(ctx)
	if err != nil {
		return err
	}
	if err := delete(ctx); err != nil {
		return err
	}
	return a.recordDeletedRows(ctx, before)
}

func (a audit[T]) recordCreated(ctx context.Context, id uuid.UUID) error {
	if a.store == nil || id == uuid.Nil {
		return nil
	}
	return a.store.RecordCreate(ctx, a.table, id)
}

func (a audit[T]) recordDeletedRows(ctx context.Context, rows map[uuid.UUID]domainHistory.JSONB) error {
	if a.store == nil {
		return nil
	}
	for id, snapshot := range rows {
		if err := a.store.RecordDelete(ctx, a.table, id, snapshot); err != nil {
			return err
		}
	}
	return nil
}

func idOf[T any](entity *T) uuid.UUID {
	if entity == nil {
		return uuid.Nil
	}
	value := reflect.ValueOf(entity)
	if value.Kind() == reflect.Pointer {
		value = value.Elem()
	}
	if !value.IsValid() || value.Kind() != reflect.Struct {
		return uuid.Nil
	}
	base := value.FieldByName("Base")
	if base.IsValid() {
		id := base.FieldByName("ID")
		if id.IsValid() {
			if out, ok := id.Interface().(uuid.UUID); ok {
				return out
			}
		}
	}
	id := value.FieldByName("ID")
	if id.IsValid() {
		if out, ok := id.Interface().(uuid.UUID); ok {
			return out
		}
	}
	return uuid.Nil
}

func idsOf[T any](entities []*T) []uuid.UUID {
	ids := make([]uuid.UUID, 0, len(entities))
	for _, entity := range entities {
		if id := idOf(entity); id != uuid.Nil {
			ids = append(ids, id)
		}
	}
	return ids
}
