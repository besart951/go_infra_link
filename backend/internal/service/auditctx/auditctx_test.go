package auditctx

import (
	"context"
	"testing"

	"github.com/google/uuid"
)

func TestActorAndBatchMetadataRemainIndependent(t *testing.T) {
	actorID := uuid.New()
	batchID := uuid.New()
	ctx := WithBatchID(WithActorID(context.Background(), actorID), batchID)

	gotActorID, ok := ActorID(ctx)
	if !ok || gotActorID == nil || *gotActorID != actorID {
		t.Fatalf("actor: got %v, %t; want %s", gotActorID, ok, actorID)
	}
	gotBatchID, ok := BatchID(ctx)
	if !ok || gotBatchID == nil || *gotBatchID != batchID {
		t.Fatalf("batch: got %v, %t; want %s", gotBatchID, ok, batchID)
	}
}

func TestAuditContextIgnoresNilIdentifiers(t *testing.T) {
	ctx := WithBatchID(WithActorID(context.Background(), uuid.Nil), uuid.Nil)

	if actorID, ok := ActorID(ctx); ok || actorID != nil {
		t.Fatalf("expected no actor, got %v, %t", actorID, ok)
	}
	if batchID, ok := BatchID(ctx); ok || batchID != nil {
		t.Fatalf("expected no batch, got %v, %t", batchID, ok)
	}
}
