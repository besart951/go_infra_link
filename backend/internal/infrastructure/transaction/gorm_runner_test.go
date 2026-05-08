package transaction_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	apptransaction "github.com/besart951/go_infra_link/backend/internal/application/transaction"
	infratransaction "github.com/besart951/go_infra_link/backend/internal/infrastructure/transaction"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type gormRunnerTestRecord struct {
	ID   string `gorm:"primaryKey"`
	Name string
}

func TestGormRunnerRollsBackOnErrorAfterPartialWrite(t *testing.T) {
	db := newGormRunnerTestDB(t)
	runner := infratransaction.NewGormRunner(db)
	stepErr := errors.New("second step failed")

	err := runner(context.Background(), func(runCtx context.Context, unit apptransaction.UnitOfWork) error {
		tx, err := infratransaction.GormDB(unit)
		if err != nil {
			return err
		}
		if err := tx.WithContext(runCtx).Create(&gormRunnerTestRecord{ID: "one", Name: "created"}).Error; err != nil {
			return err
		}
		return stepErr
	})

	if !errors.Is(err, stepErr) {
		t.Fatalf("expected step error, got %v", err)
	}
	assertGormRunnerRecordCount(t, db, 0)
}

func TestGormRunnerCommitsOnSuccess(t *testing.T) {
	db := newGormRunnerTestDB(t)
	runner := infratransaction.NewGormRunner(db)

	err := runner(context.Background(), func(runCtx context.Context, unit apptransaction.UnitOfWork) error {
		tx, err := infratransaction.GormDB(unit)
		if err != nil {
			return err
		}
		return tx.WithContext(runCtx).Create(&gormRunnerTestRecord{ID: "one", Name: "created"}).Error
	})

	if err != nil {
		t.Fatalf("expected commit to succeed, got %v", err)
	}
	assertGormRunnerRecordCount(t, db, 1)
}

func TestGormRunnerPropagatesContextToCallback(t *testing.T) {
	db := newGormRunnerTestDB(t)
	runner := infratransaction.NewGormRunner(db)
	type contextKey struct{}
	ctx := context.WithValue(context.Background(), contextKey{}, "expected")

	err := runner(ctx, func(runCtx context.Context, unit apptransaction.UnitOfWork) error {
		if got := runCtx.Value(contextKey{}); got != "expected" {
			return fmt.Errorf("expected context value to propagate, got %v", got)
		}
		if _, err := infratransaction.GormDB(unit); err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		t.Fatalf("expected transaction to succeed, got %v", err)
	}
}

func TestGormDBRejectsUnsupportedUnit(t *testing.T) {
	if _, err := infratransaction.GormDB(struct{}{}); err == nil {
		t.Fatal("expected unsupported unit error")
	}
}

func newGormRunnerTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", strings.NewReplacer("/", "_", " ", "_", "#", "_").Replace(t.Name()))
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{DisableForeignKeyConstraintWhenMigrating: true})
	if err != nil {
		t.Fatalf("expected sqlite db to open, got %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("expected sql db handle, got %v", err)
	}
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() {
		_ = sqlDB.Close()
	})

	if err := db.AutoMigrate(&gormRunnerTestRecord{}); err != nil {
		t.Fatalf("expected test table to migrate, got %v", err)
	}

	return db
}

func assertGormRunnerRecordCount(t *testing.T, db *gorm.DB, want int64) {
	t.Helper()

	var count int64
	if err := db.Model(&gormRunnerTestRecord{}).Count(&count).Error; err != nil {
		t.Fatalf("expected count to succeed, got %v", err)
	}
	if count != want {
		t.Fatalf("expected %d records, got %d", want, count)
	}
}
