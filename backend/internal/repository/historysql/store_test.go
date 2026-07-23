package historysql

import (
	"context"
	"fmt"
	"strings"
	"testing"

	domainHistory "github.com/besart951/go_infra_link/backend/internal/domain/history"
	"github.com/besart951/go_infra_link/backend/internal/service/auditctx"
	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestRecordMutationUsesAuditBatchContextUnlessMutationOverridesIt(t *testing.T) {
	contextBatchID := uuid.New()
	explicitBatchID := uuid.New()
	actorID := uuid.New()

	tests := []struct {
		name      string
		mutation  Mutation
		wantBatch uuid.UUID
	}{
		{
			name: "context batch",
			mutation: Mutation{
				Action:      domainHistory.ActionUpdate,
				EntityTable: "buildings",
				EntityID:    uuid.New(),
				BeforeJSON:  domainHistory.JSONB(`{"name":"before"}`),
				AfterJSON:   domainHistory.JSONB(`{"name":"after"}`),
			},
			wantBatch: contextBatchID,
		},
		{
			name: "explicit mutation batch takes precedence",
			mutation: Mutation{
				Action:      domainHistory.ActionUpdate,
				EntityTable: "buildings",
				EntityID:    uuid.New(),
				BeforeJSON:  domainHistory.JSONB(`{"name":"before"}`),
				AfterJSON:   domainHistory.JSONB(`{"name":"after"}`),
				BatchID:     &explicitBatchID,
			},
			wantBatch: explicitBatchID,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			db := newHistoryStoreTestDB(t)
			store := NewStore(db)
			ctx := auditctx.WithBatchID(
				auditctx.WithActorID(context.Background(), actorID),
				contextBatchID,
			)

			if err := store.RecordMutation(ctx, test.mutation); err != nil {
				t.Fatalf("RecordMutation returned error: %v", err)
			}

			var event domainHistory.ChangeEvent
			if err := db.Where(
				"entity_table = ? AND entity_id = ?",
				test.mutation.EntityTable,
				test.mutation.EntityID,
			).First(&event).Error; err != nil {
				t.Fatalf("load history event: %v", err)
			}
			if event.BatchID == nil || *event.BatchID != test.wantBatch {
				t.Fatalf("batch: got %v, want %s", event.BatchID, test.wantBatch)
			}
			if event.ActorID == nil || *event.ActorID != actorID {
				t.Fatalf("actor: got %v, want %s", event.ActorID, actorID)
			}
		})
	}
}

func TestChunkUUIDArgumentsBoundsTheSingleIDSlice(t *testing.T) {
	ids := make([]uuid.UUID, historyWriteBatchSize+1)
	for i := range ids {
		ids[i] = uuid.New()
	}
	args := []any{"project", ids}

	batches := chunkUUIDArguments(args, historyWriteBatchSize)
	if len(batches) != 2 {
		t.Fatalf("batches: got %d, want 2", len(batches))
	}
	for i, wantLength := range []int{historyWriteBatchSize, 1} {
		if got := batches[i][0]; got != "project" {
			t.Fatalf("batch %d scalar arg: got %v, want project", i, got)
		}
		chunk, ok := batches[i][1].([]uuid.UUID)
		if !ok {
			t.Fatalf("batch %d UUID arg has type %T", i, batches[i][1])
		}
		if len(chunk) != wantLength {
			t.Fatalf(
				"batch %d UUID count: got %d, want %d",
				i,
				len(chunk),
				wantLength,
			)
		}
	}
	if got := len(args[1].([]uuid.UUID)); got != len(ids) {
		t.Fatalf("source args mutated: got %d IDs, want %d", got, len(ids))
	}
}

func newHistoryStoreTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	dsn := fmt.Sprintf(
		"file:%s?mode=memory&cache=shared",
		strings.NewReplacer("/", "_", " ", "_", "#", "_").Replace(t.Name()),
	)
	db, err := gorm.Open(
		sqlite.Open(dsn),
		&gorm.Config{DisableForeignKeyConstraintWhenMigrating: true},
	)
	if err != nil {
		t.Fatalf("open sqlite database: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("get sql database: %v", err)
	}
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() {
		_ = sqlDB.Close()
	})

	if err := db.AutoMigrate(
		&domainHistory.ChangeEvent{},
		&domainHistory.ChangeEventScope{},
		&domainHistory.EntityVersion{},
	); err != nil {
		t.Fatalf("migrate history tables: %v", err)
	}
	return db
}
