package facilityjobsql

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	facilityjobs "github.com/besart951/go_infra_link/backend/internal/application/facilityjobs"
	apptransaction "github.com/besart951/go_infra_link/backend/internal/application/transaction"
	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type copiedRoot struct {
	ID uuid.UUID `gorm:"type:uuid;primaryKey"`
}

func TestStepStoreCompletesMutationAndMappingAtomically(t *testing.T) {
	db := openStepStoreTestDB(t)
	store := NewStepStore(db)
	step := testStep()
	mutationCalls := 0

	first, resumed, err := store.Execute(context.Background(), step, func(_ context.Context, unit apptransaction.UnitOfWork) (facilityjobs.StepResult, error) {
		mutationCalls++
		tx := unit.(*gorm.DB)
		targetID := uuid.New()
		return facilityjobs.StepResult{TargetID: targetID, Result: json.RawMessage(`{"ok":true}`)}, tx.Create(&copiedRoot{ID: targetID}).Error
	})
	if err != nil || resumed {
		t.Fatalf("first execution failed: resumed=%v err=%v", resumed, err)
	}

	second, resumed, err := store.Execute(context.Background(), step, func(context.Context, apptransaction.UnitOfWork) (facilityjobs.StepResult, error) {
		mutationCalls++
		return facilityjobs.StepResult{}, nil
	})
	if err != nil || !resumed {
		t.Fatalf("resume failed: resumed=%v err=%v", resumed, err)
	}
	if mutationCalls != 1 || second.TargetID != first.TargetID {
		t.Fatalf("mutation reran or target changed: calls=%d first=%s second=%s", mutationCalls, first.TargetID, second.TargetID)
	}
}

func TestStepStoreRollsBackDomainMutationAndPersistsFailure(t *testing.T) {
	db := openStepStoreTestDB(t)
	store := NewStepStore(db)
	step := testStep()
	wantErr := errors.New("injected failure")

	_, _, err := store.Execute(context.Background(), step, func(_ context.Context, unit apptransaction.UnitOfWork) (facilityjobs.StepResult, error) {
		tx := unit.(*gorm.DB)
		if createErr := tx.Create(&copiedRoot{ID: uuid.New()}).Error; createErr != nil {
			return facilityjobs.StepResult{}, createErr
		}
		return facilityjobs.StepResult{}, wantErr
	})
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected injected error, got %v", err)
	}
	var roots int64
	if err := db.Model(&copiedRoot{}).Count(&roots).Error; err != nil || roots != 0 {
		t.Fatalf("domain mutation escaped transaction: count=%d err=%v", roots, err)
	}
	item, err := store.GetItem(context.Background(), step.Key)
	if err != nil || item.Status != facilityjobs.ItemStatusFailed {
		t.Fatalf("failure item not persisted: item=%+v err=%v", item, err)
	}
}

func TestStepStorePreparedPlanResumesWithoutIDMapping(t *testing.T) {
	db := openStepStoreTestDB(t)
	store := NewStepStore(db)
	step := testStep()
	step.PersistIDMapping = false
	step.Input = json.RawMessage(`{"group":[1,2]}`)
	if err := store.Prepare(t.Context(), []facilityjobs.Step{step}); err != nil {
		t.Fatalf("Prepare() error = %v", err)
	}

	calls := 0
	mutation := func(context.Context, apptransaction.UnitOfWork) (facilityjobs.StepResult, error) {
		calls++
		return facilityjobs.StepResult{Result: json.RawMessage(`{"updated":2}`)}, nil
	}
	if _, resumed, err := store.Execute(t.Context(), step, mutation); err != nil || resumed {
		t.Fatalf("first Execute() resumed=%v error=%v", resumed, err)
	}
	result, resumed, err := store.Execute(t.Context(), step, mutation)
	if err != nil || !resumed || calls != 1 || string(result.Result) != `{"updated":2}` {
		t.Fatalf("resumed result=%s resumed=%v calls=%d error=%v", result.Result, resumed, calls, err)
	}
	if _, err := store.GetMapping(t.Context(), step); !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("unexpected ID mapping error = %v", err)
	}
}

func openStepStoreTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:"+uuid.NewString()+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&itemRecord{}, &mappingRecord{}, &copiedRoot{}); err != nil {
		t.Fatal(err)
	}
	return db
}

func testStep() facilityjobs.Step {
	return facilityjobs.Step{
		Key:        facilityjobs.ItemKey{OwnerID: uuid.New(), JobID: uuid.New(), Ordinal: 0},
		EntityType: "field_device", SourceID: uuid.New(), Input: json.RawMessage(`{"source":"test"}`),
		PersistIDMapping: true,
	}
}
