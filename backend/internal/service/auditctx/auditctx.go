package auditctx

import (
	"context"

	"github.com/google/uuid"
)

type actorKey struct{}

func WithActorID(ctx context.Context, actorID uuid.UUID) context.Context {
	if actorID == uuid.Nil {
		return ctx
	}
	return context.WithValue(ctx, actorKey{}, actorID)
}

func ActorID(ctx context.Context) (*uuid.UUID, bool) {
	id, ok := ctx.Value(actorKey{}).(uuid.UUID)
	if !ok || id == uuid.Nil {
		return nil, false
	}
	return &id, true
}
