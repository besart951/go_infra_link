package realtime

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const projectDraftLease = 60 * time.Second

// SQLProjectCollaborationStore makes revision snapshots and draft state shared
// between backend nodes. Draft values are short-lived and are never copied to
// the durable audit/change stream.
type SQLProjectCollaborationStore struct {
	db  *gorm.DB
	now func() time.Time
}

type projectDraftRecord struct {
	ProjectID uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserID    uuid.UUID `gorm:"type:uuid;primaryKey"`
	Entries   []byte    `gorm:"type:jsonb;not null"`
	UpdatedAt time.Time `gorm:"not null"`
	ExpiresAt time.Time `gorm:"not null;index"`
}

type projectPresenceRecord struct {
	ConnectionID uuid.UUID `gorm:"type:uuid;primaryKey"`
	ProjectID    uuid.UUID `gorm:"type:uuid;not null;index"`
	UserID       uuid.UUID `gorm:"type:uuid;not null;index"`
	ConnectedAt  time.Time `gorm:"not null"`
	LastSeenAt   time.Time `gorm:"not null"`
	ExpiresAt    time.Time `gorm:"not null;index"`
}

func (projectPresenceRecord) TableName() string { return "project_collaboration_sessions" }

func (projectDraftRecord) TableName() string { return "project_collaboration_drafts" }

func NewSQLProjectCollaborationStore(db *gorm.DB) *SQLProjectCollaborationStore {
	return &SQLProjectCollaborationStore{db: db, now: func() time.Time { return time.Now().UTC() }}
}

func AutoMigrateProjectCollaboration(db *gorm.DB) error {
	return db.AutoMigrate(&projectDraftRecord{}, &projectPresenceRecord{})
}

func (s *SQLProjectCollaborationStore) SaveProjectPresence(ctx context.Context, connectionID, projectID, userID uuid.UUID, connectedAt time.Time) error {
	if s == nil || s.db == nil {
		return nil
	}
	now := s.now()
	record := projectPresenceRecord{ConnectionID: connectionID, ProjectID: projectID, UserID: userID, ConnectedAt: connectedAt, LastSeenAt: now, ExpiresAt: now.Add(projectDraftLease)}
	return s.db.WithContext(ctx).Save(&record).Error
}

func (s *SQLProjectCollaborationStore) DeleteProjectPresence(ctx context.Context, connectionID uuid.UUID) error {
	if s == nil || s.db == nil {
		return nil
	}
	return s.db.WithContext(ctx).Where("connection_id = ?", connectionID).Delete(&projectPresenceRecord{}).Error
}

func (s *SQLProjectCollaborationStore) LoadProjectPresence(ctx context.Context, projectID uuid.UUID) ([]ProjectCollaboratorPresence, error) {
	if s == nil || s.db == nil {
		return nil, nil
	}
	now := s.now()
	if err := s.db.WithContext(ctx).Where("expires_at <= ?", now).Delete(&projectPresenceRecord{}).Error; err != nil {
		return nil, fmt.Errorf("prune project presence: %w", err)
	}
	var records []projectPresenceRecord
	if err := s.db.WithContext(ctx).Where("project_id = ? AND expires_at > ?", projectID, now).Order("connected_at ASC").Find(&records).Error; err != nil {
		return nil, fmt.Errorf("load project presence: %w", err)
	}
	byUser := make(map[uuid.UUID]ProjectCollaboratorPresence, len(records))
	for _, record := range records {
		item, ok := byUser[record.UserID]
		if !ok || record.ConnectedAt.Before(item.ConnectedAt) {
			item.ConnectedAt = record.ConnectedAt
		}
		if !ok || record.LastSeenAt.After(item.LastSeenAt) {
			item.LastSeenAt = record.LastSeenAt
		}
		item.UserID = record.UserID
		byUser[record.UserID] = item
	}
	items := make([]ProjectCollaboratorPresence, 0, len(byUser))
	for _, item := range byUser {
		items = append(items, item)
	}
	return items, nil
}

func (s *SQLProjectCollaborationStore) CurrentRevision(ctx context.Context, projectID uuid.UUID) (int64, error) {
	if s == nil || s.db == nil {
		return 0, nil
	}
	var revision int64
	err := s.db.WithContext(ctx).Table("project_change_cursors").
		Select("current_revision").Where("project_id = ?", projectID).Scan(&revision).Error
	return revision, err
}

func (s *SQLProjectCollaborationStore) LoadProjectDraftStates(ctx context.Context, projectID uuid.UUID) ([]ProjectDraftState, error) {
	if s == nil || s.db == nil {
		return nil, nil
	}
	now := s.now()
	if err := s.db.WithContext(ctx).Where("expires_at <= ?", now).Delete(&projectDraftRecord{}).Error; err != nil {
		return nil, fmt.Errorf("prune project drafts: %w", err)
	}
	var records []projectDraftRecord
	if err := s.db.WithContext(ctx).Where("project_id = ? AND expires_at > ?", projectID, now).Order("updated_at ASC").Find(&records).Error; err != nil {
		return nil, fmt.Errorf("load project drafts: %w", err)
	}
	states := make([]ProjectDraftState, 0, len(records))
	for _, record := range records {
		var entries []ProjectDraftEntry
		if err := json.Unmarshal(record.Entries, &entries); err != nil {
			return nil, fmt.Errorf("decode project draft: %w", err)
		}
		states = append(states, ProjectDraftState{UserID: record.UserID, Entries: entries, UpdatedAt: record.UpdatedAt})
	}
	return states, nil
}

func (s *SQLProjectCollaborationStore) SaveProjectDraftState(ctx context.Context, projectID, userID uuid.UUID, entries []ProjectDraftEntry) error {
	if s == nil || s.db == nil {
		return nil
	}
	if len(entries) == 0 {
		return s.ClearProjectDraft(ctx, projectID, userID, nil)
	}
	payload, err := json.Marshal(entries)
	if err != nil {
		return fmt.Errorf("encode project draft: %w", err)
	}
	now := s.now()
	record := projectDraftRecord{ProjectID: projectID, UserID: userID, Entries: payload, UpdatedAt: now, ExpiresAt: now.Add(projectDraftLease)}
	return s.db.WithContext(ctx).Save(&record).Error
}

func (s *SQLProjectCollaborationStore) ClearProjectDraft(ctx context.Context, projectID, userID uuid.UUID, selector *ProjectDraftSelector) error {
	if s == nil || s.db == nil {
		return nil
	}
	if selector == nil {
		return s.db.WithContext(ctx).Where("project_id = ? AND user_id = ?", projectID, userID).Delete(&projectDraftRecord{}).Error
	}
	states, err := s.LoadProjectDraftStates(ctx, projectID)
	if err != nil {
		return err
	}
	for _, state := range states {
		if state.UserID != userID {
			continue
		}
		kept := make([]ProjectDraftEntry, 0, len(state.Entries))
		for _, entry := range state.Entries {
			if entry.ProjectDraftSelector.selectorKey() != selector.selectorKey() {
				kept = append(kept, entry)
			}
		}
		return s.SaveProjectDraftState(ctx, projectID, userID, kept)
	}
	return nil
}
