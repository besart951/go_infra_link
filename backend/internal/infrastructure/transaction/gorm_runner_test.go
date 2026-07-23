package transaction_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	apptransaction "github.com/besart951/go_infra_link/backend/internal/application/transaction"
	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	domainHistory "github.com/besart951/go_infra_link/backend/internal/domain/history"
	infratransaction "github.com/besart951/go_infra_link/backend/internal/infrastructure/transaction"
	"github.com/besart951/go_infra_link/backend/internal/repository/facilitysql"
	"github.com/besart951/go_infra_link/backend/internal/repository/historycapture"
	"github.com/besart951/go_infra_link/backend/internal/repository/historysql"
	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type gormRunnerTestRecord struct {
	ID   string `gorm:"primaryKey"`
	Name string
}

type gormRunnerFacilityRecord struct {
	ID   string `gorm:"primaryKey"`
	Name string
}

type gormRunnerHistoryRecord struct {
	ID       string `gorm:"primaryKey"`
	EntityID string
}

type failingDeleteHistoryStore struct {
	historycapture.ChangeStore
	db               *gorm.DB
	model            any
	before           map[uuid.UUID]domainHistory.JSONB
	err              error
	recordSawDeleted bool
}

type failingObjectDataUpdateHistoryStore struct {
	historycapture.ChangeStore
	db               *gorm.DB
	err              error
	recordSawUpdated bool
}

func (s *failingObjectDataUpdateHistoryStore) LoadRow(
	context.Context,
	string,
	uuid.UUID,
) (domainHistory.JSONB, bool, error) {
	return domainHistory.JSONB(`{"is_active":false}`), true, nil
}

func (s *failingObjectDataUpdateHistoryStore) RecordUpdate(
	ctx context.Context,
	_ string,
	id uuid.UUID,
	_ domainHistory.JSONB,
) error {
	var objectData domainFacility.ObjectData
	if err := s.db.WithContext(ctx).Select("id", "is_active").First(&objectData, "id = ?", id).Error; err != nil {
		return err
	}
	s.recordSawUpdated = objectData.IsActive
	return s.err
}

func (s *failingDeleteHistoryStore) LoadRows(
	context.Context,
	string,
	[]uuid.UUID,
) (map[uuid.UUID]domainHistory.JSONB, error) {
	return s.before, nil
}

func (s *failingDeleteHistoryStore) RecordDeletes(
	ctx context.Context,
	_ string,
	_ map[uuid.UUID]domainHistory.JSONB,
) error {
	var count int64
	if err := s.db.WithContext(ctx).Model(s.model).Count(&count).Error; err != nil {
		return err
	}
	s.recordSawDeleted = count == 0
	return s.err
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

func TestGormRunnerCommitsAndRollsBackFacilityAndHistoryTogether(t *testing.T) {
	tests := []struct {
		name        string
		callbackErr error
		wantCount   int64
	}{
		{name: "commit", wantCount: 1},
		{
			name:        "rollback",
			callbackErr: errors.New("history persistence failed"),
			wantCount:   0,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			db := newGormRunnerTestDB(t)
			runner := infratransaction.NewGormRunner(db)

			err := runner(
				context.Background(),
				func(runCtx context.Context, unit apptransaction.UnitOfWork) error {
					tx, err := infratransaction.GormDB(unit)
					if err != nil {
						return err
					}
					if err := tx.WithContext(runCtx).Create(&gormRunnerFacilityRecord{
						ID:   "facility-one",
						Name: "updated",
					}).Error; err != nil {
						return err
					}
					if err := tx.WithContext(runCtx).Create(&gormRunnerHistoryRecord{
						ID:       "history-one",
						EntityID: "facility-one",
					}).Error; err != nil {
						return err
					}
					return test.callbackErr
				},
			)

			if !errors.Is(err, test.callbackErr) {
				t.Fatalf("transaction error: got %v, want %v", err, test.callbackErr)
			}
			assertGormRunnerModelCount(t, db, &gormRunnerFacilityRecord{}, test.wantCount)
			assertGormRunnerModelCount(t, db, &gormRunnerHistoryRecord{}, test.wantCount)
		})
	}
}

func TestNestedGormRunnerRollsBackFailedItemAndKeepsLaterSuccess(t *testing.T) {
	db := newGormRunnerTestDB(t)
	outer := infratransaction.NewGormRunner(db)
	itemErr := errors.New("BACnet validation failed")

	err := outer(context.Background(), func(runCtx context.Context, unit apptransaction.UnitOfWork) error {
		tx, txErr := infratransaction.GormDB(unit)
		if txErr != nil {
			return txErr
		}
		perItem := infratransaction.NewGormRunner(tx)
		failedErr := perItem(runCtx, func(itemCtx context.Context, itemUnit apptransaction.UnitOfWork) error {
			itemTx, itemTxErr := infratransaction.GormDB(itemUnit)
			if itemTxErr != nil {
				return itemTxErr
			}
			if createErr := itemTx.WithContext(itemCtx).Create(&gormRunnerFacilityRecord{
				ID:   "failed-field-device",
				Name: "staged root",
			}).Error; createErr != nil {
				return createErr
			}
			if createErr := itemTx.WithContext(itemCtx).Create(&gormRunnerHistoryRecord{
				ID:       "failed-history",
				EntityID: "failed-field-device",
			}).Error; createErr != nil {
				return createErr
			}
			return itemErr
		})
		if !errors.Is(failedErr, itemErr) {
			return fmt.Errorf("failed item: got %v, want %w", failedErr, itemErr)
		}

		return perItem(runCtx, func(itemCtx context.Context, itemUnit apptransaction.UnitOfWork) error {
			itemTx, itemTxErr := infratransaction.GormDB(itemUnit)
			if itemTxErr != nil {
				return itemTxErr
			}
			if createErr := itemTx.WithContext(itemCtx).Create(&gormRunnerFacilityRecord{
				ID:   "successful-field-device",
				Name: "committed root",
			}).Error; createErr != nil {
				return createErr
			}
			return itemTx.WithContext(itemCtx).Create(&gormRunnerHistoryRecord{
				ID:       "successful-history",
				EntityID: "successful-field-device",
			}).Error
		})
	})
	if err != nil {
		t.Fatalf("outer partial-success transaction: %v", err)
	}

	var facilities []gormRunnerFacilityRecord
	if err := db.Order("id").Find(&facilities).Error; err != nil {
		t.Fatalf("load facility rows: %v", err)
	}
	if len(facilities) != 1 || facilities[0].ID != "successful-field-device" {
		t.Fatalf("facility rows: %+v", facilities)
	}
	var history []gormRunnerHistoryRecord
	if err := db.Order("id").Find(&history).Error; err != nil {
		t.Fatalf("load history rows: %v", err)
	}
	if len(history) != 1 || history[0].EntityID != "successful-field-device" {
		t.Fatalf("history rows: %+v", history)
	}
}

func TestGormRunnerRollsBackAlarmValueReplacementWhenHistoryWriteFails(t *testing.T) {
	db := newGormRunnerTestDB(t)
	if err := db.AutoMigrate(&domainFacility.BacnetObjectAlarmValue{}); err != nil {
		t.Fatalf("migrate alarm values: %v", err)
	}
	bacnetObjectID := uuid.New()
	oldValueID := uuid.New()
	oldValue := domainFacility.BacnetObjectAlarmValue{
		Base:             domain.Base{ID: oldValueID},
		BacnetObjectID:   bacnetObjectID,
		AlarmTypeFieldID: uuid.New(),
		Source:           domainFacility.AlarmValueSourceUser,
	}
	if err := db.Create(&oldValue).Error; err != nil {
		t.Fatalf("seed alarm value: %v", err)
	}
	replacement := []domainFacility.BacnetObjectAlarmValue{{
		AlarmTypeFieldID: uuid.New(),
		Source:           domainFacility.AlarmValueSourceUser,
	}}

	runner := infratransaction.NewGormRunner(db)
	err := runner(context.Background(), func(runCtx context.Context, unit apptransaction.UnitOfWork) error {
		tx, txErr := infratransaction.GormDB(unit)
		if txErr != nil {
			return txErr
		}
		repository := historycapture.WrapBacnetObjectAlarmValue(
			facilitysql.NewBacnetObjectAlarmValueRepository(tx),
			historysql.NewStore(tx),
		)
		return repository.ReplaceForBacnetObject(runCtx, bacnetObjectID, replacement)
	})
	if err == nil {
		t.Fatal("expected missing history table to fail replacement")
	}

	var persisted []domainFacility.BacnetObjectAlarmValue
	if err := db.Where("bacnet_object_id = ?", bacnetObjectID).Find(&persisted).Error; err != nil {
		t.Fatalf("load alarm values after rollback: %v", err)
	}
	if len(persisted) != 1 || persisted[0].ID != oldValueID {
		t.Fatalf("alarm replacement escaped history rollback: %+v", persisted)
	}
}

func TestGormRunnerRollsBackObjectDataActivationWhenHistoryWriteFails(t *testing.T) {
	db := newGormRunnerTestDB(t)
	if err := db.AutoMigrate(&domainFacility.ObjectData{}); err != nil {
		t.Fatalf("migrate ObjectData: %v", err)
	}
	objectDataID := uuid.New()
	objectData := domainFacility.ObjectData{
		Base:        domain.Base{ID: objectDataID},
		Description: "AHU",
		Version:     "1",
		IsActive:    false,
	}
	if err := db.Omit("Project", "BacnetObjects", "Apparats").Create(&objectData).Error; err != nil {
		t.Fatalf("seed ObjectData: %v", err)
	}
	if err := db.Model(&domainFacility.ObjectData{}).
		Where("id = ?", objectDataID).
		Update("is_active", false).Error; err != nil {
		t.Fatalf("force inactive ObjectData fixture: %v", err)
	}
	var seeded domainFacility.ObjectData
	if err := db.Select("id", "is_active").First(&seeded, "id = ?", objectDataID).Error; err != nil {
		t.Fatalf("verify ObjectData fixture: %v", err)
	}
	if seeded.IsActive {
		t.Fatal("ObjectData rollback fixture must start inactive")
	}
	historyErr := errors.New("history write failed")
	store := &failingObjectDataUpdateHistoryStore{err: historyErr}

	runner := infratransaction.NewGormRunner(db)
	err := runner(context.Background(), func(runCtx context.Context, unit apptransaction.UnitOfWork) error {
		tx, txErr := infratransaction.GormDB(unit)
		if txErr != nil {
			return txErr
		}
		store.db = tx
		repository := historycapture.WrapObjectData(
			facilitysql.NewObjectDataRepository(tx),
			store,
		)
		loaded, loadErr := domain.GetByID(runCtx, repository, objectDataID)
		if loadErr != nil {
			return loadErr
		}
		loaded.IsActive = true
		return repository.Update(runCtx, loaded)
	})
	if !errors.Is(err, historyErr) {
		t.Fatalf("transaction error: got %v, want %v", err, historyErr)
	}
	if !store.recordSawUpdated {
		t.Fatal("history write did not observe the staged ObjectData activation")
	}
	var persisted domainFacility.ObjectData
	if err := db.Select("id", "is_active").First(&persisted, "id = ?", objectDataID).Error; err != nil {
		t.Fatalf("load ObjectData after rollback: %v", err)
	}
	if persisted.IsActive {
		t.Fatal("ObjectData activation escaped history rollback")
	}
}

func TestGormRunnerRollsBackSPSControllerDeleteWhenHistoryWriteFails(t *testing.T) {
	db := newGormRunnerTestDB(t)
	if err := db.AutoMigrate(&domainFacility.SPSController{}); err != nil {
		t.Fatalf("migrate SPS controllers: %v", err)
	}
	controllerID := uuid.New()
	controller := domainFacility.SPSController{
		Base:             domain.Base{ID: controllerID},
		ControlCabinetID: uuid.New(),
		DeviceName:       "SPS-01",
	}
	if err := db.Omit("ControlCabinet", "SPSControllerSystemTypes").Create(&controller).Error; err != nil {
		t.Fatalf("seed SPS controller: %v", err)
	}
	historyErr := errors.New("history write failed")
	store := &failingDeleteHistoryStore{
		model: &domainFacility.SPSController{},
		before: map[uuid.UUID]domainHistory.JSONB{
			controllerID: domainHistory.JSONB(`{"id":"` + controllerID.String() + `"}`),
		},
		err: historyErr,
	}

	runner := infratransaction.NewGormRunner(db)
	err := runner(context.Background(), func(runCtx context.Context, unit apptransaction.UnitOfWork) error {
		tx, txErr := infratransaction.GormDB(unit)
		if txErr != nil {
			return txErr
		}
		store.db = tx
		repository := historycapture.WrapSPSController(
			facilitysql.NewSPSControllerRepository(tx),
			store,
		)
		return repository.DeleteByIds(runCtx, []uuid.UUID{controllerID})
	})
	if !errors.Is(err, historyErr) {
		t.Fatalf("transaction error: got %v, want %v", err, historyErr)
	}
	if !store.recordSawDeleted {
		t.Fatal("history write was not attempted after the transactional delete")
	}
	assertGormRunnerModelCount(t, db, &domainFacility.SPSController{}, 1)
}

func TestGormRunnerRollsBackControlCabinetDeleteWhenHistoryWriteFails(t *testing.T) {
	db := newGormRunnerTestDB(t)
	if err := db.AutoMigrate(&domainFacility.ControlCabinet{}); err != nil {
		t.Fatalf("migrate control cabinets: %v", err)
	}
	cabinetID := uuid.New()
	cabinet := domainFacility.ControlCabinet{
		Base:       domain.Base{ID: cabinetID},
		BuildingID: uuid.New(),
	}
	if err := db.Omit("Building", "SPSControllers").Create(&cabinet).Error; err != nil {
		t.Fatalf("seed control cabinet: %v", err)
	}
	historyErr := errors.New("history write failed")
	store := &failingDeleteHistoryStore{
		model: &domainFacility.ControlCabinet{},
		before: map[uuid.UUID]domainHistory.JSONB{
			cabinetID: domainHistory.JSONB(`{"id":"` + cabinetID.String() + `"}`),
		},
		err: historyErr,
	}

	runner := infratransaction.NewGormRunner(db)
	err := runner(context.Background(), func(runCtx context.Context, unit apptransaction.UnitOfWork) error {
		tx, txErr := infratransaction.GormDB(unit)
		if txErr != nil {
			return txErr
		}
		store.db = tx
		repository := historycapture.WrapControlCabinet(
			facilitysql.NewControlCabinetRepository(tx),
			store,
		)
		return repository.DeleteByIds(runCtx, []uuid.UUID{cabinetID})
	})
	if !errors.Is(err, historyErr) {
		t.Fatalf("transaction error: got %v, want %v", err, historyErr)
	}
	if !store.recordSawDeleted {
		t.Fatal("history write was not attempted after the transactional delete")
	}
	assertGormRunnerModelCount(t, db, &domainFacility.ControlCabinet{}, 1)
}

func TestGormRunnerRollsBackSPSControllerSystemTypeDeleteWhenHistoryWriteFails(t *testing.T) {
	db := newGormRunnerTestDB(t)
	if err := db.AutoMigrate(&domainFacility.SPSControllerSystemType{}); err != nil {
		t.Fatalf("migrate SPS controller system types: %v", err)
	}
	systemTypeID := uuid.New()
	assignment := domainFacility.SPSControllerSystemType{
		Base:            domain.Base{ID: systemTypeID},
		SPSControllerID: uuid.New(),
		SystemTypeID:    uuid.New(),
	}
	if err := db.Omit("SPSController", "SystemType", "FieldDevices").Create(&assignment).Error; err != nil {
		t.Fatalf("seed SPS controller system type: %v", err)
	}
	historyErr := errors.New("history write failed")
	store := &failingDeleteHistoryStore{
		model: &domainFacility.SPSControllerSystemType{},
		before: map[uuid.UUID]domainHistory.JSONB{
			systemTypeID: domainHistory.JSONB(`{"id":"` + systemTypeID.String() + `"}`),
		},
		err: historyErr,
	}

	runner := infratransaction.NewGormRunner(db)
	err := runner(context.Background(), func(runCtx context.Context, unit apptransaction.UnitOfWork) error {
		tx, txErr := infratransaction.GormDB(unit)
		if txErr != nil {
			return txErr
		}
		store.db = tx
		repository := historycapture.WrapSPSControllerSystemType(
			facilitysql.NewSPSControllerSystemTypeRepository(tx),
			store,
		)
		return repository.DeleteByIds(runCtx, []uuid.UUID{systemTypeID})
	})
	if !errors.Is(err, historyErr) {
		t.Fatalf("transaction error: got %v, want %v", err, historyErr)
	}
	if !store.recordSawDeleted {
		t.Fatal("history write was not attempted after the transactional delete")
	}
	assertGormRunnerModelCount(t, db, &domainFacility.SPSControllerSystemType{}, 1)
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

	if err := db.AutoMigrate(
		&gormRunnerTestRecord{},
		&gormRunnerFacilityRecord{},
		&gormRunnerHistoryRecord{},
	); err != nil {
		t.Fatalf("expected test table to migrate, got %v", err)
	}

	return db
}

func assertGormRunnerModelCount(
	t *testing.T,
	db *gorm.DB,
	model any,
	want int64,
) {
	t.Helper()

	var count int64
	if err := db.Model(model).Count(&count).Error; err != nil {
		t.Fatalf("expected count to succeed, got %v", err)
	}
	if count != want {
		t.Fatalf("expected %d records, got %d", want, count)
	}
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
