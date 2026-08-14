package projectchange

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	domainProject "github.com/besart951/go_infra_link/backend/internal/domain/project"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	DefaultLimit = 100
	MaxLimit     = 500
	Retention    = 30 * 24 * time.Hour
)

type cursorRecord struct {
	ProjectID       uuid.UUID `gorm:"type:uuid;primaryKey"`
	CurrentRevision uint64    `gorm:"not null;default:0"`
	UpdatedAt       time.Time `gorm:"not null"`
}

func (cursorRecord) TableName() string { return "project_change_cursors" }

type changeRecord struct {
	EventID       uuid.UUID  `gorm:"type:uuid;primaryKey"`
	ProjectID     uuid.UUID  `gorm:"type:uuid;not null;uniqueIndex:idx_project_changes_revision;index:idx_project_changes_project_time"`
	Revision      uint64     `gorm:"not null;uniqueIndex:idx_project_changes_revision"`
	AggregateType string     `gorm:"size:100;not null"`
	AggregateID   *uuid.UUID `gorm:"type:uuid"`
	Action        string     `gorm:"size:50;not null"`
	ActorID       *uuid.UUID `gorm:"type:uuid"`
	ChangedFields []byte     `gorm:"type:jsonb;not null"`
	ParentRefs    []byte     `gorm:"type:jsonb;not null"`
	OccurredAt    time.Time  `gorm:"not null;index:idx_project_changes_project_time"`
}

func (changeRecord) TableName() string { return "project_changes" }

type Store struct {
	db        *gorm.DB
	now       func() time.Time
	retention time.Duration
}

func NewStore(db *gorm.DB) *Store {
	return &Store{db: db, now: func() time.Time { return time.Now().UTC() }, retention: Retention}
}

func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(&cursorRecord{}, &changeRecord{})
}

func (s *Store) Append(ctx context.Context, input domainProject.NewChange) (*domainProject.Change, error) {
	changes, err := s.AppendBatch(ctx, []domainProject.NewChange{input})
	if err != nil {
		return nil, err
	}
	return &changes[0], nil
}

func (s *Store) AppendBatch(ctx context.Context, inputs []domainProject.NewChange) ([]domainProject.Change, error) {
	if len(inputs) == 0 {
		return []domainProject.Change{}, nil
	}
	projectID := inputs[0].ProjectID
	now := s.now()
	records := make([]changeRecord, len(inputs))
	changes := make([]domainProject.Change, len(inputs))
	for i, input := range inputs {
		if input.ProjectID == uuid.Nil || input.ProjectID != projectID || input.AggregateType == "" || input.Action == "" {
			return nil, fmt.Errorf("append project change batch: invalid semantic event at index %d", i)
		}
		occurredAt := input.OccurredAt.UTC()
		if input.OccurredAt.IsZero() {
			occurredAt = now
		}
		eventID, err := uuid.NewV7()
		if err != nil {
			return nil, fmt.Errorf("allocate project change id: %w", err)
		}
		changedFields, err := json.Marshal(nonNilStrings(input.ChangedFields))
		if err != nil {
			return nil, fmt.Errorf("marshal changed fields: %w", err)
		}
		parentRefs, err := json.Marshal(nonNilRefs(input.ParentRefs))
		if err != nil {
			return nil, fmt.Errorf("marshal parent refs: %w", err)
		}
		records[i] = changeRecord{EventID: eventID, ProjectID: projectID, AggregateType: input.AggregateType, AggregateID: cloneUUID(input.AggregateID), Action: string(input.Action), ActorID: cloneUUID(input.ActorID), ChangedFields: changedFields, ParentRefs: parentRefs, OccurredAt: occurredAt}
		changes[i] = domainProject.Change{EventID: eventID, ProjectID: projectID, AggregateType: input.AggregateType, AggregateID: cloneUUID(input.AggregateID), Action: input.Action, ActorID: cloneUUID(input.ActorID), ChangedFields: nonNilStrings(input.ChangedFields), ParentRefs: nonNilRefs(input.ParentRefs), OccurredAt: occurredAt}
	}

	var lastRevision uint64
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Raw(`
			INSERT INTO project_change_cursors (project_id, current_revision, updated_at)
			VALUES (?, ?, ?)
			ON CONFLICT (project_id) DO UPDATE
			SET current_revision = project_change_cursors.current_revision + excluded.current_revision,
				updated_at = excluded.updated_at
			RETURNING current_revision
		`, projectID, len(records), now).Scan(&lastRevision).Error; err != nil {
			return fmt.Errorf("allocate project revision: %w", err)
		}
		firstRevision := lastRevision - uint64(len(records)) + 1
		for i := range records {
			records[i].Revision = firstRevision + uint64(i)
			changes[i].Revision = records[i].Revision
		}
		if err := tx.CreateInBatches(records, 500).Error; err != nil {
			return fmt.Errorf("insert project change batch: %w", err)
		}
		cutoff := now.Add(-s.retention)
		if err := tx.Where("occurred_at < ?", cutoff).
			Delete(&changeRecord{}).Error; err != nil {
			return fmt.Errorf("prune project changes: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return changes, nil
}

func (s *Store) ListAfter(ctx context.Context, projectID uuid.UUID, afterRevision uint64, limit int) (*domainProject.ChangePage, error) {
	limit = normalizeLimit(limit)
	page := &domainProject.ChangePage{ProjectID: projectID, Events: []domainProject.Change{}}

	var cursor cursorRecord
	err := s.db.WithContext(ctx).First(&cursor, "project_id = ?", projectID).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("load project change cursor: %w", err)
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return page, nil
	}
	page.CurrentRevision = cursor.CurrentRevision
	if afterRevision > page.CurrentRevision {
		page.ResetRequired = true
		return page, nil
	}

	var earliest uint64
	if err := s.db.WithContext(ctx).Model(&changeRecord{}).
		Where("project_id = ?", projectID).
		Select("COALESCE(MIN(revision), 0)").Scan(&earliest).Error; err != nil {
		return nil, fmt.Errorf("load earliest project revision: %w", err)
	}
	if (earliest == 0 && afterRevision < page.CurrentRevision) || (earliest > 0 && afterRevision < earliest-1) {
		page.ResetRequired = true
		return page, nil
	}

	var records []changeRecord
	if err := s.db.WithContext(ctx).
		Where("project_id = ? AND revision > ?", projectID, afterRevision).
		Order("revision ASC").Limit(limit + 1).Find(&records).Error; err != nil {
		return nil, fmt.Errorf("list project changes: %w", err)
	}
	if len(records) > limit {
		page.HasMore = true
		records = records[:limit]
	}

	page.Events = make([]domainProject.Change, len(records))
	for i := range records {
		change, err := records[i].domain()
		if err != nil {
			return nil, err
		}
		page.Events[i] = change
	}
	return page, nil
}

func (r changeRecord) domain() (domainProject.Change, error) {
	var changedFields []string
	if err := json.Unmarshal(r.ChangedFields, &changedFields); err != nil {
		return domainProject.Change{}, fmt.Errorf("decode project change %s changed fields: %w", r.EventID, err)
	}
	var parentRefs map[string]uuid.UUID
	if err := json.Unmarshal(r.ParentRefs, &parentRefs); err != nil {
		return domainProject.Change{}, fmt.Errorf("decode project change %s parent refs: %w", r.EventID, err)
	}
	return domainProject.Change{
		EventID: r.EventID, ProjectID: r.ProjectID, Revision: r.Revision,
		AggregateType: r.AggregateType, AggregateID: r.AggregateID,
		Action: domainProject.ChangeAction(r.Action), ActorID: r.ActorID,
		ChangedFields: nonNilStrings(changedFields), ParentRefs: nonNilRefs(parentRefs),
		OccurredAt: r.OccurredAt,
	}, nil
}

func normalizeLimit(limit int) int {
	if limit <= 0 {
		return DefaultLimit
	}
	return min(limit, MaxLimit)
}

func nonNilStrings(values []string) []string {
	if values == nil {
		return []string{}
	}
	return values
}

func nonNilRefs(values map[string]uuid.UUID) map[string]uuid.UUID {
	if values == nil {
		return map[string]uuid.UUID{}
	}
	return values
}

func cloneUUID(value *uuid.UUID) *uuid.UUID {
	if value == nil {
		return nil
	}
	clone := *value
	return &clone
}
