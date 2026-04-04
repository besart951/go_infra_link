package domain

import (
	"context"

	"github.com/google/uuid"
)

// Small, composable repository interfaces (ISP-friendly).
//
// In Go, prefer defining the interface at the consumer side.
// These are provided as building blocks so each module can depend
// only on the capabilities it actually needs.

// GetByID is a convenience wrapper that fetches a single entity by ID
// using any Reader implementation. Returns ErrNotFound when absent.
func GetByID[T any](ctx context.Context, r Reader[T], id uuid.UUID) (*T, error) {
	items, err := r.GetByIds(ctx, []uuid.UUID{id})
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, ErrNotFound
	}
	return items[0], nil
}

type Reader[T any] interface {
	GetByIds(ctx context.Context, ids []uuid.UUID) ([]*T, error)
}

type Creator[T any] interface {
	Create(ctx context.Context, entity *T) error
}

type Updater[T any] interface {
	Update(ctx context.Context, entity *T) error
}

type Deleter[T any] interface {
	DeleteByIds(ctx context.Context, ids []uuid.UUID) error
}

type Paginator[T any] interface {
	GetPaginatedList(ctx context.Context, params PaginationParams) (*PaginatedList[T], error)
}

// Repository is the common CRUD + pagination contract.
//
// Keep method naming consistent with the current codebase.

type Repository[T any] interface {
	Reader[T]
	Creator[T]
	Updater[T]
	Deleter[T]
	Paginator[T]
}

// AppendOnlyRepository is useful for audit/history tables that should not be updated/deleted.

type AppendOnlyRepository[T any] interface {
	Reader[T]
	Creator[T]
	Paginator[T]
}
