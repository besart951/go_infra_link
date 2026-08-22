package project

import (
	"context"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	"github.com/google/uuid"
)

type aggregateVersionLocker interface {
	LockAtVersion(context.Context, uuid.UUID, uint64) error
}

func lockAggregateVersion(ctx context.Context, repository any, id uuid.UUID, value uint64) error {
	version, err := domain.NewAggregateVersion(value)
	if err != nil {
		return err
	}
	locker, ok := repository.(aggregateVersionLocker)
	if !ok {
		return domain.ErrInvalidArgument
	}
	return locker.LockAtVersion(ctx, id, version.Uint64())
}
