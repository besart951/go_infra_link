package facilityjobsql

import (
	"context"
	"errors"
	"time"

	facilityjobs "github.com/besart951/go_infra_link/backend/internal/application/facilityjobs"
	apptransaction "github.com/besart951/go_infra_link/backend/internal/application/transaction"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type StepStore struct {
	db *gorm.DB
}

func NewStepStore(db *gorm.DB) *StepStore {
	return &StepStore{db: db}
}

func (s *StepStore) Prepare(ctx context.Context, steps []facilityjobs.Step) error {
	if len(steps) == 0 {
		return nil
	}
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var count int64
		key := steps[0].Key
		if err := tx.Model(&itemRecord{}).
			Where("owner_id = ? AND job_id = ?", key.OwnerID, key.JobID).
			Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return nil
		}
		now := time.Now().UTC()
		records := make([]itemRecord, len(steps))
		for index, step := range steps {
			records[index] = preparedItemRecord(step, now)
		}
		return tx.CreateInBatches(records, 500).Error
	})
}

func (s *StepStore) ListItems(ctx context.Context, ownerID, jobID uuid.UUID) ([]facilityjobs.Item, error) {
	var records []itemRecord
	err := s.db.WithContext(ctx).
		Where("owner_id = ? AND job_id = ?", ownerID, jobID).
		Order("ordinal ASC").Find(&records).Error
	items := make([]facilityjobs.Item, len(records))
	for index, record := range records {
		items[index] = record.toDomain()
	}
	return items, err
}

func (s *StepStore) Execute(
	ctx context.Context,
	step facilityjobs.Step,
	mutate facilityjobs.Mutation,
) (facilityjobs.StepResult, bool, error) {
	var result facilityjobs.StepResult
	var resumed bool
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		stored, found, err := lockOrCreateItem(tx, step)
		if err != nil {
			return err
		}
		if found && facilityjobs.ItemStatus(stored.Status) == facilityjobs.ItemStatusCompleted {
			result, resumed, err = completedResult(tx, step, stored)
			return err
		}
		result, err = mutate(ctx, apptransaction.UnitOfWork(tx))
		if err != nil {
			return err
		}
		return completeStep(tx, step, result)
	})
	if err != nil {
		s.recordFailure(ctx, step, err)
	}
	return result, resumed, err
}

func (s *StepStore) GetItem(ctx context.Context, key facilityjobs.ItemKey) (facilityjobs.Item, error) {
	var record itemRecord
	err := s.db.WithContext(ctx).Where(itemKeyQuery(key)).First(&record).Error
	return record.toDomain(), err
}

func (s *StepStore) GetMapping(ctx context.Context, step facilityjobs.Step) (facilityjobs.IDMapping, error) {
	var record mappingRecord
	err := s.db.WithContext(ctx).Where(mappingQuery(step)).First(&record).Error
	return record.toDomain(), err
}

func (s *StepStore) recordFailure(ctx context.Context, step facilityjobs.Step, cause error) {
	now := time.Now().UTC()
	record := itemRecordFromStep(step, now)
	record.Status = string(facilityjobs.ItemStatusFailed)
	record.Error = cause.Error()
	updates := map[string]any{
		"status": record.Status, "error_message": record.Error,
		"attempts": gorm.Expr("facility_job_items.attempts + 1"), "updated_at": now,
	}
	_ = s.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "owner_id"}, {Name: "job_id"}, {Name: "ordinal"}},
		DoUpdates: clause.Assignments(updates),
	}).Create(&record).Error
}

func lockOrCreateItem(tx *gorm.DB, step facilityjobs.Step) (itemRecord, bool, error) {
	var record itemRecord
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where(itemKeyQuery(step.Key)).First(&record).Error
	if err == nil {
		if facilityjobs.ItemStatus(record.Status) == facilityjobs.ItemStatusCompleted {
			return record, true, nil
		}
		err = tx.Model(&record).Updates(map[string]any{
			"status": facilityjobs.ItemStatusRunning, "attempts": gorm.Expr("attempts + 1"), "updated_at": time.Now().UTC(),
		}).Error
		return record, true, err
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return itemRecord{}, false, err
	}
	now := time.Now().UTC()
	record = itemRecordFromStep(step, now)
	return record, false, tx.Create(&record).Error
}

func completedResult(tx *gorm.DB, step facilityjobs.Step, item itemRecord) (facilityjobs.StepResult, bool, error) {
	if !step.PersistIDMapping {
		return facilityjobs.StepResult{Result: item.Result}, true, nil
	}
	var mapping mappingRecord
	if err := tx.Where(mappingQuery(step)).First(&mapping).Error; err != nil {
		return facilityjobs.StepResult{}, false, err
	}
	return facilityjobs.StepResult{TargetID: mapping.TargetID, Result: item.Result}, true, nil
}

func completeStep(tx *gorm.DB, step facilityjobs.Step, result facilityjobs.StepResult) error {
	if step.PersistIDMapping {
		if result.TargetID == uuid.Nil {
			return errors.New("facility job step target ID is required")
		}
		if err := saveMapping(tx, step, result.TargetID); err != nil {
			return err
		}
	}
	updates := map[string]any{
		"status": facilityjobs.ItemStatusCompleted, "result": []byte(result.Result),
		"error_message": "", "updated_at": time.Now().UTC(),
	}
	return tx.Model(&itemRecord{}).Where(itemKeyQuery(step.Key)).Updates(updates).Error
}

func saveMapping(tx *gorm.DB, step facilityjobs.Step, targetID uuid.UUID) error {
	now := time.Now().UTC()
	record := mappingRecordFromStep(step, targetID, now)
	result := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&record)
	if result.Error != nil || result.RowsAffected == 1 {
		return result.Error
	}
	var existing mappingRecord
	if err := tx.Where(mappingQuery(step)).First(&existing).Error; err != nil {
		return err
	}
	if existing.TargetID != targetID {
		return facilityjobs.ErrMappingConflict
	}
	return nil
}

func itemKeyQuery(key facilityjobs.ItemKey) map[string]any {
	return map[string]any{"owner_id": key.OwnerID, "job_id": key.JobID, "ordinal": key.Ordinal}
}

func mappingQuery(step facilityjobs.Step) map[string]any {
	return map[string]any{
		"owner_id": step.Key.OwnerID, "job_id": step.Key.JobID,
		"entity_type": step.EntityType, "source_id": step.SourceID,
	}
}
