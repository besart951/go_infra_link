package historysql

import (
	"context"
	"testing"
	"time"

	domainHistory "github.com/besart951/go_infra_link/backend/internal/domain/history"
	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestRecordMutationDualWritesHistoryV2(t *testing.T) {
	db := historyV2TestDB(t)
	store := NewStore(db)
	mutation := Mutation{
		Action: domainHistory.ActionCreate, EntityTable: "units", EntityID: uuid.New(),
		AfterJSON: domainHistory.JSONB(`{"id":"test"}`),
	}
	if err := store.RecordMutation(context.Background(), mutation); err != nil {
		t.Fatal(err)
	}
	assertHistoryCount(t, db, &domainHistory.ChangeEvent{}, 1)
	assertHistoryCount(t, db, &changeEventV2Record{}, 1)
	assertHistoryCount(t, db, &domainHistory.EntityVersion{}, 1)
	assertHistoryCount(t, db, &entityVersionV2Record{}, 1)
}

func TestRecordMutationRollsBackV1WhenV2WriteFails(t *testing.T) {
	db := historyV2TestDB(t)
	if err := db.Migrator().DropTable(&changeEventV2Record{}); err != nil {
		t.Fatal(err)
	}
	store := NewStore(db)
	err := store.RecordMutation(context.Background(), Mutation{
		Action: domainHistory.ActionCreate, EntityTable: "units", EntityID: uuid.New(),
		AfterJSON: domainHistory.JSONB(`{"id":"test"}`),
	})
	if err == nil {
		t.Fatal("expected V2 write failure")
	}
	assertHistoryCount(t, db, &domainHistory.ChangeEvent{}, 0)
	assertHistoryCount(t, db, &domainHistory.EntityVersion{}, 0)
}

func TestVerifiedCutoverReadsTimelineFromHistoryV2(t *testing.T) {
	db := historyV2TestDB(t)
	store := NewStore(db)
	entityID := uuid.New()
	if err := store.RecordMutation(context.Background(), Mutation{
		Action: domainHistory.ActionCreate, EntityTable: "units", EntityID: entityID,
		AfterJSON: domainHistory.JSONB(`{"id":"test"}`),
	}); err != nil {
		t.Fatal(err)
	}
	var source domainHistory.ChangeEvent
	if err := db.Where("entity_id = ?", entityID).Take(&source).Error; err != nil {
		t.Fatal(err)
	}
	if err := store.VerifyAndEnableV2Reads(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := db.Where("id = ?", source.ID).Delete(&domainHistory.ChangeEvent{}).Error; err != nil {
		t.Fatal(err)
	}
	event, err := store.GetEvent(context.Background(), source.ID)
	if err != nil || event.EntityID != entityID {
		t.Fatalf("V2 event=%+v err=%v", event, err)
	}
}

func historyV2TestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:"+uuid.NewString()+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := AutoMigrateV2(db, time.Now().UTC()); err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&domainHistory.ChangeEvent{}, &domainHistory.ChangeEventScope{}, &domainHistory.EntityVersion{}); err != nil {
		t.Fatal(err)
	}
	return db
}

func assertHistoryCount(t *testing.T, db *gorm.DB, model any, want int64) {
	t.Helper()
	var got int64
	if err := db.Model(model).Count(&got).Error; err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("count %T = %d, want %d", model, got, want)
	}
}
