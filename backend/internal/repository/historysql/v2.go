package historysql

import (
	"context"
	"fmt"
	"time"

	domainHistory "github.com/besart951/go_infra_link/backend/internal/domain/history"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const historyV2ReaderStateKey = "timeline_reader"

type historyV2StateRecord struct {
	Key          string    `gorm:"primaryKey;size:64"`
	ReadsEnabled bool      `gorm:"not null;default:false"`
	UpdatedAt    time.Time `gorm:"not null"`
}

func (historyV2StateRecord) TableName() string { return "history_v2_state" }

type changeEventV2Record struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey"`
	OccurredAt   time.Time `gorm:"primaryKey;not null"`
	ActorID      *uuid.UUID
	Action       domainHistory.Action `gorm:"type:varchar(32);not null"`
	EntityTable  string               `gorm:"type:varchar(96);not null"`
	EntityID     uuid.UUID            `gorm:"type:uuid;not null"`
	BatchID      *uuid.UUID
	Summary      *string
	BeforeJSON   domainHistory.JSONB `gorm:"type:jsonb"`
	AfterJSON    domainHistory.JSONB `gorm:"type:jsonb"`
	DiffJSON     domainHistory.JSONB `gorm:"type:jsonb"`
	MetadataJSON domainHistory.JSONB `gorm:"type:jsonb"`
}

func (changeEventV2Record) TableName() string { return "change_events_v2" }

type changeEventScopeV2Record struct {
	ID            uuid.UUID `gorm:"type:uuid;primaryKey"`
	OccurredAt    time.Time `gorm:"primaryKey;not null"`
	ChangeEventID uuid.UUID `gorm:"type:uuid;not null"`
	ScopeType     string    `gorm:"type:varchar(96);not null"`
	ScopeID       uuid.UUID `gorm:"type:uuid;not null"`
}

func (changeEventScopeV2Record) TableName() string { return "change_event_scopes_v2" }

type entityVersionV2Record struct {
	ID            uuid.UUID            `gorm:"type:uuid;primaryKey"`
	VersionAt     time.Time            `gorm:"primaryKey;not null"`
	ChangeEventID uuid.UUID            `gorm:"type:uuid;not null"`
	EntityTable   string               `gorm:"type:varchar(96);not null"`
	EntityID      uuid.UUID            `gorm:"type:uuid;not null"`
	Action        domainHistory.Action `gorm:"type:varchar(32);not null"`
	SnapshotJSON  domainHistory.JSONB  `gorm:"type:jsonb"`
}

func (entityVersionV2Record) TableName() string { return "entity_versions_v2" }

func AutoMigrateV2(db *gorm.DB, now time.Time) error {
	if err := migrateHistoryV2State(db, now); err != nil {
		return err
	}
	if db.Dialector == nil || db.Dialector.Name() != "postgres" {
		return autoMigrateV2Tables(db)
	}
	if err := createPartitionedHistoryTables(db); err != nil {
		return err
	}
	return EnsureHistoryV2Partitions(db, now)
}

func migrateHistoryV2State(db *gorm.DB, now time.Time) error {
	if err := db.AutoMigrate(&historyV2StateRecord{}); err != nil {
		return err
	}
	state := historyV2StateRecord{Key: historyV2ReaderStateKey, UpdatedAt: now.UTC()}
	return db.Clauses(clause.OnConflict{DoNothing: true}).Create(&state).Error
}

func (s *Store) V2ReadsEnabled(ctx context.Context) (bool, error) {
	if !s.db.Migrator().HasTable(&historyV2StateRecord{}) {
		return false, nil
	}
	var state historyV2StateRecord
	err := s.db.WithContext(ctx).Where("key = ?", historyV2ReaderStateKey).Take(&state).Error
	if err == gorm.ErrRecordNotFound {
		return false, nil
	}
	return state.ReadsEnabled, err
}

func (s *Store) EnableV2Reads(ctx context.Context) error {
	result := s.db.WithContext(ctx).Model(&historyV2StateRecord{}).
		Where("key = ? AND reads_enabled = ?", historyV2ReaderStateKey, false).
		Updates(map[string]any{"reads_enabled": true, "updated_at": time.Now().UTC()})
	if result.Error != nil || result.RowsAffected > 0 {
		return result.Error
	}
	enabled, err := s.V2ReadsEnabled(ctx)
	if err != nil || enabled {
		return err
	}
	return fmt.Errorf("history V2 reader state is missing")
}

func autoMigrateV2Tables(db *gorm.DB) error {
	return db.AutoMigrate(&changeEventV2Record{}, &changeEventScopeV2Record{}, &entityVersionV2Record{})
}

func createPartitionedHistoryTables(db *gorm.DB) error {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS change_events_v2 (
			id uuid NOT NULL, occurred_at timestamptz NOT NULL, actor_id uuid,
			action varchar(32) NOT NULL, entity_table varchar(96) NOT NULL, entity_id uuid NOT NULL,
			batch_id uuid, summary text, before_json jsonb, after_json jsonb, diff_json jsonb, metadata_json jsonb,
			PRIMARY KEY (occurred_at, id)) PARTITION BY RANGE (occurred_at)`,
		`CREATE TABLE IF NOT EXISTS change_event_scopes_v2 (
			id uuid NOT NULL, occurred_at timestamptz NOT NULL, change_event_id uuid NOT NULL,
			scope_type varchar(96) NOT NULL, scope_id uuid NOT NULL,
			PRIMARY KEY (occurred_at, id)) PARTITION BY RANGE (occurred_at)`,
		`CREATE TABLE IF NOT EXISTS entity_versions_v2 (
			id uuid NOT NULL, version_at timestamptz NOT NULL, change_event_id uuid NOT NULL,
			entity_table varchar(96) NOT NULL, entity_id uuid NOT NULL, action varchar(32) NOT NULL,
			snapshot_json jsonb, PRIMARY KEY (version_at, id)) PARTITION BY RANGE (version_at)`,
		`CREATE INDEX IF NOT EXISTS idx_change_events_v2_timeline ON change_events_v2 (occurred_at DESC, id DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_change_events_v2_entity ON change_events_v2 (entity_table, entity_id, occurred_at DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_change_events_v2_actor ON change_events_v2 (actor_id, occurred_at DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_change_events_v2_batch ON change_events_v2 (batch_id, occurred_at DESC) WHERE batch_id IS NOT NULL`,
		`CREATE INDEX IF NOT EXISTS idx_change_event_scopes_v2_lookup ON change_event_scopes_v2 (scope_type, scope_id, occurred_at DESC, change_event_id)`,
		`CREATE INDEX IF NOT EXISTS idx_entity_versions_v2_entity ON entity_versions_v2 (entity_table, entity_id, version_at DESC)`,
	}
	for _, statement := range statements {
		if err := db.Exec(statement).Error; err != nil {
			return err
		}
	}
	return nil
}

func EnsureHistoryV2Partitions(db *gorm.DB, now time.Time) error {
	start := monthStart(now.UTC()).AddDate(-7, 0, 0)
	end := monthStart(now.UTC()).AddDate(0, 3, 0)
	for current := start; current.Before(end); current = current.AddDate(0, 1, 0) {
		if err := createHistoryMonthPartitions(db, current); err != nil {
			return err
		}
	}
	return nil
}

func createHistoryMonthPartitions(db *gorm.DB, start time.Time) error {
	end := start.AddDate(0, 1, 0)
	suffix := start.Format("2006_01")
	tables := []struct {
		parent string
		column string
	}{
		{"change_events_v2", "occurred_at"},
		{"change_event_scopes_v2", "occurred_at"},
		{"entity_versions_v2", "version_at"},
	}
	for _, table := range tables {
		statement := fmt.Sprintf(
			"CREATE TABLE IF NOT EXISTS %s_%s PARTITION OF %s FOR VALUES FROM ('%s') TO ('%s')",
			table.parent, suffix, table.parent, start.Format(time.RFC3339), end.Format(time.RFC3339),
		)
		if err := db.Exec(statement).Error; err != nil {
			return err
		}
	}
	return nil
}

func monthStart(value time.Time) time.Time {
	return time.Date(value.Year(), value.Month(), 1, 0, 0, 0, 0, time.UTC)
}
